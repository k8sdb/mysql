package controller

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	kutildb "github.com/appscode/kutil/kubedb/v1alpha1"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/docker"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	"github.com/k8sdb/apimachinery/pkg/storage"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	apps "k8s.io/client-go/pkg/apis/apps/v1beta1"
	batch "k8s.io/client-go/pkg/apis/batch/v1"
)

const (
	// Duration in Minute
	// Check whether pod under StatefulSet is running or not
	// Continue checking for this duration until failure
	durationCheckStatefulSet = time.Minute * 30
)

func (c *Controller) findService(mysql *tapi.MySQL) (bool, error) {
	name := mysql.OffshootName()
	service, err := c.Client.CoreV1().Services(mysql.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if service.Spec.Selector[tapi.LabelDatabaseName] != name {
		return false, fmt.Errorf(`Intended service "%v" already exists`, name)
	}

	return true, nil
}

func (c *Controller) createService(mysql *tapi.MySQL) error {
	svc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   mysql.OffshootName(),
			Labels: mysql.OffshootLabels(),
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "db",
					Port:       3306,
					TargetPort: intstr.FromString("db"),
				},
			},
			Selector: mysql.OffshootLabels(),
		},
	}
	if mysql.Spec.Monitor != nil &&
		mysql.Spec.Monitor.Agent == tapi.AgentCoreosPrometheus &&
		mysql.Spec.Monitor.Prometheus != nil {
		svc.Spec.Ports = append(svc.Spec.Ports, apiv1.ServicePort{
			Name:       tapi.PrometheusExporterPortName,
			Port:       tapi.PrometheusExporterPortNumber,
			TargetPort: intstr.FromString(tapi.PrometheusExporterPortName),
		})
	}

	if _, err := c.Client.CoreV1().Services(mysql.Namespace).Create(svc); err != nil {
		return err
	}
	return nil
}

func (c *Controller) findStatefulSet(mysql *tapi.MySQL) (bool, error) {
	// SatatefulSet for MySQL database
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(mysql.Namespace).Get(mysql.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if statefulSet.Labels[tapi.LabelDatabaseKind] != tapi.ResourceKindMySQL {
		return false, fmt.Errorf(`Intended statefulSet "%v" already exists`, mysql.OffshootName())
	}

	return true, nil
}

func (c *Controller) createStatefulSet(mysql *tapi.MySQL) (*apps.StatefulSet, error) {
	// SatatefulSet for MySQL database
	statefulSet := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        mysql.OffshootName(),
			Namespace:   mysql.Namespace,
			Labels:      mysql.StatefulSetLabels(),
			Annotations: mysql.StatefulSetAnnotations(),
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    types.Int32P(1),
			ServiceName: c.opt.GoverningService,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: mysql.OffshootLabels(),
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  tapi.ResourceNameMySQL,
							Image: fmt.Sprintf("%s:%s", docker.ImageMySQL, mysql.Spec.Version),
							// todo:	ImagePullPolicy: apiv1.PullIfNotPresent,
							ImagePullPolicy: "Always", //Testing
							Ports: []apiv1.ContainerPort{
								{
									Name:          "db",
									ContainerPort: 3306,
								},
							},
							Resources: mysql.Spec.Resources,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/mysql", //Volume path of mysql, https://github.com/docker-library/mysql/blob/86431f073b3d2f963d21e33cb8943f0bdcdf143d/8.0/Dockerfile#L48
								},
							},
						},
					},
					NodeSelector:  mysql.Spec.NodeSelector,
					Affinity:      mysql.Spec.Affinity,
					SchedulerName: mysql.Spec.SchedulerName,
					Tolerations:   mysql.Spec.Tolerations,
				},
			},
		},
	}

	if mysql.Spec.Monitor != nil &&
		mysql.Spec.Monitor.Agent == tapi.AgentCoreosPrometheus &&
		mysql.Spec.Monitor.Prometheus != nil {
		exporter := apiv1.Container{
			Name: "exporter",
			Args: []string{
				"export",
				fmt.Sprintf("--address=:%d", tapi.PrometheusExporterPortNumber),
				"--v=3",
			},
			Image:           docker.ImageOperator + ":" + c.opt.ExporterTag,
			ImagePullPolicy: apiv1.PullIfNotPresent,
			Ports: []apiv1.ContainerPort{
				{
					Name:          tapi.PrometheusExporterPortName,
					Protocol:      apiv1.ProtocolTCP,
					ContainerPort: int32(tapi.PrometheusExporterPortNumber),
				},
			},
		}
		statefulSet.Spec.Template.Spec.Containers = append(statefulSet.Spec.Template.Spec.Containers, exporter)
	}

	if mysql.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(mysql)
		if err != nil {
			return nil, err
		}

		_mysql, err := kutildb.TryPatchMySQL(c.ExtClient, mysql.ObjectMeta, func(in *tapi.MySQL) *tapi.MySQL {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
		if err != nil {
			c.recorder.Eventf(mysql.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return nil, err
		}
		mysql = _mysql
	}

	//Set root user password from Secret
	setEnvFromSecret(statefulSet, mysql.Spec.DatabaseSecret)

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, mysql.Spec.Storage)

	// ---> Start #LATER
	//TODO: Use following if supported
	// otherwise remove
	// Add InitialScript to run at startup
	if mysql.Spec.Init != nil && mysql.Spec.Init.ScriptSource != nil {
		addInitialScript(statefulSet, mysql.Spec.Init.ScriptSource)
	}
	// ---> End

	if c.opt.EnableRbac {
		// Ensure ClusterRoles for database statefulsets
		if err := c.createRBACStuff(mysql); err != nil {
			return nil, err
		}

		statefulSet.Spec.Template.Spec.ServiceAccountName = mysql.Name
	}

	if _, err := c.Client.AppsV1beta1().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
}

// Set root user password from Secret, Through Env.
func setEnvFromSecret(statefulSet *apps.StatefulSet, secSource *apiv1.SecretVolumeSource) {
	statefulSet.Spec.Template.Spec.Containers[0].Env = append(statefulSet.Spec.Template.Spec.Containers[0].Env,
		apiv1.EnvVar{
			Name: "MYSQL_ROOT_PASSWORD",
			ValueFrom: &apiv1.EnvVarSource{
				SecretKeyRef: &apiv1.SecretKeySelector{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: secSource.SecretName,
					},
					Key: ".admin",
				},
			},
		},
	)
}

func (c *Controller) findSecret(secretName, namespace string) (bool, error) {
	secret, err := c.Client.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	if secret == nil {
		return false, nil
	}

	return true, nil
}

func (c *Controller) createDatabaseSecret(mysql *tapi.MySQL) (*apiv1.SecretVolumeSource, error) {
	authSecretName := mysql.Name + "-admin-auth"

	found, err := c.findSecret(authSecretName, mysql.Namespace)
	if err != nil {
		return nil, err
	}

	if !found {

		MYSQL_PASSWORD := fmt.Sprintf("%s", rand.GeneratePassword())
		data := map[string][]byte{
			".admin": []byte(MYSQL_PASSWORD),
		}

		secret := &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: authSecretName,
				Labels: map[string]string{
					tapi.LabelDatabaseKind: tapi.ResourceKindMySQL,
				},
			},
			Type: apiv1.SecretTypeOpaque,
			Data: data, // Add secret data
		}
		if _, err := c.Client.CoreV1().Secrets(mysql.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}

	return &apiv1.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

func addDataVolume(statefulSet *apps.StatefulSet, pvcSpec *apiv1.PersistentVolumeClaimSpec) {
	if pvcSpec != nil {
		if len(pvcSpec.AccessModes) == 0 {
			pvcSpec.AccessModes = []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
			}
			log.Infof(`Using "%v" as AccessModes in "%v"`, apiv1.ReadWriteOnce, *pvcSpec)
		}
		// volume claim templates
		// Dynamically attach volume
		statefulSet.Spec.VolumeClaimTemplates = []apiv1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
					Annotations: map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					},
				},
				Spec: *pvcSpec,
			},
		}
	} else {
		// Attach Empty directory
		statefulSet.Spec.Template.Spec.Volumes = append(
			statefulSet.Spec.Template.Spec.Volumes,
			apiv1.Volume{
				Name: "data",
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{},
				},
			},
		)
	}
}

// ---> Start
//TODO: Use this method to add initial script, if supported
// Otherwise, remove it
func addInitialScript(statefulSet *apps.StatefulSet, script *tapi.ScriptSourceSpec) {
	statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts,
		apiv1.VolumeMount{
			Name:      "initial-script",
			MountPath: "/docker-entrypoint-initdb.d",
		},
	)
	//fmt.Println("\n\n>>>> ScriptPath",script.ScriptPath)
	//statefulSet.Spec.Template.Spec.Containers[0].Args = []string{
	//	// Add additional args
	//	script.ScriptPath,
	//}

	statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes,
		apiv1.Volume{
			Name:         "initial-script",
			VolumeSource: script.VolumeSource,
		},
	)
}

// ---> End

func (c *Controller) createDormantDatabase(mysql *tapi.MySQL) (*tapi.DormantDatabase, error) {
	dormantDb := &tapi.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysql.Name,
			Namespace: mysql.Namespace,
			Labels: map[string]string{
				tapi.LabelDatabaseKind: tapi.ResourceKindMySQL,
			},
		},
		Spec: tapi.DormantDatabaseSpec{
			Origin: tapi.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        mysql.Name,
					Namespace:   mysql.Namespace,
					Labels:      mysql.Labels,
					Annotations: mysql.Annotations,
				},
				Spec: tapi.OriginSpec{
					MySQL: &mysql.Spec,
				},
			},
		},
	}

	initSpec, _ := json.Marshal(mysql.Spec.Init)
	if initSpec != nil {
		dormantDb.Annotations = map[string]string{
			tapi.MySQLInitSpec: string(initSpec),
		}
	}

	dormantDb.Spec.Origin.Spec.MySQL.Init = nil

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}

func (c *Controller) reCreateMySQL(mysql *tapi.MySQL) error {
	_mysql := &tapi.MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:        mysql.Name,
			Namespace:   mysql.Namespace,
			Labels:      mysql.Labels,
			Annotations: mysql.Annotations,
		},
		Spec:   mysql.Spec,
		Status: mysql.Status,
	}

	if _, err := c.ExtClient.MySQLs(_mysql.Namespace).Create(_mysql); err != nil {
		return err
	}

	return nil
}

const (
	SnapshotProcess_Restore  = "restore"
	snapshotType_DumpRestore = "dump-restore"
)

func (c *Controller) createRestoreJob(mysql *tapi.MySQL, snapshot *tapi.Snapshot) (*batch.Job, error) {
	databaseName := mysql.Name
	jobName := snapshot.OffshootName()
	jobLabel := map[string]string{
		tapi.LabelDatabaseName: databaseName,
		tapi.LabelJobType:      SnapshotProcess_Restore,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(mysql.Spec.Storage, jobName, mysql.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, _ := snapshot.Location()

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   jobName,
			Labels: jobLabel,
		},
		Spec: batch.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  SnapshotProcess_Restore,
							Image: fmt.Sprintf("%s:%s", docker.ImageMySQL, mysql.Spec.Version),
							Args: []string{
								fmt.Sprintf(`--process=%s`, SnapshotProcess_Restore),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							Resources: snapshot.Spec.Resources,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      persistentVolume.Name,
									MountPath: "/var/" + snapshotType_DumpRestore + "/",
								},
								{
									Name:      "osmconfig",
									MountPath: storage.SecretMountPath,
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []apiv1.Volume{
						{
							Name:         persistentVolume.Name,
							VolumeSource: persistentVolume.VolumeSource,
						},
						{
							Name: "osmconfig",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: snapshot.Name,
								},
							},
						},
					},
					RestartPolicy: apiv1.RestartPolicyNever,
				},
			},
		},
	}
	if snapshot.Spec.SnapshotStorageSpec.Local != nil {
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, apiv1.VolumeMount{
			Name:      "local",
			MountPath: snapshot.Spec.SnapshotStorageSpec.Local.Path,
		})
		volume := apiv1.Volume{
			Name:         "local",
			VolumeSource: snapshot.Spec.SnapshotStorageSpec.Local.VolumeSource,
		}
		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, volume)
	}
	return c.Client.BatchV1().Jobs(mysql.Namespace).Create(job)
}

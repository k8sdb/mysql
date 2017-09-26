package controller

import (
	"fmt"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
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

	_, err := c.Client.CoreV1().Services(mysql.Namespace).Create(svc)
	if err != nil {
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
							Name:            tapi.ResourceNameMySQL,
							Image:           fmt.Sprintf("%s:%s", "mysql", mysql.Spec.Version), //<<<<<<<<< image name Needs update later #LATER
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "db",
									ContainerPort: 3306,
								},
							},
							Resources: mysql.Spec.Resources,
							//VolumeMounts: []apiv1.VolumeMount{
							//	{
							//		Name:      "secret",
							//		MountPath: "/srv/" + tapi.ResourceNameMySQL + "/secrets",
							//	},
							//	{
							//		Name:      "data",
							//		MountPath: "/var/pv",
							//	},
							//},
							//Args: []string{modeBasic},
							Env: []apiv1.EnvVar{
								{Name: "MYSQL_ROOT_PASSWORD", Value: "test"}, // #Later
								// Root password is set to test
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

	//if mysql.Spec.Monitor != nil &&
	//	mysql.Spec.Monitor.Agent == tapi.AgentCoreosPrometheus &&
	//	mysql.Spec.Monitor.Prometheus != nil {
	//	exporter := apiv1.Container{
	//		Name: "exporter",
	//		Args: []string{
	//			"export",
	//			fmt.Sprintf("--address=:%d", tapi.PrometheusExporterPortNumber),
	//			"--v=3",
	//		},
	//		Image:           docker.ImageOperator + ":" + c.opt.ExporterTag,
	//		ImagePullPolicy: apiv1.PullIfNotPresent,
	//		Ports: []apiv1.ContainerPort{
	//			{
	//				Name:          tapi.PrometheusExporterPortName,
	//				Protocol:      apiv1.ProtocolTCP,
	//				ContainerPort: int32(tapi.PrometheusExporterPortNumber),
	//			},
	//		},
	//	}
	//	statefulSet.Spec.Template.Spec.Containers = append(statefulSet.Spec.Template.Spec.Containers, exporter)
	//}

	//if mysql.Spec.DatabaseSecret == nil {
	//	secretVolumeSource, err := c.createDatabaseSecret(mysql)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	_mysql, err := kutildb.TryPatchPostgres(c.ExtClient, mysql.ObjectMeta, func(in *tapi.MySQL) *tapi.MySQL {
	//		in.Spec.DatabaseSecret = secretVolumeSource
	//		return in
	//	})
	//	if err != nil {
	//		c.recorder.Eventf(mysql.ObjectReference(), apiv1.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
	//		return nil, err
	//	}
	//	mysql = _mysql
	//}

	// Add secretVolume for authentication
	//addSecretVolume(statefulSet, mysql.Spec.DatabaseSecret)
	//
	//// Add Data volume for StatefulSet
	addDataVolume(statefulSet, mysql.Spec.Storage)
	//
	//// Add InitialScript to run at startup
	//if mysql.Spec.Init != nil && mysql.Spec.Init.ScriptSource != nil {
	//	addInitialScript(statefulSet, mysql.Spec.Init.ScriptSource)
	//}
	//
	//if c.opt.EnableRbac {
	//	// Ensure ClusterRoles for database statefulsets
	//	if err := c.createRBACStuff(mysql); err != nil {
	//		return nil, err
	//	}
	//
	//	statefulSet.Spec.Template.Spec.ServiceAccountName = mysql.Name
	//}

	if _, err := c.Client.AppsV1beta1().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
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

	return nil,nil

	// #LATER
	//databaseName := mysql.Name
	//jobName := snapshot.OffshootName()
	//jobLabel := map[string]string{
	//	tapi.LabelDatabaseName: databaseName,
	//	tapi.LabelJobType:      SnapshotProcess_Restore,
	//}
	//backupSpec := snapshot.Spec.SnapshotStorageSpec
	//bucket, err := backupSpec.Container()
	//if err != nil {
	//	return nil, err
	//}
	//
	//// Get PersistentVolume object for Backup Util pod.
	//persistentVolume, err := c.getVolumeForSnapshot(mysql.Spec.Storage, jobName, mysql.Namespace)
	//if err != nil {
	//	return nil, err
	//}
	//
	//// Folder name inside Cloud bucket where backup will be uploaded
	//folderName, _ := snapshot.Location()
	//
	//job := &batch.Job{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:   jobName,
	//		Labels: jobLabel,
	//	},
	//	Spec: batch.JobSpec{
	//		Template: apiv1.PodTemplateSpec{
	//			ObjectMeta: metav1.ObjectMeta{
	//				Labels: jobLabel,
	//			},
	//			Spec: apiv1.PodSpec{
	//				Containers: []apiv1.Container{
	//					{
	//						Name: SnapshotProcess_Restore,
	//						//TODO: Use appropriate image
	//						Image: fmt.Sprintf("%s:%s", docker.ImageMySQL, mysql.Spec.Version),
	//						Args: []string{
	//							fmt.Sprintf(`--process=%s`, SnapshotProcess_Restore),
	//							fmt.Sprintf(`--host=%s`, databaseName),
	//							fmt.Sprintf(`--bucket=%s`, bucket),
	//							fmt.Sprintf(`--folder=%s`, folderName),
	//							fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
	//						},
	//						Resources: snapshot.Spec.Resources,
	//						VolumeMounts: []apiv1.VolumeMount{
	//							//TODO: Mount secret volume if necessary
	//							{
	//								Name:      persistentVolume.Name,
	//								MountPath: "/var/" + snapshotType_DumpRestore + "/",
	//							},
	//							{
	//								Name:      "osmconfig",
	//								MountPath: storage.SecretMountPath,
	//								ReadOnly:  true,
	//							},
	//						},
	//					},
	//				},
	//				Volumes: []apiv1.Volume{
	//					//TODO: Add secret volume if necessary
	//					// Check postgres repository for example
	//					{
	//						Name:         persistentVolume.Name,
	//						VolumeSource: persistentVolume.VolumeSource,
	//					},
	//					{
	//						Name: "osmconfig",
	//						VolumeSource: apiv1.VolumeSource{
	//							Secret: &apiv1.SecretVolumeSource{
	//								SecretName: snapshot.Name,
	//							},
	//						},
	//					},
	//				},
	//				RestartPolicy: apiv1.RestartPolicyNever,
	//			},
	//		},
	//	},
	//}
	//if snapshot.Spec.SnapshotStorageSpec.Local != nil {
	//	job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, apiv1.VolumeMount{
	//		Name:      "local",
	//		MountPath: snapshot.Spec.SnapshotStorageSpec.Local.Path,
	//	})
	//	volume := apiv1.Volume{
	//		Name:         "local",
	//		VolumeSource: snapshot.Spec.SnapshotStorageSpec.Local.VolumeSource,
	//	}
	//	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, volume)
	//}
	//return c.Client.BatchV1().Jobs(mysql.Namespace).Create(job)
}
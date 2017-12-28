package controller

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/docker"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/apimachinery/pkg/storage"
	apps "k8s.io/api/apps/v1beta1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// Duration in Minute
	// Check whether pod under StatefulSet is running or not
	// Continue checking for this duration until failure
	durationCheckStatefulSet = time.Minute * 30
)








// Set root user password from Secret, Through Env.
func setEnvFromSecret(statefulSet *apps.StatefulSet, secSource *core.SecretVolumeSource) {
	statefulSet.Spec.Template.Spec.Containers[0].Env = append(statefulSet.Spec.Template.Spec.Containers[0].Env,
		core.EnvVar{
			Name: "MYSQL_ROOT_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: secSource.SecretName,
					},
					Key: ".admin",
				},
			},
		},
	)
}

func addDataVolume(statefulSet *apps.StatefulSet, pvcSpec *core.PersistentVolumeClaimSpec) {
	if pvcSpec != nil {
		if len(pvcSpec.AccessModes) == 0 {
			pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
				core.ReadWriteOnce,
			}
			log.Infof(`Using "%v" as AccessModes in "%v"`, core.ReadWriteOnce, *pvcSpec)
		}
		// volume claim templates
		// Dynamically attach volume
		statefulSet.Spec.VolumeClaimTemplates = []core.PersistentVolumeClaim{
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
			core.Volume{
				Name: "data",
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{},
				},
			},
		)
	}
}

func addInitialScript(statefulSet *apps.StatefulSet, script *api.ScriptSourceSpec) {
	statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts,
		core.VolumeMount{
			Name:      "initial-script",
			MountPath: "/docker-entrypoint-initdb.d",
		},
	)

	statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes,
		core.Volume{
			Name:         "initial-script",
			VolumeSource: script.VolumeSource,
		},
	)
}

func (c *Controller) createDormantDatabase(mysql *api.MySQL) (*api.DormantDatabase, error) {
	dormantDb := &api.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysql.Name,
			Namespace: mysql.Namespace,
			Labels: map[string]string{
				api.LabelDatabaseKind: api.ResourceKindMySQL,
			},
		},
		Spec: api.DormantDatabaseSpec{
			Origin: api.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        mysql.Name,
					Namespace:   mysql.Namespace,
					Labels:      mysql.Labels,
					Annotations: mysql.Annotations,
				},
				Spec: api.OriginSpec{
					MySQL: &mysql.Spec,
				},
			},
		},
	}

	initSpec, _ := json.Marshal(mysql.Spec.Init)
	if mysql.Spec.Init != nil {
		dormantDb.Annotations = map[string]string{
			api.MySQLInitSpec: string(initSpec),
		}
	}

	dormantDb.Spec.Origin.Spec.MySQL.Init = nil

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}

func (c *Controller) reCreateMySQL(mysql *api.MySQL) error {
	_mysql := &api.MySQL{
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

func (c *Controller) createRestoreJob(mysql *api.MySQL, snapshot *api.Snapshot) (*batch.Job, error) {
	databaseName := mysql.Name
	jobName := snapshot.OffshootName()
	jobLabel := map[string]string{
		api.LabelDatabaseName: databaseName,
		api.LabelJobType:      SnapshotProcess_Restore,
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
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name: SnapshotProcess_Restore,
							//Image: fmt.Sprintf("%s:%s-util", docker.ImageMySQL, mysql.Spec.Version), //todo
							Image: fmt.Sprintf("kubedb/mysql:8.0-util"),
							Args: []string{
								fmt.Sprintf(`--process=%s`, SnapshotProcess_Restore),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							Resources: snapshot.Spec.Resources,

							VolumeMounts: []core.VolumeMount{
								{
									Name:      "secret",
									MountPath: "/srv/" + api.ResourceNameMySQL + "/secrets",
								},
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
					Volumes: []core.Volume{
						{
							Name: "secret",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: mysql.Spec.DatabaseSecret.SecretName,
								},
							},
						},
						{
							Name:         persistentVolume.Name,
							VolumeSource: persistentVolume.VolumeSource,
						},
						{
							Name: "osmconfig",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: snapshot.OSMSecretName(),
								},
							},
						},
					},
					RestartPolicy: core.RestartPolicyNever,
				},
			},
		},
	}
	if snapshot.Spec.SnapshotStorageSpec.Local != nil {
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "local",
			MountPath: snapshot.Spec.SnapshotStorageSpec.Local.Path,
		})
		volume := core.Volume{
			Name:         "local",
			VolumeSource: snapshot.Spec.SnapshotStorageSpec.Local.VolumeSource,
		}
		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, volume)
	}
	return c.Client.BatchV1().Jobs(mysql.Namespace).Create(job)
}

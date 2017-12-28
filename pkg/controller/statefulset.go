package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/appscode/kutil"
	app_util "github.com/appscode/kutil/apps/v1beta1"
	core_util "github.com/appscode/kutil/core/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/apimachinery/pkg/eventer"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Controller) ensureStatefulSet(mysql *api.MySQL) (kutil.VerbType, error) {
	if err := c.checkStatefulSet(mysql); err != nil {
		return kutil.VerbUnchanged, err
	}

	if err := c.ensureDatabaseSecret(mysql) ; err!=nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for MySQL database
	statefulSet, vt, err := c.createStatefulSet(mysql)
	if err != nil {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create StatefulSet. Reason: %v",
			err,
		)
		return err
	}

	// Check StatefulSet Pod status
	if err := c.CheckStatefulSetPodStatus(statefulSet, durationCheckStatefulSet); err != nil {
		c.recorder.Eventf(
			mysql.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToStart,
			`Failed to create StatefulSet. Reason: %v`,
			err,
		)
		return err
	} else {
		c.recorder.Event(
			mysql.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully created StatefulSet",
		)
	}

	return nil
}

func (c *Controller) checkStatefulSet(mysql *api.MySQL) error {
	// SatatefulSet for MySQL database
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(mysql.Namespace).Get(mysql.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return  nil
		} else {
			return err
		}
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindMySQL {
		return  fmt.Errorf(`Intended statefulSet "%v" already exists`, mysql.OffshootName())
	}

	return nil
}

func (c *Controller) createStatefulSet(mysql *api.MySQL) (*apps.StatefulSet, kutil.VerbType, error) {
	statefulSetMeta := metav1.ObjectMeta{
		Name:      mysql.OffshootName(),
		Namespace: mysql.Namespace,
	}
	st,vt,err := app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.Labels = core_util.UpsertMap(in.Labels, mysql.StatefulSetLabels())
		in.Annotations = core_util.UpsertMap(in.Annotations, mysql.StatefulSetAnnotations())

		in.Spec.Replicas = types.Int32P(1)
		in.Spec.Template = core.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: in.ObjectMeta.Labels,
			},
		}

		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
			Name:            api.ResourceNameMySQL,
			Image:           c.opt.Docker.GetImageWithTag(mysql),
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          "db",
					ContainerPort: 3306,
					Protocol:      core.ProtocolTCP,
				},
			},
			Resources: mysql.Spec.Resources,
		})
		if mysql.Spec.Monitor != nil &&
			mysql.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
			mysql.Spec.Monitor.Prometheus != nil {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(in.Spec.Template.Spec.Containers, core.Container{
				Name: "exporter",
				Args: []string{
					"export",
					fmt.Sprintf("--address=:%d", mysql.Spec.Monitor.Prometheus.Port),
					"--v=3",
				},
				Image:           c.opt.Docker.GetOperatorImageWithTag(mysql),
				ImagePullPolicy: core.PullIfNotPresent,
				Ports: []core.ContainerPort{
					{
						Name:          api.PrometheusExporterPortName,
						Protocol:      core.ProtocolTCP,
						ContainerPort: mysql.Spec.Monitor.Prometheus.Port,
					},
				},
			})
		}

		in = upsertDataVolume(in, mysql)

		in.Spec.Template.Spec.NodeSelector = mysql.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = mysql.Spec.Affinity
		in.Spec.Template.Spec.SchedulerName = mysql.Spec.SchedulerName
		in.Spec.Template.Spec.Tolerations = mysql.Spec.Tolerations
		if c.opt.EnableRbac {
			in.Spec.Template.Spec.ServiceAccountName = mysql.Name
		}

		in.Spec.UpdateStrategy.Type = apps.RollingUpdateStatefulSetStrategyType

		return in
	})
	return st,vt,err


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
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: mysql.OffshootLabels(),
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            api.ResourceNameMySQL,
							Image:           fmt.Sprintf("%s:%s", docker.ImageMySQL, mysql.Spec.Version),
							ImagePullPolicy: core.PullIfNotPresent,
							Ports: []core.ContainerPort{
								{
									Name:          "db",
									ContainerPort: 3306,
								},
							},
							Resources: mysql.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								{
									Name:      "data",
									MountPath: "/var/lib/mysql", //Volume path of mysql, ref: https://github.com/docker-library/mysql/blob/86431f073b3d2f963d21e33cb8943f0bdcdf143d/8.0/Dockerfile#L48
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
		mysql.Spec.Monitor.Agent == api.AgentCoreosPrometheus &&
		mysql.Spec.Monitor.Prometheus != nil {
		exporter := core.Container{
			Name: "exporter",
			Args: []string{
				"export",
				fmt.Sprintf("--address=:%d", mysql.Spec.Monitor.Prometheus.Port),
				"--v=3",
			},
			Image:           docker.ImageOperator + ":" + c.opt.ExporterTag,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          api.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: mysql.Spec.Monitor.Prometheus.Port,
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

		_mysql,vt, err := util.PatchMySQL(c.ExtClient, mysql, func(in *api.MySQL) *api.MySQL {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
		if err != nil {
			c.recorder.Eventf(mysql.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return nil, err
		}
		mysql.Spec.DatabaseSecret = _mysql.Spec.DatabaseSecret
	}

	//Set root user password from Secret
	setEnvFromSecret(statefulSet, mysql.Spec.DatabaseSecret)

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, mysql.Spec.Storage)

	if mysql.Spec.Init != nil && mysql.Spec.Init.ScriptSource != nil {
		addInitialScript(statefulSet, mysql.Spec.Init.ScriptSource)
	}

	if c.opt.EnableRbac {
		// Ensure ClusterRoles for database statefulsets
		if err := c.ensureRBACStuff(mysql); err != nil {
			return nil, err
		}

		statefulSet.Spec.Template.Spec.ServiceAccountName = mysql.Name
	}

	if _, err := c.Client.AppsV1beta1().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
}

func upsertDataVolume(statefulSet *apps.StatefulSet, mysql *api.MySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNameMySQL {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: "/var/lib/mysql",
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := mysql.Spec.Storage
			if pvcSpec != nil {
				if len(pvcSpec.AccessModes) == 0 {
					pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
						core.ReadWriteOnce,
					}
					log.Infof(`Using "%v" as AccessModes in mysql.Spec.Storage`, core.ReadWriteOnce)
				}

				volumeClaim := core.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: *pvcSpec,
				}
				if pvcSpec.StorageClassName != nil {
					volumeClaim.Annotations = map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					}
				}
				volumeClaims := statefulSet.Spec.VolumeClaimTemplates
				volumeClaims = core_util.UpsertVolumeClaim(volumeClaims, volumeClaim)
				statefulSet.Spec.VolumeClaimTemplates = volumeClaims
			} else {
				volume := core.Volume{
					Name: "data",
					VolumeSource: core.VolumeSource{
						EmptyDir: &core.EmptyDirVolumeSource{},
					},
				}
				volumes := statefulSet.Spec.Template.Spec.Volumes
				volumes = core_util.UpsertVolume(volumes, volume)
				statefulSet.Spec.Template.Spec.Volumes = volumes
				return statefulSet
			}
			break
		}
	}
	return statefulSet
}

func upsertEnv(statefulSet *apps.StatefulSet, mysql *api.MySQL, envs []core.EnvVar) *apps.StatefulSet {
	envList := []core.EnvVar{
		{
			Name: "MYSQL_ROOT_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: mysql.Spec.DatabaseSecret.SecretName,
					},
					Key: ".admin",
				},
			},
		},
	}

	envList = append(envList, envs...)

	// To do this, Upsert Container first
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceNameMySQL {
			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envList...)
			return statefulSet
		}
	}

	return statefulSet
}

func (c *Controller) checkStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
	err := core_util.WaitUntilPodRunningBySelector(
		c.Client,
		statefulSet.Namespace,
		statefulSet.Spec.Selector,
		int(types.Int32(statefulSet.Spec.Replicas)),
	)
	if err != nil {
		return err
	}
	return nil
}

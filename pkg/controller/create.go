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
	modeBasic = "basic"
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
	//if mysql.Spec.Monitor != nil &&
	//	mysql.Spec.Monitor.Agent == tapi.AgentCoreosPrometheus &&
	//	mysql.Spec.Monitor.Prometheus != nil {
	//	svc.Spec.Ports = append(svc.Spec.Ports, apiv1.ServicePort{
	//		Name:       tapi.PrometheusExporterPortName,
	//		Port:       tapi.PrometheusExporterPortNumber,
	//		TargetPort: intstr.FromString(tapi.PrometheusExporterPortName),
	//	})
	//}

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
							Name:            tapi.ResourceNameMySQL,
							Image:           fmt.Sprintf("%s:%s", "mysql", mysql.Spec.Version), //<<<<<<<<< Needs update later
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
	addSecretVolume(statefulSet, mysql.Spec.DatabaseSecret)

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, mysql.Spec.Storage)

	// Add InitialScript to run at startup
	if mysql.Spec.Init != nil && mysql.Spec.Init.ScriptSource != nil {
		addInitialScript(statefulSet, mysql.Spec.Init.ScriptSource)
	}

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
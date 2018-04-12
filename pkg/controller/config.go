package controller

import (
	"github.com/appscode/go/log/golog"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	snapc "github.com/kubedb/apimachinery/pkg/controller/snapshot"
	"github.com/kubedb/apimachinery/pkg/eventer"
	"github.com/kubedb/mysql/pkg/docker"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	AnalyticsClientID string
	EnableAnalytics   = true
	LoggerOptions     golog.Options
)

type OperatorConfig struct {
	amc.Config

	ClientConfig     *rest.Config
	KubeClient       kubernetes.Interface
	APIExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	DBClient         cs.Interface
	PromClient       pcm.MonitoringV1Interface
	CronController   snapc.CronControllerInterface
	Docker           docker.Docker
}

func NewOperatorConfig(clientConfig *rest.Config) *OperatorConfig {
	return &OperatorConfig{
		ClientConfig: clientConfig,
	}
}

func (c *OperatorConfig) New() (*Controller, error) {
	ctrl := &Controller{
		Controller: &amc.Controller{
			Client:           c.KubeClient,
			ExtClient:        c.DBClient.KubedbV1alpha1(),
			ApiExtKubeClient: c.APIExtKubeClient,
		},
		Config:         c.Config,
		docker:         c.Docker,
		promClient:     c.PromClient,
		cronController: c.CronController,
		recorder:       eventer.NewEventRecorder(c.KubeClient, "mysql operator"),
	}

	if err := ctrl.Init(); err != nil {
		return nil, err
	}

	return ctrl, nil
}

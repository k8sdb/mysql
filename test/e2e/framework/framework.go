package framework

import (
	"github.com/appscode/go/crypto/rand"
	tcs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	clientset "k8s.io/client-go/kubernetes"
)

type Framework struct {
	kubeClient   clientset.Interface
	extClient    tcs.KubedbV1alpha1Interface
	namespace    string
	name         string
	StorageClass string
}

func New(kubeClient clientset.Interface, extClient tcs.KubedbV1alpha1Interface, storageClass string) *Framework {
	return &Framework{
		kubeClient:   kubeClient,
		extClient:    extClient,
		name:         "mysql-operator",
		namespace:    rand.WithUniqSuffix("mysql"),
		StorageClass: storageClass,
	}
}

func (f *Framework) Invoke() *Invocation {
	return &Invocation{
		Framework: f,
		app:       rand.WithUniqSuffix("mysql-e2e"),
	}
}

type Invocation struct {
	*Framework
	app string
}

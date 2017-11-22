package unit_tests

import (
	"log"
	"path/filepath"
	"testing"

	"github.com/appscode/go/hold"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/mysql/pkg/controller"
	"github.com/mitchellh/go-homedir"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestController_Run(t *testing.T) {
	ctrl := GetNewController()
	ctrl.Run()
	hold.Hold()
}

func GetNewController() *controller.Controller {
	userHome, err := homedir.Dir()
	if err != nil {
		log.Fatalln(err)
	}

	// Kubernetes config
	kubeconfigPath := filepath.Join(userHome, ".kube/config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatalln(err)
	}
	// Clients
	kubeClient := clientset.NewForConfigOrDie(config)
	apiExtKubeClient := crd_cs.NewForConfigOrDie(config)
	extClient := cs.NewForConfigOrDie(config)
	// Framework

	cronController := amc.NewCronController(kubeClient, extClient)
	// Start Cron
	cronController.StartCron()

	// Controller
	//return New(kubeClient, apiExtKubeClient, extClient, nil, cronController, opt)
	//ctrl.Run()
	//root.EventuallyCRD().Should(Succeed())
	//return ctrl
	return controller.New(kubeClient, apiExtKubeClient, extClient, nil, cronController, controller.Options{
		GoverningService: api.DatabaseNamePrefix,
	})
}

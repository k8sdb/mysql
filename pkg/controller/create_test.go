package controller

import (
	"log"
	"path/filepath"
	"testing"

	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//apiv1 "k8s.io/client-go/pkg/api/v1"
	"fmt"
	tcs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/mitchellh/go-homedir"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"encoding/json"
)

func TestController_Run(t *testing.T) {
	ctrl := GetNewController()
	ctrl.Run()
}

func TestGetMS(t *testing.T) {
	c := GetNewController()
	_, _ = c.ExtClient.MySQLs("default").Create(&tapi.MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name: "t1",
		},
		Spec: tapi.MySQLSpec{
			Init: &tapi.InitSpec{
				ScriptSource: &tapi.ScriptSourceSpec{
					ScriptPath: "/var/var",
				},
			},
		},
	},
	)

	my, _ := c.ExtClient.MySQLs("default").Get("t1", metav1.GetOptions{})

	data, _ := json.MarshalIndent(my.Spec,"","  ")
	fmt.Println(string(data))
}

func GetNewController() *Controller {
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
	apiExtKubeClient := apiextensionsclient.NewForConfigOrDie(config)
	extClient := tcs.NewForConfigOrDie(config)
	// Framework

	cronController := amc.NewCronController(kubeClient, extClient)
	// Start Cron
	cronController.StartCron()

	// Controller
	//return New(kubeClient, apiExtKubeClient, extClient, nil, cronController, opt)
	//ctrl.Run()
	//root.EventuallyCRD().Should(Succeed())
	//return ctrl
	return New(kubeClient, apiExtKubeClient, extClient, nil, cronController, Options{
		GoverningService: tapi.DatabaseNamePrefix,
	})
}

func DemoMySQL() *tapi.MySQL {
	//var rscList map[apiv1.ResourceName]resource.Quantity
	//
	//rscList["storage"] = resource.Quantity{50 * 2^20}

	return &tapi.MySQL{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubedb.com/v1alpha1",
			Kind:       "MySQL",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "m12",
			Namespace: "demo",
		},
		Spec: tapi.MySQLSpec{
			Version:    "8.0",
			Replicas:   1,
			DoNotPause: true,
			//Storage: &apiv1.PersistentVolumeClaimSpec{
			//	StorageClassName: &"standard",
			//	AccessModes: []apiv1.PersistentVolumeAccessMode {
			//		"ReadWriteOnce",
			//	},
			//	Resources: apiv1.ResourceRequirements{
			//		Requests:
			//	},
			//},
		},
	}
}

func TestController_ensureCRD(t *testing.T) {
	ctrl := GetNewController()
	ctrl.ensureCustomResourceDefinition()
	log.Println("Done ensuring CRD.")
}

func TestController_CreateService(t *testing.T) {
	ctrl := GetNewController()
	log.Println("Starting Test!!!!!!!!!!!")
	ctrl.ensureCustomResourceDefinition()
	log.Println("Done ensuring CRD.")

	log.Println("Creating Service!")
	msql := DemoMySQL()
	if err := ctrl.createService(msql); err != nil {
		log.Println("error whiloe creating service", err)
	} else {
		log.Println("Service created!!")
	}
}

func TestController_FindService(t *testing.T) {
	ctrl := GetNewController()

	log.Println("Finding Service!")
	msql := DemoMySQL()
	if flag, err := ctrl.findService(msql); flag {
		log.Println("Service Found!!")
	} else {
		log.Println("No services!! >", err)
	}

}

func TestController_createStatefulSet(t *testing.T) {
	ctrl := GetNewController()

	log.Println("Creating stateful Set")
	msql := DemoMySQL()

	rsl, err := ctrl.createStatefulSet(msql)
	if err != nil {
		log.Println("Error while creating Statefulset", err)
	} else {
		log.Println("statefulset creation successful", rsl)
	}

}

func TestController_findStatefulSet(t *testing.T) {
	ctrl := GetNewController()

	log.Println("Finding stateful Set")
	msql := DemoMySQL()

	rsl, err := ctrl.findStatefulSet(msql)
	if rsl {
		log.Println("Stateful set exists!!")
	} else {
		log.Println("no such statefulset! > ", err)
	}
}

func TestController_reCreateMySQL(t *testing.T) {
	ctrl := GetNewController()

	log.Println("Re-creating mysql")
	msql := DemoMySQL()

	err := ctrl.reCreateMySQL(msql)
	if err != nil {
		log.Println("Error while re-creating MySQL", err)
	} else {
		log.Println("re-creation successful")
	}
}


func TestController_create(t *testing.T) {

	ctrl := GetNewController()
	msql := DemoMySQL()
	log.Println("creating!!")
	if err := ctrl.create(msql); err != nil {
		log.Println("error creating MySQL", err)
	} else {
		log.Println("Creating task succesfull")
	}
}
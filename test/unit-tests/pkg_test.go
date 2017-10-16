package unit_tests

import (
	"log"
	"path/filepath"
	"testing"

	apiv1 "k8s.io/client-go/pkg/api/v1"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"encoding/json"
	"fmt"

	tcs "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1"
	amc "github.com/k8sdb/apimachinery/pkg/controller"
	"github.com/k8sdb/mysql/pkg/controller"
	"github.com/mitchellh/go-homedir"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/appscode/go/hold"
)

func TestController_Run(t *testing.T) {
	ctrl := GetNewController()
	ctrl.Run()
	hold.Hold()
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

	data, _ := json.MarshalIndent(my.Spec, "", "  ")
	fmt.Println(string(data))
}

func TestGetSecretName(t *testing.T)  {
	c := GetNewController()
	_, _ = c.ExtClient.MySQLs("default").Create(&tapi.MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name: "t2",
		},
		Spec: tapi.MySQLSpec{
			Init: &tapi.InitSpec{
				ScriptSource: &tapi.ScriptSourceSpec{
					ScriptPath: "/var/var",
				},
			},
			DatabaseSecret: &apiv1.SecretVolumeSource{
				SecretName: "t1-admin-auth",
			},

		},
	},
	)

	my, _ := c.ExtClient.MySQLs("default").Get("t2", metav1.GetOptions{})

	data, _ := json.MarshalIndent(my, "", "  ")
	fmt.Println(string(data))
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
	return controller.New(kubeClient, apiExtKubeClient, extClient, nil, cronController, controller.Options{
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
		},
	}
}

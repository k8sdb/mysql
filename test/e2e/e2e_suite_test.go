package e2e_test

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/appscode/go/homedir"
	"github.com/appscode/go/log"
	logs "github.com/appscode/go/log/golog"
	pcm "github.com/coreos/prometheus-operator/pkg/client/monitoring/v1"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	snapc "github.com/kubedb/apimachinery/pkg/controller/snapshot"
	"github.com/kubedb/mysql/pkg/controller"
	"github.com/kubedb/mysql/pkg/docker"
	"github.com/kubedb/mysql/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	storageClass   string
	dockerRegistry string

	prometheusCrdGroup = pcm.Group
	prometheusCrdKinds = pcm.DefaultCrdKinds
)

func init() {
	flag.StringVar(&storageClass, "storageclass", "standard", "Kubernetes StorageClass name")
	flag.StringVar(&dockerRegistry, "docker-registry", "kubedb", "User provided docker repository")
}

const (
	TIMEOUT = 20 * time.Minute
)

var (
	ctrl *controller.Controller
	root *framework.Framework
)

func TestE2e(t *testing.T) {
	logs.InitLogs()
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(TIMEOUT)

	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "e2e Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {

	userHome := homedir.HomeDir()

	// Kubernetes config
	kubeconfigPath := filepath.Join(userHome, ".kube/config")
	By("Using kubeconfig from " + kubeconfigPath)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	Expect(err).NotTo(HaveOccurred())
	// raise throttling time. ref: https://github.com/appscode/voyager/issues/640
	config.Burst = 100
	config.QPS = 100

	// Clients
	kubeClient := kubernetes.NewForConfigOrDie(config)
	apiExtKubeClient := crd_cs.NewForConfigOrDie(config)
	extClient := cs.NewForConfigOrDie(config)
	promClient, err := pcm.NewForConfig(&prometheusCrdKinds, prometheusCrdGroup, config)
	if err != nil {
		log.Fatalln(err)
	}
	// Framework
	root = framework.New(config, kubeClient, extClient, storageClass)

	By("Using namespace " + root.Namespace())

	// Create namespace
	err = root.CreateNamespace()
	Expect(err).NotTo(HaveOccurred())

	cronController := snapc.NewCronController(kubeClient, extClient)
	// Start Cron
	cronController.StartCron()

	opt := controller.Options{
		Docker: docker.Docker{
			Registry: dockerRegistry,
		},
		EnableAnalytics:   true,
		OperatorNamespace: root.Namespace(),
		GoverningService:  api.DatabaseNamePrefix,
		AnalyticsClientID: "$kubedb$mysql$e2e",
	}

	// Controller
	ctrl = controller.New(kubeClient, apiExtKubeClient, extClient, promClient, cronController, opt)
	err = ctrl.Setup()
	if err != nil {
		log.Fatalln(err)
	}
	ctrl.Run()
	root.EventuallyCRD().Should(Succeed())
})

var _ = AfterSuite(func() {
	root.CleanMySQL()
	root.CleanDormantDatabase()
	root.CleanSnapshot()
	err := root.DeleteNamespace()
	Expect(err).NotTo(HaveOccurred())
	By("Deleted namespace")
})

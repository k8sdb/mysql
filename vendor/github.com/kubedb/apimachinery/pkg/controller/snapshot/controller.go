package snapshot

import (
	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	api_listers "github.com/kubedb/apimachinery/client/listers/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	jobc "github.com/kubedb/apimachinery/pkg/controller/job"
	"github.com/kubedb/apimachinery/pkg/eventer"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Controller struct {
	*amc.Controller
	amc.Config
	// Snapshotter interface
	snapshotter amc.Snapshotter
	// tweakListOptions for watcher
	tweakListOptions func(*metav1.ListOptions)
	// Event Recorder
	eventRecorder record.EventRecorder
	// Snapshot
	snQueue    *queue.Worker
	snInformer cache.SharedIndexInformer
	snLister   api_listers.SnapshotLister
}

// NewController creates a new Controller
func NewController(
	controller *amc.Controller,
	snapshotter amc.Snapshotter,
	config amc.Config,
	tweakListOptions func(*metav1.ListOptions),
) *Controller {
	// return new DormantDatabase Controller
	return &Controller{
		Controller:       controller,
		snapshotter:      snapshotter,
		Config:           config,
		tweakListOptions: tweakListOptions,
		eventRecorder:    eventer.NewEventRecorder(controller.Client, "Job Controller"),
	}
}

func (c *Controller) EnsureCustomResourceDefinitions() error {
	crd := []*crd_api.CustomResourceDefinition{
		api.Snapshot{}.CustomResourceDefinition(),
	}
	return crdutils.RegisterCRDs(c.ApiExtKubeClient, crd)
}

// Init ensures snapshot watcher and returns queue.Worker.
// So, it is possible to start queue.run from other package/repositories
// Return type: Snapshot queue as 1st parameter and Job.Queue as 2nd.
func (c *Controller) Init() (*queue.Worker, *queue.Worker) {
	c.initWatcher()
	jobQueue := jobc.NewController(c.Controller, c.snapshotter, c.Config, c.tweakListOptions).Init()
	return c.snQueue, jobQueue
}

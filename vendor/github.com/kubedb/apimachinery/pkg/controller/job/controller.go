package job

import (
	"github.com/appscode/kutil/tools/queue"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	batch_listers "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Controller struct {
	*amc.Controller
	amc.Config
	// SnapshotDoer interface
	snapshotter amc.Snapshotter
	// tweakListOptions for watcher
	tweakListOptions func(*metav1.ListOptions)
	// Event Recorder
	eventRecorder record.EventRecorder
	// Job
	jobQueue    *queue.Worker
	jobInformer cache.SharedIndexInformer
	jobLister   batch_listers.JobLister
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

func (c *Controller) Init() *queue.Worker {
	c.initWatcher()
	return c.jobQueue
}

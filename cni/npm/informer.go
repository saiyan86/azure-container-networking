package npm

import (
	"flag"
	"fmt"
	"time"

	"github.com/golang/glog"

	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	//"k8s.io/client-go/pkg/api/v1"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// PodLoggingController logs the name and namespace of pods that are added,
// deleted, or updated
type PodLoggingController struct {
	informerFactory informers.SharedInformerFactory
	podInformer     coreinformers.PodInformer
}

// Run starts shared informers and waits for the shared informer cache to
// synchronize.
func (c *PodLoggingController) Run(stopCh chan struct{}) error {
	// Starts all the shared informers that have been created by the factory so
	// far.
	c.informerFactory.Start(stopCh)
	// wait for the initial synchronization of the local cache.
	if !cache.WaitForCacheSync(stopCh, c.podInformer.Informer().HasSynced) {
		return fmt.Errorf("Failed to sync")
	}
	return nil
}

func (c *PodLoggingController) podAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	glog.Infof("POD CREATED: %s/%s", pod.Namespace, pod.Name)
}

func (c *PodLoggingController) podUpdate(old, new interface{}) {
	oldPod := old.(*v1.Pod)
	newPod := new.(*v1.Pod)
	glog.Infof(
		"POD UPDATED. %s/%s %s",
		oldPod.Namespace, oldPod.Name, newPod.Status.Phase,
	)
}

func (c *PodLoggingController) podDelete(obj interface{}) {
	pod := obj.(*v1.Pod)
	glog.Infof("POD DELETED: %s/%s", pod.Namespace, pod.Name)
}

// NewPodLoggingController creates a PodLoggingController
func NewPodLoggingController(informerFactory informers.SharedInformerFactory) *PodLoggingController {
	podInformer := informerFactory.Core().V1().Pods()

	c := &PodLoggingController{
		informerFactory: informerFactory,
		podInformer:     podInformer,
	}
	podInformer.Informer().AddEventHandler(
		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			// Called on creation
			AddFunc: c.podAdd,
			// Called on resource update and every resyncPeriod on existing resources.
			UpdateFunc: c.podUpdate,
			// Called on resource deletion.
			DeleteFunc: c.podDelete,
		},
	)
	return c
}

func init() {
	flag.Parse()
}

func im() {
	// Creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
		panic(err.Error())
	}

	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
		panic(err.Error())
	}

	factory := informers.NewSharedInformerFactory(clientset, time.Hour*24)
	controller := NewPodLoggingController(factory)
	stop := make(chan struct{})
	defer close(stop)
	err = controller.Run(stop)
	if err != nil {
		glog.Fatal(err)
	}
	select {}
}
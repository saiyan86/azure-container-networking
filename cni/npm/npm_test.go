package npm

import (
	"testing"

	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func TestRun(t *testing.T) {

	// Creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// Creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("[cni-npm] clientset creation failed with error %v.\n", err)
		panic(err.Error())
	}

	factory := informers.NewSharedInformerFactory(clientset, time.Hour*24)
	npMgr := NewNetworkPolicyManager(clientset, factory)
	err = npMgr.Run(wait.NeverStop)
	if err != nil {
		fmt.Printf("[cni-npm] npm failed with error %v.\n", err)
		panic(err.Error)
	}
}

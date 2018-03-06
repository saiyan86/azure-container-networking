package npm

import (
	"fmt"

	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	networkinginformers "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/tools/cache"
)

// NetworkPolicyManager contains informers for pod, namespace and networkpolicy.
type NetworkPolicyManager struct {
	informerFactory informers.SharedInformerFactory
	podInformer     coreinformers.PodInformer
	nsInformer      coreinformers.NamespaceInformer
	npInformer      networkinginformers.NetworkPolicyInformer
}

// Run starts shared informers and waits for the shared informer cache to sync.
func (npMgr *NetworkPolicyManager) Run(stopCh chan struct{}) error {
	// Starts all informers manufactured by npMgr's informerFactory.
	npMgr.informerFactory.Start(stopCh)
	// Wait for the initial sync of local cache.
	if !cache.WaitForCacheSync(stopCh, npMgr.podInformer.Informer().HasSynced) {
		return fmt.Errorf("Pod informer failed to sync")
	}

	if !cache.WaitForCacheSync(stopCh, npMgr.nsInformer.Informer().HasSynced) {
		return fmt.Errorf("Namespace informer failed to sync")
	}

	if !cache.WaitForCacheSync(stopCh, npMgr.npInformer.Informer().HasSynced) {
		return fmt.Errorf("Namespace informer failed to sync")
	}

	return nil
}

// NewNetworkPolicyManager creates a NetworkPolicyManager
func NewNetworkPolicyManager(informerFactory informers.SharedInformerFactory) *NetworkPolicyManager {
	podInformer := informerFactory.Core().V1().Pods()
	nsInformer := informerFactory.Core().V1().Namespaces()
	npInformer := informerFactory.Networking().V1().NetworkPolicies()

	npMgr := &NetworkPolicyManager{
		informerFactory: informerFactory,
		podInformer:     podInformer,
		npInformer:      npInformer,
		nsInformer:      nsInformer,
	}

	podInformer.Informer().AddEventHandler(
		// Pod event handlers
		cache.ResourceEventHandlerFuncs{
			AddFunc:    npMgr.AddPod,
			UpdateFunc: npMgr.UpdatePod,
			DeleteFunc: npMgr.DeletePod,
		},
	)

	return npMgr
}

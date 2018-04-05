package npm

import (
	"fmt"

	"github.com/Azure/azure-container-networking/cni/npm/ipsm"
	"github.com/Azure/azure-container-networking/cni/npm/iptm"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type namespace struct {
	name    string
	podMap  map[types.UID]*corev1.Pod
	npQueue []*networkingv1.NetworkPolicy // TODO: Optimize to ordered map.
	ipsMgr  *ipsm.IpsetManager
	iptMgr  *iptm.IptablesManager
}

// newNS constructs a new namespace object.
func newNs(name string) (*namespace, error) {
	ns := &namespace{
		name:    name,
		podMap:  make(map[types.UID]*corev1.Pod),
		npQueue: []*networkingv1.NetworkPolicy{},
		ipsMgr:  ipsm.NewIpsetManager(),
		iptMgr:  iptm.NewIptablesManager(),
	}

	return ns, nil
}

func isSystemNs(nsObj *corev1.Namespace) bool {
	return nsObj.ObjectMeta.Name == "kube-system"
}

// AddNamespace handles add name space.
func (npMgr *NetworkPolicyManager) AddNamespace(nsObj *corev1.Namespace) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	// Don't deal with kube-system namespace
	if isSystemNs(nsObj) {
		return nil
	}

	nsName, nsNs := nsObj.ObjectMeta.Name, nsObj.ObjectMeta.Namespace
	fmt.Printf("NAMESPACE CREATED: %s/%s\n", nsName, nsNs)

	ns, exists := npMgr.nsMap[nsName]
	if !exists {
		newns, err := newNs(nsName)
		if err != nil {
			return err
		}
		npMgr.nsMap[nsName] = newns
		ns = newns
	}

	// Create ipset list for the namespace.
	ipsMgr := ns.ipsMgr
	if err := ipsMgr.CreateList(nsName); err != nil {
		fmt.Printf("Error creating ipset list %s.\n", nsName)
		return err
	}

	return nil
}

// UpdateNamespace handles update name space.
func (npMgr *NetworkPolicyManager) UpdateNamespace(oldNsObj *corev1.Namespace, newNsObj *corev1.Namespace) error {
	npMgr.Lock()

	oldNsName, newNsName := oldNsObj.ObjectMeta.Name, newNsObj.ObjectMeta.Name
	fmt.Printf("NAMESPACE UPDATED. %s/%s", oldNsName, newNsName)

	npMgr.Unlock()
	npMgr.DeleteNamespace(oldNsObj)
	npMgr.AddNamespace(newNsObj)

	return nil
}

// DeleteNamespace handles delete name space.
func (npMgr *NetworkPolicyManager) DeleteNamespace(nsObj *corev1.Namespace) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	nsName, nsNs := nsObj.ObjectMeta.Name, nsObj.ObjectMeta.Namespace
	fmt.Printf("NAMESPACE DELETED: %s/%s\n", nsName, nsNs)

	_, exists := npMgr.nsMap[nsName]
	if !exists {
		return nil
	}

	delete(npMgr.nsMap, nsName)

	return nil
}

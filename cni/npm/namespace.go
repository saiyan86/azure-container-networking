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

// AddNamespace handles add name space.
func (npMgr *NetworkPolicyManager) AddNamespace(nsObj *corev1.Namespace) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	nsName, nsNs := nsObj.ObjectMeta.Name, nsObj.ObjectMeta.Namespace
	fmt.Printf("NAMESPACE CREATED: %s/%s\n", nsName, nsNs)

	_, exists := npMgr.nsMap[nsName]
	if !exists {
		newns, err := newNs(nsName)
		if err != nil {
			return err
		}
		npMgr.nsMap[nsName] = newns
	}

	return nil
}

// UpdateNamespace handles update name space.
func (npMgr *NetworkPolicyManager) UpdateNamespace(oldNs *corev1.Namespace, newNs *corev1.Namespace) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	oldNsName, newNsName := oldNs.ObjectMeta.Name, newNs.ObjectMeta.Name
	fmt.Printf("NAMESPACE UPDATED. %s/%s", oldNsName, newNsName)

	return nil
}

// DeleteNamespace handles delete name space.
func (npMgr *NetworkPolicyManager) DeleteNamespace(ns *corev1.Namespace) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	nsName, nsNs := ns.ObjectMeta.Name, ns.ObjectMeta.Namespace
	fmt.Printf("NAMESPACE DELETED: %s/%s\n", nsName, nsNs)

	return nil
}

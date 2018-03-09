package npm

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type namespace struct {
	name   string
	podMap map[string]*corev1.Pod
}

// newNS constructs a new namespace object.
func newNs(name string) (*namespace, error) {
	ns := &namespace{
		name:   name,
		podMap: make(map[string]*corev1.Pod),
	}

	return ns, nil
}

// AddNamespace handles add name space.
func (npMgr *NetworkPolicyManager) AddNamespace(ns *corev1.Namespace) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	nsName, nsNs := ns.ObjectMeta.Name, ns.ObjectMeta.Namespace
	fmt.Printf("NAMESPACE CREATED: %s/%s\n", nsName, nsNs)

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

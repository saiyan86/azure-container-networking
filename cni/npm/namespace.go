package npm

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type namespace struct {
	name     string
	nodeName string
	nsObj    *corev1.Namespace
}

// AddNamespace handles add name space.
func (npMgr *NetworkPolicyManager) AddNamespace(ns *corev1.Namespace) {
	fmt.Printf("NAMESPACE CREATED: %s\n", ns.Name)
}

// UpdateNamespace handles update name space.
func (npMgr *NetworkPolicyManager) UpdateNamespace(oldNs *corev1.Namespace, newNs *corev1.Namespace) {
	fmt.Printf("NAMESPACE UPDATED. %s/%s", oldNs.Name, newNs.Name)
}

// DeleteNamespace handles delete name space.
func (npMgr *NetworkPolicyManager) DeleteNamespace(ns *corev1.Namespace) {
	fmt.Printf("NAMESPACE DELETED. %s/%s", ns.Name)
}

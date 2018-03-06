package npm

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
)

/*
type npMgr interface {
	AddNetworkPolicy(obj *networkingv1.NetworkPolicy) error
	UpdateNetworkPolicy(old *networkingv1.NetworkPolicy, new *networkingv1.NetworkPolicy) error
	DeleteNetworkPolicy(obj *networkingv1.NetworkPolicy) error
}
*/
// AddNetworkPolicy adds network policy.
func (npMgr *NetworkPolicyManager) AddNetworkPolicy(np *networkingv1.NetworkPolicy) {
	fmt.Printf("NETWORK POLICY CREATED: %s/%s/%s\n", np.Namespace, np.Name, np.Spec.PolicyTypes)
}

// UpdateNetworkPolicy updates network policy.
func (npMgr *NetworkPolicyManager) UpdateNetworkPolicy(oldNp *networkingv1.NetworkPolicy, newNp *networkingv1.NetworkPolicy) {
	fmt.Printf("NETWORK POLICY UPDATED: %s/%s/%s\n", oldNp.Namespace, oldNp.Name, oldNp.Spec.PolicyTypes)
}

// DeleteNetworkPolicy deletes network policy.
func (npMgr *NetworkPolicyManager) DeleteNetworkPolicy(np *networkingv1.NetworkPolicy) {
	fmt.Printf("NETWORK POLICY DELETED: %s/%s/%s\n", np.Namespace, np.Name, np.Spec.PolicyTypes)
}

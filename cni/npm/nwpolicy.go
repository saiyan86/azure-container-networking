package npm

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
)

// AddNetworkPolicy adds network policy.
func (npMgr *NetworkPolicyManager) AddNetworkPolicy(npObj *networkingv1.NetworkPolicy) error {

	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName := npObj.ObjectMeta.Namespace, npObj.ObjectMeta.Name
	fmt.Printf("NETWORK POLICY CREATED: %s/%s\n", npNs, npName)

	selector := npObj.Spec.PodSelector
	fmt.Printf("podSelector:%+v\n", selector)

	ns, exists := npMgr.nsMap[npNs]
	if !exists {
		newns, err := newNs(npNs)
		if err != nil {
			return err
		}
		npMgr.nsMap[npNs] = newns
		ns = newns
	}

	ns.npQueue = append(ns.npQueue, npObj) //Didn't check for duplicate yet. Assuming duplicate is handled by k8s.

	// Creates ipset for specified labels.
	ipsMgr := npMgr.ipsMgr
	var labelKeys []string
	for podLabelKey, podLabelVal := range selector.MatchLabels {
		labelKey := podLabelKey + podLabelVal
		if !ipsMgr.Exists(labelKey, "") {
			labelKeys = append(labelKeys, labelKey)
			fmt.Printf("Creating ipset %s\n", labelKey)

			if err := ipsMgr.Create(labelKey); err != nil {
				fmt.Printf("Error creating ipset %s.\n", labelKey)
				return err
			}
		}
	}

	iptMgr := npMgr.iptMgr
	for _, labelKey := range labelKeys {
		fmt.Printf("!!!!!!!       %s        !!!!!!!\n", labelKey)
		// Create rule for all matching labels.
		if err := iptMgr.Add(labelKey, npObj); err != nil {
			fmt.Printf("Error creating iptables rule.\n")
			return err
		}
	}

	return nil
}

// UpdateNetworkPolicy updates network policy.
func (npMgr *NetworkPolicyManager) UpdateNetworkPolicy(oldNp *networkingv1.NetworkPolicy, newNp *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	oldNpNs, oldNpName := oldNp.ObjectMeta.Namespace, oldNp.ObjectMeta.Name
	fmt.Printf("NETWORK POLICY UPDATED: %s/%s\n", oldNpNs, oldNpName)

	return nil
}

// DeleteNetworkPolicy deletes network policy.
func (npMgr *NetworkPolicyManager) DeleteNetworkPolicy(npObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName := npObj.ObjectMeta.Namespace, npObj.ObjectMeta.Name
	fmt.Printf("NETWORK POLICY DELETED: %s/%s\n", npNs, npName)

	return nil
}

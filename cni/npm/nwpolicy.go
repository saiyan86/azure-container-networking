package npm

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
)

// AddNetworkPolicy adds network policy.
func (npMgr *NetworkPolicyManager) AddNetworkPolicy(npObj *networkingv1.NetworkPolicy) error {

	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName, selector := npObj.ObjectMeta.Namespace, npObj.ObjectMeta.Name, npObj.Spec.PodSelector
	fmt.Printf("NETWORK POLICY CREATED: %s/%s\n", npNs, npName)

	ns, exists := npMgr.nsMap[npNs]
	if !exists {
		newns, err := newNs(npNs)
		if err != nil {
			return err
		}
		npMgr.nsMap[npNs] = newns
		ns = newns
	}

	ns.npQueue = append(ns.npQueue, npObj) //No check for duplicate yet. Assuming duplicate is handled by k8s.

	// Creates ipset for specified labels.
	ipsMgr := ns.ipsMgr
	var labelKeys []string
	for podLabelKey, podLabelVal := range selector.MatchLabels {
		labelKey := npNs + "-" + podLabelKey + ":" + podLabelVal
		if err := ipsMgr.CreateList(npNs); err != nil {
			fmt.Printf("Error creating ipset list %s.\n", npNs)
			return err
		}

		if err := ipsMgr.Create(npNs, labelKey); err != nil {
			fmt.Printf("Error creating ipset %s.\n", labelKey)
			return err
		}

		labelKeys = append(labelKeys, labelKey)
	}

	iptMgr := ns.iptMgr
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
func (npMgr *NetworkPolicyManager) UpdateNetworkPolicy(oldNpObj *networkingv1.NetworkPolicy, newNpObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	oldNpNs, oldNpName := oldNpObj.ObjectMeta.Namespace, oldNpObj.ObjectMeta.Name
	fmt.Printf("NETWORK POLICY UPDATED: %s/%s\n", oldNpNs, oldNpName)

	return nil
}

// DeleteNetworkPolicy deletes network policy.
func (npMgr *NetworkPolicyManager) DeleteNetworkPolicy(npObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName, selector := npObj.ObjectMeta.Namespace, npObj.ObjectMeta.Name, npObj.Spec.PodSelector
	fmt.Printf("NETWORK POLICY DELETED: %s/%s\n", npNs, npName)

	//Gather labels associated with this network policy.
	var labelKeys []string
	for podLabelKey, podLabelVal := range selector.MatchLabels {
		labelKeys = append(labelKeys, npNs+"-"+podLabelKey+":"+podLabelVal)
	}

	ns, exists := npMgr.nsMap[npNs]
	if !exists {
		newns, err := newNs(npNs)
		if err != nil {
			return err
		}
		npMgr.nsMap[npNs] = newns
		ns = newns
	}

	//Remove iptables rules associated with those labels.
	iptMgr := ns.iptMgr
	for _, labelKey := range labelKeys {
		fmt.Printf("!!!!!!!       %s        !!!!!!!\n", labelKey)
		// Create rule for all matching labels.
		if err := iptMgr.Delete(labelKey, npObj); err != nil {
			fmt.Printf("Error deleting iptables rule.\n")
			return err
		}
	}

	return nil
}

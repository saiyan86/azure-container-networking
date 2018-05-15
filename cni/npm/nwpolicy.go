package npm

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
)

var isAzureNpmChainCreated = false

// AddNetworkPolicy adds network policy.
func (npMgr *NetworkPolicyManager) AddNetworkPolicy(npObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName := npObj.ObjectMeta.Namespace, npObj.ObjectMeta.Name
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

	if !isAzureNpmChainCreated {
		if err := ns.iptMgr.InitNpmChains(); err != nil {
			fmt.Printf("Error initialize azure-npm chains.\n")
			return err
		}
		isAzureNpmChainCreated = true
	}

	podSets, nsLists, iptEntries := parsePolicy(npObj)

	ipsMgr := ns.ipsMgr
	for _, set := range podSets {
		if err := ipsMgr.CreateSet(set); err != nil {
			fmt.Printf("Error creating ipset %s-%s\n", npNs, set)
			return err
		}
	}

	for _, list := range nsLists {
		if err := ipsMgr.CreateList(list); err != nil {
			fmt.Printf("Error creating ipset list %s-%s\n", npNs, list)
			return err
		}
	}

	if err := npMgr.InitAllNsList(ns); err != nil {
		fmt.Printf("Error initializing all-namespace ipset list.\n")
		return err
	}

	iptMgr := ns.iptMgr
	for _, iptEntry := range iptEntries {
		if err := iptMgr.Add(iptEntry); err != nil {
			fmt.Printf("Error applying iptables rule\n")
			fmt.Printf("%+v\n", iptEntry)
			return err
		}
	}

	ns.npMap[npName] = npObj //No check for duplicate yet. Assuming duplicate is handled by k8s.

	return nil
}

// UpdateNetworkPolicy updates network policy.
func (npMgr *NetworkPolicyManager) UpdateNetworkPolicy(oldNpObj *networkingv1.NetworkPolicy, newNpObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()

	oldNpNs, oldNpName := oldNpObj.ObjectMeta.Namespace, oldNpObj.ObjectMeta.Name
	fmt.Printf("NETWORK POLICY UPDATED: %s/%s\n", oldNpNs, oldNpName)

	npMgr.Unlock()
	npMgr.DeleteNetworkPolicy(oldNpObj)

	npMgr.AddNetworkPolicy(newNpObj)

	return nil
}

// DeleteNetworkPolicy deletes network policy.
func (npMgr *NetworkPolicyManager) DeleteNetworkPolicy(npObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName := npObj.ObjectMeta.Namespace, npObj.ObjectMeta.Name
	fmt.Printf("NETWORK POLICY DELETED: %s/%s\n", npNs, npName)

	ns, exists := npMgr.nsMap[npNs]
	if !exists {
		newns, err := newNs(npNs)
		if err != nil {
			return err
		}
		npMgr.nsMap[npNs] = newns
		ns = newns
	}

	_, _, iptEntries := parsePolicy(npObj)

	iptMgr := ns.iptMgr
	for _, iptEntry := range iptEntries {
		if err := iptMgr.Delete(iptEntry); err != nil {
			fmt.Printf("Error applying iptables rule\n")
			fmt.Printf("%+v\n", iptEntry)
			return err
		}
	}

	ipsMgr := ns.ipsMgr
	delete(ns.npMap, npName)
	if len(ns.npMap) == 0 {
		if err := ipsMgr.Clean(); err != nil {
			fmt.Printf("Error cleaning ipset\n")
			return err
		}

		if err := ns.iptMgr.UninitNpmChains(); err != nil {
			fmt.Printf("Error uninitialize azure-npm chains.\n")
			return err
		}
		isAzureNpmChainCreated = false

		if err := npMgr.UninitAllNsList(ns); err != nil {
			fmt.Printf("Error uninitializing all-namespace ipset list.\n")
			return err
		}
	}

	return nil
}

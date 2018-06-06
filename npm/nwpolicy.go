package npm

import (
	"github.com/Azure/azure-container-networking/log"
	networkingv1 "k8s.io/api/networking/v1"
)

// AddNetworkPolicy adds network policy.
func (npMgr *NetworkPolicyManager) AddNetworkPolicy(npObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName := npObj.ObjectMeta.Namespace, npObj.ObjectMeta.Name
	log.Printf("NETWORK POLICY CREATED: %s/%s\n", npNs, npName)

	ns, exists := npMgr.nsMap[npNs]
	if !exists {
		newns, err := newNs(npNs)
		if err != nil {
			return err
		}
		npMgr.nsMap[npNs] = newns
		ns = newns
	}

	if !npMgr.isAzureNpmChainCreated {
		if err := ns.iptMgr.InitNpmChains(); err != nil {
			log.Printf("Error initialize azure-npm chains.\n")
			return err
		}
		npMgr.isAzureNpmChainCreated = true
	}

	podSets, nsLists, iptEntries := parsePolicy(npObj)

	ipsMgr := ns.ipsMgr
	for _, set := range podSets {
		if err := ipsMgr.CreateSet(set); err != nil {
			log.Printf("Error creating ipset %s-%s\n", npNs, set)
			return err
		}
	}

	for _, list := range nsLists {
		if err := ipsMgr.CreateList(list); err != nil {
			log.Printf("Error creating ipset list %s-%s\n", npNs, list)
			return err
		}
	}

	if err := npMgr.InitAllNsList(); err != nil {
		log.Printf("Error initializing all-namespace ipset list.\n")
		return err
	}

	iptMgr := ns.iptMgr
	for _, iptEntry := range iptEntries {
		if err := iptMgr.Add(iptEntry); err != nil {
			log.Printf("Error applying iptables rule\n")
			log.Printf("%+v\n", iptEntry)
			return err
		}
	}

	ns.npMap[npName] = npObj //No check for duplicate yet. Assuming duplicate is handled by k8s.

	npMgr.numPolicies++
	log.Printf("numPolicies: %d", npMgr.numPolicies)

	return nil
}

// UpdateNetworkPolicy updates network policy.
func (npMgr *NetworkPolicyManager) UpdateNetworkPolicy(oldNpObj *networkingv1.NetworkPolicy, newNpObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()

	oldNpNs, oldNpName := oldNpObj.ObjectMeta.Namespace, oldNpObj.ObjectMeta.Name
	log.Printf("NETWORK POLICY UPDATED: %s/%s\n", oldNpNs, oldNpName)

	npMgr.Unlock()
	npMgr.DeleteNetworkPolicy(oldNpObj)

	if newNpObj.ObjectMeta.DeletionTimestamp == nil && newNpObj.ObjectMeta.DeletionGracePeriodSeconds == nil {
		npMgr.AddNetworkPolicy(newNpObj)
	}

	return nil
}

// DeleteNetworkPolicy deletes network policy.
func (npMgr *NetworkPolicyManager) DeleteNetworkPolicy(npObj *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName := npObj.ObjectMeta.Namespace, npObj.ObjectMeta.Name
	log.Printf("NETWORK POLICY DELETED: %s/%s\n", npNs, npName)

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
			log.Printf("Error applying iptables rule\n")
			log.Printf("%+v\n", iptEntry)
			return err
		}
	}

	delete(ns.npMap, npName)

	npMgr.numPolicies--

	log.Printf("numPolicies: %d", npMgr.numPolicies)
	if npMgr.numPolicies == 0 {
		if err := iptMgr.UninitNpmChains(); err != nil {
			log.Printf("Error uninitialize azure-npm chains.\n")
			return err
		}
		npMgr.isAzureNpmChainCreated = false
	}

	return nil
}

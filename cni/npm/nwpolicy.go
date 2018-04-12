package npm

import (
	"fmt"

	"github.com/Azure/azure-container-networking/cni/npm/iptm"

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
	ns.npQueue = append(ns.npQueue, npObj) //No check for duplicate yet. Assuming duplicate is handled by k8s.

	if !isAzureNpmChainCreated {
		if err := ns.iptMgr.AddChain(iptm.AzureIptablesChain); err != nil {
			fmt.Printf("Error creating iptables chain %s\n.", iptm.AzureIptablesChain)
			return err
		}
		isAzureNpmChainCreated = true
	}

	sets, iptEntries := parsePolicy(npObj)

	ipsMgr := ns.ipsMgr
	for _, set := range sets {
		if err := ipsMgr.Create(npNs, set); err != nil {
			fmt.Printf("Error creating ipset %s-%s\n", npNs, set)
			return err
		}
	}

	iptMgr := ns.iptMgr
	for _, iptEntry := range iptEntries {
		if err := iptMgr.Add(iptEntry); err != nil {
			fmt.Printf("Error applying iptables rule\n")
			fmt.Printf("%+v\n", iptEntry)
			return err
		}
	}

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

	_, iptEntries := parsePolicy(npObj)

	iptMgr := ns.iptMgr
	for _, iptEntry := range iptEntries {
		if err := iptMgr.Delete(iptEntry); err != nil {
			fmt.Printf("Error applying iptables rule\n")
			fmt.Printf("%+v\n", iptEntry)
			return err
		}
	}

	return nil
}

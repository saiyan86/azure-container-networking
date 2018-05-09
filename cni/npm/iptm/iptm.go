package iptm

import (
	"fmt"
	"os/exec"

	"github.com/Azure/azure-container-networking/cni/npm/util"
)

// IptEntry represents an iptables rule.
type IptEntry struct {
	Name       string
	HashedName string
	Chain      string
	Flag       string
	Specs      []string
}

// IptablesManager stores iptables entries.
type IptablesManager struct {
	entryMap      map[string][]*IptEntry
	OperationFlag string
}

// NewIptablesManager creates a new instance for IptablesManager object.
func NewIptablesManager() *IptablesManager {
	iptMgr := &IptablesManager{
		entryMap: make(map[string][]*IptEntry),
	}

	return iptMgr
}

// InitNpmChains initializes Azure NPM chains in iptables.
func (iptMgr *IptablesManager) InitNpmChains() error {
	iptMgr.OperationFlag = util.IptablesChainCreationFlag
	entry := &IptEntry{
		Chain: util.IptablesAzureChain,
	}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error creating iptables chain %s\n", util.IptablesAzureChain)
		return err
	}

	// Add default rule to FORWARD chain.
	iptMgr.OperationFlag = util.IptablesInsertionFlag
	defaultBlock := &IptEntry{
		Chain: util.IptablesForwardChain,
		Specs: []string{
			util.IptablesJumpFlag,
			util.IptablesReject,
		},
	}
	if err := iptMgr.Run(defaultBlock); err != nil {
		fmt.Printf("Error adding default rule to FORWARD chain\n")
		return err
	}

	// Insert AZURE-NPM chain to FORWARD chain.
	entry.Chain = util.IptablesForwardChain
	entry.Specs = []string{util.IptablesJumpFlag, util.IptablesAzureChain}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error adding AZURE-NPM chain to FORWARD chain\n")
		return err
	}

	// Add default rule to AZURE-NPM chain.
	entry.Chain = util.IptablesAzureChain
	entry.Specs = []string{
		util.IptablesMatchFlag,
		util.IptablesStateFlag,
		util.IPtablesMatchStateFlag,
		util.IptablesRelatedState + "," + util.IptablesEstablishedState,
		util.IptablesJumpFlag,
		util.IptablesAccept,
	}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error adding default rule to AZURE-NPM chain\n")
		return err
	}

	// Create AZURE-NPM-INGRESS-PORT chain.
	iptMgr.OperationFlag = util.IptablesChainCreationFlag
	entry = &IptEntry{
		Chain: util.IptablesAzureIngressPortChain,
	}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error creating iptables chain %s\n", util.IptablesAzureIngressPortChain)
		return err
	}

	// Insert AZURE-NPM-INGRESS-PORT chain to AZURE-NPM chain.
	iptMgr.OperationFlag = util.IptablesAppendFlag
	entry.Chain = util.IptablesAzureChain
	entry.Specs = []string{util.IptablesJumpFlag, util.IptablesAzureIngressPortChain}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error adding AZURE-NPM-INGRESS-PORT chain to AZURE-NPM chain\n")
		return err
	}

	// Create AZURE-NPM-INGRESS-FROM chain.
	iptMgr.OperationFlag = util.IptablesChainCreationFlag
	entry = &IptEntry{
		Chain: util.IptablesAzureIngressFromChain,
	}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error creating iptables chain %s\n", util.IptablesAzureIngressFromChain)
		return err
	}

	// Create AZURE-NPM-EGRESS-PORT chain.
	iptMgr.OperationFlag = util.IptablesChainCreationFlag
	entry = &IptEntry{
		Chain: util.IptablesAzureEgressPortChain,
	}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error creating iptables chain %s\n", util.IptablesAzureEgressPortChain)
		return err
	}

	// Insert AZURE-NPM-EGRESS-PORT chain to AZURE-NPM chain.
	iptMgr.OperationFlag = util.IptablesAppendFlag
	entry.Chain = util.IptablesAzureChain
	entry.Specs = []string{util.IptablesJumpFlag, util.IptablesAzureEgressPortChain}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error adding AZURE-NPM-EGRESS-PORT chain to AZURE-NPM chain\n")
		return err
	}

	// Create AZURE-NPM-EGRESS-FROM chain.
	iptMgr.OperationFlag = util.IptablesChainCreationFlag
	entry = &IptEntry{
		Chain: util.IptablesAzureEgressToChain,
	}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error creating iptables chain %s\n", util.IptablesAzureEgressToChain)
		return err
	}

	return nil
}

// UninitNpmChains uninitializes Azure NPM chains in iptables.
func (iptMgr *IptablesManager) UninitNpmChains() error {
	IptablesAzureChainList := []string{
		util.IptablesAzureChain,
		util.IptablesAzureIngressPortChain,
		util.IptablesAzureIngressFromChain,
		util.IptablesAzureEgressPortChain,
		util.IptablesAzureEgressToChain,
	}

	iptMgr.OperationFlag = util.IptablesFlushFlag
	for _, chain := range IptablesAzureChainList {
		entry := &IptEntry{
			Chain: chain,
		}
		if err := iptMgr.Run(entry); err != nil {
			fmt.Printf("Error flushing iptables chain %s\n", chain)
		}
	}

	iptMgr.OperationFlag = util.IptablesDeletionFlag
	for _, chain := range IptablesAzureChainList {
		entry := &IptEntry{
			Chain: chain,
		}
		if err := iptMgr.Run(entry); err != nil {
			fmt.Printf("Error deleting iptables chain %s\n", chain)
		}
	}

	return nil
}

// Add creates an entry in entryMap, and add corresponding rule in iptables.
func (iptMgr *IptablesManager) Add(entry *IptEntry) error {
	// Create iptables rules for every entry in the entryMap.
	iptMgr.OperationFlag = util.IptablesAppendFlag
	fmt.Printf("%+v\n", entry)
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error creating iptables rules.\n")
		return err
	}

	return nil
}

// Delete removes an entry from entryMap, and deletes the corresponding iptables rule.
func (iptMgr *IptablesManager) Delete(entry *IptEntry) error {
	// Create iptables rules for every entry in the entryMap.
	iptMgr.OperationFlag = util.IptablesDeletionFlag
	fmt.Printf("%+v\n", entry)
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error creating iptables rules.\n")
		return err
	}

	return nil
}

// Run execute an iptables command to update iptables.
func (iptMgr *IptablesManager) Run(entry *IptEntry) error {
	cmdName := util.Iptables
	cmdArgs := append([]string{iptMgr.OperationFlag, entry.Chain}, entry.Specs...)
	var (
		cmdOut []byte
		err    error
	)
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Printf("There was an error running command: %s\nArguments:%+v", err, cmdArgs)
		fmt.Printf("%s", string(cmdOut))
		return err
	}

	fmt.Printf("%s", string(cmdOut))
	return nil
}

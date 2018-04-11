package iptm

import (
	"fmt"
	"os/exec"
)

const IptablesChainCreationFlag string = "-N"
const IptablesInsertionFlag string = "-I"
const IptablesAppendFlag string = "-A"
const IptablesDeletionFlag string = "-D"

const IptablesJumpFlag string = "-j"

const IptablesAccept string = "ACCEPT"
const IptablesReject string = "REJECT"
const IptablesDrop string = "DROP"

const IptablesRelatedState string = "RELATED"
const IptablesEstablishedState string = "ESTABLISHED"

// AzureIptablesChain specifies the name of azure-npm created chain in iptables.
const AzureIptablesChain string = "AZURE-NPM"
const ForwardChain string = "FORWARD"

// IptEntry represents an iptables rule.
type IptEntry struct {
	Name          string
	HashedName    string
	OperationFlag string
	Chain         string
	Flag          string
	Specs         []string
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

// AddChain adds a chain to iptables
func (iptMgr *IptablesManager) AddChain(chainName string) error {
	iptMgr.OperationFlag = IptablesChainCreationFlag
	entry := &IptEntry{
		Chain: AzureIptablesChain,
	}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error creating iptables chain %s\n", chainName)
		return err
	}

	if chainName != AzureIptablesChain {
		return nil
	}

	// Add default rule to FORWARD chain.
	iptMgr.OperationFlag = IptablesInsertionFlag
	defaultBlock := &IptEntry{
		Chain: ForwardChain,
		Specs: []string{
			IptablesJumpFlag,
			IptablesReject,
		},
	}
	if err := iptMgr.Run(defaultBlock); err != nil {
		fmt.Printf("Error adding default rule to FORWARD chain\n")
		return err
	}

	// Insert AZURE-NPM chain to FORWARD chain.
	entry.Chain = ForwardChain
	entry.Specs = []string{IptablesJumpFlag, AzureIptablesChain}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error adding AZURE-NPM chain to FORWARD\n")
		return err
	}

	// Add default rule to AZURE-NPM chain.
	entry.Chain = AzureIptablesChain
	entry.Specs = []string{
		"-m",
		"state",
		"--state",
		IptablesRelatedState + "," + IptablesEstablishedState,
		IptablesJumpFlag,
		IptablesAccept,
	}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error adding default rule to AZURE-NPM chain\n")
		return err
	}

	return nil
}

// Add creates an entry in entryMap, and add corresponding rule in iptables.
func (iptMgr *IptablesManager) Add(entry *IptEntry) error {
	// Create iptables rules for every entry in the entryMap.
	iptMgr.OperationFlag = IptablesAppendFlag
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
	iptMgr.OperationFlag = IptablesDeletionFlag
	fmt.Printf("%+v\n", entry)
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error creating iptables rules.\n")
		return err
	}

	return nil
}

// Run execute an iptables command to update iptables.
func (iptMgr *IptablesManager) Run(entry *IptEntry) error {
	cmdName := "iptables"
	cmdArgs := append([]string{iptMgr.OperationFlag, entry.Chain}, entry.Specs...)
	var (
		cmdOut []byte
		err    error
	)
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Printf("There was an error running command: %s\nArguments:%+v", err, cmdArgs)
		return err
	}
	fmt.Printf("%s", string(cmdOut))

	return nil
}

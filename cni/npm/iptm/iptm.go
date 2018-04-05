package iptm

import (
	"fmt"
	"os/exec"

	networkingv1 "k8s.io/api/networking/v1"
)

const iptablesChainCreationFlag string = "-N"
const iptablesInsertionFlag string = "-I"
const iptablesAppendFlag string = "-A"
const iptablesDeletionFlag string = "-D"

const iptablesJumpFlag string = "-j"

const iptablesAccept string = "ACCEPT"
const iptablesReject string = "REJECT"
const iptablesDrop string = "DROP"

const iptablesRelatedState string = "RELATED"
const iptablesEstablishedState string = "ESTABLISHED"

// AzureIptablesChain specifies the name of azure-npm created chain in iptables.
const AzureIptablesChain string = "AZURE-NPM"
const forwardChain string = "FORWARD"

type iptEntry struct {
	name          string
	hashedName    string
	operationFlag string
	chain         string
	flag          string
	specs         []string
}

// IptablesManager stores iptables entries.
type IptablesManager struct {
	entryMap      map[string][]*iptEntry
	operationFlag string
}

// NewIptablesManager creates a new instance for IptablesManager object.
func NewIptablesManager() *IptablesManager {
	iptMgr := &IptablesManager{
		entryMap: make(map[string][]*iptEntry),
	}

	return iptMgr
}

// AddChain adds a chain to iptables
func (iptMgr *IptablesManager) AddChain(chainName string) error {
	iptMgr.operationFlag = iptablesChainCreationFlag
	entry := &iptEntry{
		chain: AzureIptablesChain,
	}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error creating iptables chain %s\n", chainName)
		return err
	}

	if chainName != AzureIptablesChain {
		return nil
	}

	// Insert AZURE-NPM chain to FORWARD chain.
	iptMgr.operationFlag = iptablesInsertionFlag
	entry.chain = forwardChain
	entry.specs = []string{iptablesJumpFlag, AzureIptablesChain}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error adding AZURE-NPM chain to FORWARD\n")
		return err
	}

	// Add default rule to FORWARD chain.
	defaultBlock := &iptEntry{
		operationFlag: iptMgr.operationFlag,
		chain:         forwardChain,
		specs: []string{
			iptablesJumpFlag,
			iptablesReject,
		},
	}
	if err := iptMgr.Run(defaultBlock); err != nil {
		fmt.Printf("Error adding default rule to FORWARD chain\n")
		return err
	}

	// Add default rule to AZURE-NPM chain.
	entry.chain = AzureIptablesChain
	entry.specs = []string{
		"-m",
		"state",
		"--state",
		iptablesRelatedState + "," + iptablesEstablishedState,
		iptablesJumpFlag,
		iptablesAccept,
	}
	if err := iptMgr.Run(entry); err != nil {
		fmt.Printf("Error adding default rule to AZURE-NPM chain\n")
		return err
	}

	return nil
}

// Add creates an entry in entryMap, and add corresponding rule in iptables.
func (iptMgr *IptablesManager) Add(entryName string, np *networkingv1.NetworkPolicy) error {
	key := np.ObjectMeta.Namespace + "-" + np.ObjectMeta.Name
	_, exists := iptMgr.entryMap[key]
	if !exists {
		if err := iptMgr.parsePolicy(entryName, np); err != nil {
			fmt.Printf("Error parsing network policy for iptables.\n")
		}
	}

	// Create iptables rules for every entry in the entryMap.
	iptMgr.operationFlag = iptablesAppendFlag
	for _, entry := range iptMgr.entryMap[key] {
		fmt.Printf("%+v\n", entry)
		if err := iptMgr.Run(entry); err != nil {
			fmt.Printf("Error creating iptables rules.\n")
			return err
		}
	}

	return nil
}

// Delete removes an entry from entryMap, and deletes the corresponding iptables rule.
func (iptMgr *IptablesManager) Delete(entryName string, np *networkingv1.NetworkPolicy) error {
	key := np.ObjectMeta.Namespace + "-" + np.ObjectMeta.Name
	_, exists := iptMgr.entryMap[key]
	if !exists {
		if err := iptMgr.parsePolicy(entryName, np); err != nil {
			fmt.Printf("Error parsing network policy for iptables.\n")
		}
	}

	// Create iptables rules for every entry in the entryMap.
	iptMgr.operationFlag = iptablesDeletionFlag
	for _, entry := range iptMgr.entryMap[key] {
		fmt.Printf("%+v\n", entry)
		if err := iptMgr.Run(entry); err != nil {
			fmt.Printf("Error creating iptables rules.\n")
			return err
		}
	}
	delete(iptMgr.entryMap, key)

	return nil
}

// Run execute an iptables command to update iptables.
func (iptMgr *IptablesManager) Run(entry *iptEntry) error {
	cmdName := "iptables"
	cmdArgs := append([]string{iptMgr.operationFlag, entry.chain}, entry.specs...)
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

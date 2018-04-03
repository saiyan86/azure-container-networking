package iptm

import (
	"fmt"
	"os/exec"

	networkingv1 "k8s.io/api/networking/v1"
)

const iptablesInsertionFlag string = "-I"
const iptablesDeletionFlag string = "-D"

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
	iptMgr.operationFlag = iptablesInsertionFlag
	for _, entry := range iptMgr.entryMap[key] {
		fmt.Printf("%+v\n", entry)
		if err := iptMgr.Run(entry); err != nil {
			fmt.Printf("Error creating ipset rules.\n")
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
			fmt.Printf("Error creating ipset rules.\n")
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
		fmt.Printf("There was an error running command: %s\n", err)
		return err
	}
	fmt.Printf("%s", string(cmdOut))

	return nil
}

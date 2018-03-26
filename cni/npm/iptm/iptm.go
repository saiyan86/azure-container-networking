package iptm

import (
	"fmt"
	"os/exec"

	networkingv1 "k8s.io/api/networking/v1"
)

type iptEntry struct {
	name          string
	operationFlag string
	chain         string
	flag          string
	specs         []string
}

// IptablesManager stores iptables entries.
type IptablesManager struct {
	entryMap map[string][]*iptEntry
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
	var err error

	_, exists := iptMgr.entryMap[entryName]
	if !exists {
		err = iptMgr.parsePolicy(entryName, np)
		if err != nil {
			fmt.Printf("Error parsing network policy for iptables.\n")
		}
	}

	// Create iptables rules for every entry in the entryMap.
	for _, entry := range iptMgr.entryMap[entryName] {
		fmt.Printf("%+v\n", entry)
		if err := iptMgr.Run(entry); err != nil {
			fmt.Printf("Error creating ipset rules.\n")
			return err
		}
	}

	return nil
}

// Run execute an iptables command to update iptables.
func (iptMgr *IptablesManager) Run(entry *iptEntry) error {
	cmdName := "iptables"
	cmdArgs := append([]string{entry.operationFlag, entry.chain}, entry.specs...)
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

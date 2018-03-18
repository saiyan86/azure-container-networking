package iptm

import (
	"fmt"
	"os"
	"os/exec"

	networkingv1 "k8s.io/api/networking/v1"
)

// IPManager interface manages iptables and ipset.
type IPManager interface {
	Create() error
	//	Delete() error
	//	Apply() error
}

type iptEntry struct {
	name          string
	operationFlag string
	chain         string
	flag          string
	specs         []string
}

// IptablesManager stores iptables entries.
type IptablesManager struct {
	entryMap map[string]*iptEntry
}

// NewIptablesManager creates a new instance for IptablesManager object.
func NewIptablesManager() *IptablesManager {
	iptMgr := &IptablesManager{
		entryMap: make(map[string]*iptEntry),
	}

	return iptMgr
}

// Add creates an entry in entryMap, and add corresponding rule in iptables.
func (iptMgr *IptablesManager) Add(entryName string, np *networkingv1.NetworkPolicy) error {
	_, exists := iptMgr.entryMap[entryName]
	if !exists {
		iptMgr.entryMap[entryName] = &iptEntry{
			name:          entryName,
			operationFlag: "-I",
			chain:         "FORWARD",              //TODO: take dependency on ingress/egress. We also need create our own chain.
			specs:         []string{"-j", "DROP"}, //TODO: take dependency on network policy.
		}
	}

	if err := iptMgr.create(iptMgr.entryMap[entryName]); err != nil {
		fmt.Printf("Error creating ipset rules.\n")
		return err
	}

	return nil
}

// create execute an iptables command to update iptables.
func (iptMgr *IptablesManager) create(entry *iptEntry) error {
	cmdName := "iptables"
	cmdArgs := append([]string{entry.operationFlag, entry.chain}, entry.specs...)
	var (
		cmdOut []byte
		err    error
	)
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Println(os.Stderr, "There was an error running command: ", err)
		return err
	}
	fmt.Printf("%s", string(cmdOut))

	return nil
}

package iptm

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/npm/util"
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
	OperationFlag string
}

// NewIptablesManager creates a new instance for IptablesManager object.
func NewIptablesManager() *IptablesManager {
	iptMgr := &IptablesManager{
		OperationFlag: "",
	}

	return iptMgr
}

// InitNpmChains initializes Azure NPM chains in iptables.
func (iptMgr *IptablesManager) InitNpmChains() error {
	log.Printf("Initializing AZURE-NPM")

	if err := iptMgr.AddChain(util.IptablesAzureChain); err != nil {
		return err
	}

	// Insert AZURE-NPM chain to FORWARD chain.
	entry := &IptEntry{
		Chain: util.IptablesForwardChain,
		Specs: []string{
			util.IptablesJumpFlag,
			util.IptablesAzureChain,
		},
	}
	exists, err := iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if !exists {
		iptMgr.OperationFlag = util.IptablesInsertionFlag
		if _, err = iptMgr.Run(entry); err != nil {
			log.Printf("Error adding AZURE-NPM chain to FORWARD chain\n")
			return err
		}
	}

	// Add default allow CONNECTED/RELATED rule to AZURE-NPM chain.
	entry.Chain = util.IptablesAzureChain
	entry.Specs = []string{
		util.IptablesMatchFlag,
		util.IptablesStateFlag,
		util.IPtablesMatchStateFlag,
		util.IptablesRelatedState + "," + util.IptablesEstablishedState,
		util.IptablesJumpFlag,
		util.IptablesAccept,
	}
	exists, err = iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if !exists {
		iptMgr.OperationFlag = util.IptablesInsertionFlag
		if _, err = iptMgr.Run(entry); err != nil {
			log.Printf("Error adding default allow CONNECTED/RELATED rule to AZURE-NPM chain\n")
			return err
		}
	}

	// Add default allow kube-system rules to AZURE-NPM chain.
	entry.Specs = []string{
		util.IptablesMatchFlag,
		util.IptablesSetFlag,
		util.IptablesMatchSetFlag,
		util.GetHashedName(util.KubeSystemFlag),
		util.IptablesDstFlag,
		util.IptablesJumpFlag,
		util.IptablesAccept,
	}
	exists, err = iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if !exists {
		iptMgr.OperationFlag = util.IptablesAppendFlag
		if _, err := iptMgr.Run(entry); err != nil {
			log.Printf("Error adding default allow kube-system rule to AZURE-NPM chain\n")
			return err
		}
	}

	entry.Specs = []string{
		util.IptablesMatchFlag,
		util.IptablesSetFlag,
		util.IptablesMatchSetFlag,
		util.GetHashedName(util.KubeSystemFlag),
		util.IptablesSrcFlag,
		util.IptablesJumpFlag,
		util.IptablesAccept,
	}
	exists, err = iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if !exists {
		iptMgr.OperationFlag = util.IptablesAppendFlag
		if _, err := iptMgr.Run(entry); err != nil {
			log.Printf("Error adding default allow kube-system rule to AZURE-NPM chain\n")
			return err
		}
	}
	// Create AZURE-NPM-INGRESS-PORT chain.
	if err := iptMgr.AddChain(util.IptablesAzureIngressPortChain); err != nil {
		return err
	}

	// Insert AZURE-NPM-INGRESS-PORT chain to AZURE-NPM chain.
	entry.Chain = util.IptablesAzureChain
	entry.Specs = []string{util.IptablesJumpFlag, util.IptablesAzureIngressPortChain}
	exists, err = iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if !exists {
		iptMgr.OperationFlag = util.IptablesAppendFlag
		if _, err := iptMgr.Run(entry); err != nil {
			log.Printf("Error adding AZURE-NPM-INGRESS-PORT chain to AZURE-NPM chain\n")
			return err
		}
	}

	// Create AZURE-NPM-INGRESS-FROM chain.
	if err := iptMgr.AddChain(util.IptablesAzureIngressFromChain); err != nil {
		return err
	}

	// Create AZURE-NPM-EGRESS-PORT chain.
	if err := iptMgr.AddChain(util.IptablesAzureEgressPortChain); err != nil {
		return err
	}

	// Insert AZURE-NPM-EGRESS-PORT chain to AZURE-NPM chain.
	entry.Chain = util.IptablesAzureChain
	entry.Specs = []string{util.IptablesJumpFlag, util.IptablesAzureEgressPortChain}
	exists, err = iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if !exists {
		iptMgr.OperationFlag = util.IptablesAppendFlag
		if _, err := iptMgr.Run(entry); err != nil {
			log.Printf("Error adding AZURE-NPM-EGRESS-PORT chain to AZURE-NPM chain\n")
			return err
		}
	}

	// Create AZURE-NPM-EGRESS-FROM chain.
	if err := iptMgr.AddChain(util.IptablesAzureEgressToChain); err != nil {
		return err
	}

	// Create AZURE-NPM-TARGET-SETS chain.
	if err := iptMgr.AddChain(util.IptablesAzureTargetSetsChain); err != nil {
		return err
	}

	// Insert AZURE-NPM-TARGET-SETS chain to AZURE-NPM chain.
	entry.Chain = util.IptablesAzureChain
	entry.Specs = []string{util.IptablesJumpFlag, util.IptablesAzureTargetSetsChain}
	exists, err = iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if !exists {
		iptMgr.OperationFlag = util.IptablesAppendFlag
		if _, err := iptMgr.Run(entry); err != nil {
			log.Printf("Error adding AZURE-NPM-TARGET-SETS chain to AZURE-NPM chain\n")
			return err
		}
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
		util.IptablesAzureTargetSetsChain,
	}

	// Remove AZURE-NPM chain from FORWARD chain.
	entry := &IptEntry{
		Chain: util.IptablesForwardChain,
		Specs: []string{
			util.IptablesJumpFlag,
			util.IptablesAzureChain,
		},
	}
	iptMgr.OperationFlag = util.IptablesDeletionFlag
	errCode, err := iptMgr.Run(entry)
	if errCode != 1 && err != nil {
		log.Printf("Error removing default rule from FORWARD chain\n")
		return err
	}

	iptMgr.OperationFlag = util.IptablesFlushFlag
	for _, chain := range IptablesAzureChainList {
		entry := &IptEntry{
			Chain: chain,
		}
		if _, err := iptMgr.Run(entry); err != nil {
			log.Printf("Error flushing iptables chain %s\n", chain)
		}
	}

	for _, chain := range IptablesAzureChainList {
		if err := iptMgr.DeleteChain(chain); err != nil {
			return err
		}
	}

	return nil
}

// Exists checks if a rule exists in iptables.
func (iptMgr *IptablesManager) Exists(entry *IptEntry) (bool, error) {
	iptMgr.OperationFlag = util.IptablesCheckFlag
	returnCode, err := iptMgr.Run(entry)
	if err == nil {
		log.Printf("Rule exists. %+v\n", entry)
		return true, nil
	}

	if returnCode == 1 {
		log.Printf("Rule doesn't exist. %+v\n", entry)
		return false, nil
	}

	return false, err
}

// AddChain adds a chain to iptables.
func (iptMgr *IptablesManager) AddChain(chain string) error {
	entry := &IptEntry{
		Chain: chain,
	}
	iptMgr.OperationFlag = util.IptablesChainCreationFlag
	errCode, err := iptMgr.Run(entry)
	if errCode == 1 && err != nil {
		log.Printf("Chain already exists %s\n", entry.Chain)
		return nil
	}

	if err != nil {
		log.Printf("Error creating iptables chain %s\n", entry.Chain)
		return err
	}

	return nil
}

// DeleteChain deletes a chain from iptables.
func (iptMgr *IptablesManager) DeleteChain(chain string) error {
	entry := &IptEntry{
		Chain: chain,
	}
	iptMgr.OperationFlag = util.IptablesDestroyFlag
	errCode, err := iptMgr.Run(entry)
	if errCode == 1 && err != nil {
		log.Printf("Chain doesn't exist %s\n", entry.Chain)
		return nil
	}

	if err != nil {
		log.Printf("Error deleting iptables chain %s\n", entry.Chain)
		return err
	}

	return nil
}

// Add adds a rule in iptables.
func (iptMgr *IptablesManager) Add(entry *IptEntry) error {
	log.Printf("%+v\n", entry)

	exists, err := iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	iptMgr.OperationFlag = util.IptablesAppendFlag
	if _, err := iptMgr.Run(entry); err != nil {
		log.Printf("Error creating iptables rules.\n")
		return err
	}

	return nil
}

// Delete removes a rule in iptables.
func (iptMgr *IptablesManager) Delete(entry *IptEntry) error {
	log.Printf("%+v\n", entry)

	exists, err := iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	iptMgr.OperationFlag = util.IptablesDeletionFlag
	if _, err := iptMgr.Run(entry); err != nil {
		log.Printf("Error deleting iptables rules.\n")
		return err
	}

	return nil
}

// Run execute an iptables command to update iptables.
func (iptMgr *IptablesManager) Run(entry *IptEntry) (int, error) {
	cmdName := util.Iptables
	cmdArgs := append([]string{iptMgr.OperationFlag, entry.Chain}, entry.Specs...)
	var (
		errCode int
	)
	cmdOut, err := exec.Command(cmdName, cmdArgs...).Output()
	log.Printf("%s\n", string(cmdOut))

	if msg, failed := err.(*exec.ExitError); failed {
		errCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
		if errCode > 1 {
			log.Printf("There was an error running command: %s\nArguments:%+v", err, cmdArgs)
		}

		return errCode, err
	}

	return 0, nil
}

// Save saves current iptables configuration to /var/log/iptables.conf
func (iptMgr *IptablesManager) Save(configFile string) error {
	if len(configFile) == 0 {
		configFile = util.IptablesConfigFile
	}

	cmd := exec.Command(util.IptablesSave)

	// create the config file for writing
	f, err := os.Create(configFile)
	if err != nil {
		log.Printf("Error opening file: %s.", configFile)
		return err
	}
	defer f.Close()
	cmd.Stdout = f

	if err := cmd.Start(); err != nil {
		log.Printf("Error running iptables-save.\n")
		return err
	}
	cmd.Wait()

	return nil
}

// Restore restores iptables configuration from /var/log/iptables.conf
func (iptMgr *IptablesManager) Restore(configFile string) error {
	if len(configFile) == 0 {
		configFile = util.IptablesConfigFile
	}

	cmd := exec.Command(util.IptablesRestore)

	// open the config file for reading
	f, err := os.Open(configFile)
	if err != nil {
		log.Printf("Error opening file: %s.", configFile)
		return err
	}
	defer f.Close()
	cmd.Stdin = f

	if err := cmd.Start(); err != nil {
		log.Printf("Error running iptables-restore.\n")
		return err
	}
	cmd.Wait()

	return nil
}

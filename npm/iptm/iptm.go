/*

Part of this file is modified from iptables package from Kuberenetes.
https://github.com/kubernetes/kubernetes/blob/master/pkg/util/iptables

*/
package iptm

import (
	"os"
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/npm/util"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	defaultlockWaitTimeInSeconds = "60"
)

// IptEntry represents an iptables rule.
type IptEntry struct {
	Command               string
	Name                  string
	HashedName            string
	Chain                 string
	Flag                  string
	LockWaitTimeInSeconds string
	Specs                 []string
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
	log.Printf("Initializing AZURE-NPM chains.")

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
			log.Errorf("Error: failed to add AZURE-NPM chain to FORWARD chain.")
			return err
		}
	}

	// Add default allow CONNECTED/RELATED rule to AZURE-NPM chain.
	entry.Chain = util.IptablesAzureChain
	entry.Specs = []string{
		util.IptablesMatchFlag,
		util.IptablesStateFlag,
		util.IptablesMatchStateFlag,
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
			log.Printf("Error: failed to add default allow CONNECTED/RELATED rule to AZURE-NPM chain.")
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
			log.Errorf("Error: failed to add AZURE-NPM-INGRESS-PORT chain to AZURE-NPM chain.")
			return err
		}
	}

	// Create AZURE-NPM-INGRESS-FROM-NS chain.
	if err = iptMgr.AddChain(util.IptablesAzureIngressFromNsChain); err != nil {
		return err
	}

	// Create AZURE-NPM-INGRESS-FROM-POD chain.
	if err = iptMgr.AddChain(util.IptablesAzureIngressFromPodChain); err != nil {
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
			log.Errorf("Error: failed to add AZURE-NPM-EGRESS-PORT chain to AZURE-NPM chain.")
			return err
		}
	}

	// Create AZURE-NPM-EGRESS-TO-NS chain.
	if err = iptMgr.AddChain(util.IptablesAzureEgressToNsChain); err != nil {
		return err
	}

	// Create AZURE-NPM-EGRESS-TO-POD chain.
	if err = iptMgr.AddChain(util.IptablesAzureEgressToPodChain); err != nil {
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
			log.Errorf("Error: failed to add AZURE-NPM-TARGET-SETS chain to AZURE-NPM chain.")
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
		util.IptablesAzureIngressFromNsChain,
		util.IptablesAzureIngressFromPodChain,
		util.IptablesAzureEgressPortChain,
		util.IptablesAzureEgressToNsChain,
		util.IptablesAzureEgressToPodChain,
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
		log.Errorf("Error: failed to remove default rule from FORWARD chain.")
		return err
	}

	iptMgr.OperationFlag = util.IptablesFlushFlag
	for _, chain := range IptablesAzureChainList {
		entry := &IptEntry{
			Chain: chain,
		}
		if _, err := iptMgr.Run(entry); err != nil {
			log.Errorf("Error: failed to flush iptables chain %s.", chain)
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
		log.Printf("Rule exists. %+v.", entry)
		return true, nil
	}

	if returnCode == 1 {
		log.Printf("Rule doesn't exist. %+v.", entry)
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
	if err != nil {
		if errCode == 1 {
			log.Printf("Chain already exists %s.", entry.Chain)
			return nil
		}

		log.Errorf("Error: failed to create iptables chain %s.", entry.Chain)
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
	if err != nil {
		if errCode == 1 {
			log.Printf("Chain doesn't exist %s.", entry.Chain)
			return nil
		}
		log.Errorf("Error: failed to delete iptables chain %s.", entry.Chain)
		return err
	}

	return nil
}

// Add adds a rule in iptables.
func (iptMgr *IptablesManager) Add(entry *IptEntry) error {
	log.Printf("Add iptables entry: %+v.", entry)

	exists, err := iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	iptMgr.OperationFlag = util.IptablesInsertionFlag
	if _, err := iptMgr.Run(entry); err != nil {
		log.Errorf("Error: failed to create iptables rules.")
		return err
	}

	return nil
}

// Delete removes a rule in iptables.
func (iptMgr *IptablesManager) Delete(entry *IptEntry) error {
	log.Printf("Deleting iptables entry: %+v", entry)

	exists, err := iptMgr.Exists(entry)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	iptMgr.OperationFlag = util.IptablesDeletionFlag
	if _, err := iptMgr.Run(entry); err != nil {
		log.Errorf("Error: failed to delete iptables rules.")
		return err
	}

	return nil
}

// Run execute an iptables command to update iptables.
func (iptMgr *IptablesManager) Run(entry *IptEntry) (int, error) {
	cmdName := entry.Command
	if cmdName == "" {
		cmdName = util.Iptables
	}

	if entry.LockWaitTimeInSeconds == "" {
		entry.LockWaitTimeInSeconds = defaultlockWaitTimeInSeconds
	}

	cmdArgs := append([]string{util.IptablesWaitFlag, entry.LockWaitTimeInSeconds, iptMgr.OperationFlag, entry.Chain}, entry.Specs...)
	log.Printf("Executing iptables command %s %v", cmdName, cmdArgs)
	_, err := exec.Command(cmdName, cmdArgs...).Output()

	if msg, failed := err.(*exec.ExitError); failed {
		errCode := msg.Sys().(syscall.WaitStatus).ExitStatus()
		if errCode > 1 {
			log.Errorf("Error: There was an error running command: %s %s Arguments:%v", err, cmdName, cmdArgs)
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

	l, err := grabIptablesLocks()
	if err != nil {
		return err
	}

	defer func(l *os.File) {
		if err = l.Close(); err != nil {
			log.Printf("Failed to close iptables locks")
		}
	}(l)

	// create the config file for writing
	f, err := os.Create(configFile)
	if err != nil {
		log.Errorf("Error: failed to open file: %s.", configFile)
		return err
	}
	defer f.Close()

	cmd := exec.Command(util.IptablesSave)
	cmd.Stdout = f
	if err := cmd.Start(); err != nil {
		log.Errorf("Error: failed to run iptables-save.")
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

	l, err := grabIptablesLocks()
	if err != nil {
		return err
	}

	defer func(l *os.File) {
		if err = l.Close(); err != nil {
			log.Printf("Failed to close iptables locks")
		}
	}(l)

	// open the config file for reading
	f, err := os.Open(configFile)
	if err != nil {
		log.Errorf("Error: failed to open file: %s.", configFile)
		return err
	}
	defer f.Close()

	cmd := exec.Command(util.IptablesRestore)
	cmd.Stdin = f
	if err := cmd.Start(); err != nil {
		log.Errorf("Error: failed to run iptables-restore.")
		return err
	}
	cmd.Wait()

	return nil
}

// grabs iptables v1.6 xtable lock
func grabIptablesLocks() (*os.File, error) {
	var success bool

	l := &os.File{}
	defer func(l *os.File) {
		// Clean up immediately on failure
		if !success {
			l.Close()
		}
	}(l)

	// Grab 1.6.x style lock.
	l, err := os.OpenFile(util.IptablesLockFile, os.O_CREATE, 0600)
	if err != nil {
		log.Printf("Error: failed to open iptables lock file %s.", util.IptablesLockFile)
		return nil, err
	}

	if err := wait.PollImmediate(200*time.Millisecond, 2*time.Second, func() (bool, error) {
		if err := grabIptablesFileLock(l); err != nil {
			return false, nil
		}

		return true, nil
	}); err != nil {
		log.Printf("Error: failed to acquire new iptables lock: %v.", err)
		return nil, err
	}

	success = true
	return l, nil
}

func grabIptablesFileLock(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
}

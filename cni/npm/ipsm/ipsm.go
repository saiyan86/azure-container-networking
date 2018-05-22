package ipsm

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/Azure/azure-container-networking/cni/npm/util"
)

type ipsEntry struct {
	operationFlag string
	name          string
	set           string
	spec          string
}

// IpsetManager stores ipset states.
type IpsetManager struct {
	listMap map[string]*Ipset //tracks all set lists.
	setMap  map[string]*Ipset //label -> []ip
}

// Ipset represents one ipset entry.
type Ipset struct {
	name       string
	elements   []string
	referCount int
}

// NewIpset creates a new instance for Ipset object.
func NewIpset(setName string) *Ipset {
	return &Ipset{
		name: setName,
	}
}

// NewIpsetManager creates a new instance for IpsetManager object.
func NewIpsetManager() *IpsetManager {
	return &IpsetManager{
		listMap: make(map[string]*Ipset),
		setMap:  make(map[string]*Ipset),
	}
}

// Exists checks if an element exists in setMap/listMap.
func (ipsMgr *IpsetManager) Exists(key string, val string, kind string) bool {
	m := ipsMgr.setMap
	if kind == util.IpsetSetListFlag {
		m = ipsMgr.listMap
	}

	if _, exists := m[key]; !exists {
		return false
	}

	for _, elem := range m[key].elements {
		if elem == val {
			return true
		}
	}

	return false
}

func getHashedName(name string) string {
	return util.AzureNpmPrefix + util.Hash(name)
}

func isNsSet(setName string) bool {
	return !strings.Contains(setName, "-") && !strings.Contains(setName, ":")
}

// IncrementReferCount increases referCount of a specific ipset by one.
func (ipsMgr *IpsetManager) IncrementReferCount(setName string) error {
	if _, exists := ipsMgr.setMap[setName]; !exists {
		return fmt.Errorf("set %s doesn't exist, can't increment", setName)
	}

	ipsMgr.setMap[setName].referCount++

	return nil
}

// DecrementReferCount decreases referCount of a specific ipset by one.
func (ipsMgr *IpsetManager) DecrementReferCount(setName string) error {
	if _, exists := ipsMgr.setMap[setName]; !exists {
		return fmt.Errorf("set %s doesn't exist, can't increment", setName)
	}

	ipsMgr.setMap[setName].referCount--

	return nil
}

// NotReferredByNwPolicy checks if a specific ipset is referred by any network policy.
func (ipsMgr *IpsetManager) NotReferredByNwPolicy(setName string) bool {
	return ipsMgr.setMap[setName].referCount == 0
}

// CreateList creates an ipset list. npm maintains one setlist per namespace label.
func (ipsMgr *IpsetManager) CreateList(listName string) error {
	// Ignore system pods.
	if listName == util.KubeSystemFlag {
		return nil
	}

	if _, exists := ipsMgr.listMap[listName]; exists {
		return nil
	}

	entry := &ipsEntry{
		name:          listName,
		operationFlag: util.IpsetCreationFlag,
		set:           getHashedName(listName),
		spec:          util.IpsetSetListFlag,
	}
	fmt.Printf("Creating List: %+v\n", entry)
	if _, err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error creating ipset list %s.\n", listName)
		return err
	}

	ipsMgr.listMap[listName] = NewIpset(listName)

	return nil
}

// DeleteList removes an ipset list.
func (ipsMgr *IpsetManager) DeleteList(listName string) error {
	entry := &ipsEntry{
		operationFlag: util.IpsetDestroyFlag,
		set:           getHashedName(listName),
	}

	errCode, err := ipsMgr.Run(entry)
	if errCode == 1 && err != nil {
		fmt.Printf("Cannot delete list %s as it's being referred or doesn't exist.\n", listName)
		return nil
	}

	if err != nil {
		fmt.Printf("Error deleting ipset %s", listName)
		fmt.Printf("%+v\n", entry)
		return err
	}

	delete(ipsMgr.listMap, listName)

	return nil
}

// AddToList inserts an ipset to an ipset list.
func (ipsMgr *IpsetManager) AddToList(listName string, setName string) error {
	if ipsMgr.Exists(listName, setName, util.IpsetSetListFlag) {
		return nil
	}

	if err := ipsMgr.CreateList(listName); err != nil {
		return err
	}

	entry := &ipsEntry{
		operationFlag: util.IpsetAppendFlag,
		set:           getHashedName(listName),
		spec:          getHashedName(setName),
	}

	if _, err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error creating ipset rules.\n")
		fmt.Printf("rule: %+v\n", entry)
		return err
	}

	ipsMgr.listMap[listName].elements = append(ipsMgr.listMap[listName].elements, setName)

	return nil
}

// DeleteFromList removes an ipset to an ipset list.
func (ipsMgr *IpsetManager) DeleteFromList(listName string, setName string) error {
	if _, exists := ipsMgr.listMap[listName]; !exists {
		return fmt.Errorf("ipset list with name %s not found", listName)
	}

	for i, val := range ipsMgr.listMap[listName].elements {
		if val == setName {
			ipsMgr.listMap[listName].elements = append(ipsMgr.listMap[listName].elements[:i], ipsMgr.listMap[listName].elements[i+1:]...)
		}
	}

	hashedListName, hashedSetName := getHashedName(listName), getHashedName(setName)
	entry := &ipsEntry{
		operationFlag: util.IpsetDeletionFlag,
		set:           hashedListName,
		spec:          hashedSetName,
	}
	errCode, err := ipsMgr.Run(entry)
	if errCode > 1 && err != nil {
		fmt.Printf("Error deleting ipset entry.\n")
		fmt.Printf("%+v\n", entry)
		return err
	}

	if len(ipsMgr.listMap[listName].elements) == 0 {
		if err := ipsMgr.DeleteList(listName); err != nil {
			fmt.Printf("Error deleting ipset list %s.\n", listName)
			return err
		}
	}

	return nil
}

// CreateSet creates an ipset.
func (ipsMgr *IpsetManager) CreateSet(setName string) error {
	if _, exists := ipsMgr.setMap[setName]; exists {
		return nil
	}

	entry := &ipsEntry{
		name:          setName,
		operationFlag: util.IpsetCreationFlag,
		// Use hashed string for set name to avoid string length limit of ipset.
		set:  getHashedName(setName),
		spec: util.IpsetNetHashFlag,
	}
	fmt.Printf("Creating Set: %+v\n", entry)
	if _, err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error creating ipset.\n")
		return err
	}

	ipsMgr.setMap[setName] = NewIpset(setName)
	return nil
}

// DeleteSet removes a set from ipset.
func (ipsMgr *IpsetManager) DeleteSet(setName string) error {
	if _, exists := ipsMgr.setMap[setName]; !exists {
		return fmt.Errorf("ipset with name %s not found", setName)
	}

	if len(ipsMgr.setMap[setName].elements) > 0 {
		return nil
	}

	entry := &ipsEntry{
		operationFlag: util.IpsetDestroyFlag,
		set:           getHashedName(setName),
	}
	errCode, err := ipsMgr.Run(entry)
	if errCode == 1 && err != nil {
		fmt.Printf("Cannot delete set %s as it's being referred.\n", setName)
		return nil
	}

	if err != nil {
		fmt.Printf("Error deleting ipset %s", setName)
		fmt.Printf("%+v\n", entry)
		return err
	}

	delete(ipsMgr.setMap, setName)

	return nil
}

// AddToSet inserts an ip to an entry in setMap, and creates/updates the corresponding ipset.
func (ipsMgr *IpsetManager) AddToSet(setName string, ip string) error {
	if ipsMgr.Exists(setName, ip, util.IpsetNetHashFlag) {
		return nil
	}

	if err := ipsMgr.CreateSet(setName); err != nil {
		return err
	}

	entry := &ipsEntry{
		operationFlag: util.IpsetAppendFlag,
		set:           getHashedName(setName),
		spec:          ip,
	}

	if _, err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error creating ipset rules.\n")
		fmt.Printf("rule: %+v\n", entry)
		return err
	}
	ipsMgr.setMap[setName].elements = append(ipsMgr.setMap[setName].elements, ip)

	return nil
}

// DeleteFromSet removes an ip from an entry in setMap, and delete/update the corresponding ipset.
func (ipsMgr *IpsetManager) DeleteFromSet(setName string, ip string) error {
	if _, exists := ipsMgr.setMap[setName]; !exists {
		return fmt.Errorf("ipset with name %s not found", setName)
	}

	for i, val := range ipsMgr.setMap[setName].elements {
		if val == ip {
			ipsMgr.setMap[setName].elements = append(ipsMgr.setMap[setName].elements[:i], ipsMgr.setMap[setName].elements[i+1:]...)
		}
	}

	entry := &ipsEntry{
		operationFlag: util.IpsetDeletionFlag,
		set:           getHashedName(setName),
		spec:          ip,
	}
	if _, err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error deleting ipset entry.\n")
		fmt.Printf("%+v\n", entry)
		return err
	}

	return nil
}

// Clean removes all the empty sets & lists under the namespace.
func (ipsMgr *IpsetManager) Clean() error {
	for setName, set := range ipsMgr.setMap {
		if len(set.elements) > 0 {
			continue
		}

		if err := ipsMgr.DeleteSet(setName); err != nil {
			fmt.Printf("Error cleaning ipset\n")
			return err
		}
	}

	for listName, list := range ipsMgr.listMap {
		if len(list.elements) > 0 {
			continue
		}

		if err := ipsMgr.DeleteList(listName); err != nil {
			fmt.Printf("Error cleaning ipset list\n")
			return err
		}
	}

	return nil
}

// Destroy completely cleans ipset.
func (ipsMgr *IpsetManager) Destroy() error {
	entry := &ipsEntry{
		operationFlag: util.IpsetFlushFlag,
	}
	if _, err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error flushing ipset\n")
		return err
	}

	entry.operationFlag = util.IpsetDestroyFlag
	if _, err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error destroying ipset\n")
		return err
	}

	return nil
}

// Run execute an ipset command to update ipset.
func (ipsMgr *IpsetManager) Run(entry *ipsEntry) (int, error) {
	cmdName := util.Ipset
	cmdArgs := []string{entry.operationFlag, util.IpsetExistFlag}
	if len(entry.set) > 0 {
		cmdArgs = append(cmdArgs, entry.set)
	}
	if len(entry.spec) > 0 {
		cmdArgs = append(cmdArgs, entry.spec)
	}

	var (
		errCode int
	)
	cmdOut, err := exec.Command(cmdName, cmdArgs...).Output()
	fmt.Printf("%s\n", string(cmdOut))

	if msg, failed := err.(*exec.ExitError); failed {
		errCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
		if errCode > 1 {
			fmt.Printf("There was an error running command: %s\nArguments:%+v", err, cmdArgs)
		}

		return errCode, err
	}

	return 0, nil
}

// Save saves ipset to file.
func (ipsMgr *IpsetManager) Save() error {
	cmd := exec.Command(util.Ipset, util.IpsetSaveFlag, util.IpsetFileFlag, util.IpsetConfigFile)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error saving ipset to file.\n")
		return err
	}
	cmd.Wait()

	return nil
}

// Restore restores ipset from file.
func (ipsMgr *IpsetManager) Restore() error {
	cmd := exec.Command(util.Ipset, util.IpsetRestoreFlag, util.IpsetFileFlag, util.IpsetConfigFile)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error restoring ipset from file.\n")
		return err
	}
	cmd.Wait()

	return nil
}

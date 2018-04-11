package ipsm

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Azure/azure-container-networking/cni/npm/util"
)

type ipsEntry struct {
	operationFlag string
	name          string
	set           string
	spec          string
}

// IpsetManager stores ipset entries.
type IpsetManager struct {
	listMap  map[string][]string //tracks all set lists.
	entryMap map[string]*ipsEntry
	labelMap map[string][]string //label -> []ip
}

// ExistsInSet checks if an element exists in labelMap/listMap.
func (ipsMgr *IpsetManager) Exists(key string, val string, flag string) bool {
	m := ipsMgr.labelMap
	if flag == "setlist" {
		m = ipsMgr.listMap
	}

	_, exists := m[key]
	if !exists {
		return false
	}

	for _, elem := range m[key] {
		if elem == val {
			return true
		}
	}

	return false
}

// CreateList creates an ipset list. npm maintains one setlist per namespace label.
func (ipsMgr *IpsetManager) CreateList(setListName string) error {
	// Ignore system pods.
	if setListName == "kube-system" {
		return nil
	}

	hashedName := "azure-npm-" + util.Hash(setListName)
	_, exists := ipsMgr.listMap[setListName]
	if exists {
		return nil
	}

	ipsMgr.entryMap[setListName] = &ipsEntry{
		operationFlag: "-N",
		set:           hashedName,
		spec:          "setlist",
	}
	if err := ipsMgr.Run(ipsMgr.entryMap[setListName]); err != nil {
		fmt.Printf("Error creating ipset list %s.\n", setListName)
		return err
	}

	return nil
}

// Create creates an ipset.
func (ipsMgr *IpsetManager) Create(namespace string, setName string) error {
	// Use hashed string for set name to avoid string length limit of ipset.
	hashedName := "azure-npm-" + util.Hash(setName)
	_, exists := ipsMgr.entryMap[setName]
	if exists {
		return nil
	}

	ipsMgr.entryMap[setName] = &ipsEntry{
		operationFlag: "-N",
		set:           hashedName,
		spec:          "nethash",
	}
	if err := ipsMgr.Run(ipsMgr.entryMap[setName]); err != nil {
		fmt.Printf("Error creating ipset.\n")
		return err
	}

	// Add this ipset to the namespace's ipset list.
	/*
		entry := &ipsEntry{
			operationFlag: "-A",
			set:           namespace,
			spec:          hashedName,
		}
		if err := ipsMgr.Run(entry); err != nil {
			fmt.Printf("Error creating ipset.\n")
			return err
		}
	*/
	return nil
}

// AddToList inserts an ipset to an ipset list.
func (ipsMgr *IpsetManager) AddToList(setName string, listName string) error {
	if ipsMgr.Exists(listName, setName, "setlist") {
		return nil
	}

	if err := ipsMgr.CreateList(listName); err != nil {
		return err
	}

	ipsMgr.entryMap[listName].operationFlag = "-A"
	ipsMgr.entryMap[listName].spec = "azure-npm-" + util.Hash(setName)

	if err := ipsMgr.Run(ipsMgr.entryMap[listName]); err != nil {
		fmt.Printf("Error creating ipset rules.\n")
		fmt.Printf("rule: %+v\n", ipsMgr.entryMap[setName])
		return err
	}
	ipsMgr.listMap[listName] = append(ipsMgr.listMap[listName], setName)

	return nil
}

// AddToSet inserts an ip to an entry in labelMap, and creates/updates the corresponding ipset.
func (ipsMgr *IpsetManager) AddToSet(namespace string, setName string, ip string) error {
	if ipsMgr.Exists(setName, ip, "nethash") {
		return nil
	}

	if err := ipsMgr.Create(namespace, setName); err != nil {
		return err
	}

	ipsMgr.entryMap[setName].operationFlag = "-A"
	ipsMgr.entryMap[setName].spec = ip

	if err := ipsMgr.Run(ipsMgr.entryMap[setName]); err != nil {
		fmt.Printf("Error creating ipset rules.\n")
		fmt.Printf("rule: %+v\n", ipsMgr.entryMap[setName])
		return err
	}
	ipsMgr.labelMap[setName] = append(ipsMgr.labelMap[setName], ip)

	return nil
}

// DeleteFromSet removes an ip from an entry in labelMap, and delete/update the corresponding ipset.
func (ipsMgr *IpsetManager) DeleteFromSet(setName string, ip string) error {
	isEmpty := false

	_, exists := ipsMgr.labelMap[setName]
	if !exists {
		return fmt.Errorf("ipset with name %s not found", setName)
	}

	for i, val := range ipsMgr.labelMap[setName] {
		if val == ip {
			ipsMgr.labelMap[setName] = append(ipsMgr.labelMap[setName][:i], ipsMgr.labelMap[setName][i+1:]...)
		}
	}

	isEmpty = len(ipsMgr.labelMap[setName]) == 0

	entry := &ipsEntry{
		operationFlag: "-D",
		set:           setName,
		spec:          ip,
	}
	if err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error deleting ipset entry.\n")
		return err
	}

	if isEmpty {
		if err := ipsMgr.DeleteSet(setName); err != nil {
			fmt.Printf("Error deleting ipset %s.\n", setName)
			return err
		}
	}

	return nil
}

// DeleteSet removes a set from ipset.
func (ipsMgr *IpsetManager) DeleteSet(setName string) error {
	entry := &ipsEntry{
		operationFlag: "-D",
		set:           setName,
	}

	if err := ipsMgr.Run(entry); err != nil {
		return fmt.Errorf("Error deleting ipset %s", setName)
	}

	delete(ipsMgr.listMap, setName)

	return nil
}

// Run execute an ipset command to update ipset.
func (ipsMgr *IpsetManager) Run(entry *ipsEntry) error {
	cmdName := "ipset"
	cmdArgs := []string{entry.operationFlag, entry.set, entry.spec}
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

// NewIpsetManager creates a new instance for IpsetManager object.
func NewIpsetManager() *IpsetManager {

	ipsMgr := &IpsetManager{
		listMap:  make(map[string][]string),
		entryMap: make(map[string]*ipsEntry),
		labelMap: make(map[string][]string),
	}

	return ipsMgr
}

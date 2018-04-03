package ipsm

import (
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
)

type ipsEntry struct {
	operationFlag string
	name          string
	set           string
	spec          string
}

// IpsetManager stores ipset entries.
type IpsetManager struct {
	entryMap map[string]*ipsEntry
	labelMap map[string][]string //label -> []ip
}

// hash hashes a string to another string with length <= 32.
func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprint(h.Sum32())
}

// Exists checks if the ip exists in LabelMap.
func (ipsMgr *IpsetManager) Exists(key string, val string) bool {
	_, exists := ipsMgr.labelMap[key]
	if !exists {
		return false
	}

	for _, elem := range ipsMgr.labelMap[key] {
		if elem == val {
			return true
		}
	}

	return false
}

// CreateList creates an ipset list. npm maintains one setlist per namespace.
func (ipsMgr *IpsetManager) CreateList(setListName string) error {
	entry := &ipsEntry{
		operationFlag: "-N",
		set:           setListName,
		spec:          "setlist",
	}
	if err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error creating ipset.\n")
		return err
	}

	return nil
}

// Create creates an ipset.
func (ipsMgr *IpsetManager) Create(namespace string, setName string) error {
	// Use hashed string for set name, and annotate the real set name.
	hashedName := hash(setName)
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
	entry := &ipsEntry{
		operationFlag: "-A",
		set:           namespace,
		spec:          hashedName,
	}
	if err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error creating ipset.\n")
		return err
	}

	return nil
}

// Add inserts an ip to an entry in labelMap, and creates/updates the corresponding ipset.
func (ipsMgr *IpsetManager) Add(namespace string, setName string, ip string) error {
	ipsMgr.labelMap[setName] = append(ipsMgr.labelMap[setName], ip)

	if err := ipsMgr.Create(namespace, setName); err != nil {
		return err
	}

	ipsMgr.entryMap[setName].operationFlag = "-A"
	ipsMgr.entryMap[setName].spec = ip //This only holds one ip for now. Actually there will be multiple IPs under one setName.

	if err := ipsMgr.Run(ipsMgr.entryMap[setName]); err != nil {
		fmt.Printf("Error creating ipset rules.\n")
		fmt.Printf("rule: %+v\n", ipsMgr.entryMap[setName])
		return err
	}

	return nil
}

// DeleteFromSet removes an ip from an entry in labelMap, and delete/update the corresponding ipset.
func (ipsMgr *IpsetManager) DeleteFromSet(setName string, ip string) (bool, error) {
	isEmpty := false

	_, exists := ipsMgr.labelMap[setName]
	if !exists {
		return false, fmt.Errorf("ipset with name %s not found", setName)
	}

	for i, val := range ipsMgr.labelMap[setName] {
		if val == ip {
			ipsMgr.labelMap[setName] = append(ipsMgr.labelMap[setName][:i], ipsMgr.labelMap[setName][i+1:]...)
		}
	}

	if len(ipsMgr.labelMap[setName]) == 0 {
		isEmpty = true
	}

	entry := &ipsEntry{
		operationFlag: "-D",
		set:           setName,
		spec:          ip,
	}

	if err := ipsMgr.Run(entry); err != nil {
		fmt.Printf("Error deleing ipset entry.\n")
		return isEmpty, err
	}

	return isEmpty, nil
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
		entryMap: make(map[string]*ipsEntry),
		labelMap: make(map[string][]string),
	}

	return ipsMgr
}

package ipsm

import (
	"fmt"
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

// Add insert an ip to an entry in labelMap, and create/update the corresponding ipset.
func (ipsMgr *IpsetManager) Add(setName string, ip string) error {
	ipsMgr.labelMap[setName] = append(ipsMgr.labelMap[setName], ip)

	_, exists := ipsMgr.entryMap[setName]
	if !exists {
		ipsMgr.entryMap[setName] = &ipsEntry{
			operationFlag: "-N",
			set:           setName,
			spec:          "nethash",
		}
		if err := ipsMgr.create(ipsMgr.entryMap[setName]); err != nil {
			fmt.Printf("Error creating ipset.\n")
			return err
		}
	}

	ipsMgr.entryMap[setName].operationFlag = "-A"
	ipsMgr.entryMap[setName].spec = ip

	if err := ipsMgr.create(ipsMgr.entryMap[setName]); err != nil {
		fmt.Printf("Error creating ipset rules.\n")
		return err
	}

	return nil
}

// create execute an ipset command to update ipset.
func (ipsMgr *IpsetManager) create(entry *ipsEntry) error {
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

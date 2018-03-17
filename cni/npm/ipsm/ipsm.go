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
func (ipsMgr *IpsetManager) Add(key string, val string) error {
	ipsMgr.labelMap[key] = append(ipsMgr.labelMap[key], val)

	_, exists := ipsMgr.entryMap[key]
	if !exists {
		ipsMgr.entryMap[key] = &ipsEntry{
			operationFlag: "-N",
			set:           key,
			spec:          "nethash",
		}
		if err := ipsMgr.create(ipsMgr.entryMap[key]); err != nil {
			fmt.Printf("Error creating ipset.\n")
			return err
		}
	}

	ipsMgr.entryMap[key].operationFlag = "-A"
	ipsMgr.entryMap[key].spec = val

	if err := ipsMgr.create(ipsMgr.entryMap[key]); err != nil {
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

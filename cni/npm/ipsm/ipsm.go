package ipsm

import "fmt"

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

// ExistsInLabelMap checks if the ip exists in LabelMap.
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
func (ipsMgr *IpsetManager) Add(key string, val string) {
	ipsMgr.labelMap[key] = append(ipsMgr.labelMap[key], val)

	_, exists := ipsMgr.entryMap[key]
	if !exists {
		ipsMgr.entryMap[key] = &ipsEntry{
			operationFlag: "add",
			set:           key,
			spec:          val,
		}
	} else {
		ipsMgr.entryMap[key].spec += val
	}

	fmt.Printf("~~~~~~~~~~~~~~~~~~~~~~~~~~\n%+v\n", ipsMgr.entryMap[key])

}

// NewIpsetManager creates a new instance for IpsetManager object.
func NewIpsetManager() *IpsetManager {

	ipsMgr := &IpsetManager{
		entryMap: make(map[string]*ipsEntry),
		labelMap: make(map[string][]string),
	}

	return ipsMgr
}

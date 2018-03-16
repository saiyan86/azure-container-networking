package ipsm

type ipsEntry struct {
	name          string
	operationFlag string
	set           string
	spec          string
}

// IpsetManager stores ipset entries.
type IpsetManager struct {
	entryMap map[string]*ipsEntry
	labelMap map[string][]string //label -> []ip
}

// ExistsInLabelMap checks if the ip exists in LabelMap.
func (ipsMgr *IpsetManager) ExistsInLabelMap(key string, val string) bool {
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

// AddToLabelMap insert an entry to labelMap.
func (ipsMgr *IpsetManager) AddToLabelMap(key string, val string) {
	ipsMgr.labelMap[key] = append(ipsMgr.labelMap[key], val)
}

// NewIpsetManager creates a new instance for IpsetManager object.
func NewIpsetManager() *IpsetManager {

	ipsMgr := &IpsetManager{
		entryMap: make(map[string]*ipsEntry),
		labelMap: make(map[string][]string),
	}

	return ipsMgr
}

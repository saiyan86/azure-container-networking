package iptm

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
	spec          string
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

// Create creates an iptables rule from network policy and ipset.
/*
func (*IptablesManager) Create() {

}
*/

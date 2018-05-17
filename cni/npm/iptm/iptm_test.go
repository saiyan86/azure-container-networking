package iptm

import "testing"

func TestSave(t *testing.T) {
	iptMgr := &IptablesManager{}
	if err := iptMgr.Save(); err != nil {
		t.Errorf("TestSave failed @ iptMgr.Save()")
	}
}

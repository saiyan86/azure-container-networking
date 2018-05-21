package iptm

import "testing"

func TestSave(t *testing.T) {
	iptMgr := &IptablesManager{}
	if err := iptMgr.Save(); err != nil {
		t.Errorf("TestSave failed @ iptMgr.Save")
	}
}

func TestRestore(t *testing.T) {
	iptMgr := &IptablesManager{}
	if err := iptMgr.Restore(); err != nil {
		t.Errorf("TestRestore failed @ iptMgr.Restore")
	}
}

func TestInitNpmChains(t *testing.T) {
	iptMgr := &IptablesManager{}

	if err := iptMgr.Save(); err != nil {
		t.Errorf("TestInitNpmChains failed @ iptMgr.Save")
	}

	if err := iptMgr.InitNpmChains(); err != nil {
		t.Errorf("TestInitNpmChains @ iptMgr.InitNpmChains")
	}

	if err := iptMgr.Restore(); err != nil {
		t.Errorf("TestInitNpmChains failed @ iptMgr.Restore")
	}
}

func TestUninitNpmChains(t *testing.T) {
	iptMgr := &IptablesManager{}

	if err := iptMgr.Save(); err != nil {
		t.Errorf("TestUninitNpmChains failed @ iptMgr.Save")
	}

	if err := iptMgr.InitNpmChains(); err != nil {
		t.Errorf("TestUninitNpmChains @ iptMgr.InitNpmChains")
	}

	if err := iptMgr.UninitNpmChains(); err != nil {
		t.Errorf("TestUninitNpmChains @ iptMgr.UninitNpmChains")
	}

	if err := iptMgr.Restore(); err != nil {
		t.Errorf("TestUninitNpmChains failed @ iptMgr.Restore")
	}
}

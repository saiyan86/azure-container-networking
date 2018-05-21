package iptm

import (
	"testing"

	"github.com/Azure/azure-container-networking/cni/npm/util"
)

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

func TestExists(t *testing.T) {
	iptMgr := &IptablesManager{}
	if err := iptMgr.Save(); err != nil {
		t.Errorf("TestExists failed @ iptMgr.Save")
	}

	iptMgr.OperationFlag = util.IptablesCheckFlag
	entry := &IptEntry{
		Chain: util.IptablesForwardChain,
		Specs: []string{
			util.IptablesJumpFlag,
			util.IptablesAccept,
		},
	}
	if _, err := iptMgr.Exists(entry); err != nil {
		t.Errorf("TestExists failed @ iptMgr.Exists")
	}

	if err := iptMgr.Restore(); err != nil {
		t.Errorf("TestExists failed @ iptMgr.Restore")
	}
}

func TestAddChain(t *testing.T) {
	iptMgr := &IptablesManager{}
	if err := iptMgr.Save(); err != nil {
		t.Errorf("TestAddChain failed @ iptMgr.Save")
	}

	if err := iptMgr.AddChain("TEST-CHAIN"); err != nil {
		t.Errorf("TestAddChain failed @ iptMgr.AddChain")
	}

	if err := iptMgr.Restore(); err != nil {
		t.Errorf("TestAddChain failed @ iptMgr.Restore")
	}
}

func TestDeleteChain(t *testing.T) {
	iptMgr := &IptablesManager{}
	if err := iptMgr.Save(); err != nil {
		t.Errorf("TestDeleteChain failed @ iptMgr.Save")
	}

	if err := iptMgr.AddChain("TEST-CHAIN"); err != nil {
		t.Errorf("TestDeleteChain failed @ iptMgr.AddChain")
	}

	if err := iptMgr.DeleteChain("TEST-CHAIN"); err != nil {
		t.Errorf("TestDeleteChain failed @ iptMgr.DeleteChain")
	}

	if err := iptMgr.Restore(); err != nil {
		t.Errorf("TestDeleteChain failed @ iptMgr.Restore")
	}
}

func TestAdd(t *testing.T) {
	iptMgr := &IptablesManager{}
	if err := iptMgr.Save(); err != nil {
		t.Errorf("TestAdd failed @ iptMgr.Save")
	}

	entry := &IptEntry{
		Chain: util.IptablesForwardChain,
		Specs: []string{
			util.IptablesJumpFlag,
			util.IptablesReject,
		},
	}
	if err := iptMgr.Add(entry); err != nil {
		t.Errorf("TestAdd failed @ iptMgr.Add")
	}

	if err := iptMgr.Restore(); err != nil {
		t.Errorf("TestAdd failed @ iptMgr.Restore")
	}
}

func TestDelete(t *testing.T) {
	iptMgr := &IptablesManager{}
	if err := iptMgr.Save(); err != nil {
		t.Errorf("TestDelete failed @ iptMgr.Save")
	}

	entry := &IptEntry{
		Chain: util.IptablesForwardChain,
		Specs: []string{
			util.IptablesJumpFlag,
			util.IptablesReject,
		},
	}
	if err := iptMgr.Add(entry); err != nil {
		t.Errorf("TestDelete failed @ iptMgr.Add")
	}

	if err := iptMgr.Delete(entry); err != nil {
		t.Errorf("TestDelete failed @ iptMgr.Delete")
	}

	/*
		if err := iptMgr.Restore(); err != nil {
			t.Errorf("TestDelete failed @ iptMgr.Restore")
		}
	*/
}

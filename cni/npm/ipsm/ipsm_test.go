package ipsm

import (
	"testing"
)

func TestSave(t *testing.T) {
	ipsMgr := NewIpsetManager()
	if err := ipsMgr.Save(); err != nil {
		t.Errorf("TestSave failed @ ipsMgr.Save")
	}
}

func TestRestore(t *testing.T) {
	ipsMgr := NewIpsetManager()
	if err := ipsMgr.Save(); err != nil {
		t.Errorf("TestRestore failed @ ipsMgr.Save")
	}

	if err := ipsMgr.Restore(); err != nil {
		t.Errorf("TestRestore failed @ ipsMgr.Restore")
	}
}

func TestCreateList(t *testing.T) {
	ipsMgr := NewIpsetManager()
	if err := ipsMgr.Save(); err != nil {
		t.Errorf("TestCreateList failed @ ipsMgr.Save")
	}

	defer func() {
		if err := ipsMgr.Restore(); err != nil {
			t.Errorf("TestCreateList failed @ ipsMgr.Restore")
		}
	}()
}

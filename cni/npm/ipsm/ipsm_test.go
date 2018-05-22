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

	if err := ipsMgr.CreateList("test-list"); err != nil {
		t.Errorf("TestCreateList failed @ ipsMgr.CreateList")
	}
}

func TestDeleteList(t *testing.T) {
	ipsMgr := NewIpsetManager()
	if err := ipsMgr.Save(); err != nil {
		t.Errorf("TestDeleteList failed @ ipsMgr.Save")
	}

	defer func() {
		if err := ipsMgr.Restore(); err != nil {
			t.Errorf("TestDeleteList failed @ ipsMgr.Restore")
		}
	}()
}

func TestAddToList(t *testing.T) {
	ipsMgr := NewIpsetManager()
	if err := ipsMgr.Save(); err != nil {
		t.Errorf("TestAddToList failed @ ipsMgr.Save")
	}

	defer func() {
		if err := ipsMgr.Restore(); err != nil {
			t.Errorf("TestAddToList failed @ ipsMgr.Restore")
		}
	}()

	if err := ipsMgr.CreateSet("test-set"); err != nil {
		t.Errorf("TestAddToList failed @ ipsMgr.CreateSet")
	}

	if err := ipsMgr.AddToList("test-list", "test-set"); err != nil {
		t.Errorf("TestAddToList failed @ ipsMgr.AddToList")
	}
}

func TestMain(m *testing.M) {
	ipsMgr := NewIpsetManager()
	ipsMgr.Save()

	m.Run()

	ipsMgr.Restore()
}

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

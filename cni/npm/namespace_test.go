package npm

import (
	"fmt"
	"testing"
)

func TestnewNs(t *testing.T) {
	if _, err := newNs("test"); err != nil {
		t.Errorf("TestnewNs failed @ newNs")
	}
}

func TestAllNsList(t *testing.T) {

	fmt.Printf("hit")

	ns, err := newNs("test")
	if err != nil {
		t.Errorf("TestAllNsList failed @ newNs")
	}

	npMgr := &NetworkPolicyManager{}

	if err := npMgr.InitAllNsList(ns); err != nil {
		t.Errorf("TestAllNsList failed @ InitAllNsList")
	}

	if err := npMgr.UninitAllNsList(ns); err != nil {
		t.Errorf("TestAllNsList failed @ UninitAllNsList")
	}
}

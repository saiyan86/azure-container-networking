package npm

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestnewNs(t *testing.T) {
	if _, err := newNs("test"); err != nil {
		t.Errorf("TestnewNs failed @ newNs")
	}
}

func TestAllNsList(t *testing.T) {
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

func TestAddNamespace(t *testing.T) {
	npMgr := &NetworkPolicyManager{
		nsMap: make(map[string]*namespace),
	}

	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Labels: map[string]string{
				"app": "test",
			},
		},
	}

	if err := npMgr.AddNamespace(nsObj); err != nil {
		fmt.Errorf("TestAddNamespace @ AddNamespace")
	}
}

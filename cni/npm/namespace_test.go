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
			Name: "test-namespace",
			Labels: map[string]string{
				"app": "test-namespace",
			},
		},
	}

	if err := npMgr.AddNamespace(nsObj); err != nil {
		fmt.Errorf("TestAddNamespace @ npMgr.AddNamespace")
	}

	ns, err := newNs("test")
	if err != nil {
		t.Errorf("TestAddNamespace failed @ newNs")
	}

	if err := ns.ipsMgr.Destroy(); err != nil {
		t.Errorf("TestAddNamespace failed @ ns.ipsMgr.Destroy")
	}
}

func TestUpdateNamespace(t *testing.T) {
	npMgr := &NetworkPolicyManager{
		nsMap: make(map[string]*namespace),
	}

	oldNsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"app": "old-test-namespace",
			},
		},
	}

	now := metav1.Now()
	gracePeriod := int64(1)
	newNsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"app": "new-test-namespace",
			},
			DeletionTimestamp:          &now,
			DeletionGracePeriodSeconds: &gracePeriod,
		},
	}

	if err := npMgr.AddNamespace(oldNsObj); err != nil {
		t.Errorf("TestUpdateNamespace failed @ npMgr.AddNamespace")
	}

	if err := npMgr.UpdateNamespace(oldNsObj, newNsObj); err != nil {
		t.Errorf("TestUpdateNamespace failed @ npMgr.UpdateNamespace")
	}

	ns, err := newNs("test-namespace")
	if err != nil {
		t.Errorf("TestAddNamespace failed @ newNs")
	}

	if err := ns.ipsMgr.Destroy(); err != nil {
		t.Errorf("TestAddNamespace failed @ ns.ipsMgr.Destroy")
	}
}

func TestDeleteNamespace(t *testing.T) {
	npMgr := &NetworkPolicyManager{
		nsMap: make(map[string]*namespace),
	}

	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Labels: map[string]string{
				"app": "test-namespace",
			},
		},
	}

	if err := npMgr.AddNamespace(nsObj); err != nil {
		fmt.Errorf("TestDeleteNamespace @ npMgr.AddNamespace")
	}

	ns, err := newNs("test-namespace")
	if err != nil {
		t.Errorf("TestDeleteNamespace failed @ newNs")
	}

	if err := npMgr.DeleteNamespace(nsObj); err != nil {
		fmt.Errorf("TestDeleteNamespace @ npMgr.DeleteNamespace")
	}

	if err := ns.ipsMgr.Destroy(); err != nil {
		t.Errorf("TestDeleteNamespace failed @ ns.ipsMgr.Destroy")
	}
}

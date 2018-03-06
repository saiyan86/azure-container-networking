package npm

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type podMgr interface {
	AddPod(ns *corev1.Pod) error
	UpdatePod(ns *corev1.Pod) error
	DeletePod(ns *corev1.Pod) error
}

// AddPod handles add pod.
func (npMgr *NetworkPolicyManager) AddPod(obj interface{}) {
	pod := obj.(*corev1.Pod)
	fmt.Printf("POD CREATED: %s/%s/%s\n", pod.Namespace, pod.Name, pod.Spec.NodeName)
}

// UpdatePod handles update pod.
func (npMgr *NetworkPolicyManager) UpdatePod(old, new interface{}) {
	oldPod := old.(*corev1.Pod)
	newPod := new.(*corev1.Pod)
	fmt.Printf(
		"POD UPDATED. %s/%s %s",
		oldPod.Namespace, oldPod.Name, newPod.Status.Phase,
	)
}

// DeletePod handles delete pod.
func (npMgr *NetworkPolicyManager) DeletePod(obj interface{}) {
	pod := obj.(*corev1.Pod)
	fmt.Printf("POD DELETED: %s/%s/%s\n", pod.Namespace, pod.Name, pod.Spec.NodeName)
}

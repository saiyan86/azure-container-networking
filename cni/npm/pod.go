package npm

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

/*
type podMgr interface {
	AddPod(obj *corev1.Pod) error
	UpdatePod(old *corev1.Pod, new *corev1.Pod) error
	DeletePod(obj *corev1.Pod) error
}
*/
// func (npc *controller) AddPod(obj *coreapi.Pod) error {

// AddPod handles add pod.
func (npMgr *NetworkPolicyManager) AddPod(pod *corev1.Pod) {
	fmt.Printf("POD CREATED: %s/%s/%s\n", pod.Namespace, pod.Name, pod.Spec.NodeName)
}

// UpdatePod handles update pod.
func (npMgr *NetworkPolicyManager) UpdatePod(oldPod, newPod *corev1.Pod) {
	fmt.Printf(
		"POD UPDATED. %s/%s %s",
		oldPod.Namespace, oldPod.Name, newPod.Status.Phase,
	)
}

// DeletePod handles delete pod.
func (npMgr *NetworkPolicyManager) DeletePod(pod *corev1.Pod) {
	fmt.Printf("POD DELETED: %s/%s", pod.Namespace, pod.Name)
}

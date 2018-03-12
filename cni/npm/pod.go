package npm

import (
	"fmt"
	"time"

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

func isRunning(pod *corev1.Pod) bool {
	return pod.Status.Phase != "Failed" &&
		pod.Status.Phase != "Succeeded" &&
		pod.Status.Phase != "Unknown"
}

// AddPod handles add pod.
func (npMgr *NetworkPolicyManager) AddPod(pod *corev1.Pod) error {
	time.Sleep(10000 * time.Millisecond)

	npMgr.Lock()
	defer npMgr.Unlock()

	podNs, podName, podNodeName, podLabel := pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, pod.Spec.NodeName, pod.ObjectMeta.Labels
	fmt.Printf("POD CREATED: %s/%s/%s%+v\n", podNs, podName, podNodeName, podLabel)

	if !isRunning(pod) {
		return nil
	}

	ns, exists := npMgr.nsMap[podNs]
	if !exists {
		newns, err := newNs(podNs)
		if err != nil {
			return err
		}
		npMgr.nsMap[podNs] = newns
		ns = newns
	}

	ns.podMap[podName] = pod

	return nil
}

// UpdatePod handles update pod.
func (npMgr *NetworkPolicyManager) UpdatePod(oldPod, newPod *corev1.Pod) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	oldPodNs, oldPodName, newPodStatus := oldPod.ObjectMeta.Namespace, oldPod.ObjectMeta.Name, newPod.Status.Phase

	fmt.Printf(
		"POD UPDATED. %s/%s %s",
		oldPodNs, oldPodName, newPodStatus,
	)

	return nil
}

// DeletePod handles delete pod.
func (npMgr *NetworkPolicyManager) DeletePod(pod *corev1.Pod) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	podNs, podName, podNodeName := pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, pod.Spec.NodeName
	fmt.Printf("POD DELETED: %s/%s/%s\n", podNs, podName, podNodeName)

	return nil
}

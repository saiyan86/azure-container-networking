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

func isRunning(podObj *corev1.Pod) bool {
	return podObj.Status.Phase != "Failed" &&
		podObj.Status.Phase != "Succeeded" &&
		podObj.Status.Phase != "Unknown"
}

// AddPod handles add pod.
func (npMgr *NetworkPolicyManager) AddPod(podObj *corev1.Pod) error {

	npMgr.Lock()
	defer npMgr.Unlock()

	// Check if the pod is local
	if podObj.Spec.NodeName == npMgr.nodeName {
		return nil
	}

	podNs, podName, podNodeName, podLabel := podObj.ObjectMeta.Namespace, podObj.ObjectMeta.Name, podObj.Spec.NodeName, podObj.ObjectMeta.Labels
	fmt.Printf("POD CREATED: %s/%s/%s%+v\n", podNs, podName, podNodeName, podLabel)

	if !isRunning(podObj) {
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

	ns.podMap[podName] = podObj

	exists = false
	for podLabelType, podLabelValue := range podLabel {
		fmt.Printf("podLabelType: %s/ podLabelValue:%s\n", podLabelType, podLabelValue)
		for _, np := range ns.npMap {
			if np.Spec.PodSelector.MatchLabels[podLabelType] == podLabelValue {
				fmt.Printf("found matching policy\n")
			}
		}
	}

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
func (npMgr *NetworkPolicyManager) DeletePod(podObj *corev1.Pod) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	podNs, podName, podNodeName := podObj.ObjectMeta.Namespace, podObj.ObjectMeta.Name, podObj.Spec.NodeName
	fmt.Printf("POD DELETED: %s/%s/%s\n", podNs, podName, podNodeName)

	return nil
}

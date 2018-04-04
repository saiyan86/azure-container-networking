package npm

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func isValidPod(podObj *corev1.Pod) bool {
	return podObj.Status.Phase != "Failed" &&
		podObj.Status.Phase != "Succeeded" &&
		podObj.Status.Phase != "Unknown" &&
		len(podObj.Status.PodIP) > 0
}

func isSystemPod(podObj *corev1.Pod) bool {
	return podObj.ObjectMeta.Namespace == "kube-system"
}

// AddPod handles add pod.
func (npMgr *NetworkPolicyManager) AddPod(podObj *corev1.Pod) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	// Don't deal with system pods.
	if isSystemPod(podObj) {
		return nil
	}

	if !isValidPod(podObj) {
		return nil
	}

	podNs, podName, podNodeName, podLabels, podIP := podObj.ObjectMeta.Namespace, podObj.ObjectMeta.Name, podObj.Spec.NodeName, podObj.ObjectMeta.Labels, podObj.Status.PodIP
	fmt.Printf("POD CREATED: %s/%s/%s%+v%s\n", podNs, podName, podNodeName, podLabels, podIP)

	ns, exists := npMgr.nsMap[podNs]
	if !exists {
		newns, err := newNs(podNs)
		if err != nil {
			return err
		}
		npMgr.nsMap[podNs] = newns
		ns = newns
	}
	ns.podMap[podObj.ObjectMeta.UID] = podObj

	// Add pod to ipset
	ipsMgr := ns.ipsMgr
	var labelKeys []string
	for podLabelKey, podLabelVal := range podLabels {
		labelKey := podNs + "-" + podLabelKey + ":" + podLabelVal
		fmt.Printf("Adding pod %s to ipset %s\n", podIP, labelKey)
		if err := ipsMgr.Add(podNs, labelKey, podIP); err != nil {
			fmt.Printf("Error Adding pod to ipset.\n")
			return err
		}
		labelKeys = append(labelKeys, labelKey)
	}

	// Check if the pod is local
	/*
		if podObj.Spec.NodeName != npMgr.nodeName {
			return nil
		}

		/*
			iptMgr := ns.iptMgr
			exists = false


				for _, np := range ns.npQueue {
					selector, err := metav1.LabelSelectorAsSelector(&np.Spec.PodSelector)
					if err != nil {
						fmt.Printf("Error converting label selector\n")
						return err
					}
					if selector.Matches(labels.Set(podLabels)) {
						fmt.Printf("--------------found matching policy-----------------\n")

						for _, labelKey := range labelKeys {
							fmt.Printf("!!!!!!!       %s        !!!!!!!\n", labelKey)
							// Create rule for all matching labels.
							if err := iptMgr.Add(labelKey, np); err != nil {
								fmt.Printf("Error creating iptables rule.\n")
								return err
							}
						}
					}
				}
	*/
	return nil
}

// UpdatePod handles update pod.
func (npMgr *NetworkPolicyManager) UpdatePod(oldPod, newPod *corev1.Pod) error {
	npMgr.Lock()

	// Don't deal with system pods.
	if isSystemPod(newPod) {
		npMgr.Unlock()
		return nil
	}

	if !isValidPod(newPod) {
		npMgr.Unlock()
		return nil
	}

	oldPodNs, oldPodName, newPodStatus, newPodIP := oldPod.ObjectMeta.Namespace, oldPod.ObjectMeta.Name, newPod.Status.Phase, newPod.Status.PodIP

	fmt.Printf(
		"POD UPDATED. %s/%s %s %s\n",
		oldPodNs, oldPodName, newPodStatus, newPodIP,
	)

	npMgr.Unlock()
	npMgr.AddPod(newPod)

	return nil
}

// DeletePod handles delete pod.
func (npMgr *NetworkPolicyManager) DeletePod(podObj *corev1.Pod) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	// Don't deal with system pods.
	if isSystemPod(podObj) {
		return nil
	}

	podNs, podName, podNodeName, podLabels := podObj.ObjectMeta.Namespace, podObj.ObjectMeta.Name, podObj.Spec.NodeName, podObj.ObjectMeta.Labels
	fmt.Printf("POD DELETED: %s/%s/%s\n", podNs, podName, podNodeName)

	ns, exists := npMgr.nsMap[podNs]
	if !exists {
		newns, err := newNs(podNs)
		if err != nil {
			return err
		}
		npMgr.nsMap[podNs] = newns
		ns = newns
	}
	delete(ns.podMap, podObj.ObjectMeta.UID)

	// Delete pod from ipset
	podIP := podObj.Status.PodIP
	ipsMgr := ns.ipsMgr
	for podLabelKey, podLabelVal := range podLabels {
		labelKey := podNs + "-" + podLabelKey + ":" + podLabelVal
		if err := ipsMgr.DeleteFromSet(labelKey, podIP); err != nil {
			fmt.Printf("Error deleting pod from ipset.\n")
			return err
		}
	}

	return nil
}

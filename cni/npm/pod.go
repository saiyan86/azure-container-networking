package npm

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func isRunning(podObj *corev1.Pod) bool {
	return podObj.Status.Phase != "Failed" &&
		podObj.Status.Phase != "Succeeded" &&
		podObj.Status.Phase != "Unknown"
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

	podNs, podName, podNodeName, podLabels := podObj.ObjectMeta.Namespace, podObj.ObjectMeta.Name, podObj.Spec.NodeName, podObj.ObjectMeta.Labels
	fmt.Printf("POD CREATED: %s/%s/%s%+v\n", podNs, podName, podNodeName, podLabels)

	// Add pod to ipset
	podIP := podObj.Status.PodIP
	ipsMgr := npMgr.ipsMgr
	var labelKeys []string
	for podLabelKey, podLabelVal := range podLabels {
		labelKey := podLabelKey + podLabelVal
		if ipsMgr.Exists(labelKey, podIP) {
			return nil
		}
		labelKeys = append(labelKeys, labelKey)
		if err := ipsMgr.Add(labelKey, podIP); err != nil {
			fmt.Printf("Error Adding pod to ipset.\n")
			return err
		}
	}

	// Check if the pod is local
	if podObj.Spec.NodeName != npMgr.nodeName {
		return nil
	}

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

	ns.podMap[podObj.ObjectMeta.UID] = podObj

	iptMgr := npMgr.iptMgr
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

	// Don't deal with system pods.
	if isSystemPod(podObj) {
		return nil
	}

	podNs, podName, podNodeName, podLabels := podObj.ObjectMeta.Namespace, podObj.ObjectMeta.Name, podObj.Spec.NodeName, podObj.ObjectMeta.Labels
	fmt.Printf("POD DELETED: %s/%s/%s\n", podNs, podName, podNodeName)

	// Delete pod from ipset
	podIP := podObj.Status.PodIP
	ipsMgr := npMgr.ipsMgr
	for podLabelKey, podLabelVal := range podLabels {
		labelKey := podLabelKey + podLabelVal
		if ipsMgr.Exists(labelKey, podIP) {
			if err := ipsMgr.Delete(labelKey, podIP); err != nil {
				fmt.Printf("Error deleting pod from ipset.\n")
				return err
			}
		}
	}	

	return nil
}

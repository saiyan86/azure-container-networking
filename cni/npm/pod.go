package npm

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

	podNs, podName, podNodeName, podLabels := podObj.ObjectMeta.Namespace, podObj.ObjectMeta.Name, podObj.Spec.NodeName, podObj.ObjectMeta.Labels
	fmt.Printf("POD CREATED: %s/%s/%s%+v\n", podNs, podName, podNodeName, podLabels)

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

	ipsMgr := npMgr.ipsMgr
	//iptMgr := npMgr.iptMgr
	exists = false
	podIP := podObj.Status.PodIP

	var labelKeys []string
	for podLabelKey, podLabelVal := range podLabels {
		labelKey := podLabelKey + podLabelVal
		if ipsMgr.ExistsInLabelMap(labelKey, podIP) {
			return nil
		}
		labelKeys = append(labelKeys, labelKey)
		ipsMgr.AddToLabelMap(labelKey, podIP)
	}

	for _, np := range ns.npQueue {
		selector, err := metav1.LabelSelectorAsSelector(&np.Spec.PodSelector)
		if err != nil {
			fmt.Printf("Error converting label selector\n")
		}
		if selector.Matches(labels.Set(podLabels)) {
			fmt.Printf("--------------found matching policy-----------------\n")

			for _, labelKey := range labelKeys {
				fmt.Printf("!!!!!!!       %s        !!!!!!!\n", labelKey)
			}
			// Create rule for all matching labels.

			/*
				rule, err = iptMgr.Create(np, npMgr.ipSet)
				if err != nil {
					fmt.Printf("Error creating rule.\n")
				}

				if err = iptMgr.Apply(rule); err != nil {
					fmt.Printf("Error applying rule.\n")
				}
			*/
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

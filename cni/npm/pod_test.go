package npm

import (
	"testing"

	"github.com/Azure/azure-container-networking/cni/npm/ipsm"

	"github.com/Azure/azure-container-networking/cni/npm/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestisValidPod(t *testing.T) {
	podObj := &corev1.Pod{
		Status: corev1.PodStatus{
			Phase: "Running",
			PodIP: "1.2.3.4",
		},
	}
	if ok := isValidPod(podObj); !ok {
		t.Errorf("TestisValidPod failed @ isValidPod")
	}
}

func TestisSystemPod(t *testing.T) {
	podObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: util.KubeSystemFlag,
		},
	}
	if ok := isSystemPod(podObj); !ok {
		t.Errorf("TestisSystemPod failed @ isSystemPod")
	}
}

func TestAddPod(t *testing.T) {
	npMgr := &NetworkPolicyManager{
		nsMap: make(map[string]*namespace),
	}

	podObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-pod",
			Labels: map[string]string{
				"app": "test-pod",
			},
		},
		Status: corev1.PodStatus{
			Phase: "Running",
			PodIP: "1.2.3.4",
		},
	}
	if err := npMgr.AddPod(podObj); err != nil {
		t.Errorf("TestAddPod failed @ AddPod")
	}

	ipsMgr := &ipsm.IpsetManager{}
	if err := ipsMgr.Destroy(); err != nil {
		t.Errorf("TestAddPod failed @ ns.ipsMgr.Destroy")
	}
}

func TestUpdatePod(t *testing.T) {
	npMgr := &NetworkPolicyManager{
		nsMap: make(map[string]*namespace),
	}

	oldPodObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "old-test-pod",
			Labels: map[string]string{
				"app": "old-test-pod",
			},
		},
		Status: corev1.PodStatus{
			Phase: "Running",
			PodIP: "1.2.3.4",
		},
	}

	newPodObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "new-test-pod",
			Labels: map[string]string{
				"app": "new-test-pod",
			},
		},
		Status: corev1.PodStatus{
			Phase: "Running",
			PodIP: "4.3.2.1",
		},
	}

	if err := npMgr.AddPod(oldPodObj); err != nil {
		t.Errorf("TestUpdatePod failed @ AddPod")
	}

	if err := npMgr.UpdatePod(oldPodObj, newPodObj); err != nil {
		t.Errorf("TestUpdatePod failed @ UpdatePod")
	}

	ipsMgr := &ipsm.IpsetManager{}
	if err := ipsMgr.Destroy(); err != nil {
		t.Errorf("TestUpdatePod failed @ ns.ipsMgr.Destroy")
	}
}

func TestDeletePod(t *testing.T) {
	npMgr := &NetworkPolicyManager{
		nsMap: make(map[string]*namespace),
	}

	podObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-pod",
			Labels: map[string]string{
				"app": "test-pod",
			},
		},
		Status: corev1.PodStatus{
			Phase: "Running",
			PodIP: "1.2.3.4",
		},
	}
	if err := npMgr.AddPod(podObj); err != nil {
		t.Errorf("TestDeletePod failed @ AddPod")
	}

	if err := npMgr.DeletePod(podObj); err != nil {
		t.Errorf("TestDeletePod failed @ DeletePod")
	}

	ipsMgr := &ipsm.IpsetManager{}
	if err := ipsMgr.Destroy(); err != nil {
		t.Errorf("TestDeletePod failed @ ns.ipsMgr.Destroy")
	}
}

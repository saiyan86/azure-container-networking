package npm

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddPod(t *testing.T) {
	var npMgr NetworkPolicyManager

	fooPodIP := "1.2.3.4"
	podFoo := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "foo",
			Namespace: "default",
			Name:      "foo",
			Labels:    map[string]string{"run": "foo"}},
		Status: corev1.PodStatus{PodIP: fooPodIP},
	}

	if err := npMgr.AddPod(podFoo); err != nil {
		t.Errorf("Addpod failed")
	}

}

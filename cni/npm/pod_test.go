package npm

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestisValidPod(t *testing.T) {
	podObj := &corev1.Pod{
		Status: corev1.PodStatus{
			Phase: "Running",
			PodIP: "1.2.3.4",
		},
	}
	if valid := isValidPod(podObj); !valid {
		t.Errorf("TestisValidPod failed @ isValidPod")
	}
}

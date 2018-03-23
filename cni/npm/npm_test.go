package npm

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddPod(t *testing.T) {
	
	npMgr := &NetworkPolicyManager{
		clientset:       clientset,
		informerFactory: informerFactory,
		podInformer:     podInformer,
		nsInformer:      nsInformer,
		npInformer:      npInformer,
		nodeName:        os.Getenv("HOSTNAME"),
		nsMap:           make(map[string]*namespace),
		ipsMgr:          ipsm.NewIpsetManager(),
		iptMgr:          iptm.NewIptablesManager(),
	}

	testPodIP := "1.2.3.4"
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "test",
			Namespace: "default",
			Name:      "test",
			Labels:    map[string]string{"app": "test"}},
		Status: corev1.PodStatus{PodIP: testPodIP},
	}

	if err := npMgr.AddPod(testPod); err != nil {
		t.Errorf("TestAddPod failed")
	}
}
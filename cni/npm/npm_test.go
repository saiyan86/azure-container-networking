package npm

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	//"github.com/Azure/azure-container-networking/cni/npm/ipsm"
	"github.com/Azure/azure-container-networking/cni/npm/iptm"
)

func TestAddPod(t *testing.T) {
	var npMgr NetworkPolicyManager

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

func TestParseIngress(t *testing.T) {
	var iptMgr iptm.IptablesManager

	ruleOne := &networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{{
			PodSelector : &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		}},
		Ports: []networkingv1.NetworkPolicyPort{{
			Protocol: &networkingv1.Protocol{"tcp"},
			Port: &intstr.IntOrString{
				StrVal: "8000",
			},
		}},
	}

	ruleTwo := &networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{{
			PodSelector : &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		}},
		Ports: []networkingv1.NetworkPolicyPort{{
			Protocol: &networkingv1.Protocol{"udp"},
			Port: &intstr.IntOrString{
				StrVal: "8001",
			},
		}},
	}	

	var rules []networkingv1.NetworkPolicyIngressRule
	rules = append(rules, ruleOne)
	rules = append(rules, ruleTwo)

	ipsetName := "testIpsetName"

	if err := iptMgr.parseIngress(ipsetName, rules); err != nil {
		t.Errorf("TestParseIngress failed")
	}
}
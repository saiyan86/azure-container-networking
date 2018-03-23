package iptm

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

)

func TestParseIngress(t *testing.T) {
	var iptMgr IptablesManager

	ruleOne := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{{
			PodSelector : &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		}},
		Ports: []networkingv1.NetworkPolicyPort{{
			Port: &intstr.IntOrString{
				StrVal: "8000",
			},
		}},
	}

	ruleTwo := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{{
			PodSelector : &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		}},
		Ports: []networkingv1.NetworkPolicyPort{{
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
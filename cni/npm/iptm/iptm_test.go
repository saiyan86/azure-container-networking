func TestParseIngress(t *testing.T) {
	var iptMgr iptm.IptablesManager

	ruleOne := networkingv1.NetworkPolicyIngressRule{
		From: []networkingv1.NetworkPolicyPeer{{
			PodSelector : &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		}},
		Ports: []networkingv1.NetworkPolicyPort{{
			Protocol: &corev1.Protocol{"tcp"},
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
			Protocol: &corev1.Protocol{"udp"},
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
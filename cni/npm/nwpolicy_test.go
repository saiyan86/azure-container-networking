package npm

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-container-networking/cni/npm/iptm"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestAddNetworkPolicy(t *testing.T) {
	npMgr := &NetworkPolicyManager{
		nsMap: make(map[string]*namespace),
	}

	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-nwpolicy",
			Labels: map[string]string{
				"app": "test-namespace",
			},
		},
	}

	if err := npMgr.AddNamespace(nsObj); err != nil {
		fmt.Errorf("TestAddNetworkPolicy @ npMgr.AddNamespace")
	}

	tcp := corev1.ProtocolTCP
	allow := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-ingress",
			Namespace: "test-nwpolicy",
		},
		Spec: networkingv1.NetworkPolicySpec{
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				networkingv1.NetworkPolicyIngressRule{
					From: []networkingv1.NetworkPolicyPeer{{
						PodSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{"app": "test"},
						},
					}},
					Ports: []networkingv1.NetworkPolicyPort{{
						Protocol: &tcp,
						Port: &intstr.IntOrString{
							StrVal: "8000",
						},
					}},
				},
			},
		},
	}

	iptMgr := iptm.IptablesManager{}
	if err := iptMgr.Save(); err != nil {
		t.Errorf("TestAddNetworkPolicy failed @ iptMgr.Save")
	}

	if err := npMgr.AddNetworkPolicy(allow); err != nil {
		t.Errorf("TestAddNetworkPolicy failed @ AddNetworkPolicy")
	}

	/*
			if err := iptMgr.Restore(); err != nil {
				t.Errorf("TestAddNetworkPolicy failed @ iptMgr.Restore")
			}

				ipsMgr := &ipsm.IpsetManager{}
		if err := ipsMgr.Destroy(); err != nil {
			t.Errorf("TestAddNamespace failed @ ns.ipsMgr.Destroy")
		}
	*/
}

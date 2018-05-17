package npm

import (
	networkingv1 "k8s.io/api/networking/v1"
)

func TestAddNetworkPolicy(t *testing.T) {
	npMgr := &NetworkPolicyManager{
		nsMap: make(map[string]*namespace),
	}

	allow := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "allow-ingress",
			Namespace: ""
		}
	}
}
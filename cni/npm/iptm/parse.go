package iptm

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
)

type policyInfo struct {
	name  string
	ports []networkingv1.NetworkPolicyPort
}

type portsInfo struct {
	protocol string
	port     string
}

func (iptMgr *IptablesManager) parseIngress(ipsetName string, rules []networkingv1.NetworkPolicyIngressRule) error {
	var protAndPortsSlice []*portsInfo
	var podLabels []string
	//TODO: handle NamesapceSelector & IPBlock
	for _, rule := range rules {
		for _, portRule := range rule.Ports {
			protAndPortsSlice = append(protAndPortsSlice,
				&portsInfo{
					protocol: string(*portRule.Protocol),
					port:     fmt.Sprint(portRule.Port.IntVal),
				})
		}
		for _, fromRule := range rule.From {
			for podLabelKey, podLabelVal := range fromRule.PodSelector.MatchLabels {
				podLabels = append(podLabels, podLabelKey+podLabelVal)
			}
		}
	}

	for _, protAndPorts := range protAndPortsSlice {
		srcEntry := &iptEntry{
			name:          ipsetName,
			operationFlag: "-I",
			chain:         "FORWARD",
			specs:         []string{"-p", protAndPorts.protocol, "--sport", protAndPorts.port, "-m", "set", "--match-set", ipsetName, "src", "-j", "ACCEPT"},
		}
		iptMgr.entryMap[ipsetName] = append(iptMgr.entryMap[ipsetName], srcEntry)

		dstEntry := &iptEntry{
			name:          ipsetName,
			operationFlag: "-I",
			chain:         "FORWARD",
			specs:         []string{"-p", protAndPorts.protocol, "--dport", protAndPorts.port, "-m", "set", "--match-set", ipsetName, "dst", "-j", "ACCEPT"},
		}
		iptMgr.entryMap[ipsetName] = append(iptMgr.entryMap[ipsetName], dstEntry)
	}

	// Handle PodSelector field of NetworkPolicyPeer.
	for _, label := range podLabels {
		entry := &iptEntry{
			name:          label,
			operationFlag: "-I",
			chain:         "FORWARD",
			specs:         []string{"-m", "set", "--match-set", ipsetName, "src", "-j", "ACCEPT"},
		}
		iptMgr.entryMap[label] = append(iptMgr.entryMap[label], entry)
	}

	// TODO: Handle NamespaceSelector field of NetworkPolicyPeer.
	// TODO: Handle IPBlock field of NetworkPolicyPeer.

	return nil
}

func (iptMgr *IptablesManager) parseEgress(ipsetName string, rules []networkingv1.NetworkPolicyEgressRule) error {

	return nil
}

// ParsePolicy parses network policy.
func (iptMgr *IptablesManager) parsePolicy(ipsetName string, np *networkingv1.NetworkPolicy) error {
	if err := iptMgr.parseIngress(ipsetName, np.Spec.Ingress); err != nil {
		fmt.Printf("Error parsing ingress rule for iptables\n")
		return err
	}

	if err := iptMgr.parseEgress(ipsetName, np.Spec.Egress); err != nil {
		fmt.Printf("Error parsing egress rule for iptables\n")
		return err
	}

	return nil
}

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
	for _, rule := range rules {
		for _, portInfoFromRule := range rule.Ports {
			protAndPortsSlice = append(protAndPortsSlice,
				&portsInfo{
					protocol: string(*portInfoFromRule.Protocol),
					port:     fmt.Sprint(portInfoFromRule.Port.IntVal),
				})
		}
	}

	for _, protAndPorts := range protAndPortsSlice {
		entry := &iptEntry{
			name:          ipsetName,
			operationFlag: "-A",
			chain:         "FORWARD",
			specs:         []string{"-p", protAndPorts.protocol, protAndPorts.port, "-m", "set", "--match-set", ipsetName, "src", "-j", "REJECT"},
		}
		iptMgr.entryMap[ipsetName] = append(iptMgr.entryMap[ipsetName], entry)
	}

	for _, ent := range iptMgr.entryMap[ipsetName] {
		fmt.Printf("%+v\n", ent)
	}

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

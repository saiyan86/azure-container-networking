package iptm

import (
	"fmt"
	"hash/fnv"

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

func hash(s string) string {
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprint(h.Sum32())
}

func (iptMgr *IptablesManager) parseIngress(ipsetName string, rules []networkingv1.NetworkPolicyIngressRule) error {
	var protPortPairSlice []*portsInfo
	var podLabels []string
	//TODO: handle NamesapceSelector & IPBlock
	for _, rule := range rules {
		for _, portRule := range rule.Ports {
			protPortPairSlice = append(protPortPairSlice,
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

	// Use hashed string for set name, and annotate the real set name.
	hashedName := hash(ipsetName)
	for _, protPortPair := range protPortPairSlice {
		srcEntry := &iptEntry{
			name:          hashedName,
			operationFlag: iptMgr.operationFlag,
			chain:         "FORWARD",
			specs:         []string{"-p", protPortPair.protocol, "--sport", protPortPair.port, "-m", "set", "--match-set", hashedName, "src", "-j", "comment", "\"", hashedName, "\"", "REJECT"},
		}
		iptMgr.entryMap[hashedName] = append(iptMgr.entryMap[hashedName], srcEntry)

		dstEntry := &iptEntry{
			name:          hashedName,
			operationFlag: iptMgr.operationFlag,
			chain:         "FORWARD",
			specs:         []string{"-p", protPortPair.protocol, "--dport", protPortPair.port, "-m", "set", "--match-set", hashedName, "dst", "-j", "comment", "\"", hashedName, "\"", "REJECT"},
		}
		iptMgr.entryMap[hashedName] = append(iptMgr.entryMap[hashedName], dstEntry)
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

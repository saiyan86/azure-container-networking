package iptm

import (
	"fmt"

	"github.com/Azure/azure-container-networking/cni/npm/util"
	networkingv1 "k8s.io/api/networking/v1"
)

const defaultPolicyKey = "default-azure-policy"

type policyInfo struct {
	name  string
	ports []networkingv1.NetworkPolicyPort
}

type portsInfo struct {
	protocol string
	port     string
}

func (iptMgr *IptablesManager) parseIngress(ipsetName string, npName string, rules []networkingv1.NetworkPolicyIngressRule) error {
	// By default block all traffic.
	defaultBlock := &iptEntry{
		operationFlag: iptMgr.operationFlag,
		chain:         "FORWARD",
		specs:         []string{"-j", "REJECT"},
	}
	iptMgr.entryMap[defaultPolicyKey] = append(iptMgr.entryMap[defaultPolicyKey], defaultBlock)

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

	// Use hashed string for ipset name to avoid string length limit of ipset.
	hashedName := "azure-npm-" + util.Hash(ipsetName)
	for _, protPortPair := range protPortPairSlice {
		srcEntry := &iptEntry{
			name:          ipsetName,
			hashedName:    hashedName,
			operationFlag: iptMgr.operationFlag,
			chain:         "FORWARD",
			specs:         []string{"-p", protPortPair.protocol, "--sport", protPortPair.port, "-m", "set", "--match-set", hashedName, "src", "-j", "ACCEPT"},
		}
		iptMgr.entryMap[npName] = append(iptMgr.entryMap[npName], srcEntry)

		dstEntry := &iptEntry{
			name:          ipsetName,
			hashedName:    hashedName,
			operationFlag: iptMgr.operationFlag,
			chain:         "FORWARD",
			specs:         []string{"-p", protPortPair.protocol, "--dport", protPortPair.port, "-m", "set", "--match-set", hashedName, "dst", "-j", "ACCEPT"},
		}
		iptMgr.entryMap[npName] = append(iptMgr.entryMap[npName], dstEntry)
	}

	// Handle PodSelector field of NetworkPolicyPeer.
	for _, label := range podLabels {
		entry := &iptEntry{
			name:          label,
			hashedName:    hashedName,
			operationFlag: "-I",
			chain:         "FORWARD",
			specs:         []string{"-m", "set", "--match-set", hashedName, "src", "-j", "ACCEPT"},
		}
		iptMgr.entryMap[npName] = append(iptMgr.entryMap[npName], entry)
	}

	// TODO: Handle NamespaceSelector field of NetworkPolicyPeer.
	// TODO: Handle IPBlock field of NetworkPolicyPeer.

	return nil
}

func (iptMgr *IptablesManager) parseEgress(ipsetName string, npName string, rules []networkingv1.NetworkPolicyEgressRule) error {

	return nil
}

// ParsePolicy parses network policy.
func (iptMgr *IptablesManager) parsePolicy(ipsetName string, np *networkingv1.NetworkPolicy) error {
	if err := iptMgr.parseIngress(ipsetName, np.ObjectMeta.Namespace+"-"+np.ObjectMeta.Name, np.Spec.Ingress); err != nil {
		fmt.Printf("Error parsing ingress rule for iptables\n")
		return err
	}

	if err := iptMgr.parseEgress(ipsetName, np.ObjectMeta.Namespace+"-"+np.ObjectMeta.Name, np.Spec.Egress); err != nil {
		fmt.Printf("Error parsing egress rule for iptables\n")
		return err
	}

	return nil
}

package npm

import (
	"fmt"

	"github.com/Azure/azure-container-networking/cni/npm/iptm"
	"github.com/Azure/azure-container-networking/cni/npm/util"
	networkingv1 "k8s.io/api/networking/v1"
)

// azureNpmPrefix defines prefix for ipset.
const azureNpmPrefix string = "azure-npm-"

type policyInfo struct {
	name  string
	ports []networkingv1.NetworkPolicyPort
}

type portsInfo struct {
	protocol string
	port     string
}

func parseIngress(ns string, targetSets []string, rules []networkingv1.NetworkPolicyIngressRule) ([]string, []*iptm.IptEntry) {
	var (
		protPortPairSlice []*portsInfo
		podRuleSets       []string // pod sets listed in Ingress rules.
		nsRuleSets        []string // namespace sets listed in Ingress rules
		entries           []*iptm.IptEntry
	)
	//TODO: handle IPBlock
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
				podRuleSets = append(podRuleSets, ns+"-"+podLabelKey+":"+podLabelVal)
			}

			for nsLabelKey, nsLabelVal := range fromRule.NamespaceSelector.MatchLabels {
				nsRuleSets = append(nsRuleSets, ns+"-"+nsLabelKey+":"+nsLabelVal)
			}
		}
	}

	// Use hashed string for ipset name to avoid string length limit of ipset.
	for _, targetSet := range targetSets {
		hashedTargetSetName := azureNpmPrefix + util.Hash(targetSet)
		for _, protPortPair := range protPortPairSlice {
			entry := &iptm.IptEntry{
				Name:       targetSet,
				HashedName: hashedTargetSetName,
				Chain:      iptm.IptablesAzureChain,
				Specs: []string{
					iptm.IptablesPortFlag, protPortPair.protocol,
					iptm.IptablesDstPortFlag, protPortPair.port,
					iptm.IptablesMatchFlag,
					iptm.IptablesSetFlag,
					iptm.IptablesMatchSetFlag,
					hashedTargetSetName,
					iptm.IptablesDstFlag,
					iptm.IptablesJumpFlag,
					iptm.IptablesAccept,
				},
			}
			entries = append(entries, entry)
		}

		// Handle PodSelector field of NetworkPolicyPeer.
		for _, podRuleSet := range podRuleSets {
			hashedRuleSetName := azureNpmPrefix + util.Hash(podRuleSet)
			entry := &iptm.IptEntry{
				Name:       podRuleSet,
				HashedName: hashedRuleSetName,
				Chain:      iptm.IptablesAzureChain,
				Specs: []string{
					iptm.IptablesMatchFlag,
					iptm.IptablesSetFlag,
					iptm.IptablesMatchSetFlag,
					hashedRuleSetName,
					iptm.IptablesSrcFlag,
					iptm.IptablesMatchFlag,
					iptm.IptablesSetFlag,
					iptm.IptablesMatchSetFlag,
					hashedTargetSetName,
					iptm.IptablesDstFlag,
					iptm.IptablesJumpFlag,
					iptm.IptablesAccept,
				},
			}
			entries = append(entries, entry)
		}

		// TODO: Handle NamespaceSelector field of NetworkPolicyPeer. Use namespace selector to match corresponding namespaces.
		for _, nsRuleSet := range nsRuleSets {
			hashedRuleSetName := azureNpmPrefix + util.Hash(nsRuleSet)
			entry := &iptm.IptEntry{
				Name:       nsRuleSet,
				HashedName: hashedRuleSetName,
				Chain:      iptm.IptablesAzureChain,
				Specs: []string{
					iptm.IptablesMatchFlag,
					iptm.IptablesSetFlag,
					iptm.IptablesMatchSetFlag,
					hashedRuleSetName,
					iptm.IptablesSrcFlag,
					iptm.IptablesMatchFlag,
					iptm.IptablesSetFlag,
					iptm.IptablesMatchSetFlag,
					hashedTargetSetName,
					iptm.IptablesDstFlag,
					iptm.IptablesJumpFlag,
					iptm.IptablesAccept,
				},
			}
			entries = append(entries, entry)
		}

		// TODO: Handle IPBlock field of NetworkPolicyPeer.
	}
	return podRuleSets, entries
}

func parseEgress(rules []networkingv1.NetworkPolicyEgressRule) ([]string, []*iptm.IptEntry) {

	return nil, nil
}

// ParsePolicy parses network policy.
func parsePolicy(npObj *networkingv1.NetworkPolicy) ([]string, []*iptm.IptEntry) {

	var (
		sets    []string
		entries []*iptm.IptEntry
	)

	// Get affected pods.
	npNs, selector := npObj.ObjectMeta.Namespace, npObj.Spec.PodSelector.MatchLabels
	for podLabelKey, podLabelVal := range selector {
		set := npNs + "-" + podLabelKey + ":" + podLabelVal
		sets = append(sets, set)
	}

	ingressSets, ingressEntries := parseIngress(npNs, sets, npObj.Spec.Ingress)
	sets = append(sets, ingressSets...)
	entries = append(entries, ingressEntries...)

	/*
		egressSets, egressEntries := parseEgress(np.Spec.Egress)
		append(sets, egressSets)
		append(entries, egressEntries)
	*/

	return sets, entries
}

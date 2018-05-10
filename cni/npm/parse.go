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
	if len(rules) == 0 {
		return nil, nil
	}

	var (
		portRuleExists    = false
		fromRuleExists    = false
		protPortPairSlice []*portsInfo
		podRuleSets       []string // pod sets listed in Ingress rules.
		nsRuleSets        []string // namespace sets listed in Ingress rules
		entries           []*iptm.IptEntry
		ipblock           *networkingv1.IPBlock
	)
	//TODO: handle IPBlock
	for _, rule := range rules {
		for _, portRule := range rule.Ports {
			protPortPairSlice = append(protPortPairSlice,
				&portsInfo{
					protocol: string(*portRule.Protocol),
					port:     fmt.Sprint(portRule.Port.IntVal),
				})

			portRuleExists = true
		}

		for _, fromRule := range rule.From {
			if fromRule.PodSelector != nil {
				for podLabelKey, podLabelVal := range fromRule.PodSelector.MatchLabels {
					podRuleSets = append(podRuleSets, ns+"-"+podLabelKey+":"+podLabelVal)
				}
			}

			if fromRule.NamespaceSelector != nil {
				for nsLabelKey, nsLabelVal := range fromRule.NamespaceSelector.MatchLabels {
					nsRuleSets = append(nsRuleSets, "ns-"+nsLabelKey+":"+nsLabelVal)
				}
			}

			if fromRule.IPBlock != nil {
				ipblock = fromRule.IPBlock
			}

			fromRuleExists = true
		}
	}

	// Use hashed string for ipset name to avoid string length limit of ipset.
	for _, targetSet := range targetSets {
		hashedTargetSetName := azureNpmPrefix + util.Hash(targetSet)

		if !portRuleExists {
			entry := &iptm.IptEntry{
				Name:       targetSet,
				HashedName: hashedTargetSetName,
				Chain:      util.IptablesAzureIngressPortChain,
				Specs: []string{
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedTargetSetName,
					util.IptablesDstFlag,
					util.IptablesJumpFlag,
					util.IptablesAzureIngressFromChain,
				},
			}
			entries = append(entries, entry)
		} else {
			for _, protPortPair := range protPortPairSlice {
				entry := &iptm.IptEntry{
					Name:       targetSet,
					HashedName: hashedTargetSetName,
					Chain:      util.IptablesAzureIngressPortChain,
					Specs: []string{
						util.IptablesPortFlag,
						protPortPair.protocol,
						util.IptablesDstPortFlag,
						protPortPair.port,
						util.IptablesMatchFlag,
						util.IptablesSetFlag,
						util.IptablesMatchSetFlag,
						hashedTargetSetName,
						util.IptablesDstFlag,
						util.IptablesJumpFlag,
						util.IptablesAzureIngressFromChain,
					},
				}
				entries = append(entries, entry)
			}
		}

		if !fromRuleExists {
			entry := &iptm.IptEntry{
				Name:       targetSet,
				HashedName: hashedTargetSetName,
				Chain:      util.IptablesAzureIngressFromChain,
				Specs: []string{
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedTargetSetName,
					util.IptablesDstFlag,
					util.IptablesJumpFlag,
					util.IptablesAccept,
				},
			}
			entries = append(entries, entry)
			continue
		}

		// Handle PodSelector field of NetworkPolicyPeer.
		for _, podRuleSet := range podRuleSets {
			hashedRuleSetName := azureNpmPrefix + util.Hash(podRuleSet)
			entry := &iptm.IptEntry{
				Name:       podRuleSet,
				HashedName: hashedRuleSetName,
				Chain:      util.IptablesAzureIngressFromChain,
				Specs: []string{
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedRuleSetName,
					util.IptablesSrcFlag,
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedTargetSetName,
					util.IptablesDstFlag,
					util.IptablesJumpFlag,
					util.IptablesAccept,
				},
			}
			entries = append(entries, entry)
		}

		// Handle NamespaceSelector field of NetworkPolicyPeer
		for _, nsRuleSet := range nsRuleSets {
			hashedRuleSetName := azureNpmPrefix + util.Hash(nsRuleSet)
			entry := &iptm.IptEntry{
				Name:       nsRuleSet,
				HashedName: hashedRuleSetName,
				Chain:      util.IptablesAzureIngressFromChain,
				Specs: []string{
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedRuleSetName,
					util.IptablesSrcFlag,
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedTargetSetName,
					util.IptablesDstFlag,
					util.IptablesJumpFlag,
					util.IptablesAccept,
				},
			}
			entries = append(entries, entry)
		}

		if ipblock == nil {
			continue
		}

		// Handle ipblock field of NetworkPolicyPeer
		if len(ipblock.Except) > 0 {
			for _, except := range ipblock.Except {
				entry := &iptm.IptEntry{
					Chain: util.IptablesAzureIngressFromChain,
					Specs: []string{
						util.IptablesSFlag,
						except,
						util.IptablesJumpFlag,
						util.IptablesReject,
					},
				}
				entries = append(entries, entry)
			}
		}

		if len(ipblock.CIDR) > 0 {
			cidrEntry := &iptm.IptEntry{
				Chain: util.IptablesAzureIngressFromChain,
				Specs: []string{
					util.IptablesSFlag,
					ipblock.CIDR,
					util.IptablesJumpFlag,
					util.IptablesAccept,
				},
			}
			entries = append(entries, cidrEntry)
		}
	}
	return podRuleSets, entries
}

func parseEgress(ns string, targetSets []string, rules []networkingv1.NetworkPolicyEgressRule) ([]string, []*iptm.IptEntry) {
	if len(rules) == 0 {
		return nil, nil
	}

	var (
		portRuleExists    = false
		toRuleExists      = false
		protPortPairSlice []*portsInfo
		podRuleSets       []string // pod sets listed in Egress rules.
		nsRuleSets        []string // namespace sets listed in Egress rules
		entries           []*iptm.IptEntry
		ipblock           *networkingv1.IPBlock
	)

	//TODO: handle IPBlock
	for _, rule := range rules {
		for _, portRule := range rule.Ports {
			protPortPairSlice = append(protPortPairSlice,
				&portsInfo{
					protocol: string(*portRule.Protocol),
					port:     fmt.Sprint(portRule.Port.IntVal),
				})

			portRuleExists = true
		}

		for _, toRule := range rule.To {
			if toRule.PodSelector != nil {
				for podLabelKey, podLabelVal := range toRule.PodSelector.MatchLabels {
					podRuleSets = append(podRuleSets, ns+"-"+podLabelKey+":"+podLabelVal)
				}
			}

			if toRule.NamespaceSelector != nil {
				for nsLabelKey, nsLabelVal := range toRule.NamespaceSelector.MatchLabels {
					nsRuleSets = append(nsRuleSets, "ns-"+nsLabelKey+":"+nsLabelVal)
				}
			}

			if toRule.IPBlock != nil {
				ipblock = toRule.IPBlock
			}

			toRuleExists = true
		}
	}

	// Use hashed string for ipset name to avoid string length limit of ipset.
	for _, targetSet := range targetSets {
		hashedTargetSetName := azureNpmPrefix + util.Hash(targetSet)

		if !portRuleExists {
			entry := &iptm.IptEntry{
				Name:       targetSet,
				HashedName: hashedTargetSetName,
				Chain:      util.IptablesAzureEgressPortChain,
				Specs: []string{
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedTargetSetName,
					util.IptablesDstFlag,
					util.IptablesJumpFlag,
					util.IptablesAzureEgressToChain,
				},
			}
			entries = append(entries, entry)
		} else {
			for _, protPortPair := range protPortPairSlice {
				entry := &iptm.IptEntry{
					Name:       targetSet,
					HashedName: hashedTargetSetName,
					Chain:      util.IptablesAzureEgressPortChain,
					Specs: []string{
						util.IptablesPortFlag,
						protPortPair.protocol,
						util.IptablesDstPortFlag,
						protPortPair.port,
						util.IptablesMatchFlag,
						util.IptablesSetFlag,
						util.IptablesMatchSetFlag,
						hashedTargetSetName,
						util.IptablesSrcFlag,
						util.IptablesJumpFlag,
						util.IptablesAzureEgressToChain,
					},
				}
				entries = append(entries, entry)
			}
		}

		if !toRuleExists {
			entry := &iptm.IptEntry{
				Name:       targetSet,
				HashedName: hashedTargetSetName,
				Chain:      util.IptablesAzureEgressToChain,
				Specs: []string{
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedTargetSetName,
					util.IptablesDstFlag,
					util.IptablesJumpFlag,
					util.IptablesAccept,
				},
			}
			entries = append(entries, entry)
			continue
		}

		// Handle PodSelector field of NetworkPolicyPeer.
		for _, podRuleSet := range podRuleSets {
			hashedRuleSetName := azureNpmPrefix + util.Hash(podRuleSet)
			entry := &iptm.IptEntry{
				Name:       podRuleSet,
				HashedName: hashedRuleSetName,
				Chain:      util.IptablesAzureEgressToChain,
				Specs: []string{
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedTargetSetName,
					util.IptablesSrcFlag,
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedRuleSetName,
					util.IptablesDstFlag,
					util.IptablesJumpFlag,
					util.IptablesAccept,
				},
			}
			entries = append(entries, entry)
		}

		// Handle NamespaceSelector field of NetworkPolicyPeer
		for _, nsRuleSet := range nsRuleSets {
			hashedRuleSetName := azureNpmPrefix + util.Hash(nsRuleSet)
			entry := &iptm.IptEntry{
				Name:       nsRuleSet,
				HashedName: hashedRuleSetName,
				Chain:      util.IptablesAzureEgressToChain,
				Specs: []string{
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedTargetSetName,
					util.IptablesSrcFlag,
					util.IptablesMatchFlag,
					util.IptablesSetFlag,
					util.IptablesMatchSetFlag,
					hashedRuleSetName,
					util.IptablesDstFlag,
					util.IptablesJumpFlag,
					util.IptablesAccept,
				},
			}
			entries = append(entries, entry)
		}

		if ipblock == nil {
			continue
		}

		// Handle ipblock field of NetworkPolicyPeer
		if len(ipblock.Except) > 0 {
			for _, except := range ipblock.Except {
				entry := &iptm.IptEntry{
					Chain: util.IptablesAzureEgressToChain,
					Specs: []string{
						util.IptablesDFlag,
						except,
						util.IptablesJumpFlag,
						util.IptablesReject,
					},
				}
				entries = append(entries, entry)
			}
		}

		if len(ipblock.CIDR) > 0 {
			cidrEntry := &iptm.IptEntry{
				Chain: util.IptablesAzureEgressToChain,
				Specs: []string{
					util.IptablesDFlag,
					ipblock.CIDR,
					util.IptablesJumpFlag,
					util.IptablesAccept,
				},
			}
			entries = append(entries, cidrEntry)
		}
	}

	return podRuleSets, entries
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

	egressSets, egressEntries := parseEgress(npNs, sets, npObj.Spec.Egress)
	sets = append(sets, egressSets...)
	entries = append(entries, egressEntries...)

	return util.UniqueStrSlice(sets), entries
}

package util

//kubernetes related constants.
const (
	KubeSystemFlag          string = "kube-system"
	KubePodTemplateHashFlag string = "pod-template-hash"
	KubeAllPodsFlag         string = "all-pod"
	KubeAllNamespacesFlag   string = "all-namespace"
)

//iptables related constants.
const (
	Iptables                      string = "iptables"
	IptablesChainCreationFlag     string = "-N"
	IptablesInsertionFlag         string = "-I"
	IptablesAppendFlag            string = "-A"
	IptablesDeletionFlag          string = "-D"
	IptablesFlushFlag             string = "-F"
	IptablesCheckFlag             string = "-C"
	IptablesDestroyFlag           string = "-X"
	IptablesJumpFlag              string = "-j"
	IptablesAccept                string = "ACCEPT"
	IptablesReject                string = "REJECT"
	IptablesDrop                  string = "DROP"
	IptablesSrcFlag               string = "src"
	IptablesDstFlag               string = "dst"
	IptablesPortFlag              string = "-p"
	IptablesSFlag                 string = "-s"
	IptablesDFlag                 string = "-d"
	IptablesDstPortFlag           string = "--dport"
	IptablesMatchFlag             string = "-m"
	IptablesSetFlag               string = "set"
	IptablesMatchSetFlag          string = "--match-set"
	IptablesStateFlag             string = "state"
	IPtablesMatchStateFlag        string = "--state"
	IptablesRelatedState          string = "RELATED"
	IptablesEstablishedState      string = "ESTABLISHED"
	IptablesAzureChain            string = "AZURE-NPM"
	IptablesAzureIngressPortChain string = "AZURE-NPM-INGRESS-PORT"
	IptablesAzureIngressFromChain string = "AZURE-NPM-INGRESS-FROM"
	IptablesAzureEgressPortChain  string = "AZURE-NPM-EGRESS-PORT"
	IptablesAzureEgressToChain    string = "AZURE-NPM-EGRESS-TO"
	IptablesForwardChain          string = "FORWARD"
)

//ipset related constants.
const (
	Ipset             string = "ipset"
	IpsetCreationFlag string = "-N"
	IpsetAppendFlag   string = "-A"
	IpsetDeletionFlag string = "-D"
	IpsetDestroyFlag  string = "-X"

	IpsetExistFlag string = "-exist"

	IpsetSetListFlag string = "setlist"
	IpsetNetHashFlag string = "nethash"
	AzureNpmPrefix   string = "azure-npm-"
)

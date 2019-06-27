package network

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Azure/azure-container-networking/cni"
	"github.com/Azure/azure-container-networking/cns"
	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/network"
	"github.com/Azure/azure-container-networking/network/policy"
	"github.com/Microsoft/hcsshim"

	cniTypes "github.com/containernetworking/cni/pkg/types"
	cniTypesCurr "github.com/containernetworking/cni/pkg/types/current"
)

/* handleConsecutiveAdd handles consecutive add calls for infrastructure containers on Windows platform.
 * This is a temporary work around for issue #57253 of Kubernetes.
 * We can delete this if statement once they fix it.
 * Issue link: https://github.com/kubernetes/kubernetes/issues/57253
 */
func handleConsecutiveAdd(containerId, endpointId string, nwInfo *network.NetworkInfo, nwCfg *cni.NetworkConfig) (*cniTypesCurr.Result, error) {
	hnsEndpoint, err := hcsshim.GetHNSEndpointByName(endpointId)
	if hnsEndpoint != nil {
		log.Printf("[net] Found existing endpoint through hcsshim: %+v", hnsEndpoint)
		log.Printf("[net] Attaching ep %v to container %v", hnsEndpoint.Id, containerId)

		err := hcsshim.HotAttachEndpoint(containerId, hnsEndpoint.Id)
		if err != nil {
			log.Printf("[cni-net] Failed to hot attach shared endpoint[%v] to container [%v], err:%v.", hnsEndpoint.Id, containerId, err)
			return nil, err
		}

		// Populate result.
		address := nwInfo.Subnets[0].Prefix
		address.IP = hnsEndpoint.IPAddress
		result := &cniTypesCurr.Result{
			IPs: []*cniTypesCurr.IPConfig{
				{
					Version: "4",
					Address: address,
					Gateway: net.ParseIP(hnsEndpoint.GatewayAddress),
				},
			},
			Routes: []*cniTypes.Route{
				{
					Dst: net.IPNet{net.IPv4zero, net.IPv4Mask(0, 0, 0, 0)},
					GW:  net.ParseIP(hnsEndpoint.GatewayAddress),
				},
			},
		}

		// Populate DNS servers.
		result.DNS.Nameservers = nwCfg.DNS.Nameservers

		return result, nil
	}

	err = fmt.Errorf("GetHNSEndpointByName for %v returned nil with err %v", endpointId, err)
	return nil, err
}

func addDefaultRoute(gwIPString string, epInfo *network.EndpointInfo, result *cniTypesCurr.Result) {
}

func addInfraRoutes(azIpamResult *cniTypesCurr.Result, result *cniTypesCurr.Result, epInfo *network.EndpointInfo) {
}

func setNetworkOptions(cnsNwConfig *cns.GetNetworkContainerResponse, nwInfo *network.NetworkInfo) {
	if cnsNwConfig != nil && cnsNwConfig.MultiTenancyInfo.ID != 0 {
		log.Printf("Setting Network Options")
		vlanMap := make(map[string]interface{})
		vlanMap[network.VlanIDKey] = strconv.Itoa(cnsNwConfig.MultiTenancyInfo.ID)
		nwInfo.Options[dockerNetworkOption] = vlanMap
	}
}

func setEndpointOptions(cnsNwConfig *cns.GetNetworkContainerResponse, epInfo *network.EndpointInfo, vethName string) {
	if cnsNwConfig != nil && cnsNwConfig.MultiTenancyInfo.ID != 0 {
		log.Printf("Setting Endpoint Options")
		var cnetAddressMap []string
		for _, ipSubnet := range cnsNwConfig.CnetAddressSpace {
			cnetAddressMap = append(cnetAddressMap, ipSubnet.IPAddress+"/"+strconv.Itoa(int(ipSubnet.PrefixLength)))
		}
		epInfo.Data[network.CnetAddressSpace] = cnetAddressMap
	}
}

func addSnatInterface(nwCfg *cni.NetworkConfig, result *cniTypesCurr.Result) {
}

func updateSubnetPrefix(cnsNwConfig *cns.GetNetworkContainerResponse, subnetPrefix *net.IPNet) error {
	if cnsNwConfig != nil && cnsNwConfig.MultiTenancyInfo.ID != 0 {
		ipconfig := cnsNwConfig.IPConfiguration
		ipAddr := net.ParseIP(ipconfig.IPSubnet.IPAddress)
		if ipAddr.To4() != nil {
			*subnetPrefix = net.IPNet{Mask: net.CIDRMask(int(ipconfig.IPSubnet.PrefixLength), 32)}
		} else if ipAddr.To16() != nil {
			*subnetPrefix = net.IPNet{Mask: net.CIDRMask(int(ipconfig.IPSubnet.PrefixLength), 128)}
		} else {
			return fmt.Errorf("[cni-net] Failed to get mask from CNS network configuration")
		}

		subnetPrefix.IP = ipAddr.Mask(subnetPrefix.Mask)
		log.Printf("Updated subnetPrefix: %s", subnetPrefix.String())
	}

	return nil
}

func getNetworkName(podName, podNs, ifName string, nwCfg *cni.NetworkConfig) (networkName string, err error) {
	networkName = nwCfg.Name
	err = nil
	if nwCfg.MultiTenancy {
		if len(strings.TrimSpace(podName)) == 0 || len(strings.TrimSpace(podNs)) == 0 {
			err = fmt.Errorf("POD info cannot be empty. PodName: %s, PodNamespace: %s", podName, podNs)
			return
		}

		_, cnsNetworkConfig, _, err := getContainerNetworkConfiguration(nwCfg, podName, podNs, ifName)
		if err != nil {
			log.Printf("GetContainerNetworkConfiguration failed for podname %v namespace %v with error %v", podName, podNs, err)
		} else {
			var subnet net.IPNet
			if err = updateSubnetPrefix(cnsNetworkConfig, &subnet); err == nil {
				// networkName will look like ~ azure-vlan1-172-28-1-0_24
				networkName = strings.Replace(subnet.String(), ".", "-", -1)
				networkName = strings.Replace(networkName, "/", "_", -1)
				networkName = fmt.Sprintf("%s-vlan%v-%v", nwCfg.Name, cnsNetworkConfig.MultiTenancyInfo.ID, networkName)
			}
		}
	}

	return
}

func setupInfraVnetRoutingForMultitenancy(
	nwCfg *cni.NetworkConfig,
	azIpamResult *cniTypesCurr.Result,
	epInfo *network.EndpointInfo,
	result *cniTypesCurr.Result) {
}

func getNetworkDNSSettings(nwCfg *cni.NetworkConfig, result *cniTypesCurr.Result, namespace string) (network.DNSInfo, error) {
	var nwDNS network.DNSInfo

	// use custom dns if present
	nwDNS = getCustomDNS(nwCfg)
	if len(nwDNS.Servers) > 0 || nwDNS.Suffix != "" {
		return nwDNS, nil
	}

	if (len(nwCfg.DNS.Search) == 0) != (len(nwCfg.DNS.Nameservers) == 0) {
		err := fmt.Errorf("Wrong DNS configuration: %+v", nwCfg.DNS)
		return nwDNS, err
	}

	nwDNS = network.DNSInfo{
		Servers: nwCfg.DNS.Nameservers,
	}

	return nwDNS, nil
}

func getEndpointDNSSettings(nwCfg *cni.NetworkConfig, result *cniTypesCurr.Result, namespace string) (network.DNSInfo, error) {
	var epDNS network.DNSInfo

	// use custom dns if present
	epDNS = getCustomDNS(nwCfg)
	if len(epDNS.Servers) > 0 || epDNS.Suffix != "" {
		return epDNS, nil
	}

	if (len(nwCfg.DNS.Search) == 0) != (len(nwCfg.DNS.Nameservers) == 0) {
		err := fmt.Errorf("Wrong DNS configuration: %+v", nwCfg.DNS)
		return epDNS, err
	}

	if len(nwCfg.DNS.Search) > 0 {
		epDNS = network.DNSInfo{
			Servers: nwCfg.DNS.Nameservers,
			Suffix:  namespace + "." + strings.Join(nwCfg.DNS.Search, ","),
		}
	} else {
		epDNS = network.DNSInfo{
			Suffix:  result.DNS.Domain,
			Servers: result.DNS.Nameservers,
		}
	}

	return epDNS, nil
}

// getPoliciesFromRuntimeCfg returns network policies from network config.
func getPoliciesFromRuntimeCfg(nwCfg *cni.NetworkConfig) []policy.Policy {
	log.Printf("[net] RuntimeConfigs: %+v", nwCfg.RuntimeConfig)
	var policies []policy.Policy
	for _, mapping := range nwCfg.RuntimeConfig.PortMappings {
		rawPolicy, _ := json.Marshal(&hcsshim.NatPolicy{
			Type:         "NAT",
			ExternalPort: uint16(mapping.HostPort),
			InternalPort: uint16(mapping.ContainerPort),
			Protocol:     mapping.Protocol,
		})

		policy := policy.Policy{
			Type: policy.EndpointPolicy,
			Data: rawPolicy,
		}
		log.Printf("[net] Creating port mapping policy: %+v", policy)

		policies = append(policies, policy)
	}

	return policies
}

func getCustomDNS(nwCfg *cni.NetworkConfig) network.DNSInfo {
	log.Printf("[net] RuntimeConfigs: %+v", nwCfg.RuntimeConfig)

	var search string
	if len(nwCfg.RuntimeConfig.DNS.Searches) > 0 {
		search = strings.Join(nwCfg.RuntimeConfig.DNS.Searches, ",")
	}

	return network.DNSInfo{
		Servers: nwCfg.RuntimeConfig.DNS.Servers,
		Suffix:  search,
	}
}

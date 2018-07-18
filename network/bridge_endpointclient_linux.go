package network

import (
	"net"

	"github.com/Azure/azure-container-networking/ebtables"
	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/netlink"
)

type LinuxBridgeEndpointClient struct {
	bridgeName        string
	hostPrimaryIfName string
	hostVethName      string
	containerVethName string
	hostPrimaryMac    net.HardwareAddr
	containerMac      net.HardwareAddr
	mode              string
}

func NewLinuxBridgeEndpointClient(
	extIf *externalInterface,
	hostVethName string,
	containerVethName string,
	mode string,
) *LinuxBridgeEndpointClient {

	client := &LinuxBridgeEndpointClient{
		bridgeName:        extIf.BridgeName,
		hostPrimaryIfName: extIf.Name,
		hostVethName:      hostVethName,
		containerVethName: containerVethName,
		hostPrimaryMac:    extIf.MacAddress,
		mode:              mode,
	}

	return client
}

func (client *LinuxBridgeEndpointClient) AddEndpoints(epInfo *EndpointInfo) error {
	if err := createEndpoint(client.hostVethName, client.containerVethName); err != nil {
		return err
	}

	containerIf, err := net.InterfaceByName(client.containerVethName)
	if err != nil {
		return err
	}

	client.containerMac = containerIf.HardwareAddr
	return nil
}

func (client *LinuxBridgeEndpointClient) AddEndpointRules(epInfo *EndpointInfo) error {
	var err error

	log.Printf("[net] Setting link %v master %v.", client.hostVethName, client.bridgeName)
	if err := netlink.SetLinkMaster(client.hostVethName, client.bridgeName); err != nil {
		return err
	}

	for _, ipAddr := range epInfo.IPAddresses {
		// Add ARP reply rule.
		log.Printf("[net] Adding ARP reply rule for IP address %v", ipAddr.String())
		if err = ebtables.SetArpReply(ipAddr.IP, client.getArpReplyAddress(client.containerMac), ebtables.Append); err != nil {
			return err
		}

		// Add MAC address translation rule.
		log.Printf("[net] Adding MAC DNAT rule for IP address %v", ipAddr.String())
		if err := ebtables.SetDnatForIPAddress(client.hostPrimaryIfName, ipAddr.IP, client.containerMac, ebtables.Append); err != nil {
			return err
		}
	}

	return nil
}

func (client *LinuxBridgeEndpointClient) DeleteEndpointRules(ep *endpoint) {
	// Delete rules for IP addresses on the container interface.
	for _, ipAddr := range ep.IPAddresses {
		// Delete ARP reply rule.
		log.Printf("[net] Deleting ARP reply rule for IP address %v on %v.", ipAddr.String(), ep.Id)
		err := ebtables.SetArpReply(ipAddr.IP, client.getArpReplyAddress(ep.MacAddress), ebtables.Delete)
		if err != nil {
			log.Printf("[net] Failed to delete ARP reply rule for IP address %v: %v.", ipAddr.String(), err)
		}

		// Delete MAC address translation rule.
		log.Printf("[net] Deleting MAC DNAT rule for IP address %v on %v.", ipAddr.String(), ep.Id)
		err = ebtables.SetDnatForIPAddress(client.hostPrimaryIfName, ipAddr.IP, ep.MacAddress, ebtables.Delete)
		if err != nil {
			log.Printf("[net] Failed to delete MAC DNAT rule for IP address %v: %v.", ipAddr.String(), err)
		}
	}
}

// getArpReplyAddress returns the MAC address to use in ARP replies.
func (client *LinuxBridgeEndpointClient) getArpReplyAddress(epMacAddress net.HardwareAddr) net.HardwareAddr {
	var macAddress net.HardwareAddr

	if client.mode == opModeTunnel {
		// In tunnel mode, resolve all IP addresses to the virtual MAC address for hairpinning.
		macAddress, _ = net.ParseMAC(virtualMacAddress)
	} else {
		// Otherwise, resolve to actual MAC address.
		macAddress = epMacAddress
	}

	return macAddress
}

func (client *LinuxBridgeEndpointClient) MoveEndpointsToContainerNS(epInfo *EndpointInfo, nsID uintptr) error {
	// Move the container interface to container's network namespace.
	log.Printf("[net] Setting link %v netns %v.", client.containerVethName, epInfo.NetNsPath)
	if err := netlink.SetLinkNetNs(client.containerVethName, nsID); err != nil {
		return err
	}

	return nil
}

func (client *LinuxBridgeEndpointClient) SetupContainerInterfaces(epInfo *EndpointInfo) error {
	if err := setupContainerInterface(client.containerVethName, epInfo.IfName); err != nil {
		return err
	}

	client.containerVethName = epInfo.IfName
	return nil
}

func (client *LinuxBridgeEndpointClient) ConfigureContainerInterfacesAndRoutes(epInfo *EndpointInfo) error {
	if err := assignIPToInterface(client.containerVethName, epInfo.IPAddresses); err != nil {
		return err
	}

	if err := addRoutes(client.containerVethName, epInfo.Routes); err != nil {
		return err
	}

	return nil
}

func (client *LinuxBridgeEndpointClient) DeleteEndpoints(ep *endpoint) error {
	log.Printf("[net] Deleting veth pair %v %v.", ep.HostIfName, ep.IfName)
	err := netlink.DeleteLink(ep.HostIfName)
	if err != nil {
		log.Printf("[net] Failed to delete veth pair %v: %v.", ep.HostIfName, err)
		return err
	}

	return nil
}

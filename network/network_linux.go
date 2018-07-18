// Copyright 2017 Microsoft. All rights reserved.
// MIT License

// +build linux

package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/netlink"
	"golang.org/x/sys/unix"
)

const (
	// Prefix for bridge names.
	bridgePrefix = "azure"

	// Virtual MAC address used by Azure VNET.
	virtualMacAddress = "12:34:56:78:9a:bc"

	genericData = "com.docker.network.generic"

	SnatBridgeIPKey = "snatBridgeIP"

	LocalIPKey = "localIP"

	OptVethName = "vethname"
)

// Linux implementation of route.
type route netlink.Route

// NewNetworkImpl creates a new container network.
func (nm *networkManager) newNetworkImpl(nwInfo *NetworkInfo, extIf *externalInterface) (*network, error) {
	// Connect the external interface.
	var vlanid int
	opt, _ := nwInfo.Options[genericData].(map[string]interface{})
	log.Printf("opt %+v options %+v", opt, nwInfo.Options)

	switch nwInfo.Mode {
	case opModeTunnel:
		fallthrough
	case opModeBridge:
		log.Printf("create bridge")
		if err := nm.connectExternalInterface(extIf, nwInfo); err != nil {
			return nil, err
		}

		if opt != nil && opt[VlanIDKey] != nil {
			vlanid, _ = strconv.Atoi(opt[VlanIDKey].(string))
		}

	default:
		return nil, errNetworkModeInvalid
	}

	// Create the network object.
	nw := &network{
		Id:               nwInfo.Id,
		Mode:             nwInfo.Mode,
		Endpoints:        make(map[string]*endpoint),
		extIf:            extIf,
		VlanId:           vlanid,
		EnableSnatOnHost: nwInfo.EnableSnatOnHost,
	}

	return nw, nil
}

// DeleteNetworkImpl deletes an existing container network.
func (nm *networkManager) deleteNetworkImpl(nw *network) error {
	var networkClient NetworkClient

	if nw.VlanId != 0 {
		networkClient = NewOVSClient(nw.extIf.BridgeName, nw.extIf.Name, "", nw.EnableSnatOnHost)
	} else {
		networkClient = NewLinuxBridgeClient(nw.extIf.BridgeName, nw.extIf.Name, nw.Mode)
	}

	// Disconnect the interface if this was the last network using it.
	if len(nw.extIf.Networks) == 1 {
		nm.disconnectExternalInterface(nw.extIf, networkClient)
	}

	return nil
}

//  SaveIPConfig saves the IP configuration of an interface.
func (nm *networkManager) saveIPConfig(hostIf *net.Interface, extIf *externalInterface) error {
	// Save the default routes on the interface.
	routes, err := netlink.GetIpRoute(&netlink.Route{Dst: &net.IPNet{}, LinkIndex: hostIf.Index})
	if err != nil {
		log.Printf("[net] Failed to query routes: %v.", err)
		return err
	}

	for _, r := range routes {
		if r.Dst == nil {
			if r.Family == unix.AF_INET {
				extIf.IPv4Gateway = r.Gw
			} else if r.Family == unix.AF_INET6 {
				extIf.IPv6Gateway = r.Gw
			}
		}

		extIf.Routes = append(extIf.Routes, (*route)(r))
	}

	// Save global unicast IP addresses on the interface.
	addrs, err := hostIf.Addrs()
	for _, addr := range addrs {
		ipAddr, ipNet, err := net.ParseCIDR(addr.String())
		ipNet.IP = ipAddr
		if err != nil {
			continue
		}

		if !ipAddr.IsGlobalUnicast() {
			continue
		}

		extIf.IPAddresses = append(extIf.IPAddresses, ipNet)

		log.Printf("[net] Deleting IP address %v from interface %v.", ipNet, hostIf.Name)

		err = netlink.DeleteIpAddress(hostIf.Name, ipAddr, ipNet)
		if err != nil {
			break
		}
	}

	log.Printf("[net] Saved interface IP configuration %+v.", extIf)

	return err
}

// ApplyIPConfig applies a previously saved IP configuration to an interface.
func (nm *networkManager) applyIPConfig(extIf *externalInterface, targetIf *net.Interface) error {
	// Add IP addresses.
	for _, addr := range extIf.IPAddresses {
		log.Printf("[net] Adding IP address %v to interface %v.", addr, targetIf.Name)

		err := netlink.AddIpAddress(targetIf.Name, addr.IP, addr)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "file exists") {
			log.Printf("[net] Failed to add IP address %v: %v.", addr, err)
			return err
		}
	}

	// Add IP routes.
	for _, route := range extIf.Routes {
		route.LinkIndex = targetIf.Index

		log.Printf("[net] Adding IP route %+v.", route)

		err := netlink.AddIpRoute((*netlink.Route)(route))
		if err != nil {
			log.Printf("[net] Failed to add IP route %v: %v.", route, err)
			return err
		}
	}

	return nil
}

// ConnectExternalInterface connects the given host interface to a bridge.
func (nm *networkManager) connectExternalInterface(extIf *externalInterface, nwInfo *NetworkInfo) error {
	var err error
	var networkClient NetworkClient
	log.Printf("[net] Connecting interface %v.", extIf.Name)
	defer func() { log.Printf("[net] Connecting interface %v completed with err:%v.", extIf.Name, err) }()

	// Check whether this interface is already connected.
	if extIf.BridgeName != "" {
		log.Printf("[net] Interface is already connected to bridge %v.", extIf.BridgeName)
		return nil
	}

	// Find the external interface.
	hostIf, err := net.InterfaceByName(extIf.Name)
	if err != nil {
		return err
	}

	// If a bridge name is not specified, generate one based on the external interface index.
	bridgeName := nwInfo.BridgeName
	if bridgeName == "" {
		bridgeName = fmt.Sprintf("%s%d", bridgePrefix, hostIf.Index)
	}

	opt, _ := nwInfo.Options[genericData].(map[string]interface{})
	if opt != nil && opt[VlanIDKey] != nil {
		snatBridgeIP := ""

		if opt != nil && opt[SnatBridgeIPKey] != nil {
			snatBridgeIP, _ = opt[SnatBridgeIPKey].(string)
		}

		networkClient = NewOVSClient(bridgeName, extIf.Name, snatBridgeIP, nwInfo.EnableSnatOnHost)
	} else {
		networkClient = NewLinuxBridgeClient(bridgeName, extIf.Name, nwInfo.Mode)
	}

	// Check if the bridge already exists.
	bridge, err := net.InterfaceByName(bridgeName)
	if err != nil {
		// Create the bridge.

		if err := networkClient.CreateBridge(); err != nil {
			log.Printf("Error while creating bridge %+v", err)
			return err
		}

		// On failure, delete the bridge.
		defer func() {
			if err != nil {
				networkClient.DeleteBridge()
			}
		}()

		bridge, err = net.InterfaceByName(bridgeName)
		if err != nil {
			return err
		}
	} else {
		// Use the existing bridge.
		log.Printf("[net] Found existing bridge %v.", bridgeName)
	}

	// Save host IP configuration.
	err = nm.saveIPConfig(hostIf, extIf)
	if err != nil {
		log.Printf("[net] Failed to save IP configuration for interface %v: %v.", hostIf.Name, err)
	}

	// External interface down.
	log.Printf("[net] Setting link %v state down.", hostIf.Name)
	err = netlink.SetLinkState(hostIf.Name, false)
	if err != nil {
		return err
	}

	// Connect the external interface to the bridge.
	log.Printf("[net] Setting link %v master %v.", hostIf.Name, bridgeName)
	if err := networkClient.SetBridgeMasterToHostInterface(); err != nil {
		return err
	}

	// External interface up.
	log.Printf("[net] Setting link %v state up.", hostIf.Name)
	err = netlink.SetLinkState(hostIf.Name, true)
	if err != nil {
		return err
	}

	// Bridge up.
	log.Printf("[net] Setting link %v state up.", bridgeName)
	err = netlink.SetLinkState(bridgeName, true)
	if err != nil {
		return err
	}

	// Add the bridge rules.
	err = networkClient.AddL2Rules(extIf)
	if err != nil {
		return err
	}

	// External interface hairpin on.
	log.Printf("[net] Setting link %v hairpin on.", hostIf.Name)
	if err := networkClient.SetHairpinOnHostInterface(true); err != nil {
		return err
	}

	// Apply IP configuration to the bridge for host traffic.
	err = nm.applyIPConfig(extIf, bridge)
	if err != nil {
		log.Printf("[net] Failed to apply interface IP configuration: %v.", err)
	}

	extIf.BridgeName = bridgeName
	err = nil

	log.Printf("[net] Connected interface %v to bridge %v.", extIf.Name, extIf.BridgeName)

	return nil
}

// DisconnectExternalInterface disconnects a host interface from its bridge.
func (nm *networkManager) disconnectExternalInterface(extIf *externalInterface, networkClient NetworkClient) {
	log.Printf("[net] Disconnecting interface %v.", extIf.Name)

	log.Printf("[net] Deleting bridge rules")
	// Delete bridge rules set on the external interface.
	networkClient.DeleteL2Rules(extIf)

	log.Printf("[net] Deleting bridge")
	// Delete Bridge
	networkClient.DeleteBridge()

	extIf.BridgeName = ""
	log.Printf("Restoring ipconfig with primary interface %v", extIf.Name)

	// Restore IP configuration.
	hostIf, _ := net.InterfaceByName(extIf.Name)
	err := nm.applyIPConfig(extIf, hostIf)
	if err != nil {
		log.Printf("[net] Failed to apply IP configuration: %v.", err)
	}

	extIf.IPAddresses = nil
	extIf.Routes = nil

	log.Printf("[net] Disconnected interface %v.", extIf.Name)
}

func getNetworkInfoImpl(nwInfo *NetworkInfo, nw *network) {
	if nw.VlanId != 0 {
		vlanMap := make(map[string]interface{})
		vlanMap[VlanIDKey] = strconv.Itoa(nw.VlanId)
		nwInfo.Options[genericData] = vlanMap
	}
}

// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package network

import (
	"sync"
	"time"

	"github.com/Azure/azure-container-networking/common"
	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/platform"
	"github.com/Azure/azure-container-networking/store"
)

const (
	// Network store key.
	storeKey  = "Network"
	VlanIDKey = "VlanID"
)

type NetworkClient interface {
	CreateBridge() error
	DeleteBridge() error
	AddL2Rules(extIf *externalInterface) error
	DeleteL2Rules(extIf *externalInterface)
	SetBridgeMasterToHostInterface() error
	SetHairpinOnHostInterface(bool) error
}

type EndpointClient interface {
	AddEndpoints(epInfo *EndpointInfo) error
	AddEndpointRules(epInfo *EndpointInfo) error
	DeleteEndpointRules(ep *endpoint)
	MoveEndpointsToContainerNS(epInfo *EndpointInfo, nsID uintptr) error
	SetupContainerInterfaces(epInfo *EndpointInfo) error
	ConfigureContainerInterfacesAndRoutes(epInfo *EndpointInfo) error
	DeleteEndpoints(ep *endpoint) error
}

// NetworkManager manages the set of container networking resources.
type networkManager struct {
	Version            string
	TimeStamp          time.Time
	ExternalInterfaces map[string]*externalInterface
	store              store.KeyValueStore
	sync.Mutex
}

// NetworkManager API.
type NetworkManager interface {
	Initialize(config *common.PluginConfig) error
	Uninitialize()

	AddExternalInterface(ifName string, subnet string) error

	CreateNetwork(nwInfo *NetworkInfo) error
	DeleteNetwork(networkId string) error
	GetNetworkInfo(networkId string) (*NetworkInfo, error)

	CreateEndpoint(networkId string, epInfo *EndpointInfo) error
	DeleteEndpoint(networkId string, endpointId string) error
	GetEndpointInfo(networkId string, endpointId string) (*EndpointInfo, error)
	AttachEndpoint(networkId string, endpointId string, sandboxKey string) (*endpoint, error)
	DetachEndpoint(networkId string, endpointId string) error
}

// Creates a new network manager.
func NewNetworkManager() (NetworkManager, error) {
	nm := &networkManager{
		ExternalInterfaces: make(map[string]*externalInterface),
	}

	return nm, nil
}

// Initialize configures network manager.
func (nm *networkManager) Initialize(config *common.PluginConfig) error {
	nm.Version = config.Version
	nm.store = config.Store

	// Restore persisted state.
	err := nm.restore()
	return err
}

// Uninitialize cleans up network manager.
func (nm *networkManager) Uninitialize() {
}

// Restore reads network manager state from persistent store.
func (nm *networkManager) restore() error {
	// Skip if a store is not provided.
	if nm.store == nil {
		log.Printf("[net] network store is nil")
		return nil
	}

	rebooted := false
	// After a reboot, all address resources are implicitly released.
	// Ignore the persisted state if it is older than the last reboot time.

	// Read any persisted state.
	err := nm.store.Read(storeKey, nm)
	if err != nil {
		if err == store.ErrKeyNotFound {
			log.Printf("[net] network store key not found")
			// Considered successful.
			return nil
		} else {
			log.Printf("[net] Failed to restore state, err:%v\n", err)
			return err
		}
	}

	modTime, err := nm.store.GetModificationTime()
	if err == nil {
		rebootTime, err := platform.GetLastRebootTime()
		log.Printf("[net] reboot time %v store mod time %v", rebootTime, modTime)
		if err == nil && rebootTime.After(modTime) {
			rebooted = true
		}
	}

	// Populate pointers.
	for _, extIf := range nm.ExternalInterfaces {
		for _, nw := range extIf.Networks {
			nw.extIf = extIf
		}
	}

	// if rebooted recreate the network that existed before reboot.
	if rebooted {
		log.Printf("[net] Rehydrating network state from persistent store")
		for _, extIf := range nm.ExternalInterfaces {
			for _, nw := range extIf.Networks {
				nwInfo, err := nm.GetNetworkInfo(nw.Id)
				if err != nil {
					log.Printf("[net] Failed to fetch network info for network %v extif %v err %v. This should not happen", nw, extIf, err)
					return err
				}

				extIf.BridgeName = ""

				_, err = nm.newNetworkImpl(nwInfo, extIf)
				if err != nil {
					log.Printf("[net] Restoring network failed for nwInfo %v extif %v. This should not happen %v", nwInfo, extIf, err)
					return err
				}
			}
		}
	}

	log.Printf("[net] Restored state, %+v\n", nm)
	return nil
}

// Save writes network manager state to persistent store.
func (nm *networkManager) save() error {
	// Skip if a store is not provided.
	if nm.store == nil {
		return nil
	}

	// Update time stamp.
	nm.TimeStamp = time.Now()

	err := nm.store.Write(storeKey, nm)
	if err == nil {
		log.Printf("[net] Save succeeded.\n")
	} else {
		log.Printf("[net] Save failed, err:%v\n", err)
	}
	return err
}

//
// NetworkManager API
//
// Provides atomic stateful wrappers around core networking functionality.
//

// AddExternalInterface adds a host interface to the list of available external interfaces.
func (nm *networkManager) AddExternalInterface(ifName string, subnet string) error {
	nm.Lock()
	defer nm.Unlock()

	err := nm.newExternalInterface(ifName, subnet)
	if err != nil {
		return err
	}

	err = nm.save()
	if err != nil {
		return err
	}

	return nil
}

// CreateNetwork creates a new container network.
func (nm *networkManager) CreateNetwork(nwInfo *NetworkInfo) error {
	nm.Lock()
	defer nm.Unlock()

	_, err := nm.newNetwork(nwInfo)
	if err != nil {
		return err
	}

	err = nm.save()
	if err != nil {
		return err
	}

	return nil
}

// DeleteNetwork deletes an existing container network.
func (nm *networkManager) DeleteNetwork(networkId string) error {
	nm.Lock()
	defer nm.Unlock()

	err := nm.deleteNetwork(networkId)
	if err != nil {
		return err
	}

	err = nm.save()
	if err != nil {
		return err
	}

	return nil
}

// GetNetworkInfo returns information about the given network.
func (nm *networkManager) GetNetworkInfo(networkId string) (*NetworkInfo, error) {
	nm.Lock()
	defer nm.Unlock()

	nw, err := nm.getNetwork(networkId)
	if err != nil {
		return nil, err
	}

	nwInfo := &NetworkInfo{
		Id:      networkId,
		Subnets: nw.Subnets,
		Mode:    nw.Mode,
		Options: make(map[string]interface{}),
	}

	getNetworkInfoImpl(nwInfo, nw)

	if nw.extIf != nil {
		nwInfo.BridgeName = nw.extIf.BridgeName
	}

	return nwInfo, nil
}

// CreateEndpoint creates a new container endpoint.
func (nm *networkManager) CreateEndpoint(networkId string, epInfo *EndpointInfo) error {
	nm.Lock()
	defer nm.Unlock()

	nw, err := nm.getNetwork(networkId)
	if err != nil {
		return err
	}

	if nw.VlanId != 0 {
		if epInfo.Data[VlanIDKey] == nil {
			log.Printf("overriding endpoint vlanid with network vlanid")
			epInfo.Data[VlanIDKey] = nw.VlanId
		}
	}

	_, err = nw.newEndpoint(epInfo)
	if err != nil {
		return err
	}

	err = nm.save()
	if err != nil {
		return err
	}

	return nil
}

// DeleteEndpoint deletes an existing container endpoint.
func (nm *networkManager) DeleteEndpoint(networkId string, endpointId string) error {
	nm.Lock()
	defer nm.Unlock()

	nw, err := nm.getNetwork(networkId)
	if err != nil {
		return err
	}

	err = nw.deleteEndpoint(endpointId)
	if err != nil {
		return err
	}

	err = nm.save()
	if err != nil {
		return err
	}

	return nil
}

// GetEndpointInfo returns information about the given endpoint.
func (nm *networkManager) GetEndpointInfo(networkId string, endpointId string) (*EndpointInfo, error) {
	nm.Lock()
	defer nm.Unlock()

	nw, err := nm.getNetwork(networkId)
	if err != nil {
		return nil, err
	}

	ep, err := nw.getEndpoint(endpointId)
	if err != nil {
		return nil, err
	}

	return ep.getInfo(), nil
}

// AttachEndpoint attaches an endpoint to a sandbox.
func (nm *networkManager) AttachEndpoint(networkId string, endpointId string, sandboxKey string) (*endpoint, error) {
	nm.Lock()
	defer nm.Unlock()

	nw, err := nm.getNetwork(networkId)
	if err != nil {
		return nil, err
	}

	ep, err := nw.getEndpoint(endpointId)
	if err != nil {
		return nil, err
	}

	err = ep.attach(sandboxKey)
	if err != nil {
		return nil, err
	}

	err = nm.save()
	if err != nil {
		return nil, err
	}

	return ep, nil
}

// DetachEndpoint detaches an endpoint from its sandbox.
func (nm *networkManager) DetachEndpoint(networkId string, endpointId string) error {
	nm.Lock()
	defer nm.Unlock()

	nw, err := nm.getNetwork(networkId)
	if err != nil {
		return err
	}

	ep, err := nw.getEndpoint(endpointId)
	if err != nil {
		return err
	}

	err = ep.detach()
	if err != nil {
		return err
	}

	err = nm.save()
	if err != nil {
		return err
	}

	return nil
}

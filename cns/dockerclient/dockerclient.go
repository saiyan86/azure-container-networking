// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package dockerclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Azure/azure-container-networking/cns/common"
	"github.com/Azure/azure-container-networking/cns/imdsclient"
	"github.com/Azure/azure-container-networking/log"
)

const (
	defaultDockerConnectionURL = "http://127.0.0.1:2375"
	defaultIpamPlugin          = "azure-vnet"
	networkMode                = "com.microsoft.azure.network.mode"
	bridgeMode                 = "bridge"
)

// DockerClient specifies a client to connect to docker.
type DockerClient struct {
	connectionURL string
	imdsClient    *imdsclient.ImdsClient
}

// NewDockerClient create a new docker client.
func NewDockerClient(url string) (*DockerClient, error) {
	return &DockerClient{
		connectionURL: url,
		imdsClient:    &imdsclient.ImdsClient{},
	}, nil
}

// NewDefaultDockerClient create a new docker client.
func NewDefaultDockerClient(imdsClient *imdsclient.ImdsClient) (*DockerClient, error) {
	return &DockerClient{
		connectionURL: defaultDockerConnectionURL,
		imdsClient:    imdsClient,
	}, nil
}

// NetworkExists tries to retrieve a network from docker (if it exists).
func (dockerClient *DockerClient) NetworkExists(networkName string) error {
	log.Printf("[Azure CNS] NetworkExists")

	res, err := http.Get(
		dockerClient.connectionURL + inspectNetworkPath + networkName)

	if err != nil {
		log.Printf("[Azure CNS] Error received from http Post for docker network inspect %v %v", networkName, err.Error())
		return err
	}

	defer res.Body.Close()

	// network exists
	if res.StatusCode == 200 {
		log.Debugf("[Azure CNS] Network with name %v already exists. Docker return code: %v", networkName, res.StatusCode)
		return nil
	}

	// network not found
	if res.StatusCode == 404 {
		log.Debugf("[Azure CNS] Network with name %v does not exist. Docker return code: %v", networkName, res.StatusCode)
		return fmt.Errorf("Network not found")
	}

	return fmt.Errorf("Unknown return code from docker inspect %d", res.StatusCode)
}

// CreateNetwork creates a network using docker network create.
func (dockerClient *DockerClient) CreateNetwork(networkName string, nicInfo *imdsclient.InterfaceInfo, options map[string]interface{}) error {
	log.Printf("[Azure CNS] CreateNetwork")

	enableSnat := true

	config := &Config{
		Subnet: nicInfo.Subnet,
	}

	configs := make([]Config, 1)
	configs[0] = *config
	ipamConfig := &IPAM{
		Driver: defaultIpamPlugin,
		Config: configs,
	}

	netConfig := &NetworkConfiguration{
		Name:     networkName,
		Driver:   defaultNetworkPlugin,
		IPAM:     *ipamConfig,
		Internal: true,
	}

	if options != nil {
		if _, ok := options[OptDisableSnat]; ok {
			enableSnat = false
		}
	}

	if enableSnat {
		netConfig.Options = make(map[string]interface{})
		netConfig.Options[networkMode] = bridgeMode
	}

	log.Printf("[Azure CNS] Going to create network with config: %+v", netConfig)

	netConfigJSON := new(bytes.Buffer)
	err := json.NewEncoder(netConfigJSON).Encode(netConfig)
	if err != nil {
		return err
	}

	res, err := http.Post(
		dockerClient.connectionURL+createNetworkPath,
		"application/json; charset=utf-8",
		netConfigJSON)

	if err != nil {
		log.Printf("[Azure CNS] Error received from http Post for docker network create %v", networkName)
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != 201 {
		var createNetworkResponse DockerErrorResponse
		err = json.NewDecoder(res.Body).Decode(&createNetworkResponse)
		var ermsg string
		ermsg = ""
		if err != nil {
			ermsg = err.Error()
		}
		return fmt.Errorf("[Azure CNS] Create docker network failed with error code %v - %v - %v",
			res.StatusCode, createNetworkResponse.message, ermsg)
	}

	if enableSnat {
		err = common.SetOutboundSNAT(nicInfo.Subnet)
		if err != nil {
			log.Printf("[Azure CNS] Error setting up SNAT outbound rule %v", err)
		}
	}

	return nil
}

// DeleteNetwork creates a network using docker network create.
func (dockerClient *DockerClient) DeleteNetwork(networkName string) error {
	log.Printf("[Azure CNS] DeleteNetwork")

	url := dockerClient.connectionURL + inspectNetworkPath + networkName
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("[Azure CNS] Error received while creating http DELETE request for network delete %v %v", networkName, err.Error())
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("[Azure CNS] HTTP Post returned error %v", err.Error())
		return err
	}

	defer res.Body.Close()

	// network successfully deleted.
	if res.StatusCode == 204 {
		primaryNic, err := dockerClient.imdsClient.GetPrimaryInterfaceInfoFromHost()
		if err != nil {
			return err
		}

		cmd := fmt.Sprintf("iptables -t nat -D POSTROUTING -m iprange ! --dst-range 168.63.129.16 -m addrtype ! --dst-type local ! -d %v -j MASQUERADE",
			primaryNic.Subnet)
		err = common.ExecuteShellCommand(cmd)
		if err != nil {
			log.Printf("[Azure CNS] Error Removing Outbound SNAT rule %v", err)
		}

		return nil
	}

	// network not found.
	if res.StatusCode == 404 {
		return fmt.Errorf("[Azure CNS] Network not found %v", networkName)
	}

	return fmt.Errorf("[Azure CNS] Unknown return code from docker delete network %v: ret = %d",
		networkName, res.StatusCode)
}

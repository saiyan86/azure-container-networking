package ovsctl

import (
	"fmt"
	"net"
	"strings"

	"github.com/Azure/azure-container-networking/common"
	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/platform"
)

const (
	defaultMacForArpResponse = "12:34:56:78:9a:bc"
)

func CreateOVSBridge(bridgeName string) error {
	log.Printf("[ovs] Creating OVS Bridge %v", bridgeName)

	ovsCreateCmd := fmt.Sprintf("ovs-vsctl add-br %s", bridgeName)
	_, err := platform.ExecuteCommand(ovsCreateCmd)
	if err != nil {
		log.Printf("[ovs] Error while creating OVS bridge %v", err)
		return err
	}

	return nil
}

func DeleteOVSBridge(bridgeName string) error {
	log.Printf("[ovs] Deleting OVS Bridge %v", bridgeName)

	ovsCreateCmd := fmt.Sprintf("ovs-vsctl del-br %s", bridgeName)
	_, err := platform.ExecuteCommand(ovsCreateCmd)
	if err != nil {
		log.Printf("[ovs] Error while deleting OVS bridge %v", err)
		return err
	}

	return nil
}

func AddPortOnOVSBridge(hostIfName string, bridgeName string, vlanID int) error {
	cmd := ""

	if vlanID == 0 {
		cmd = fmt.Sprintf("ovs-vsctl add-port %s %s", bridgeName, hostIfName)
	} else {
		cmd = fmt.Sprintf("ovs-vsctl add-port %s %s tag=%d", bridgeName, hostIfName, vlanID)
	}
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Error while setting OVS as master to primary interface %v", err)
		return err
	}

	return nil
}

func GetOVSPortNumber(interfaceName string) (string, error) {
	cmd := fmt.Sprintf("ovs-vsctl get Interface %s ofport", interfaceName)
	ofport, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Get ofport failed with error %v", err)
		return "", err
	}

	return strings.Trim(ofport, "\n"), nil
}

func AddVMIpAcceptRule(bridgeName string, primaryIP string, mac string) error {
	cmd := fmt.Sprintf("ovs-ofctl add-flow %s ip,nw_dst=%s,dl_dst=%s,priority=20,actions=normal", bridgeName, primaryIP, mac)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Adding SNAT rule failed with error %v", err)
		return err
	}

	return nil
}

func AddArpSnatRule(bridgeName string, mac string, macHex string, ofport string) error {
	cmd := fmt.Sprintf(`ovs-ofctl add-flow %v table=1,priority=10,arp,arp_op=1,actions='mod_dl_src:%s,
		load:0x%s->NXM_NX_ARP_SHA[],output:%s'`, bridgeName, mac, macHex, ofport)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Adding ARP SNAT rule failed with error %v", err)
		return err
	}

	return nil
}

func AddIpSnatRule(bridgeName string, port string, mac string) error {
	cmd := fmt.Sprintf("ovs-ofctl add-flow %v priority=20,ip,in_port=%s,vlan_tci=0,actions=mod_dl_src:%s,strip_vlan,normal",
		bridgeName, port, mac)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Adding IP SNAT rule failed with error %v", err)
		return err
	}

	cmd = fmt.Sprintf("ovs-ofctl add-flow %v priority=10,ip,in_port=%s,actions=drop",
		bridgeName, port)
	_, err = platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Dropping vlantag packet rule failed with error %v", err)
		return err
	}

	return nil
}

func AddArpDnatRule(bridgeName string, port string, mac string) error {
	// Add DNAT rule to forward ARP replies to container interfaces.
	cmd := fmt.Sprintf(`ovs-ofctl add-flow %s arp,arp_op=2,in_port=%s,actions='mod_dl_dst:ff:ff:ff:ff:ff:ff,
		load:0x%s->NXM_NX_ARP_THA[],normal'`, bridgeName, port, mac)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Adding DNAT rule failed with error %v", err)
		return err
	}

	return nil
}

func AddFakeArpReply(bridgeName string, ip net.IP) error {
	// If arp fields matches, set arp reply rule for the request
	macAddrHex := strings.Replace(defaultMacForArpResponse, ":", "", -1)
	ipAddrInt := common.IpToInt(ip)

	log.Printf("[ovs] Adding ARP reply rule for IP address %v ", ip.String())
	cmd := fmt.Sprintf(`ovs-ofctl add-flow %s arp,arp_op=1,priority=20,actions='load:0x2->NXM_OF_ARP_OP[],
			move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:%s,
			move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],move:NXM_OF_ARP_TPA[]->NXM_OF_ARP_SPA[],
			load:0x%s->NXM_NX_ARP_SHA[],load:0x%x->NXM_OF_ARP_TPA[],IN_PORT'`,
		bridgeName, defaultMacForArpResponse, macAddrHex, ipAddrInt)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Adding ARP reply rule failed with error %v", err)
		return err
	}

	return nil
}

func AddArpReplyRule(bridgeName string, port string, ip net.IP, mac string, vlanid int, mode string) error {
	ipAddrInt := common.IpToInt(ip)
	macAddrHex := strings.Replace(mac, ":", "", -1)

	log.Printf("[ovs] Adding ARP reply rule to add vlan %v and forward packet to table 1 for port %v", vlanid, port)
	cmd := fmt.Sprintf(`ovs-ofctl add-flow %s arp,arp_op=1,in_port=%s,actions='mod_vlan_vid:%v,resubmit(,1)'`,
		bridgeName, port, vlanid)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Adding ARP reply rule failed with error %v", err)
		return err
	}

	// If arp fields matches, set arp reply rule for the request
	log.Printf("[ovs] Adding ARP reply rule for IP address %v and vlanid %v.", ip, vlanid)
	cmd = fmt.Sprintf(`ovs-ofctl add-flow %s table=1,arp,arp_tpa=%s,dl_vlan=%v,arp_op=1,priority=20,actions='load:0x2->NXM_OF_ARP_OP[],
			move:NXM_OF_ETH_SRC[]->NXM_OF_ETH_DST[],mod_dl_src:%s,
			move:NXM_NX_ARP_SHA[]->NXM_NX_ARP_THA[],move:NXM_OF_ARP_SPA[]->NXM_OF_ARP_TPA[],
			load:0x%s->NXM_NX_ARP_SHA[],load:0x%x->NXM_OF_ARP_SPA[],strip_vlan,IN_PORT'`,
		bridgeName, ip.String(), vlanid, mac, macAddrHex, ipAddrInt)
	_, err = platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Adding ARP reply rule failed with error %v", err)
		return err
	}

	return nil
}

func AddMacDnatRule(bridgeName string, port string, ip net.IP, mac string, vlanid int) error {
	cmd := fmt.Sprintf("ovs-ofctl add-flow %s ip,nw_dst=%s,dl_vlan=%v,in_port=%s,actions=mod_dl_dst:%s,normal",
		bridgeName, ip.String(), vlanid, port, mac)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Adding MAC DNAT rule failed with error %v", err)
		return err
	}

	return nil
}

func DeleteArpReplyRule(bridgeName string, port string, ip net.IP, vlanid int) {
	cmd := fmt.Sprintf("ovs-ofctl del-flows %s arp,arp_op=1,in_port=%s",
		bridgeName, port)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[net] Deleting ARP reply rule failed with error %v", err)
	}

	cmd = fmt.Sprintf("ovs-ofctl del-flows %s table=1,arp,arp_tpa=%s,dl_vlan=%v,arp_op=1",
		bridgeName, ip.String(), vlanid)
	_, err = platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[net] Deleting ARP reply rule failed with error %v", err)
	}
}

func DeleteIPSnatRule(bridgeName string, port string) {
	cmd := fmt.Sprintf("ovs-ofctl del-flows %v ip,in_port=%s",
		bridgeName, port)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("Error while deleting ovs rule %v error %v", cmd, err)
	}
}

func DeleteMacDnatRule(bridgeName string, port string, ip net.IP, vlanid int) {
	cmd := fmt.Sprintf("ovs-ofctl del-flows %s ip,nw_dst=%s,dl_vlan=%v,in_port=%s",
		bridgeName, ip.String(), vlanid, port)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[net] Deleting MAC DNAT rule failed with error %v", err)
	}
}

func DeletePortFromOVS(bridgeName string, interfaceName string) error {
	// Disconnect external interface from its bridge.
	cmd := fmt.Sprintf("ovs-vsctl del-port %s %s", bridgeName, interfaceName)
	_, err := platform.ExecuteCommand(cmd)
	if err != nil {
		log.Printf("[ovs] Failed to disconnect interface %v from bridge, err:%v.", interfaceName, err)
		return err
	}

	return nil
}

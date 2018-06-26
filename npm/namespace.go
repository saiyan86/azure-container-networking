package npm

import (
	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/npm/ipsm"
	"github.com/Azure/azure-container-networking/npm/iptm"
	"github.com/Azure/azure-container-networking/npm/util"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

type namespace struct {
	name   string
	setMap map[string]string
	podMap map[types.UID]*corev1.Pod
	npMap  map[string]*networkingv1.NetworkPolicy // TODO: Optimize to ordered map.
	ipsMgr *ipsm.IpsetManager
	iptMgr *iptm.IptablesManager
}

// newNS constructs a new namespace object.
func newNs(name string) (*namespace, error) {
	ns := &namespace{
		name:   name,
		setMap: make(map[string]string),
		podMap: make(map[types.UID]*corev1.Pod),
		npMap:  make(map[string]*networkingv1.NetworkPolicy),
		ipsMgr: ipsm.NewIpsetManager(),
		iptMgr: iptm.NewIptablesManager(),
	}

	return ns, nil
}

func isSystemNs(nsObj *corev1.Namespace) bool {
	return nsObj.ObjectMeta.Name == util.KubeSystemFlag
}

func getNsIpsetName(k, v string) string {
	return "ns-" + k + ":" + v
}

// InitAllNsList syncs all-namespace ipset list.
func (npMgr *NetworkPolicyManager) InitAllNsList() error {
	allNs := npMgr.nsMap[util.KubeAllNamespacesFlag]
	for nsName := range npMgr.nsMap {
		if nsName == util.KubeAllNamespacesFlag {
			continue
		}

		if err := allNs.ipsMgr.AddToList(util.KubeAllNamespacesFlag, nsName); err != nil {
			log.Printf("Error adding namespace set %s to list %s\n", nsName, util.KubeAllNamespacesFlag)
			return err
		}
	}

	return nil
}

// UninitAllNsList cleans all-namespace ipset list.
func (npMgr *NetworkPolicyManager) UninitAllNsList() error {
	allNs := npMgr.nsMap[util.KubeAllNamespacesFlag]
	for nsName := range npMgr.nsMap {
		if nsName == util.KubeAllNamespacesFlag {
			continue
		}

		if err := allNs.ipsMgr.DeleteFromList(util.KubeAllNamespacesFlag, nsName); err != nil {
			log.Printf("Error deleting namespace set %s from list %s\n", nsName, util.KubeAllNamespacesFlag)
			return err
		}
	}

	return nil
}

// AddNamespace handles add namespace.
func (npMgr *NetworkPolicyManager) AddNamespace(nsObj *corev1.Namespace) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	nsName, nsNs := nsObj.ObjectMeta.Name, nsObj.ObjectMeta.Namespace
	log.Printf("NAMESPACE CREATED: %s/%s\n", nsName, nsNs)

	allNs := npMgr.nsMap[util.KubeAllNamespacesFlag]
	// Create ipset for the namespace.
	ipsMgr := allNs.ipsMgr
	if err := ipsMgr.CreateSet(nsName); err != nil {
		log.Printf("Error creating ipset for namespace %s.\n", nsName)
		return err
	}

	if err := ipsMgr.AddToList(util.KubeAllNamespacesFlag, nsName); err != nil {
		log.Printf("Error adding %s to all-namespace ipset list.\n", nsName)
		return err
	}

	// Add the namespace to its label's ipset list.
	var labelKeys []string
	nsLabels := nsObj.ObjectMeta.Labels
	for nsLabelKey, nsLabelVal := range nsLabels {
		labelKey := getNsIpsetName(nsLabelKey, nsLabelVal)
		log.Printf("Adding namespace %s to ipset list %s\n", nsName, labelKey)
		if err := ipsMgr.AddToList(labelKey, nsName); err != nil {
			log.Printf("Error Adding namespace %s to ipset list %s\n", nsName, labelKey)
			return err
		}
		labelKeys = append(labelKeys, labelKey)
	}

	ns, err := newNs(nsName)
	if err != nil {
		log.Printf("Error creating namespace %s\n", nsName)
	}
	npMgr.nsMap[nsName] = ns

	return nil
}

// UpdateNamespace handles update namespace.
func (npMgr *NetworkPolicyManager) UpdateNamespace(oldNsObj *corev1.Namespace, newNsObj *corev1.Namespace) error {
	oldNsName, newNsName := oldNsObj.ObjectMeta.Name, newNsObj.ObjectMeta.Name
	log.Printf("NAMESPACE UPDATED. %s/%s", oldNsName, newNsName)

	npMgr.DeleteNamespace(oldNsObj)

	if newNsObj.ObjectMeta.DeletionTimestamp == nil && newNsObj.ObjectMeta.DeletionGracePeriodSeconds == nil {
		npMgr.AddNamespace(newNsObj)
	}

	return nil
}

// DeleteNamespace handles delete namespace.
func (npMgr *NetworkPolicyManager) DeleteNamespace(nsObj *corev1.Namespace) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	nsName, nsNs := nsObj.ObjectMeta.Name, nsObj.ObjectMeta.Namespace
	log.Printf("NAMESPACE DELETED: %s/%s\n", nsName, nsNs)

	_, exists := npMgr.nsMap[nsName]
	if !exists {
		return nil
	}

	// Delete the namespace from its label's ipset list.
	ipsMgr := npMgr.nsMap[util.KubeAllNamespacesFlag].ipsMgr
	var labelKeys []string
	nsLabels := nsObj.ObjectMeta.Labels
	for nsLabelKey, nsLabelVal := range nsLabels {
		labelKey := getNsIpsetName(nsLabelKey, nsLabelVal)
		log.Printf("Deleting namespace %s from ipset list %s\n", nsName, labelKey)
		if err := ipsMgr.DeleteFromList(labelKey, nsName); err != nil {
			log.Printf("Error deleting namespace %s from ipset list %s\n", nsName, labelKey)
			return err
		}
		labelKeys = append(labelKeys, labelKey)
	}

	// Delete the namespace from all-namespace ipset list.
	if err := ipsMgr.DeleteFromList(util.KubeAllNamespacesFlag, nsName); err != nil {
		log.Printf("Error deleting namespace %s from ipset list %s\n", nsName, util.KubeAllNamespacesFlag)
		return err
	}

	// Delete ipset for the namespace.
	if err := ipsMgr.DeleteSet(nsName); err != nil {
		log.Printf("Error deleting ipset for namespace %s.\n", nsName)
		return err
	}

	delete(npMgr.nsMap, nsName)

	return nil
}

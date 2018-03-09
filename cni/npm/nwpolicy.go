package npm

import (
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
)

/*
type npMgr interface {
	AddNetworkPolicy(obj *networkingv1.NetworkPolicy) error
	UpdateNetworkPolicy(old *networkingv1.NetworkPolicy, new *networkingv1.NetworkPolicy) error
	DeleteNetworkPolicy(obj *networkingv1.NetworkPolicy) error
}
*/

/*
Create:
(*v1.NetworkPolicy)(0xc4202f4d80)(&NetworkPolicyNetworkPolicy{
   ObjectMeta:k8s_io_apimachinery_pkg_apis_meta_v1.Obje
ctMeta   {
      Name:allow-tcp-443,
      GenerateName:,
      Namespace:default,
      SelfLink:/apis/networking.k8s.io/v1/namespaces/default/networkpolicies/allow-tcp-443,
      UID:477      a3c8f-219e-11e8-812d-000d3afd5b25,
      ResourceVersion:16      37785,
      Generation:1,
      CreationTimestamp:2018-03      -07      00:27:12      +0000 UTC,
      DeletionTimestamp:<nil>,
      DeletionGracePeriodSeconds:nil,
      Labels:map      [
         string
      ]      string      {

      },
      Annotations:map      [
         string
      ]      string      {
         kubectl.kubernetes.io/l
ast-applied-configuration:{
            "apiVersion":"networking.k8s.io/v1",
            "kind":"NetworkPolicy",
            "metadata":{
               "annotations":{

               },
               "name":"allow-tcp-443",
               "namespace":"default"
            },
            "spec":{
               "ingress":[
                  {
                     "from":[
                        {
                           "podSelector":{
                              "matchLabels":{
                                 "app":"web"
                              }
                           }
                        }
                     ],
                     "ports":[
                        {
                           "port":443,
                           "protocol":"TCP"
                        }
                     ]
                  }
               ],
               "podSelector":{
                  "matchLab
els":{
                     "app":"web"
                  }
               }
            }
         },

      },
      OwnerReferences:[

      ],
      Finalizers:[

      ],
      ClusterName:,
      Initializers:nil,

   },
   Spec:NetworkPolicySpec   {
      PodSelecto
r:k8s_io_apimachinery_pkg_apis_meta_v1.LabelSelector      {
         MatchLabels:map         [
            string
         ]         string         {
            app:web,

         },
         MatchEx
pressions:[

         ],

      },
      Ingress:[
         {
            [
               {
                  0                  xc42025e170 443
               }
            ]            [
               {
                  LabelSelector                  {
                     MatchLabels:map                     [
                        string
                     ]                     string                     {
                        app:web,

                     },
                     MatchExpressions:[

                     ],

                  }                  nil nil
               }
            ]
         }
      ],
      Egress:[

      ],
      PolicyTypes:[

      ],

   },

})
*/
// AddNetworkPolicy adds network policy.
func (npMgr *NetworkPolicyManager) AddNetworkPolicy(np *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName := np.ObjectMeta.Namespace, np.ObjectMeta.Name
	fmt.Printf("NETWORK POLICY CREATED: %s/%s\n", npNs, npName)

	ns, exists := npMgr.nsMap[npNs]
	if !exists {
		newns, err := newNs(npNs)
		if err != nil {
			return err
		}
		npMgr.nsMap[npNs] = newns
		ns = newns
		fmt.Printf("new namespace created: %s\n", npNs)
	}

	// debug
	for k, v := range ns.podMap {
		fmt.Printf("key[%s] value[%s]\n", k, v)
	}

	return nil
}

// UpdateNetworkPolicy updates network policy.
func (npMgr *NetworkPolicyManager) UpdateNetworkPolicy(oldNp *networkingv1.NetworkPolicy, newNp *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	oldNpNs, oldNpName := oldNp.ObjectMeta.Namespace, oldNp.ObjectMeta.Name
	fmt.Printf("NETWORK POLICY UPDATED: %s/%s\n", oldNpNs, oldNpName)

	return nil
}

// DeleteNetworkPolicy deletes network policy.
func (npMgr *NetworkPolicyManager) DeleteNetworkPolicy(np *networkingv1.NetworkPolicy) error {
	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName := np.ObjectMeta.Namespace, np.ObjectMeta.Name
	fmt.Printf("NETWORK POLICY DELETED: %s/%s\n", npNs, npName)

	return nil
}

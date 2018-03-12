package npm

import (
	"fmt"
	"time"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	time.Sleep(500 * time.Millisecond)

	npMgr.Lock()
	defer npMgr.Unlock()

	npNs, npName := np.ObjectMeta.Namespace, np.ObjectMeta.Name
	fmt.Printf("NETWORK POLICY CREATED: %s/%s\n", npNs, npName)

	selector := np.Spec.PodSelector
	fmt.Printf("podSelector:%+v\n", selector)

	clientset := npMgr.clientset
	podList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: "app=nginx",
	})
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		fmt.Printf("%s/%s/%+v", pod.ObjectMeta.Name, pod.ObjectMeta.Namespace, pod.ObjectMeta.Labels)
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

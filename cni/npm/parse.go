package npm

import (
	"github.com/davecgh/go-spew/spew"

	networkingv1 "k8s.io/api/networking/v1"
)

type policyInfo struct {
	name string
}

// ParsePolicy parses network policy.
func ParsePolicy(np *networkingv1.NetworkPolicy) {
	spew.Dump(np)
}

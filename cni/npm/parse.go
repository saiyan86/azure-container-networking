package npm

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"

	networkingv1 "k8s.io/api/networking/v1"
)

type policyInfo struct {
	name string
}

// ParsePolicy parses network policy.
func ParsePolicy(np *networkingv1.NetworkPolicy) {
	spew.Dump(np)
	b, err := json.MarshalIndent(np, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)
}

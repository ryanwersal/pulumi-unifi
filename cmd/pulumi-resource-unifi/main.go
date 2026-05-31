// Command pulumi-resource-unifi is the Pulumi plugin binary for the UniFi
// provider. Pulumi discovers plugins by the `pulumi-resource-<name>` naming
// convention, so this directory name is load-bearing.
package main

import (
	"context"
	"log"

	p "github.com/pulumi/pulumi-go-provider"

	"github.com/ryanwersal/pulumi-unifi/provider"
)

func main() {
	prov, err := provider.New()
	if err != nil {
		log.Fatalf("failed to construct provider: %v", err)
	}

	if err := p.RunProvider(context.Background(), provider.Name, provider.Version, prov); err != nil {
		log.Fatalf("provider exited with error: %v", err)
	}
}

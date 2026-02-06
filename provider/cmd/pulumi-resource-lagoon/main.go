package main

import (
	"fmt"
	"os"

	p "github.com/pulumi/pulumi-go-provider"
	lagoon "github.com/tag1consulting/pulumi-lagoon/provider/pkg/provider"
)

// Version is set at build time via ldflags.
var Version = "0.2.0-dev"

func main() {
	provider := lagoon.NewProvider(Version)
	if err := p.RunProvider("lagoon", Version, provider); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

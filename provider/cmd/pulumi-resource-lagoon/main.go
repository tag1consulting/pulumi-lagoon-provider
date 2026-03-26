package main

import (
	"context"
	"fmt"
	"os"

	lagoon "github.com/tag1consulting/pulumi-lagoon/provider/pkg/provider"
)

// Version is set at build time via ldflags.
var Version = "0.2.6"

func main() {
	provider, err := lagoon.NewProvider(Version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := provider.Run(context.Background(), "lagoon", Version); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

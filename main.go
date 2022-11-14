package main

import (
	"context"
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/scalr/terraform-provider-scalr/scalr"
	"log"
	"os"
)

const (
	scalrProviderName = "registry.scalr.io/scalr/scalr"
)

func main() {
	var debugMode bool
	ctx := context.Background()

	flag.BoolVar(&debugMode, "debug", false, "Start provider in debug mode.")
	flag.Parse()

	if debugMode {
		err := plugin.Debug(ctx, scalrProviderName,
			&plugin.ServeOpts{
				ProviderFunc: scalr.Provider,
			})
		log.Printf("[ERROR] Could not start the debug server: %v", err)
		os.Exit(1)
	} else {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: scalr.Provider})
	}
}

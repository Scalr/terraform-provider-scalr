package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	"github.com/scalr/terraform-provider-scalr/scalr"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: scalr.Provider})
}

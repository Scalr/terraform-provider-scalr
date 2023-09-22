package main

import (
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/scalr/terraform-provider-scalr/scalr"
)

// Commands to prepare auto-generated documentation.
// - format terraform example snippets:
//go:generate terraform fmt -recursive examples
// - generate the /docs content:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --rendered-website-dir docs
// - inject proper 'order' Front Matter directives so pages are always sorted alphabetically:
//go:generate go run tools/page_order.go -dir=docs/data-sources
//go:generate go run tools/page_order.go -dir=docs/resources

const (
	scalrProviderAddr = "registry.scalr.io/scalr/scalr"
)

func main() {
	var isDebug bool
	flag.BoolVar(&isDebug, "debug", false, "Start provider in debug mode.")
	flag.Parse()

	// Remove any date and time prefix in log package function output to
	// prevent duplicate timestamp and incorrect log level setting
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	schema.DescriptionKind = schema.StringMarkdown

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: scalr.Provider,
		ProviderAddr: scalrProviderAddr,
		Debug:        isDebug,
	})
}

package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/scalr/terraform-provider-scalr/internal/provider"
	"github.com/scalr/terraform-provider-scalr/scalr"
	"github.com/scalr/terraform-provider-scalr/version"
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
	ctx := context.Background()

	var isDebug bool
	flag.BoolVar(&isDebug, "debug", false, "Start provider in debug mode.")
	flag.Parse()

	// Remove any date and time prefix in log package function output to
	// prevent duplicate timestamp and incorrect log level setting
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	schema.DescriptionKind = schema.StringMarkdown

	providers := []func() tfprotov5.ProviderServer{
		// New provider implementation with Terraform Plugin Framework
		providerserver.NewProtocol5(provider.New(version.ProviderVersion)()),
		// Classic provider implementation with Terraform Plugin SDK
		scalr.Provider().GRPCProvider,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf5server.ServeOpt
	if isDebug {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	// Serve both the classic and the new provider
	err = tf5server.Serve(
		scalrProviderAddr,
		muxServer.ProviderServer,
		serveOpts...,
	)

	if err != nil {
		log.Fatal(err)
	}
}

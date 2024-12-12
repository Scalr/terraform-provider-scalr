package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const testProviderVersion = "test"

var testAccProvider provider.Provider

func init() {
	schema.DescriptionKind = schema.StringMarkdown
	testAccProvider = New(testProviderVersion)()
}

func protoV5ProviderFactories(t *testing.T) map[string]func() (tfprotov5.ProviderServer, error) {
	ctx := context.Background()

	providers := []func() tfprotov5.ProviderServer{
		providerserver.NewProtocol5(testAccProvider),
		testAccProviderSDK.GRPCProvider,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		t.Fatal(err.Error())
	}

	return map[string]func() (tfprotov5.ProviderServer, error){
		"scalr": func() (tfprotov5.ProviderServer, error) { return muxServer.ProviderServer(), nil },
	}
}

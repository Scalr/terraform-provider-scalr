package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/scalr/go-scalr"

	scalrV2 "github.com/scalr/go-scalr/v2/scalr"
)

type ResourceWithScalrClient struct {
	Client   *scalr.Client
	ClientV2 *scalrV2.Client
}

func (r *ResourceWithScalrClient) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*Clients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Clients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.Client = c.Client
	r.ClientV2 = c.ClientV2
}

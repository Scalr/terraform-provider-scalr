package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/scalr/go-scalr"
)

type ResourceWithScalrClient struct {
	Client *scalr.Client
}

func (r *ResourceWithScalrClient) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*scalr.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *scalr.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.Client = c
}

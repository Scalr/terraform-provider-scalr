package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/scalr/go-scalr"
)

type DataSourceWithScalrClient struct {
	Client *scalr.Client
}

func (d *DataSourceWithScalrClient) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*scalr.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *scalr.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.Client = c
}

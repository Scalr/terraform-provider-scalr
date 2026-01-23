package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/scalr/go-scalr"
	scalrV2 "github.com/scalr/go-scalr/v2/scalr"
)

type DataSourceWithScalrClient struct {
	Client   *scalr.Client
	ClientV2 *scalrV2.Client
}

func (d *DataSourceWithScalrClient) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*Clients)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf(
				"Expected *Clients, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)

		return
	}

	d.Client = c.Client
	d.ClientV2 = c.ClientV2
}

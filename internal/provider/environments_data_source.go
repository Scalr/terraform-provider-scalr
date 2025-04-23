package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/defaults"
)

// Compile-time interface checks
var (
	_ datasource.DataSource              = &environmentsDataSource{}
	_ datasource.DataSourceWithConfigure = &environmentsDataSource{}
)

func newEnvironmentsDataSource() datasource.DataSource {
	return &environmentsDataSource{}
}

// environmentsDataSource defines the data source implementation.
type environmentsDataSource struct {
	framework.DataSourceWithScalrClient
}

// environmentsDataSourceModel describes the data source data model.
type environmentsDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	TagIDs    types.Set    `tfsdk:"tag_ids"`
	AccountID types.String `tfsdk:"account_id"`
	IDs       types.Set    `tfsdk:"ids"`
}

func (d *environmentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

func (d *environmentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a list of environment ids by name or tags.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of this data source.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The query used in a Scalr environment name filter.",
				Optional:            true,
			},
			"tag_ids": schema.SetAttribute{
				MarkdownDescription: "List of tag IDs associated with the environment.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Scalr account, in the format `acc-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
			},
			"ids": schema.SetAttribute{
				MarkdownDescription: "The list of environment IDs, in the format [`env-xxxxxxxxxxx`, `env-yyyyyyyyy`].",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *environmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg environmentsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var accID string
	if !cfg.AccountID.IsNull() {
		accID = cfg.AccountID.ValueString()
	} else {
		var diags diag.Diagnostics
		accID, diags = defaults.GetDefaultScalrAccountID()
		resp.Diagnostics.Append(diags...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	id := strings.Builder{} // holds the string to build a unique resource id hash
	id.WriteString(accID)
	ids := make([]string, 0)

	opts := scalr.EnvironmentListOptions{
		Filter: &scalr.EnvironmentFilter{
			Account: ptr(accID),
		},
	}

	if !cfg.Name.IsNull() {
		id.WriteString(cfg.Name.ValueString())
		opts.Filter.Name = cfg.Name.ValueStringPointer()
	}

	if !cfg.TagIDs.IsNull() {
		var tagIDs []string
		resp.Diagnostics.Append(cfg.TagIDs.ElementsAs(ctx, &tagIDs, false)...)
		if len(tagIDs) > 0 {
			for _, t := range tagIDs {
				id.WriteString(t)
			}
			opts.Filter.Tag = ptr("in:" + strings.Join(tagIDs, ","))
		}
	}

	for {
		el, err := d.Client.Environments.List(ctx, opts)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving environments", err.Error())
			return
		}

		for _, e := range el.Items {
			ids = append(ids, e.ID)
		}

		if el.CurrentPage >= el.TotalPages {
			break
		}
		opts.PageNumber = el.NextPage
	}

	cfg.Id = types.StringValue(fmt.Sprintf("%d", framework.HashString(id.String())))
	cfg.AccountID = types.StringValue(accID)

	idsValue, diags := types.SetValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	cfg.IDs = idsValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

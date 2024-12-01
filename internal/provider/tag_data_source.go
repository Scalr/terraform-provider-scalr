package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource                     = &tagDataSource{}
	_ datasource.DataSourceWithConfigure        = &tagDataSource{}
	_ datasource.DataSourceWithConfigValidators = &tagDataSource{}
)

func newTagDataSource() datasource.DataSource {
	return &tagDataSource{}
}

// tagDataSource defines the data source implementation.
type tagDataSource struct {
	framework.DataSourceWithScalrClient
}

// tagDataSourceModel describes the data source data model.
type tagDataSourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	AccountID types.String `tfsdk:"account_id"`
}

func (d *tagDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (d *tagDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a tag.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the tag in the format `tag-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the tag.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Scalr account, in the format `acc-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (d *tagDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *tagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg tagDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.TagListOptions{}
	if !cfg.Id.IsNull() {
		opts.Tag = cfg.Id.ValueStringPointer()
	}
	if !cfg.Name.IsNull() {
		opts.Name = cfg.Name.ValueStringPointer()
	}

	tags, err := d.Client.Tags.List(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving tag", err.Error())
		return
	}

	// Unlikely
	if tags.TotalCount > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving tag",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if tags.TotalCount == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving tag",
			fmt.Sprintf("Could not find tag with ID '%s', name '%s'.", cfg.Id.ValueString(), cfg.Name.ValueString()),
		)
		return
	}

	tag := tags.Items[0]

	cfg.Id = types.StringValue(tag.ID)
	cfg.Name = types.StringValue(tag.Name)
	cfg.AccountID = types.StringValue(tag.Account.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

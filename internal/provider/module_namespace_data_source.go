package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/ops/module_namespace"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

var (
	_ datasource.DataSource                     = &moduleNamespaceDataSource{}
	_ datasource.DataSourceWithConfigure        = &moduleNamespaceDataSource{}
	_ datasource.DataSourceWithConfigValidators = &moduleNamespaceDataSource{}
)

func newModuleNamespaceDataSource() datasource.DataSource {
	return &moduleNamespaceDataSource{}
}

type moduleNamespaceDataSource struct {
	framework.DataSourceWithScalrClient
}

type moduleNamespaceDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	IsShared     types.Bool   `tfsdk:"is_shared"`
	Environments types.Set    `tfsdk:"environments"`
	Modules      types.Set    `tfsdk:"modules"`
	Owners       types.Set    `tfsdk:"owners"`
}

func (r *moduleNamespaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_namespace"
}

func (r *moduleNamespaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a single module namespace.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The module namespace ID.",
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the module namespace.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"is_shared": schema.BoolAttribute{
				MarkdownDescription: "Whether the module namespace is shared.",
				Computed:            true,
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "Set of environment IDs associated with the module namespace.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"modules": schema.SetAttribute{
				MarkdownDescription: "Set of module IDs in the module namespace.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"owners": schema.SetAttribute{
				MarkdownDescription: "Set of team IDs that own the module namespace.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *moduleNamespaceDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{}
}

func (r *moduleNamespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg moduleNamespaceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := module_namespace.ListModuleNamespacesOptions{
		Filter: map[string]string{"name": cfg.Name.ValueString()},
	}

	namespaces, err := r.ClientV2.ModuleNamespace.ListModuleNamespaces(ctx, &options)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving module namespaces", err.Error())
		return
	}

	if len(namespaces) > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving module namespace",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if len(namespaces) == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving module namespace",
			fmt.Sprintf("Could not find module namespace with name '%s'.",
				cfg.Name.ValueString()),
		)
		return
	}
	namespace := namespaces[0]

	cfg.ID = types.StringValue(namespace.ID)
	cfg.Name = types.StringValue(namespace.Attributes.Name)
	cfg.IsShared = types.BoolValue(namespace.Attributes.IsShared)

	environments, diags := framework.FlattenRelationshipIDsSet(
		ctx,
		namespace.Relationships.Environments,
		func(e *schemas.Environment) string { return e.ID },
		&cfg.Environments,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg.Environments = environments

	modules, diags := framework.FlattenRelationshipIDsSet(
		ctx,
		namespace.Relationships.Modules,
		func(m *schemas.Module) string { return m.ID },
		&cfg.Modules,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg.Modules = modules

	owners, diags := framework.FlattenRelationshipIDsSet(
		ctx,
		namespace.Relationships.Owners,
		func(t *schemas.Team) string { return t.ID },
		&cfg.Owners,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cfg.Owners = owners

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

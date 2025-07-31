package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
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
					validation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the module namespace.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
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

	filter := scalr.ModuleNamespaceFilter{
		Name: cfg.Name.ValueString(),
	}

	options := scalr.ModuleNamespacesListOptions{
		Filter: &filter,
	}

	namespaces, err := r.Client.ModuleNamespaces.List(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving module namespaces", err.Error())
		return
	}

	if len(namespaces.Items) > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving module namespace",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if len(namespaces.Items) == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving module namespace",
			fmt.Sprintf("Could not find module namespace with name '%s'.",
				cfg.Name.ValueString()),
		)
		return
	}
	namespace := namespaces.Items[0]

	cfg.ID = types.StringValue(namespace.ID)
	cfg.Name = types.StringValue(namespace.Name)
	cfg.IsShared = types.BoolValue(namespace.IsShared)

	// Set environments
	if len(namespace.Environments) > 0 {
		environmentIDs := make([]string, len(namespace.Environments))
		for i, env := range namespace.Environments {
			environmentIDs[i] = env.ID
		}
		environmentsSet, diags := types.SetValueFrom(ctx, types.StringType, environmentIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		cfg.Environments = environmentsSet
	} else {
		cfg.Environments = types.SetNull(types.StringType)
	}

	// Set modules
	if len(namespace.Modules) > 0 {
		moduleIDs := make([]string, len(namespace.Modules))
		for i, module := range namespace.Modules {
			moduleIDs[i] = module.ID
		}
		modulesSet, diags := types.SetValueFrom(ctx, types.StringType, moduleIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		cfg.Modules = modulesSet
	} else {
		cfg.Modules = types.SetNull(types.StringType)
	}

	// Set owners
	if len(namespace.Owners) > 0 {
		ownerIDs := make([]string, len(namespace.Owners))
		for i, owner := range namespace.Owners {
			ownerIDs[i] = owner.ID
		}
		ownersSet, diags := types.SetValueFrom(ctx, types.StringType, ownerIDs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		cfg.Owners = ownersSet
	} else {
		cfg.Owners = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

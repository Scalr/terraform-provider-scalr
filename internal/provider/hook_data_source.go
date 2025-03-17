package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &hookDataSource{}
	_ datasource.DataSourceWithConfigure        = &hookDataSource{}
	_ datasource.DataSourceWithConfigValidators = &hookDataSource{}
)

func newHookDataSource() datasource.DataSource {
	return &hookDataSource{}
}

// hookDataSource defines the data source implementation.
type hookDataSource struct {
	framework.DataSourceWithScalrClient
}

// hookDataSourceModel describes the data source data model.
type hookDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Interpreter    types.String `tfsdk:"interpreter"`
	ScriptfilePath types.String `tfsdk:"scriptfile_path"`
	VcsProviderId  types.String `tfsdk:"vcs_provider_id"`
	VcsRepo        types.List   `tfsdk:"vcs_repo"`
}

// hookDataSourceVcsRepoModel maps the vcs_repo nested schema data.
type hookDataSourceVcsRepoModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Branch     types.String `tfsdk:"branch"`
}

func (d *hookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hook"
}

func (d *hookDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a hook.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the hook in the format `hook-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the hook.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the hook.",
				Computed:            true,
			},
			"interpreter": schema.StringAttribute{
				MarkdownDescription: "The interpreter to execute the hook script, such as 'bash', 'python3', etc.",
				Computed:            true,
			},
			"scriptfile_path": schema.StringAttribute{
				MarkdownDescription: "Path to the script file in the repository.",
				Computed:            true,
			},
			"vcs_provider_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VCS provider in the format `vcs-<RANDOM STRING>`.",
				Computed:            true,
			},
			"vcs_repo": schema.ListAttribute{
				MarkdownDescription: "Settings for the repository where the hook script is stored.",
				Computed:            true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"identifier": types.StringType,
						"branch":     types.StringType,
					},
				},
			},
		},
	}
}

func (d *hookDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *hookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg hookDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.HookListOptions{}

	if !cfg.Id.IsNull() {
		opts.Query = cfg.Id.ValueString()
	}

	if !cfg.Name.IsNull() {
		opts.Name = cfg.Name.ValueString()
	}

	hooks, err := d.Client.Hooks.List(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving hook", err.Error())
		return
	}

	// Якщо шукаємо за ID, додатково перевіряємо результати
	if !cfg.Id.IsNull() {
		idToFind := cfg.Id.ValueString()
		filteredHooks := make([]*scalr.Hook, 0)

		for _, hook := range hooks.Items {
			if hook.ID == idToFind {
				filteredHooks = append(filteredHooks, hook)
			}
		}

		// Замінюємо оригінальний список відфільтрованим
		hooks.Items = filteredHooks
		hooks.TotalCount = len(filteredHooks)
	}

	// Unlikely
	if hooks.TotalCount > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving hook",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if hooks.TotalCount == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving hook",
			fmt.Sprintf("Could not find hook with ID '%s', name '%s'.", cfg.Id.ValueString(), cfg.Name.ValueString()),
		)
		return
	}

	hook := hooks.Items[0]

	cfg.Id = types.StringValue(hook.ID)
	cfg.Name = types.StringValue(hook.Name)

	if hook.Description != "" {
		cfg.Description = types.StringValue(hook.Description)
	} else {
		cfg.Description = types.StringNull()
	}

	cfg.Interpreter = types.StringValue(hook.Interpreter)
	cfg.ScriptfilePath = types.StringValue(hook.ScriptfilePath)

	if hook.VcsProvider != nil {
		cfg.VcsProviderId = types.StringValue(hook.VcsProvider.ID)
	}

	if hook.VcsRepo != nil {
		vcsRepoModel := hookDataSourceVcsRepoModel{
			Identifier: types.StringValue(hook.VcsRepo.Identifier),
			Branch:     types.StringValue(hook.VcsRepo.Branch),
		}

		vcsRepos := []hookDataSourceVcsRepoModel{vcsRepoModel}

		vcsRepoList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"identifier": types.StringType,
				"branch":     types.StringType,
			},
		}, vcsRepos)

		if diags.HasError() {
			resp.Diagnostics.AddError("Error creating VCS repo list", diags.Errors()[0].Summary())
			return
		}

		cfg.VcsRepo = vcsRepoList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

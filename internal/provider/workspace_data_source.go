package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/ops/workspace"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &workspaceDataSource{}
	_ datasource.DataSourceWithConfigure        = &workspaceDataSource{}
	_ datasource.DataSourceWithConfigValidators = &workspaceDataSource{}
)

var (
	workspaceDataSourceVcsRepoElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"identifier":         types.StringType,
			"path":               types.StringType,
			"dry_runs_enabled":   types.BoolType,
			"ingress_submodules": types.BoolType,
		},
	}
)

func newWorkspaceDataSource() datasource.DataSource {
	return &workspaceDataSource{}
}

// workspaceDataSource defines the data source implementation.
type workspaceDataSource struct {
	framework.DataSourceWithScalrClient
}

// workspaceDataSourceModel describes the data source data model.
type workspaceDataSourceModel struct {
	Id                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	EnvironmentID             types.String `tfsdk:"environment_id"`
	VCSProviderID             types.String `tfsdk:"vcs_provider_id"`
	ModuleVersionID           types.String `tfsdk:"module_version_id"`
	AgentPoolID               types.String `tfsdk:"agent_pool_id"`
	RunnerImageVersionID      types.String `tfsdk:"runner_image_version_id"`
	AutoApply                 types.Bool   `tfsdk:"auto_apply"`
	ForceLatestRun            types.Bool   `tfsdk:"force_latest_run"`
	DeletionProtectionEnabled types.Bool   `tfsdk:"deletion_protection_enabled"`
	RemoteBackend             types.Bool   `tfsdk:"remote_backend"`
	Operations                types.Bool   `tfsdk:"operations"`
	ExecutionMode             types.String `tfsdk:"execution_mode"`
	TerraformVersion          types.String `tfsdk:"terraform_version"`
	Terragrunt                types.List   `tfsdk:"terragrunt"`
	IaCPlatform               types.String `tfsdk:"iac_platform"`
	Type                      types.String `tfsdk:"type"`
	WorkingDirectory          types.String `tfsdk:"working_directory"`
	HasResources              types.Bool   `tfsdk:"has_resources"`
	AutoQueueRuns             types.String `tfsdk:"auto_queue_runs"`
	Hooks                     types.List   `tfsdk:"hooks"`
	VCSRepo                   types.List   `tfsdk:"vcs_repo"`
	TagIDs                    types.List   `tfsdk:"tag_ids"`
	CreatedBy                 types.List   `tfsdk:"created_by"`
}

type workspaceDataSourceVcsRepoModel struct {
	Identifier        types.String `tfsdk:"identifier"`
	Path              types.String `tfsdk:"path"`
	DryRunsEnabled    types.Bool   `tfsdk:"dry_runs_enabled"`
	IngressSubmodules types.Bool   `tfsdk:"ingress_submodules"`
}

func (d *workspaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workspace"
}

func (d *workspaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves the details of a single workspace.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the workspace.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the workspace.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Required:            true,
			},
			"vcs_provider_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of a VCS provider in the format `vcs-<RANDOM STRING>`.",
				Computed:            true,
			},
			"module_version_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of a module version in the format `modver-<RANDOM STRING>`.",
				Computed:            true,
			},
			"agent_pool_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of an agent pool in the format `apool-<RANDOM STRING>`.",
				Computed:            true,
			},
			"runner_image_version_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the container image version used as the runner image on Scalr-managed agent pools." +
					" Null when the workspace uses the account default runner image.",
				Computed: true,
			},
			"auto_apply": schema.BoolAttribute{
				MarkdownDescription: "Boolean indicates if `terraform apply` will be automatically run when `terraform plan` ends without error.",
				Computed:            true,
			},
			"force_latest_run": schema.BoolAttribute{
				MarkdownDescription: "Boolean indicates if latest new run will be automatically raised in priority.",
				Computed:            true,
			},
			"deletion_protection_enabled": schema.BoolAttribute{
				MarkdownDescription: "Boolean, indicates if the workspace has the protection from an accidental state lost. If enabled and the workspace has resource, the deletion will not be allowed.",
				Computed:            true,
			},
			"remote_backend": schema.BoolAttribute{
				MarkdownDescription: "Manages if Scalr exports the remote backend configuration and state storage for your" +
					" infrastructure management. Disabling this feature will also prevent the ability to perform state locking," +
					" which ensures that concurrent operations do not conflict. Additionally, it will disable the capability to" +
					" initiate CLI-driven runs through Scalr.",
				Computed: true,
			},
			"operations": schema.BoolAttribute{
				MarkdownDescription: "Boolean indicates if the workspace is being used for remote execution.",
				Computed:            true,
			},
			"execution_mode": schema.StringAttribute{
				MarkdownDescription: "Execution mode of the workspace.",
				Computed:            true,
			},
			"terraform_version": schema.StringAttribute{
				MarkdownDescription: "The version of Terraform used for this workspace.",
				Computed:            true,
			},
			"terragrunt": schema.ListAttribute{
				MarkdownDescription: "List of terragrunt configurations in a workspace if set.",
				ElementType:         terragruntElementType,
				Computed:            true,
			},
			"iac_platform": schema.StringAttribute{
				MarkdownDescription: "The IaC platform used for this workspace.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the Scalr Workspace environment.",
				Computed:            true,
			},
			"working_directory": schema.StringAttribute{
				MarkdownDescription: "A relative path that Terraform will execute within.",
				Computed:            true,
			},
			"has_resources": schema.BoolAttribute{
				MarkdownDescription: "The presence of active terraform resources in the current state version.",
				Computed:            true,
			},
			"auto_queue_runs": schema.StringAttribute{
				MarkdownDescription: "Indicates if runs have to be queued automatically when a new configuration version is uploaded. Supported values are `skip_first`, `always`, `never`, `on_create_only`:" +
					"\n  * `skip_first` - after the very first configuration version is uploaded into the workspace the run will not be triggered. But the following configurations will do. This is the default behavior." +
					"\n  * `always` - runs will be triggered automatically on every upload of the configuration version." +
					"\n  * `never` - configuration versions are uploaded into the workspace, but runs will not be triggered." +
					"\n  * `on_create_only` - single run will be triggered only when the workspace is created and the first configuration version is uploaded. Subsequent configurations will not trigger runs.",
				Computed: true,
			},
			"hooks": schema.ListAttribute{
				MarkdownDescription: "List of custom hooks in a workspace.",
				ElementType:         hooksElementType,
				Computed:            true,
			},
			"vcs_repo": schema.ListAttribute{
				MarkdownDescription: "If a workspace is linked to a VCS repository this block shows the details, otherwise `{}`",
				ElementType:         workspaceDataSourceVcsRepoElementType,
				Computed:            true,
			},
			"tag_ids": schema.ListAttribute{
				MarkdownDescription: "List of tag IDs associated with the workspace.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"created_by": schema.ListAttribute{
				MarkdownDescription: "Details of the user that created the workspace.",
				ElementType:         userElementType,
				Computed:            true,
			},
		},
	}
}

func (d *workspaceDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *workspaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg workspaceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := workspace.GetWorkspacesOptions{
		Include: []string{"created-by"},
		Filter:  map[string]string{"environment": cfg.EnvironmentID.ValueString()},
	}
	if !cfg.Id.IsNull() {
		opts.Filter["workspace"] = cfg.Id.ValueString()
	}
	if !cfg.Name.IsNull() {
		opts.Filter["name"] = cfg.Name.ValueString()
	}

	workspaces, err := d.ClientV2.Workspace.GetWorkspaces(ctx, &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving workspace", err.Error())
		return
	}
	if len(workspaces) > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving workspace",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}
	if len(workspaces) == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving workspace",
			fmt.Sprintf(
				"Could not find workspace with ID '%s', name '%s' and environment_id '%s'",
				cfg.Id.ValueString(), cfg.Name.ValueString(), cfg.EnvironmentID.ValueString(),
			),
		)
		return
	}

	ws := workspaces[0]

	cfg.Id = types.StringValue(ws.ID)
	cfg.Name = types.StringValue(ws.Attributes.Name)
	cfg.AutoApply = types.BoolValue(ws.Attributes.AutoApply)
	cfg.ForceLatestRun = types.BoolValue(ws.Attributes.ForceLatestRun)
	cfg.DeletionProtectionEnabled = types.BoolValue(ws.Attributes.DeletionProtectionEnabled)
	cfg.Operations = types.BoolValue(ws.Attributes.Operations)
	cfg.ExecutionMode = types.StringValue(string(ws.Attributes.ExecutionMode))
	cfg.TerraformVersion = types.StringValue(ws.Attributes.TerraformVersion)
	cfg.IaCPlatform = types.StringValue(string(ws.Attributes.IacPlatform))
	cfg.Type = types.StringValue(string(ws.Attributes.EnvironmentType))
	cfg.WorkingDirectory = types.StringPointerValue(ws.Attributes.WorkingDirectory)
	cfg.HasResources = types.BoolValue(ws.Attributes.HasResources)
	cfg.AutoQueueRuns = types.StringValue(string(ws.Attributes.AutoQueueRuns))
	cfg.RemoteBackend = types.BoolValue(ws.Attributes.RemoteBackend)
	cfg.VCSProviderID = types.StringNull()
	cfg.ModuleVersionID = types.StringNull()
	cfg.AgentPoolID = types.StringNull()
	cfg.RunnerImageVersionID = types.StringNull()

	if ws.Relationships.VcsProvider != nil {
		cfg.VCSProviderID = types.StringValue(ws.Relationships.VcsProvider.ID)
	}
	if ws.Relationships.ModuleVersion != nil {
		cfg.ModuleVersionID = types.StringValue(ws.Relationships.ModuleVersion.ID)
	}
	if ws.Relationships.AgentPool != nil {
		cfg.AgentPoolID = types.StringValue(ws.Relationships.AgentPool.ID)
	}
	if ws.Relationships.RunnerImageVersion != nil {
		cfg.RunnerImageVersionID = types.StringValue(ws.Relationships.RunnerImageVersion.ID)
	}

	createdBy := make([]userModel, 0, 1)
	if ws.Relationships.CreatedBy != nil {
		createdBy = append(createdBy, *userModelFromAPIv2(ws.Relationships.CreatedBy))
	}
	createdByValue, diags := types.ListValueFrom(ctx, userElementType, createdBy)
	resp.Diagnostics.Append(diags...)
	cfg.CreatedBy = createdByValue

	vcsRepo := make([]workspaceDataSourceVcsRepoModel, 0, 1)
	if ws.Attributes.VcsRepo != nil {
		vcsRepo = append(vcsRepo, workspaceDataSourceVcsRepoModel{
			Identifier:        types.StringValue(ws.Attributes.VcsRepo.Identifier),
			Path:              types.StringPointerValue(ws.Attributes.VcsRepo.Path),
			DryRunsEnabled:    types.BoolValue(ws.Attributes.VcsRepo.DryRunsEnabled),
			IngressSubmodules: types.BoolValue(ws.Attributes.VcsRepo.IngressSubmodules),
		})
	}
	vcsRepoValue, diags := types.ListValueFrom(ctx, workspaceDataSourceVcsRepoElementType, vcsRepo)
	resp.Diagnostics.Append(diags...)
	cfg.VCSRepo = vcsRepoValue

	terragrunt := make([]terragruntModel, 0, 1)
	if ws.Attributes.Terragrunt != nil {
		terragrunt = append(terragrunt, terragruntModel{
			Version:                     types.StringValue(ws.Attributes.Terragrunt.Version),
			UseRunAll:                   types.BoolValue(ws.Attributes.Terragrunt.UseRunAll),
			IncludeExternalDependencies: types.BoolValue(ws.Attributes.Terragrunt.IncludeExternalDependencies),
		})
	}
	terragruntValue, diags := types.ListValueFrom(ctx, terragruntElementType, terragrunt)
	resp.Diagnostics.Append(diags...)
	cfg.Terragrunt = terragruntValue

	hooks := make([]hooksModel, 0, 1)
	if ws.Attributes.Hooks != nil {
		hooks = append(hooks, hooksModel{
			PreInit:   types.StringPointerValue(ws.Attributes.Hooks.PreInit),
			PrePlan:   types.StringPointerValue(ws.Attributes.Hooks.PrePlan),
			PostPlan:  types.StringPointerValue(ws.Attributes.Hooks.PostPlan),
			PreApply:  types.StringPointerValue(ws.Attributes.Hooks.PreApply),
			PostApply: types.StringPointerValue(ws.Attributes.Hooks.PostApply),
		})
	}
	hooksValue, diags := types.ListValueFrom(ctx, hooksElementType, hooks)
	resp.Diagnostics.Append(diags...)
	cfg.Hooks = hooksValue

	tags := make([]string, len(ws.Relationships.Tags))
	for i, tag := range ws.Relationships.Tags {
		tags[i] = tag.ID
	}
	sort.Strings(tags)
	tagsValue, diags := types.ListValueFrom(ctx, types.StringType, tags)
	resp.Diagnostics.Append(diags...)
	cfg.TagIDs = tagsValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

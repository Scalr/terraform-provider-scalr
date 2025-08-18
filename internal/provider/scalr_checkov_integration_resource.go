package provider

import (
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &checkovIntegrationResource{}
	_ resource.ResourceWithConfigure        = &checkovIntegrationResource{}
	_ resource.ResourceWithConfigValidators = &checkovIntegrationResource{}
	_ resource.ResourceWithImportState      = &checkovIntegrationResource{}
)

var (
	checkovVcsRepoElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"identifier": types.StringType,
			"branch":     types.StringType,
			"path":       types.StringType,
		},
	}
)

func newCheckovIntegrationResource() resource.Resource {
	return &checkovIntegrationResource{}
}

// checkovIntegrationResource defines the resource implementation.
type checkovIntegrationResource struct {
	framework.ResourceWithScalrClient
}

// checkovIntegrationResourceModel describes the resource data model.
type checkovIntegrationResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Version               types.String `tfsdk:"version"`
	CliArgs               types.String `tfsdk:"cli_args"`
	Environments          types.Set    `tfsdk:"environments"`
	VCSProviderID         types.String `tfsdk:"vcs_provider_id"`
	VCSRepo               types.List   `tfsdk:"vcs_repo"`
	ExternalChecksEnabled types.Bool   `tfsdk:"external_checks_enabled"`
}

type checkovVcsRepoModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Branch     types.String `tfsdk:"branch"`
	Path       types.String `tfsdk:"path"`
}

func (r *checkovIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_checkov_integration"
}

func (r *checkovIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of Checkov integrations in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Checkov integration.",
				Required:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of the Checkov integration to use.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cli_args": schema.StringAttribute{
				MarkdownDescription: "CLI parameters to be passed to checkov command.",
				Optional:            true,
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "List of environments this integration is linked to. Use `[\"*\"]` to allow in all environments.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"external_checks_enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether external checks should be enabled. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"vcs_provider_id": schema.StringAttribute{
				MarkdownDescription: "ID of VCS provider in the format `vcs-<RANDOM STRING>`. Required if `external_checks_enabled` is `true`.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"vcs_repo": schema.ListNestedBlock{
				MarkdownDescription: "Settings for the Checkov integration's VCS repository. Required if `external_checks_enabled` is `true`.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							MarkdownDescription: "A reference to your VCS repository." +
								" For GitHub, GitHub Enterprise and GitLab the format is `<org>/<repo>`." +
								" For Azure DevOps Services the format is `<org>/<project>/<repo>`.",
							Required: true,
						},
						"branch": schema.StringAttribute{
							MarkdownDescription: "Branch of a repository the Checkov custom checks are associated with.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "The sub-directory of the VCS repository where Checkov checks are stored." +
								" If omitted or specified as an empty string, this defaults to the repository's root.",
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func (r *checkovIntegrationResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.RequiredTogether(
			path.MatchRoot("vcs_provider_id"),
			path.MatchRoot("vcs_repo"),
		),
	}
}

func (r *checkovIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan checkovIntegrationResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.CheckovIntegrationCreateOptions{
		Name:                  plan.Name.ValueStringPointer(),
		IsShared:              ptr(false),
		ExternalChecksEnabled: plan.ExternalChecksEnabled.ValueBoolPointer(),
	}

	if !plan.Version.IsNull() {
		opts.Version = plan.Version.ValueStringPointer()
	}

	if !plan.CliArgs.IsNull() {
		opts.CliArgs = plan.CliArgs.ValueStringPointer()
	}

	if !plan.VCSProviderID.IsUnknown() && !plan.VCSProviderID.IsNull() {
		opts.VcsProvider = &scalr.VcsProvider{
			ID: plan.VCSProviderID.ValueString(),
		}
	}

	if !plan.VCSRepo.IsUnknown() && !plan.VCSRepo.IsNull() {
		var vcsRepo []checkovVcsRepoModel
		resp.Diagnostics.Append(plan.VCSRepo.ElementsAs(ctx, &vcsRepo, false)...)

		if len(vcsRepo) > 0 {
			repo := vcsRepo[0]

			opts.VCSRepo = &scalr.CheckovIntegrationVCSRepoOptions{
				Identifier: repo.Identifier.ValueStringPointer(),
				Path:       repo.Path.ValueStringPointer(),
			}

			if !repo.Branch.IsUnknown() && !repo.Branch.IsNull() {
				opts.VCSRepo.Branch = repo.Branch.ValueStringPointer()
			}
		}
	}

	if !plan.Environments.IsUnknown() && !plan.Environments.IsNull() {
		var environments []string
		resp.Diagnostics.Append(plan.Environments.ElementsAs(ctx, &environments, false)...)

		if (len(environments) == 1) && (environments[0] == "*") {
			opts.IsShared = ptr(true)
		} else if len(environments) > 0 {
			envs := make([]*scalr.Environment, len(environments))
			for i, env := range environments {
				envs[i] = &scalr.Environment{ID: env}
			}
			opts.Environments = envs
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	checkovIntegration, err := r.Client.CheckovIntegrations.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Checkov integration", err.Error())
		return
	}

	plan.Id = types.StringValue(checkovIntegration.ID)
	plan.Name = types.StringValue(checkovIntegration.Name)
	plan.Version = types.StringValue(checkovIntegration.Version)
	plan.CliArgs = types.StringValue(checkovIntegration.CliArgs)

	envs := make([]string, len(checkovIntegration.Environments))
	for i, env := range checkovIntegration.Environments {
		envs[i] = env.ID
	}
	if checkovIntegration.IsShared {
		envs = []string{"*"}
	}
	envsValues, d := types.SetValueFrom(ctx, types.StringType, envs)
	resp.Diagnostics.Append(d...)
	plan.Environments = envsValues

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *checkovIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state checkovIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	checkovIntegration, err := r.Client.CheckovIntegrations.Read(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving Checkov integration", err.Error())
		return
	}

	// Overwrite attributes with refreshed values
	state.Name = types.StringValue(checkovIntegration.Name)
	state.Version = types.StringValue(checkovIntegration.Version)
	state.CliArgs = types.StringValue(checkovIntegration.CliArgs)
	state.ExternalChecksEnabled = types.BoolValue(checkovIntegration.ExternalChecksEnabled)

	state.VCSProviderID = types.StringNull()
	if checkovIntegration.VcsProvider != nil {
		state.VCSProviderID = types.StringValue(checkovIntegration.VcsProvider.ID)
	}

	state.VCSRepo = types.ListNull(checkovVcsRepoElementType)
	if checkovIntegration.VCSRepo != nil {
		repo := checkovVcsRepoModel{
			Identifier: types.StringValue(checkovIntegration.VCSRepo.Identifier),
			Path:       types.StringValue(checkovIntegration.VCSRepo.Path),
		}

		if checkovIntegration.VCSRepo.Branch != "" {
			branch := types.StringValue(checkovIntegration.VCSRepo.Branch)
			repo.Branch = branch
		}

		repoValue, d := types.ListValueFrom(ctx, checkovVcsRepoElementType, []checkovVcsRepoModel{repo})
		resp.Diagnostics.Append(d...)
		state.VCSRepo = repoValue
	}

	if checkovIntegration.IsShared {
		envs := []string{"*"}
		envsValues, d := types.SetValueFrom(ctx, types.StringType, envs)
		resp.Diagnostics.Append(d...)
		state.Environments = envsValues
	} else {
		envs := make([]string, len(checkovIntegration.Environments))
		for i, env := range checkovIntegration.Environments {
			envs[i] = env.ID
		}
		envsValues, d := types.SetValueFrom(ctx, types.StringType, envs)
		resp.Diagnostics.Append(d...)
		state.Environments = envsValues
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *checkovIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read plan data
	var plan, state checkovIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.CheckovIntegrationUpdateOptions{}

	if !plan.Name.Equal(state.Name) {
		opts.Name = plan.Name.ValueStringPointer()
	}

	if !plan.Version.Equal(state.Version) {
		opts.Version = plan.Version.ValueStringPointer()
	}

	if !plan.CliArgs.Equal(state.CliArgs) {
		opts.CliArgs = plan.CliArgs.ValueStringPointer()
	}

	if !plan.ExternalChecksEnabled.Equal(state.ExternalChecksEnabled) {
		opts.ExternalChecksEnabled = plan.ExternalChecksEnabled.ValueBoolPointer()
	}

	if !plan.VCSProviderID.IsNull() {
		opts.VcsProvider = &scalr.VcsProvider{ID: plan.VCSProviderID.ValueString()}
	}

	if !plan.VCSRepo.IsNull() {
		var vcsRepo []checkovVcsRepoModel
		resp.Diagnostics.Append(plan.VCSRepo.ElementsAs(ctx, &vcsRepo, false)...)

		if len(vcsRepo) > 0 {
			repo := vcsRepo[0]

			opts.VCSRepo = &scalr.CheckovIntegrationVCSRepoOptions{
				Identifier: repo.Identifier.ValueStringPointer(),
				Path:       repo.Path.ValueStringPointer(),
			}

			if !repo.Branch.IsUnknown() && !repo.Branch.IsNull() {
				opts.VCSRepo.Branch = repo.Branch.ValueStringPointer()
			}
		}
	}

	var environments []string
	if !plan.Environments.Equal(state.Environments) {
		resp.Diagnostics.Append(plan.Environments.ElementsAs(ctx, &environments, false)...)
	} else {
		resp.Diagnostics.Append(state.Environments.ElementsAs(ctx, &environments, false)...)
	}
	if (len(environments) == 1) && (environments[0] == "*") {
		opts.IsShared = ptr(true)
		opts.Environments = make([]*scalr.Environment, 0)
	} else if len(environments) > 0 {
		envs := make([]*scalr.Environment, len(environments))
		for i, env := range environments {
			envs[i] = &scalr.Environment{ID: env}
		}
		opts.Environments = envs
		opts.IsShared = ptr(false)
	} else {
		opts.IsShared = ptr(false)
		opts.Environments = make([]*scalr.Environment, 0)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing resource
	checkovIntegration, err := r.Client.CheckovIntegrations.Update(ctx, plan.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Checkov integration", err.Error())
		return
	}

	// Overwrite attributes with refreshed values
	plan.Name = types.StringValue(checkovIntegration.Name)
	plan.Version = types.StringValue(checkovIntegration.Version)
	plan.CliArgs = types.StringValue(checkovIntegration.CliArgs)

	if checkovIntegration.IsShared {
		envs := []string{"*"}
		envsValues, d := types.SetValueFrom(ctx, types.StringType, envs)
		resp.Diagnostics.Append(d...)
		state.Environments = envsValues
	} else {
		envs := make([]string, len(checkovIntegration.Environments))
		for i, env := range checkovIntegration.Environments {
			envs[i] = env.ID
		}
		envsValues, d := types.SetValueFrom(ctx, types.StringType, envs)
		resp.Diagnostics.Append(d...)
		state.Environments = envsValues
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *checkovIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state checkovIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.CheckovIntegrations.Delete(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting Checkov integration", err.Error())
		return
	}
}

func (r *checkovIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &checkovIntegrationResource{}
	_ resource.ResourceWithConfigure        = &checkovIntegrationResource{}
	_ resource.ResourceWithConfigValidators = &checkovIntegrationResource{}
	_ resource.ResourceWithImportState      = &checkovIntegrationResource{}
)

var checkovVcsRepoElementType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"identifier": types.StringType,
		"branch":     types.StringType,
		"path":       types.StringType,
	},
}

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
	ScanMode              types.String `tfsdk:"scan_mode"`
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

func checkovIntegrationResourceModelFromAPI(
	ctx context.Context,
	ci *schemas.CheckovIntegration,
	prior *checkovIntegrationResourceModel,
) (*checkovIntegrationResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &checkovIntegrationResourceModel{
		Id:                    types.StringValue(ci.ID),
		Name:                  types.StringValue(ci.Attributes.Name),
		Version:               types.StringValue(ci.Attributes.Version),
		CliArgs:               types.StringValue(stringOrEmpty(ci.Attributes.CliArgs)),
		ScanMode:              types.StringValue(string(ci.Attributes.ScanMode)),
		ExternalChecksEnabled: types.BoolValue(ci.Attributes.ExternalChecksEnabled),
		VCSProviderID:         types.StringNull(),
		VCSRepo:               types.ListNull(checkovVcsRepoElementType),
		Environments:          types.SetNull(types.StringType),
	}

	if ci.Relationships.VcsProvider != nil {
		model.VCSProviderID = types.StringValue(ci.Relationships.VcsProvider.ID)
	}

	if ci.Attributes.VcsRepo != nil {
		repo := []checkovVcsRepoModel{{
			Identifier: types.StringValue(ci.Attributes.VcsRepo.Identifier),
			Branch:     types.StringValue(ci.Attributes.VcsRepo.Branch),
			Path:       types.StringValue(stringOrEmpty(ci.Attributes.VcsRepo.Path)),
		}}
		repoValue, d := types.ListValueFrom(ctx, checkovVcsRepoElementType, repo)
		diags.Append(d...)
		model.VCSRepo = repoValue
	}

	var envs types.Set
	var d diag.Diagnostics
	if ci.Attributes.IsShared {
		envs, d = types.SetValueFrom(ctx, types.StringType, []string{"*"})
	} else {
		var priorEnvs *types.Set
		if prior != nil {
			priorEnvs = &prior.Environments
		}
		envs, d = framework.FlattenRelationshipIDsSet(
			ctx,
			ci.Relationships.Environments,
			func(e *schemas.Environment) string { return e.ID },
			priorEnvs,
		)
	}
	diags.Append(d...)
	model.Environments = envs

	return model, diags
}

// expandCheckovVcsRepo converts the `vcs_repo` block of the plan into the API request options.
func expandCheckovVcsRepo(ctx context.Context, v types.List) (
	*schemas.CheckovIntegrationVcsRepoRequest,
	diag.Diagnostics,
) {
	var diags diag.Diagnostics

	if v.IsUnknown() || v.IsNull() {
		return nil, diags
	}

	var repos []checkovVcsRepoModel
	diags.Append(v.ElementsAs(ctx, &repos, false)...)
	if diags.HasError() || len(repos) == 0 {
		return nil, diags
	}

	repo := repos[0]
	return &schemas.CheckovIntegrationVcsRepoRequest{
		Identifier: value.Set(repo.Identifier.ValueString()),
		Path:       value.Set(repo.Path.ValueString()),
		Branch:     framework.SetIfKnownString(repo.Branch),
	}, diags
}

// expandCheckovEnvironments converts the `environments` attribute of the plan into the list of
// environments to link the integration to. The `["*"]` wildcard is reported back as the
// `isShared` flag, which makes the integration available in every environment of the account.
func expandCheckovEnvironments(ctx context.Context, v types.Set) (
	environments []schemas.Environment,
	isShared bool,
	diags diag.Diagnostics,
) {
	environments = make([]schemas.Environment, 0)

	if v.IsUnknown() || v.IsNull() {
		return environments, false, diags
	}

	envs, d := framework.ExpandRelationshipIDsSet(ctx, v, func(id string) schemas.Environment {
		return schemas.Environment{ID: id}
	})
	diags.Append(d...)
	if diags.HasError() {
		return environments, false, diags
	}

	if len(envs) == 1 && envs[0].ID == "*" {
		return environments, true, diags
	}

	return envs, false, diags
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
					stringvalidation.StringIsNotWhiteSpace(),
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
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"scan_mode": schema.StringAttribute{
				MarkdownDescription: "What Checkov scans." +
					" `code` scans the configuration files before the run is planned." +
					" `plan` scans the plan JSON after the plan phase, so the findings cover the resources" +
					" the run actually provisions; such a scan is skipped when the plan has no changes." +
					" Valid values are `code`, `plan`. Default `code`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(schemas.CheckovIntegrationScanModeCode)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(schemas.CheckovIntegrationScanModeCode),
						string(schemas.CheckovIntegrationScanModePlan),
					),
				},
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "List of environments this integration is linked to. Use `[\"*\"]` to allow in all environments.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidation.StringIsNotWhiteSpace()),
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

	opts := schemas.CheckovIntegrationRequest{
		Attributes: schemas.CheckovIntegrationAttributesRequest{
			Name:                  value.Set(plan.Name.ValueString()),
			Version:               framework.SetIfKnownString(plan.Version),
			CliArgs:               framework.SetIfKnownString(plan.CliArgs),
			ScanMode:              checkovScanMode(plan.ScanMode),
			ExternalChecksEnabled: framework.SetIfKnownBool(plan.ExternalChecksEnabled),
			IsShared:              value.Set(false),
		},
	}

	if !plan.VCSProviderID.IsUnknown() && !plan.VCSProviderID.IsNull() {
		opts.Relationships.VcsProvider = value.Set(schemas.VcsProvider{ID: plan.VCSProviderID.ValueString()})
	}

	vcsRepo, diags := expandCheckovVcsRepo(ctx, plan.VCSRepo)
	resp.Diagnostics.Append(diags...)
	opts.Attributes.VcsRepo = value.SetPtrMaybe(vcsRepo)

	envs, isShared, diags := expandCheckovEnvironments(ctx, plan.Environments)
	resp.Diagnostics.Append(diags...)
	if isShared {
		opts.Attributes.IsShared = value.Set(true)
	} else if len(envs) > 0 {
		opts.Relationships.Environments = value.Set(envs)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ci, err := r.ClientV2.CheckovIntegration.CreateCheckovIntegration(ctx, &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Checkov integration", err.Error())
		return
	}

	result, diags := checkovIntegrationResourceModelFromAPI(ctx, ci, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *checkovIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state checkovIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	ci, err := r.ClientV2.CheckovIntegration.GetCheckovIntegration(ctx, state.Id.ValueString(), nil)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving Checkov integration", err.Error())
		return
	}

	result, diags := checkovIntegrationResourceModelFromAPI(ctx, ci, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *checkovIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read plan and state data
	var plan, state checkovIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.CheckovIntegrationRequest{}

	if !plan.Name.Equal(state.Name) {
		opts.Attributes.Name = value.Set(plan.Name.ValueString())
	}

	if !plan.Version.Equal(state.Version) {
		opts.Attributes.Version = framework.SetIfKnownString(plan.Version)
	}

	if !plan.CliArgs.Equal(state.CliArgs) {
		opts.Attributes.CliArgs = framework.SetIfKnownString(plan.CliArgs)
	}

	if !plan.ScanMode.Equal(state.ScanMode) {
		opts.Attributes.ScanMode = checkovScanMode(plan.ScanMode)
	}

	if !plan.ExternalChecksEnabled.Equal(state.ExternalChecksEnabled) {
		opts.Attributes.ExternalChecksEnabled = framework.SetIfKnownBool(plan.ExternalChecksEnabled)
	}

	if !plan.VCSProviderID.IsNull() {
		opts.Relationships.VcsProvider = value.Set(schemas.VcsProvider{ID: plan.VCSProviderID.ValueString()})
	}

	vcsRepo, diags := expandCheckovVcsRepo(ctx, plan.VCSRepo)
	resp.Diagnostics.Append(diags...)
	opts.Attributes.VcsRepo = value.SetPtrMaybe(vcsRepo)

	// An integration that is not shared keeps only the environments the request carries, so the
	// full set goes with every update, whether or not it changed.
	environments := plan.Environments
	if plan.Environments.Equal(state.Environments) {
		environments = state.Environments
	}
	envs, isShared, diags := expandCheckovEnvironments(ctx, environments)
	resp.Diagnostics.Append(diags...)
	opts.Attributes.IsShared = value.Set(isShared)
	opts.Relationships.Environments = value.Set(envs)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing resource
	ci, err := r.ClientV2.CheckovIntegration.UpdateCheckovIntegration(ctx, plan.Id.ValueString(), &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Checkov integration", err.Error())
		return
	}

	result, diags := checkovIntegrationResourceModelFromAPI(ctx, ci, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *checkovIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state checkovIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.CheckovIntegration.DeleteCheckovIntegration(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting Checkov integration", err.Error())
		return
	}
}

func (r *checkovIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// checkovScanMode converts the `scan_mode` attribute value into the API request value.
func checkovScanMode(v types.String) *value.Value[schemas.CheckovIntegrationScanMode] {
	if v.IsUnknown() || v.IsNull() {
		return value.Unset[schemas.CheckovIntegrationScanMode]()
	}
	return value.Set(schemas.CheckovIntegrationScanMode(v.ValueString()))
}

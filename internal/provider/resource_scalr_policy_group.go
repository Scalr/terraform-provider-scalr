package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/ops/policy_group"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/defaults"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ resource.Resource                = &policyGroupResource{}
	_ resource.ResourceWithConfigure   = &policyGroupResource{}
	_ resource.ResourceWithImportState = &policyGroupResource{}
)

func newPolicyGroupResource() resource.Resource {
	return &policyGroupResource{}
}

// policyGroupResource defines the resource implementation.
type policyGroupResource struct {
	framework.ResourceWithScalrClient
}

func (r *policyGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_group"
}

func policyGroupResourceSchema() *schema.Schema {
	return &schema.Schema{
		MarkdownDescription: "Manage the state of policy groups in Scalr. Create, update and destroy.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of a policy group.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "A system status of the Policy group.",
				Computed:            true,
			},
			"error_message": schema.StringAttribute{
				MarkdownDescription: "A detailed error if Scalr failed to process the policy group.",
				Computed:            true,
			},
			"opa_version": schema.StringAttribute{
				MarkdownDescription: "The version of Open Policy Agent to run policies against." +
					" If omitted, the system default version is assigned.",
				Optional: true,
				Computed: true,
			},
			"execution_mode": schema.StringAttribute{
				MarkdownDescription: "The stage of the run the policy group is evaluated at." +
					" Valid values are `pre-plan` and `post-plan`. Defaults to `post-plan`." +
					" Changing this forces a resource to be re-created.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(policyGroupExecutionModePostPlan),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						policyGroupExecutionModePrePlan,
						policyGroupExecutionModePostPlan,
					),
				},
			},
			"common_functions_folder": schema.StringAttribute{
				MarkdownDescription: "An absolute path from the repository root to the folder that contains common rego functions.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the Scalr account, in the format `acc-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				Default:             defaults.AccountIDRequired(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vcs_provider_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of a VCS provider, in the format `vcs-<RANDOM STRING>`.",
				Required:            true,
			},
			"policies": schema.ListAttribute{
				MarkdownDescription: "A list of the OPA policies the group verifies each run.",
				ElementType:         policyGroupPolicyElementType,
				Computed:            true,
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "A list of the environments the policy group is linked to." +
					" Use `[\"*\"]` to enforce in all environments." +
					" To manage a linkage use either this attribute or the `scalr_policy_group_linkage` resource.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},

		Blocks: map[string]schema.Block{
			"vcs_repo": schema.ListNestedBlock{
				MarkdownDescription: "The VCS meta-data to create the policy from.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							MarkdownDescription: "The reference to the VCS repository in the format `:org/:repo`," +
								" this refers to the organization and repository in your VCS provider.",
							Required: true,
						},
						"branch": schema.StringAttribute{
							MarkdownDescription: "The branch of a repository the policy group is associated with." +
								" If omitted, the repository default branch will be used.",
							Optional: true,
							Computed: true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "The subdirectory of the VCS repository where OPA policies are stored." +
								" If omitted or submitted as an empty string, this defaults to the repository's root.",
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}
}

func (r *policyGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = *policyGroupResourceSchema()
}

// Values of the `execution_mode` attribute, mapped to the API `execute-as` values.
const (
	policyGroupExecutionModePrePlan  = "pre-plan"
	policyGroupExecutionModePostPlan = "post-plan"
)

var (
	policyGroupVcsRepoElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"identifier": types.StringType,
			"branch":     types.StringType,
			"path":       types.StringType,
		},
	}
	policyGroupPolicyElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":           types.StringType,
			"enabled":        types.BoolType,
			"enforced_level": types.StringType,
		},
	}
)

// policyGroupResourceModel describes the resource data model.
type policyGroupResourceModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Status                types.String `tfsdk:"status"`
	ErrorMessage          types.String `tfsdk:"error_message"`
	OpaVersion            types.String `tfsdk:"opa_version"`
	ExecutionMode         types.String `tfsdk:"execution_mode"`
	CommonFunctionsFolder types.String `tfsdk:"common_functions_folder"`
	AccountID             types.String `tfsdk:"account_id"`
	VCSProviderID         types.String `tfsdk:"vcs_provider_id"`
	VCSRepo               types.List   `tfsdk:"vcs_repo"`
	Policies              types.List   `tfsdk:"policies"`
	Environments          types.Set    `tfsdk:"environments"`
}

type policyGroupVcsRepoModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Branch     types.String `tfsdk:"branch"`
	Path       types.String `tfsdk:"path"`
}

type policyGroupPolicyModel struct {
	Name          types.String `tfsdk:"name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	EnforcedLevel types.String `tfsdk:"enforced_level"`
}

// expandPolicyGroupVcsRepo converts the `vcs_repo` block of the plan into the API request options.
func expandPolicyGroupVcsRepo(ctx context.Context, v types.List) (
	*schemas.PolicyGroupVcsRepoRequest,
	diag.Diagnostics,
) {
	var diags diag.Diagnostics

	if v.IsUnknown() || v.IsNull() {
		return nil, diags
	}

	var repos []policyGroupVcsRepoModel
	diags.Append(v.ElementsAs(ctx, &repos, false)...)
	if diags.HasError() || len(repos) == 0 {
		return nil, diags
	}

	repo := repos[0]
	opts := &schemas.PolicyGroupVcsRepoRequest{
		Identifier: value.Set(repo.Identifier.ValueString()),
		Path:       value.Set(repo.Path.ValueString()),
		Branch:     framework.SetIfKnownString(repo.Branch),
	}

	return opts, diags
}

// expandPolicyGroupEnvironments converts the `environments` attribute of the plan into the list
// of environments to link the policy group to. The `["*"]` wildcard value is reported back
// as the `isEnforced` flag, which enforces the policy group in all environments.
func expandPolicyGroupEnvironments(ctx context.Context, v types.Set) (
	environments []schemas.Environment,
	isEnforced bool,
	diags diag.Diagnostics,
) {
	environments = make([]schemas.Environment, 0)

	if v.IsUnknown() || v.IsNull() {
		return environments, false, diags
	}

	var environmentIDs []string
	diags.Append(v.ElementsAs(ctx, &environmentIDs, false)...)
	if diags.HasError() {
		return environments, false, diags
	}

	if len(environmentIDs) == 1 && environmentIDs[0] == "*" {
		return environments, true, diags
	}

	for _, environmentID := range environmentIDs {
		if environmentID == "*" {
			diags.AddAttributeError(
				path.Root("environments"),
				"Invalid environments value",
				"Impossible to enforce the policy group in all and on a limited list of environments."+
					" Please remove either wildcard or environment identifiers.",
			)
			return environments, false, diags
		}
		environments = append(environments, schemas.Environment{ID: environmentID})
	}

	return environments, false, diags
}

// readPolicyGroup retrieves the policy group along with the policies it contains.
func (r *policyGroupResource) readPolicyGroup(ctx context.Context, id string) (*schemas.PolicyGroup, error) {
	return r.ClientV2.PolicyGroup.GetPolicyGroup(ctx, id, &policy_group.GetPolicyGroupOptions{
		Include: []string{"policies"},
	})
}

func (r *policyGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyGroupResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vcsRepo, d := expandPolicyGroupVcsRepo(ctx, plan.VCSRepo)
	resp.Diagnostics.Append(d...)

	environments, isEnforced, d := expandPolicyGroupEnvironments(ctx, plan.Environments)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.PolicyGroupRequest{
		Attributes: schemas.PolicyGroupAttributesRequest{
			Name:                  value.Set(plan.Name.ValueString()),
			OpaVersion:            framework.SetIfKnownString(plan.OpaVersion),
			ExecuteAs:             policyGroupExecuteAs(plan.ExecutionMode),
			CommonFunctionsFolder: policyGroupCommonFunctionsFolder(plan.CommonFunctionsFolder),
			IsEnforced:            value.Set(isEnforced),
			VcsRepo:               value.SetPtrMaybe(vcsRepo),
		},
		Relationships: schemas.PolicyGroupRelationshipsRequest{
			Account:     value.Set(schemas.Account{ID: plan.AccountID.ValueString()}),
			VcsProvider: value.Set(schemas.VcsProvider{ID: plan.VCSProviderID.ValueString()}),
		},
	}

	pg, err := r.ClientV2.PolicyGroup.CreatePolicyGroup(ctx, &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating policy group", err.Error())
		return
	}

	if !isEnforced && len(environments) > 0 {
		err = r.ClientV2.PolicyGroup.CreatePolicyGroupEnvironments(ctx, pg.ID, environments)
		if err != nil {
			detail := fmt.Sprintf(
				"Failed to link environments to policy group %q: %s", pg.ID, err,
			)
			rollbackErr := r.ClientV2.PolicyGroup.DeletePolicyGroup(ctx, pg.ID)
			if rollbackErr != nil && !errors.Is(rollbackErr, client.ErrNotFound) {
				detail = fmt.Sprintf(
					"%s\n\nFailed to delete the policy group during rollback: %s", detail, rollbackErr,
				)
			}
			resp.Diagnostics.AddError("Error linking environments to policy group", detail)
			return
		}
	}

	// Get refreshed resource state from API
	pg, err = r.readPolicyGroup(ctx, pg.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving policy group", err.Error())
		return
	}

	result, d := policyGroupResourceModelFromAPI(ctx, pg)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *policyGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state policyGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	pg, err := r.readPolicyGroup(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving policy group", err.Error())
		return
	}

	result, d := policyGroupResourceModelFromAPI(ctx, pg)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *policyGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state policyGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.PolicyGroupRequest{}

	if !plan.Name.Equal(state.Name) {
		opts.Attributes.Name = value.Set(plan.Name.ValueString())
	}

	if !plan.OpaVersion.Equal(state.OpaVersion) {
		opts.Attributes.OpaVersion = framework.SetIfKnownString(plan.OpaVersion)
	}

	if !plan.CommonFunctionsFolder.Equal(state.CommonFunctionsFolder) {
		opts.Attributes.CommonFunctionsFolder = policyGroupCommonFunctionsFolder(plan.CommonFunctionsFolder)
	}

	if !plan.VCSRepo.Equal(state.VCSRepo) {
		vcsRepo, d := expandPolicyGroupVcsRepo(ctx, plan.VCSRepo)
		resp.Diagnostics.Append(d...)
		opts.Attributes.VcsRepo = value.SetPtrMaybe(vcsRepo)
	}

	if !plan.VCSProviderID.Equal(state.VCSProviderID) {
		opts.Relationships.VcsProvider = value.Set(schemas.VcsProvider{ID: plan.VCSProviderID.ValueString()})
	}

	var (
		environments       []schemas.Environment
		isEnforced         bool
		updateEnvironments bool
	)
	if !plan.Environments.IsUnknown() && !plan.Environments.Equal(state.Environments) {
		var d diag.Diagnostics
		environments, isEnforced, d = expandPolicyGroupEnvironments(ctx, plan.Environments)
		resp.Diagnostics.Append(d...)
		updateEnvironments = true
		opts.Attributes.IsEnforced = value.Set(isEnforced)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing resource
	_, err := r.ClientV2.PolicyGroup.UpdatePolicyGroup(ctx, plan.Id.ValueString(), &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating policy group", err.Error())
		return
	}

	if updateEnvironments && !isEnforced {
		err = r.ClientV2.PolicyGroup.UpdatePolicyGroupEnvironments(ctx, plan.Id.ValueString(), environments)
		if err != nil {
			resp.Diagnostics.AddError("Error updating environments for policy group", err.Error())
			return
		}
	}

	// Get refreshed resource state from API
	pg, err := r.readPolicyGroup(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving policy group", err.Error())
		return
	}

	result, d := policyGroupResourceModelFromAPI(ctx, pg)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *policyGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state policyGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.PolicyGroup.DeletePolicyGroup(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting policy group", err.Error())
		return
	}
}

func (r *policyGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// policyGroupExecuteAs converts the `execution_mode` attribute value into the API request value.
func policyGroupExecuteAs(v types.String) *value.Value[schemas.PolicyGroupExecuteAs] {
	if v.IsUnknown() || v.IsNull() {
		return value.Unset[schemas.PolicyGroupExecuteAs]()
	}
	return value.Set(policyGroupExecutionModeToAPI(v.ValueString()))
}

// policyGroupCommonFunctionsFolder converts the `common_functions_folder` attribute value
// into the API request value, sending an explicit null when the folder is not set.
func policyGroupCommonFunctionsFolder(v types.String) *value.Value[string] {
	if v.IsUnknown() {
		return value.Unset[string]()
	}
	if v.IsNull() || v.ValueString() == "" {
		return value.Null[string]()
	}
	return value.Set(v.ValueString())
}

func policyGroupResourceModelFromAPI(
	ctx context.Context,
	pg *schemas.PolicyGroup,
) (*policyGroupResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &policyGroupResourceModel{
		Id:           types.StringValue(pg.ID),
		Name:         types.StringValue(pg.Attributes.Name),
		Status:       types.StringValue(string(pg.Attributes.Status)),
		ErrorMessage: types.StringValue(stringOrEmpty(pg.Attributes.ErrorMessage)),
		OpaVersion:   types.StringValue(pg.Attributes.OpaVersion),
		ExecutionMode: types.StringValue(
			policyGroupExecutionModeFromAPI(pg.Attributes.ExecuteAs),
		),

		CommonFunctionsFolder: types.StringValue(stringOrEmpty(pg.Attributes.CommonFunctionsFolder)),
		AccountID:             types.StringNull(),
		VCSProviderID:         types.StringNull(),
		VCSRepo:               types.ListNull(policyGroupVcsRepoElementType),
		Policies:              types.ListNull(policyGroupPolicyElementType),
		Environments:          types.SetNull(types.StringType),
	}

	if pg.Relationships.Account != nil {
		model.AccountID = types.StringValue(pg.Relationships.Account.ID)
	}

	if pg.Relationships.VcsProvider != nil {
		model.VCSProviderID = types.StringValue(pg.Relationships.VcsProvider.ID)
	}

	repo := []policyGroupVcsRepoModel{{
		Identifier: types.StringValue(pg.Attributes.VcsRepo.Identifier),
		Branch:     types.StringPointerValue(pg.Attributes.VcsRepo.Branch),
		Path:       types.StringValue(stringOrEmpty(pg.Attributes.VcsRepo.Path)),
	}}
	repoValue, d := types.ListValueFrom(ctx, policyGroupVcsRepoElementType, repo)
	diags.Append(d...)
	model.VCSRepo = repoValue

	policies := make([]policyGroupPolicyModel, 0, len(pg.Relationships.Policies))
	for _, policy := range pg.Relationships.Policies {
		if policy == nil {
			continue
		}
		policies = append(policies, policyGroupPolicyModel{
			Name:          types.StringValue(policy.Attributes.Name),
			Enabled:       types.BoolValue(policy.Attributes.Enabled),
			EnforcedLevel: types.StringValue(string(policy.Attributes.EnforcedLevel)),
		})
	}
	policiesValue, d := types.ListValueFrom(ctx, policyGroupPolicyElementType, policies)
	diags.Append(d...)
	model.Policies = policiesValue

	environments := make([]string, 0, len(pg.Relationships.Environments))
	if pg.Attributes.IsEnforced {
		environments = append(environments, "*")
	} else {
		for _, environment := range pg.Relationships.Environments {
			if environment == nil {
				continue
			}
			environments = append(environments, environment.ID)
		}
	}
	environmentsValue, d := types.SetValueFrom(ctx, types.StringType, environments)
	diags.Append(d...)
	model.Environments = environmentsValue

	return model, diags
}

// policyGroupExecutionModeToAPI converts the `execution_mode` attribute value
// into the API `execute-as` value.
func policyGroupExecutionModeToAPI(mode string) schemas.PolicyGroupExecuteAs {
	if mode == policyGroupExecutionModePrePlan {
		return schemas.PolicyGroupExecuteAsPrePlanCheck
	}
	return schemas.PolicyGroupExecuteAsPolicyCheck
}

// policyGroupExecutionModeFromAPI converts the API `execute-as` value
// into the `execution_mode` attribute value.
func policyGroupExecutionModeFromAPI(executeAs schemas.PolicyGroupExecuteAs) string {
	switch executeAs {
	case schemas.PolicyGroupExecuteAsPrePlanCheck:
		return policyGroupExecutionModePrePlan
	case schemas.PolicyGroupExecuteAsPolicyCheck:
		return policyGroupExecutionModePostPlan
	default:
		// Keep unknown values as-is, so the drift is visible to the user.
		return string(executeAs)
	}
}

// stringOrEmpty returns the value of a string pointer, or an empty string if the pointer is nil.
func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

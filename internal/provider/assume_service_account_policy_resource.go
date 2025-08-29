package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
)

var (
	_ resource.Resource                     = &assumeServiceAccountPolicyResource{}
	_ resource.ResourceWithConfigure        = &assumeServiceAccountPolicyResource{}
	_ resource.ResourceWithConfigValidators = &assumeServiceAccountPolicyResource{}
	_ resource.ResourceWithImportState      = &assumeServiceAccountPolicyResource{}
)

func newAssumeServiceAccountPolicyResource() resource.Resource {
	return &assumeServiceAccountPolicyResource{}
}

type assumeServiceAccountPolicyResource struct {
	framework.ResourceWithScalrClient
}

type assumeServiceAccountPolicyResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	ServiceAccountID       types.String `tfsdk:"service_account_id"`
	ProviderID             types.String `tfsdk:"provider_id"`
	MaximumSessionDuration types.Int64  `tfsdk:"maximum_session_duration"`
	ClaimConditions        types.Set    `tfsdk:"claim_condition"`
}

type claimConditionModel struct {
	Claim    types.String `tfsdk:"claim"`
	Value    types.String `tfsdk:"value"`
	Operator types.String `tfsdk:"operator"`
}

func (r *assumeServiceAccountPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assume_service_account_policy"
}

func (r *assumeServiceAccountPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Assume Service Account Policy in Scalr.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Assume Service Account Policy.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Assume Service Account Policy.",
				Required:            true,
			},
			"service_account_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Service Account to which this policy is attached.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the Workload Identity Provider associated with this policy.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"maximum_session_duration": schema.Int64Attribute{
				MarkdownDescription: "The maximum session duration in seconds for the assumed role.",
				Optional:            true,
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"claim_condition": schema.SetNestedBlock{
				MarkdownDescription: "A set of claim conditions for the policy.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"claim": schema.StringAttribute{
							MarkdownDescription: "The claim to match.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The value to match for the claim.",
							Required:            true,
						},
						"operator": schema.StringAttribute{
							MarkdownDescription: "The operator to use for matching the claim value. Must be one of: 'eq', 'contains', 'startswith', or 'endswith'.",
							Optional:            true,
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("eq", "contains", "startswith", "endswith"),
							},
						},
					},
				},
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeAtMost(10),
				},
			},
		},
	}
}

func (r *assumeServiceAccountPolicyResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *assumeServiceAccountPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan assumeServiceAccountPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	claimConditions, diags := toScalrClaimConditions(ctx, plan.ClaimConditions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createOpts := scalr.AssumeServiceAccountPolicyCreateOptions{
		Name:            scalr.String(plan.Name.ValueString()),
		Provider:        &scalr.WorkloadIdentityProvider{ID: plan.ProviderID.ValueString()},
		ClaimConditions: claimConditions,
	}

	if !plan.MaximumSessionDuration.IsNull() && !plan.MaximumSessionDuration.IsUnknown() {
		createOpts.MaximumSessionDuration = scalr.Int(int(plan.MaximumSessionDuration.ValueInt64()))
	}

	policy, err := r.Client.AssumeServiceAccountPolicies.Create(ctx, plan.ServiceAccountID.ValueString(), createOpts)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Assume Service Account Policy",
			fmt.Sprintf("Could not create policy, unexpected error: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(policy.ID)
	plan.Name = types.StringValue(policy.Name)
	plan.MaximumSessionDuration = types.Int64Value(int64(policy.MaximumSessionDuration))

	tfClaimConditions, diags := toTerraformClaimConditions(ctx, policy.ClaimConditions)
	resp.Diagnostics.Append(diags...)
	plan.ClaimConditions = tfClaimConditions

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *assumeServiceAccountPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state assumeServiceAccountPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.Client.AssumeServiceAccountPolicies.Read(ctx, state.ServiceAccountID.ValueString(), state.ID.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading Assume Service Account Policy",
			fmt.Sprintf("Could not read policy, unexpected error: %s", err.Error()),
		)
		return
	}

	state.ID = types.StringValue(policy.ID)
	state.Name = types.StringValue(policy.Name)
	state.ProviderID = types.StringValue(policy.Provider.ID)
	state.MaximumSessionDuration = types.Int64Value(int64(policy.MaximumSessionDuration))

	claimConditions, diags := toTerraformClaimConditions(ctx, policy.ClaimConditions)
	resp.Diagnostics.Append(diags...)
	state.ClaimConditions = claimConditions

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *assumeServiceAccountPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan assumeServiceAccountPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	claimConditions, diags := toScalrClaimConditions(ctx, plan.ClaimConditions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateOpts := scalr.AssumeServiceAccountPolicyUpdateOptions{
		Name:            scalr.String(plan.Name.ValueString()),
		ClaimConditions: &claimConditions,
	}

	// Only set MaximumSessionDuration if it has a value
	if !plan.MaximumSessionDuration.IsNull() && !plan.MaximumSessionDuration.IsUnknown() {
		updateOpts.MaximumSessionDuration = scalr.Int(int(plan.MaximumSessionDuration.ValueInt64()))
	}

	policy, err := r.Client.AssumeServiceAccountPolicies.Update(
		ctx,
		plan.ServiceAccountID.ValueString(),
		plan.ID.ValueString(),
		updateOpts,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Assume Service Account Policy",
			fmt.Sprintf("Could not update policy, unexpected error: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(policy.ID)
	plan.Name = types.StringValue(policy.Name)
	plan.MaximumSessionDuration = types.Int64Value(int64(policy.MaximumSessionDuration))

	tfClaimConditions, diags := toTerraformClaimConditions(ctx, policy.ClaimConditions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ClaimConditions = tfClaimConditions

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *assumeServiceAccountPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state assumeServiceAccountPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.AssumeServiceAccountPolicies.Delete(ctx, state.ServiceAccountID.ValueString(), state.ID.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError(
			"Error deleting Assume Service Account Policy",
			fmt.Sprintf("Could not delete policy, unexpected error: %s", err.Error()),
		)
		return
	}
}

func (r *assumeServiceAccountPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ids := strings.Split(req.ID, ":")
	if len(ids) != 2 || ids[0] == "" || ids[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format 'service_account_id:policy_id'.",
		)
		return
	}

	serviceAccountID := ids[0]
	policyID := ids[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_account_id"), serviceAccountID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), policyID)...)
}

func toScalrClaimConditions(ctx context.Context, tfSet types.Set) ([]scalr.ClaimCondition, diag.Diagnostics) {
	var diags diag.Diagnostics
	var conditions []claimConditionModel

	diags.Append(tfSet.ElementsAs(ctx, &conditions, false)...)
	if diags.HasError() {
		return nil, diags
	}

	scalrConditions := make([]scalr.ClaimCondition, len(conditions))
	for i, condition := range conditions {
		var operator *string
		if !condition.Operator.IsNull() && !condition.Operator.IsUnknown() {
			operator = condition.Operator.ValueStringPointer()
		}

		scalrConditions[i] = scalr.ClaimCondition{
			Claim:    condition.Claim.ValueString(),
			Value:    condition.Value.ValueString(),
			Operator: operator,
		}
	}

	return scalrConditions, diags
}

func toTerraformClaimConditions(ctx context.Context, scalrConditions []scalr.ClaimCondition) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	conditions := make([]claimConditionModel, len(scalrConditions))

	for i, condition := range scalrConditions {
		conditions[i] = claimConditionModel{
			Claim:    types.StringValue(condition.Claim),
			Value:    types.StringValue(condition.Value),
			Operator: types.StringPointerValue(condition.Operator),
		}
	}

	tfSet, diags := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"claim":    types.StringType,
			"value":    types.StringType,
			"operator": types.StringType,
		},
	}, conditions)

	return tfSet, diags
}

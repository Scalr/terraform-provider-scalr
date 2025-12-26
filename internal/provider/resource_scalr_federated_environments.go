package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &federatedEnvironmentsResource{}
	_ resource.ResourceWithConfigure        = &federatedEnvironmentsResource{}
	_ resource.ResourceWithConfigValidators = &federatedEnvironmentsResource{}
)

func newFederatedEnvironmentsResource() resource.Resource {
	return &federatedEnvironmentsResource{}
}

type federatedEnvironmentsResource struct {
	framework.ResourceWithScalrClient
}

type federatedEnvironmentsResourceModel struct {
	EnvironmentId         types.String `tfsdk:"environment_id"`
	FederatedEnvironments types.Set    `tfsdk:"federated_environments"`
}

func (r *federatedEnvironmentsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_federated_environments"
}

func (r *federatedEnvironmentsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the list of federated environments of an environment in Scalr.",

		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "The ID of an environment that federates access to other environments.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"federated_environments": schema.SetAttribute{
				MarkdownDescription: "The list of environment identifiers that are allowed to access environment that federates access. Use `*` to allow all environments.",
				ElementType:         types.StringType,
				Required:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
			},
		},
	}
}

func (r *federatedEnvironmentsResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *federatedEnvironmentsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan federatedEnvironmentsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var federatedIDs []string
	resp.Diagnostics.Append(plan.FederatedEnvironments.ElementsAs(ctx, &federatedIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isShared := len(federatedIDs) == 1 && federatedIDs[0] == "*"

	if len(federatedIDs) == 0 {
		resp.Diagnostics.AddError("Error creating federated environments", "at least one environment identifier is required")
		return
	}

	environmentRequest := schemas.EnvironmentRequest{
		Attributes: schemas.EnvironmentAttributesRequest{
			IsFederatedToAccount: value.Set(isShared),
		},
	}
	_, er := r.ClientV2.Environment.UpdateEnvironment(ctx, plan.EnvironmentId.ValueString(), &environmentRequest, nil)
	if er != nil {
		resp.Diagnostics.AddError("Error updating federated environments", er.Error())
	}

	if isShared {
		federatedValue, _ := types.SetValueFrom(ctx, types.StringType, []string{"*"})
		result := federatedEnvironmentsResourceModel{
			EnvironmentId:         plan.EnvironmentId,
			FederatedEnvironments: federatedValue,
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
		return
	}

	federatedEnvironments := make([]schemas.Environment, len(federatedIDs))
	for i, envID := range federatedIDs {
		federatedEnvironments[i] = schemas.Environment{ID: envID}
	}

	err := r.ClientV2.Environment.AddFederatedEnvironments(ctx, plan.EnvironmentId.ValueString(), federatedEnvironments)
	if err != nil {
		resp.Diagnostics.AddError("Error adding federated environments", err.Error())
		return
	}

	federated, err := getFederatedEnvironments(ctx, r.Client, plan.EnvironmentId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving federated environments", err.Error())
		return
	}

	federatedValue, d := types.SetValueFrom(ctx, types.StringType, federated)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	result := federatedEnvironmentsResourceModel{
		EnvironmentId:         plan.EnvironmentId,
		FederatedEnvironments: federatedValue,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *federatedEnvironmentsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state federatedEnvironmentsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environment, err := r.ClientV2.Environment.GetEnvironment(ctx, state.EnvironmentId.ValueString(), nil)
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving environment", err.Error())
		return
	}

	if environment.Attributes.IsFederatedToAccount {
		federatedValue, _ := types.SetValueFrom(ctx, types.StringType, []string{"*"})
		result := federatedEnvironmentsResourceModel{
			EnvironmentId:         state.EnvironmentId,
			FederatedEnvironments: federatedValue,
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
		return
	}

	federated, err := getFederatedEnvironments(ctx, r.Client, state.EnvironmentId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving federated environments", err.Error())
		return
	}

	federatedValue, d := types.SetValueFrom(ctx, types.StringType, federated)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	result := federatedEnvironmentsResourceModel{
		EnvironmentId:         state.EnvironmentId,
		FederatedEnvironments: federatedValue,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *federatedEnvironmentsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state federatedEnvironmentsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planFederated, stateFederated []string
	resp.Diagnostics.Append(plan.FederatedEnvironments.ElementsAs(ctx, &planFederated, false)...)
	resp.Diagnostics.Append(state.FederatedEnvironments.ElementsAs(ctx, &stateFederated, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isShared := len(planFederated) == 1 && planFederated[0] == "*"

	environmentRequest := schemas.EnvironmentRequest{
		Attributes: schemas.EnvironmentAttributesRequest{
			IsFederatedToAccount: value.Set(isShared),
		},
	}
	_, er := r.ClientV2.Environment.UpdateEnvironment(ctx, plan.EnvironmentId.ValueString(), &environmentRequest, nil)
	if er != nil {
		resp.Diagnostics.AddError("Error updating federated environments", er.Error())
	}

	if isShared {
		federatedValue, _ := types.SetValueFrom(ctx, types.StringType, []string{"*"})
		result := federatedEnvironmentsResourceModel{
			EnvironmentId:         plan.EnvironmentId,
			FederatedEnvironments: federatedValue,
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
		return
	}

	federatedToAdd, federatedToRemove := diff(stateFederated, planFederated)

	if len(federatedToAdd) > 0 {
		e := make([]schemas.Environment, len(federatedToAdd))
		for i, env := range federatedToAdd {
			e[i] = schemas.Environment{ID: env}
		}
		err := r.ClientV2.Environment.AddFederatedEnvironments(ctx, plan.EnvironmentId.ValueString(), e)
		if err != nil {
			resp.Diagnostics.AddError("Error adding federated environments", err.Error())
			return
		}
	}

	if len(federatedToRemove) > 0 {
		e := make([]schemas.Environment, len(federatedToRemove))
		for i, env := range federatedToRemove {
			e[i] = schemas.Environment{ID: env}
		}
		err := r.ClientV2.Environment.DeleteFederatedEnvironment(ctx, plan.EnvironmentId.ValueString(), e)
		if err != nil {
			resp.Diagnostics.AddError("Error removing federated environments", err.Error())
			return
		}
	}

	federated, err := getFederatedEnvironments(ctx, r.Client, plan.EnvironmentId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving federated environments", err.Error())
		return
	}

	federatedValue, d := types.SetValueFrom(ctx, types.StringType, federated)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	result := federatedEnvironmentsResourceModel{
		EnvironmentId:         plan.EnvironmentId,
		FederatedEnvironments: federatedValue,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *federatedEnvironmentsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var state federatedEnvironmentsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var federatedIDs []string
	resp.Diagnostics.Append(state.FederatedEnvironments.ElementsAs(ctx, &federatedIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(federatedIDs) == 1 && federatedIDs[0] == "*" {
		environmentRequest := schemas.EnvironmentRequest{
			Attributes: schemas.EnvironmentAttributesRequest{
				IsFederatedToAccount: value.Set(false),
			},
		}
		_, er := r.ClientV2.Environment.UpdateEnvironment(ctx, state.EnvironmentId.ValueString(), &environmentRequest, nil)
		if er != nil {
			resp.Diagnostics.AddError("Error updating federated environments", er.Error())
		}
		return
	}

	if len(federatedIDs) > 0 {
		e := make([]schemas.Environment, len(federatedIDs))
		for i, env := range federatedIDs {
			e[i] = schemas.Environment{ID: env}
		}
		err := r.ClientV2.Environment.DeleteFederatedEnvironment(ctx, state.EnvironmentId.ValueString(), e)
		if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
			resp.Diagnostics.AddError("Error removing federated environments", err.Error())
			return
		}
	}
}

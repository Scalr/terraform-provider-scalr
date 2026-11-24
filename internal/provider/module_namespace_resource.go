package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

var (
	_ resource.Resource                     = &moduleNamespaceResource{}
	_ resource.ResourceWithConfigure        = &moduleNamespaceResource{}
	_ resource.ResourceWithConfigValidators = &moduleNamespaceResource{}
	_ resource.ResourceWithImportState      = &moduleNamespaceResource{}
	_ resource.ResourceWithModifyPlan       = &moduleNamespaceResource{}
)

func newModuleNamespaceResource() resource.Resource {
	return &moduleNamespaceResource{}
}

type moduleNamespaceResource struct {
	framework.ResourceWithScalrClient
}

type moduleNamespaceResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	IsShared     types.Bool   `tfsdk:"is_shared"`
	Environments types.Set    `tfsdk:"environments"`
	Owners       types.Set    `tfsdk:"owners"`
}

func moduleNamespaceResourceModelFromAPI(
	ctx context.Context,
	namespace *schemas.ModuleNamespace,
	prior *moduleNamespaceResourceModel,
) (*moduleNamespaceResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &moduleNamespaceResourceModel{
		ID:       types.StringValue(namespace.ID),
		Name:     types.StringValue(namespace.Attributes.Name),
		IsShared: types.BoolValue(namespace.Attributes.IsShared),
	}

	var priorEnvironments, priorOwners *types.Set
	if prior != nil {
		priorEnvironments = &prior.Environments
		priorOwners = &prior.Owners
	}

	environments, d := framework.FlattenRelationshipIDsSet(
		ctx,
		namespace.Relationships.Environments,
		func(e *schemas.Environment) string { return e.ID },
		priorEnvironments,
	)
	diags.Append(d...)
	model.Environments = environments

	owners, d := framework.FlattenRelationshipIDsSet(
		ctx,
		namespace.Relationships.Owners,
		func(t *schemas.Team) string { return t.ID },
		priorOwners,
	)
	diags.Append(d...)
	model.Owners = owners

	return model, diags
}

func (r *moduleNamespaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_namespace"
}

func (r *moduleNamespaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of module namespaces in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the module namespace.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidation.StringIsNamespaceName(),
				},
			},
			"is_shared": schema.BoolAttribute{
				MarkdownDescription: "Whether the module namespace is shared.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"environments": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Set of environment IDs associated with the module namespace.",
				DeprecationMessage: "Setting this attribute is deprecated." +
					" Use the `scalr_module_namespace_environment` resource instead.",
				Validators: []validator.Set{moduleNamespaceEnvironmentsValidator{}},
			},
			"owners": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of team IDs that own the module namespace.",
			},
		},
	}
}

func (r *moduleNamespaceResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *moduleNamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan moduleNamespaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.ModuleNamespaceRequest{
		Attributes: schemas.ModuleNamespaceAttributesRequest{
			Name:     value.Set(plan.Name.ValueString()),
			IsShared: framework.SetIfKnownBool(plan.IsShared),
		},
	}

	if !plan.Environments.IsUnknown() && !plan.Environments.IsNull() {
		environments, diags := framework.ExpandRelationshipIDsSet(
			ctx, plan.Environments, func(id string) schemas.Environment {
				return schemas.Environment{ID: id}
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		opts.Relationships.Environments = value.Set(environments)
	}

	owners, diags := framework.ExpandRelationshipIDsSet(
		ctx, plan.Owners, func(id string) schemas.Team {
			return schemas.Team{ID: id}
		},
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	opts.Relationships.Owners = value.Set(owners)

	namespace, err := r.ClientV2.ModuleNamespace.CreateModuleNamespace(ctx, &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating module namespace", err.Error())
		return
	}

	result, diags := moduleNamespaceResourceModelFromAPI(ctx, namespace, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *moduleNamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state moduleNamespaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespace, err := r.ClientV2.ModuleNamespace.GetModuleNamespace(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving module namespace", err.Error())
		return
	}

	result, diags := moduleNamespaceResourceModelFromAPI(ctx, namespace, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *moduleNamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state moduleNamespaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.ModuleNamespaceRequest{
		Attributes: schemas.ModuleNamespaceAttributesRequest{
			IsShared: framework.SetIfKnownBool(plan.IsShared),
		},
	}

	// An unchanged set is not sent, so links made elsewhere survive an update.
	if !plan.Environments.Equal(state.Environments) &&
		!plan.Environments.IsUnknown() && !plan.Environments.IsNull() {
		environments, diags := framework.ExpandRelationshipIDsSet(
			ctx, plan.Environments, func(id string) schemas.Environment {
				return schemas.Environment{ID: id}
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		opts.Relationships.Environments = value.Set(environments)
	}

	if !plan.Owners.Equal(state.Owners) {
		owners, diags := framework.ExpandRelationshipIDsSet(
			ctx, plan.Owners, func(id string) schemas.Team {
				return schemas.Team{ID: id}
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		opts.Relationships.Owners = value.Set(owners)
	}

	namespace, err := r.ClientV2.ModuleNamespace.UpdateModuleNamespace(ctx, plan.ID.ValueString(), &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating module namespace", err.Error())
		return
	}

	result, diags := moduleNamespaceResourceModelFromAPI(ctx, namespace, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *moduleNamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state moduleNamespaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.ModuleNamespace.DeleteModuleNamespace(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting module namespace", err.Error())
		return
	}
}

// ModifyPlan clears the planned environment links when the namespace becomes shared.
func (r *moduleNamespaceResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		// Create or destroy, nothing to adjust.
		return
	}

	var config, plan, state moduleNamespaceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Environments.IsNull() || state.IsShared.ValueBool() || !plan.IsShared.ValueBool() {
		return
	}
	if plan.Environments.IsUnknown() || len(plan.Environments.Elements()) == 0 {
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(
		ctx, path.Root("environments"), types.SetValueMust(types.StringType, nil),
	)...)
}

func (r *moduleNamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Compile-time interface check
var _ validator.Set = moduleNamespaceEnvironmentsValidator{}

type moduleNamespaceEnvironmentsValidator struct{}

func (v moduleNamespaceEnvironmentsValidator) Description(_ context.Context) string {
	return "environments can't be specified for a shared module namespace"
}

func (v moduleNamespaceEnvironmentsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v moduleNamespaceEnvironmentsValidator) ValidateSet(
	ctx context.Context,
	req validator.SetRequest,
	resp *validator.SetResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var isShared types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("is_shared"), &isShared)...)

	if len(req.ConfigValue.Elements()) > 0 && isShared.ValueBool() {
		resp.Diagnostics.Append(
			diag.NewAttributeErrorDiagnostic(
				req.Path,
				"Invalid Attribute Combination",
				"Environments can't be specified for a shared module namespace.",
			),
		)
	}
}

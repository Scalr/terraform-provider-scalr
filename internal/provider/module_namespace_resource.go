package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

var (
	_ resource.Resource                     = &moduleNamespaceResource{}
	_ resource.ResourceWithConfigure        = &moduleNamespaceResource{}
	_ resource.ResourceWithConfigValidators = &moduleNamespaceResource{}
	_ resource.ResourceWithImportState      = &moduleNamespaceResource{}
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
					validation.StringIsNotWhiteSpace(),
					validation.StringIsNamespaceName(),
				},
			},
			"is_shared": schema.BoolAttribute{
				MarkdownDescription: "Whether the module namespace is shared.",
				Optional:            true,
				Computed:            true,
			},
			"environments": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of environment IDs associated with the module namespace.",
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

	var environments []string
	resp.Diagnostics.Append(plan.Environments.ElementsAs(ctx, &environments, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var owners []string
	resp.Diagnostics.Append(plan.Owners.ElementsAs(ctx, &owners, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert environment IDs to Environment objects
	var environmentObjects []*scalr.Environment
	for _, envID := range environments {
		environmentObjects = append(environmentObjects, &scalr.Environment{ID: envID})
	}

	// Convert owner IDs to Team objects
	var ownerObjects []*scalr.Team
	for _, ownerID := range owners {
		ownerObjects = append(ownerObjects, &scalr.Team{ID: ownerID})
	}

	opts := scalr.ModuleNamespaceCreateOptions{
		Name:         plan.Name.ValueStringPointer(),
		IsShared:     plan.IsShared.ValueBoolPointer(),
		Environments: environmentObjects,
		Owners:       ownerObjects,
	}

	namespace, err := r.Client.ModuleNamespaces.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating module namespace", err.Error())
		return
	}

	plan.ID = types.StringValue(namespace.ID)
	plan.Name = types.StringValue(namespace.Name)
	plan.IsShared = types.BoolValue(namespace.IsShared)

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
		plan.Environments = environmentsSet
	} else {
		plan.Environments = types.SetNull(types.StringType)
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
		plan.Owners = ownersSet
	} else {
		plan.Owners = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *moduleNamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state moduleNamespaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespace, err := r.Client.ModuleNamespaces.Read(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving module namespace", err.Error())
		return
	}

	state.Name = types.StringValue(namespace.Name)
	state.IsShared = types.BoolValue(namespace.IsShared)

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
		state.Environments = environmentsSet
	} else {
		state.Environments = types.SetNull(types.StringType)
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
		state.Owners = ownersSet
	} else {
		state.Owners = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *moduleNamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan moduleNamespaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var environments []string
	resp.Diagnostics.Append(plan.Environments.ElementsAs(ctx, &environments, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var owners []string
	resp.Diagnostics.Append(plan.Owners.ElementsAs(ctx, &owners, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert environment IDs to Environment objects
	var environmentObjects []*scalr.Environment
	for _, envID := range environments {
		environmentObjects = append(environmentObjects, &scalr.Environment{ID: envID})
	}

	// Convert owner IDs to Team objects
	var ownerObjects []*scalr.Team
	for _, ownerID := range owners {
		ownerObjects = append(ownerObjects, &scalr.Team{ID: ownerID})
	}

	opts := scalr.ModuleNamespaceUpdateOptions{
		IsShared:     plan.IsShared.ValueBoolPointer(),
		Environments: environmentObjects,
		Owners:       ownerObjects,
	}

	namespace, err := r.Client.ModuleNamespaces.Update(ctx, plan.ID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating module namespace", err.Error())
		return
	}

	plan.Name = types.StringValue(namespace.Name)
	plan.IsShared = types.BoolValue(namespace.IsShared)

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
		plan.Environments = environmentsSet
	} else {
		plan.Environments = types.SetNull(types.StringType)
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
		plan.Owners = ownersSet
	} else {
		plan.Owners = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *moduleNamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state moduleNamespaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.ModuleNamespaces.Delete(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting module namespace", err.Error())
		return
	}
}

func (r *moduleNamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

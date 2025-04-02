package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &variableResource{}
	_ resource.ResourceWithConfigure        = &variableResource{}
	_ resource.ResourceWithConfigValidators = &variableResource{}
	_ resource.ResourceWithImportState      = &variableResource{}
)

func newVariableResource() resource.Resource {
	return &variableResource{}
}

// variableResource defines the resource implementation.
type variableResource struct {
	framework.ResourceWithScalrClient
}

func (r *variableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_variable"
}

func (r *variableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = *variableResourceSchema()
}

func (r *variableResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	// If needed, add config validation logic here,
	// or remove this method if no additional validation is needed.
	return []resource.ConfigValidator{}
}

func (r *variableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan variableResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.VariableCreateOptions{
		Key:         plan.Key.ValueStringPointer(),
		Value:       plan.Value.ValueStringPointer(),
		Description: plan.Description.ValueStringPointer(),
		Category:    ptr(scalr.CategoryType(plan.Category.ValueString())),
		HCL:         plan.HCL.ValueBoolPointer(),
		Sensitive:   plan.Sensitive.ValueBoolPointer(),
		Final:       plan.Final.ValueBoolPointer(),
		Account:     &scalr.Account{ID: plan.AccountID.ValueString()},
		QueryOptions: &scalr.VariableWriteQueryOptions{
			Force:   plan.Force.ValueBoolPointer(),
			Include: ptr("updated-by"),
		},
	}

	if !plan.WorkspaceID.IsUnknown() && !plan.WorkspaceID.IsNull() {
		opts.Workspace = &scalr.Workspace{ID: plan.WorkspaceID.ValueString()}
	}

	if !plan.EnvironmentID.IsUnknown() && !plan.EnvironmentID.IsNull() {
		opts.Environment = &scalr.Environment{ID: plan.EnvironmentID.ValueString()}
	}

	variable, err := r.Client.Variables.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating variable", err.Error())
		return
	}

	result, diags := variableResourceModelFromAPI(ctx, variable, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *variableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state variableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	variable, err := r.Client.Variables.Read(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving variable", err.Error())
		return
	}

	result, diags := variableResourceModelFromAPI(ctx, variable, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *variableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read plan & state data
	var plan, state variableResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.VariableUpdateOptions{
		Key:         plan.Key.ValueStringPointer(),
		Value:       plan.Value.ValueStringPointer(),
		HCL:         plan.HCL.ValueBoolPointer(),
		Sensitive:   plan.Sensitive.ValueBoolPointer(),
		Description: plan.Description.ValueStringPointer(),
		Final:       plan.Final.ValueBoolPointer(),
		QueryOptions: &scalr.VariableWriteQueryOptions{
			Force:   plan.Force.ValueBoolPointer(),
			Include: ptr("updated-by"),
		},
	}

	// Update existing resource
	variable, err := r.Client.Variables.Update(ctx, plan.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating variable", err.Error())
		return
	}

	result, diags := variableResourceModelFromAPI(ctx, variable, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *variableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state variableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.Variables.Delete(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting variable", err.Error())
		return
	}
}

func (r *variableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

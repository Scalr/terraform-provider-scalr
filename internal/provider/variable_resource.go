package provider

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
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
	_ resource.ResourceWithUpgradeState     = &variableResource{}
)

func newVariableResource() resource.Resource {
	return &variableResource{}
}

// variableResource defines the resource implementation.
type variableResource struct {
	framework.ResourceWithScalrClient
}

type privateMeta struct {
	IsWriteOnly bool `json:"is_write_only"`
}

func (r *variableResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_variable"
}

func (r *variableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = *variableResourceSchema()
}

func (r *variableResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.RequiredTogether(
			path.MatchRoot("value_wo"),
			path.MatchRoot("value_wo_version"),
		),
	}
}

func (r *variableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config, plan variableResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine which value to send: value_wo takes precedence over value
	var valueToSend *string
	isWriteOnly := !config.ValueWO.IsNull() && !config.ValueWO.IsUnknown()
	if isWriteOnly {
		valueToSend = config.ValueWO.ValueStringPointer()
	} else {
		valueToSend = plan.Value.ValueStringPointer()
	}

	opts := scalr.VariableCreateOptions{
		Key:         plan.Key.ValueStringPointer(),
		Value:       valueToSend,
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

	result, diags := variableResourceModelFromAPI(ctx, variable, &plan, isWriteOnly)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Persist private metadata
	metaBytes, err := json.Marshal(privateMeta{IsWriteOnly: isWriteOnly})
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal private metadata", err.Error())
		return
	}

	resp.Private.SetKey(ctx, "meta", metaBytes)
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

	metaBytes, diags := req.Private.GetKey(ctx, "meta")
	resp.Diagnostics.Append(diags...)

	var meta privateMeta
	if metaBytes != nil {
		_ = json.Unmarshal(metaBytes, &meta)
	}

	result, diags := variableResourceModelFromAPI(ctx, variable, &state, meta.IsWriteOnly)
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
	var config, plan, state variableResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.VariableUpdateOptions{
		Key:         plan.Key.ValueStringPointer(),
		HCL:         plan.HCL.ValueBoolPointer(),
		Sensitive:   plan.Sensitive.ValueBoolPointer(),
		Description: plan.Description.ValueStringPointer(),
		Final:       plan.Final.ValueBoolPointer(),
		QueryOptions: &scalr.VariableWriteQueryOptions{
			Force:   plan.Force.ValueBoolPointer(),
			Include: ptr("updated-by"),
		},
	}

	isWriteOnly := !config.ValueWO.IsNull() && !config.ValueWO.IsUnknown()
	if isWriteOnly {
		// Only update write-only value if the version attribute has changed
		if !plan.ValueWOVersion.Equal(state.ValueWOVersion) {
			opts.Value = config.ValueWO.ValueStringPointer()
		}
	} else if !plan.Value.Equal(state.Value) {
		opts.Value = plan.Value.ValueStringPointer()
	}

	// Update existing resource
	variable, err := r.Client.Variables.Update(ctx, plan.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating variable", err.Error())
		return
	}

	result, diags := variableResourceModelFromAPI(ctx, variable, &plan, isWriteOnly)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Persist private metadata
	metaBytes, err := json.Marshal(privateMeta{IsWriteOnly: isWriteOnly})
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal private metadata", err.Error())
		return
	}

	resp.Private.SetKey(ctx, "meta", metaBytes)
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

func (r *variableResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   variableResourceSchemaV0(),
			StateUpgrader: upgradeVariableResourceStateV0toV3(r.Client),
		},
		1: {
			PriorSchema:   variableResourceSchemaV1(),
			StateUpgrader: upgradeVariableResourceStateV1toV3(r.Client),
		},
		2: {
			PriorSchema:   variableResourceSchemaV2(),
			StateUpgrader: upgradeVariableResourceStateV2toV3(r.Client),
		},
	}
}

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/client"
	varops "github.com/scalr/go-scalr/v2/scalr/ops/variable"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

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

func isVarSetVariable(ctx context.Context, plan framework.AttrGetter) bool {
	var varID, varSetID types.String
	plan.GetAttribute(ctx, path.Root("id"), &varID)
	plan.GetAttribute(ctx, path.Root("var_set_id"), &varSetID)
	// It is a variable set variable if var_set_id is set,
	// or it has the id, and it is a variable set variable ID (vsvar-...).
	// The latter condition is to support importing variable set variables with ID passthrough to the Read method.
	return !varSetID.IsNull() || !varID.IsNull() && strings.HasPrefix(varID.ValueString(), "vsvar-")
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
		resourcevalidator.Conflicting(
			path.MatchRoot("var_set_id"),
			path.MatchRoot("environment_id"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("var_set_id"),
			path.MatchRoot("workspace_id"),
		),
	}
}

func (r *variableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if isVarSetVariable(ctx, req.Plan) {
		r.createVarSetVariable(ctx, req, resp)
	} else {
		r.createClassicVariable(ctx, req, resp)
	}
}

func (r *variableResource) createClassicVariable(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
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

	createReq := schemas.VariableRequest{
		Attributes: schemas.VariableAttributesRequest{
			Key:         value.Set(plan.Key.ValueString()),
			Value:       value.SetPtrMaybe(valueToSend),
			Description: value.SetPtrMaybe(plan.Description.ValueStringPointer()),
			Category:    value.Set(schemas.VariableCategory(plan.Category.ValueString())),
			Hcl:         value.Set(plan.HCL.ValueBool()),
			Sensitive:   value.Set(plan.Sensitive.ValueBool()),
			Final:       value.Set(plan.Final.ValueBool()),
		},
	}

	if !plan.WorkspaceID.IsUnknown() && !plan.WorkspaceID.IsNull() {
		createReq.Relationships.Workspace = value.Set(schemas.Workspace{ID: plan.WorkspaceID.ValueString()})
	}

	if !plan.EnvironmentID.IsUnknown() && !plan.EnvironmentID.IsNull() {
		createReq.Relationships.Environment = value.Set(schemas.Environment{ID: plan.EnvironmentID.ValueString()})
	}

	createOpts := varops.CreateVariableOptions{
		Force:   plan.Force.ValueBool(),
		Include: []string{"updated-by"},
	}

	variable, err := r.ClientV2.Variable.CreateVariable(ctx, &createReq, &createOpts)
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

func (r *variableResource) createVarSetVariable(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
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

	createReq := schemas.VariableSetVariableRequest{
		Attributes: schemas.VariableSetVariableAttributesRequest{
			Key:         value.Set(plan.Key.ValueString()),
			Value:       value.SetPtrMaybe(valueToSend),
			Description: value.SetPtrMaybe(plan.Description.ValueStringPointer()),
			Category:    value.Set(schemas.VariableSetVariableCategory(plan.Category.ValueString())),
			Hcl:         value.Set(plan.HCL.ValueBool()),
			Sensitive:   value.Set(plan.Sensitive.ValueBool()),
			Final:       value.Set(plan.Final.ValueBool()),
		},
	}

	if !plan.VarSetID.IsUnknown() && !plan.VarSetID.IsNull() {
		createReq.Relationships.VarSet = value.Set(schemas.VariableSet{ID: plan.VarSetID.ValueString()})
	}

	variable, err := r.ClientV2.VariableSetVariable.CreateVarSetVariable(ctx, &createReq, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating variable", err.Error())
		return
	}

	result, diags := variableSetVariableResourceModelFromAPI(ctx, variable, &plan, isWriteOnly)
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
	if isVarSetVariable(ctx, req.State) {
		r.readVarSetVariable(ctx, req, resp)
	} else {
		r.readClassicVariable(ctx, req, resp)
	}
}

func (r *variableResource) readClassicVariable(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state variableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	variable, err := r.ClientV2.Variable.GetVariable(
		ctx,
		state.Id.ValueString(),
		&varops.GetVariableOptions{Include: []string{"updated-by"}},
	)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
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

func (r *variableResource) readVarSetVariable(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	// Get current state
	var state variableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	variable, err := r.ClientV2.VariableSetVariable.GetVarSetVariable(ctx, state.Id.ValueString(), nil)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
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

	result, diags := variableSetVariableResourceModelFromAPI(ctx, variable, &state, meta.IsWriteOnly)
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
	if isVarSetVariable(ctx, req.Plan) {
		r.updateVarSetVariable(ctx, req, resp)
	} else {
		r.updateClassicVariable(ctx, req, resp)
	}
}

func (r *variableResource) updateClassicVariable(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Read plan & state data
	var config, plan, state variableResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := schemas.VariableRequest{}

	if !plan.Key.Equal(state.Key) {
		updateReq.Attributes.Key = value.Set(plan.Key.ValueString())
	}

	if !plan.Description.Equal(state.Description) {
		updateReq.Attributes.Description = value.SetPtr(plan.Description.ValueStringPointer())
	}

	if !plan.HCL.Equal(state.HCL) {
		updateReq.Attributes.Hcl = value.Set(plan.HCL.ValueBool())
	}

	if !plan.Sensitive.Equal(state.Sensitive) {
		updateReq.Attributes.Sensitive = value.Set(plan.Sensitive.ValueBool())
	}

	if !plan.Final.Equal(state.Final) {
		updateReq.Attributes.Final = value.Set(plan.Final.ValueBool())
	}

	metaBytes, diags := req.Private.GetKey(ctx, "meta")
	resp.Diagnostics.Append(diags...)

	var meta privateMeta
	if metaBytes != nil {
		_ = json.Unmarshal(metaBytes, &meta)
	}

	isWriteOnly := !config.ValueWO.IsNull() && !config.ValueWO.IsUnknown()
	isWriteOnlyChanged := isWriteOnly != meta.IsWriteOnly
	if isWriteOnly {
		// Only update write-only value if the version attribute has changed
		if !plan.ValueWOVersion.Equal(state.ValueWOVersion) || isWriteOnlyChanged {
			updateReq.Attributes.Value = value.SetPtr(config.ValueWO.ValueStringPointer())
		}
	} else if !plan.Value.Equal(state.Value) || isWriteOnlyChanged {
		updateReq.Attributes.Value = value.SetPtr(plan.Value.ValueStringPointer())
	}

	updateOpts := varops.UpdateVariableOptions{
		Force:   plan.Force.ValueBool(),
		Include: []string{"updated-by"},
	}

	// Update existing resource
	variable, err := r.ClientV2.Variable.UpdateVariable(ctx, plan.Id.ValueString(), &updateReq, &updateOpts)
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
	metaBytes, err = json.Marshal(privateMeta{IsWriteOnly: isWriteOnly})
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal private metadata", err.Error())
		return
	}

	resp.Private.SetKey(ctx, "meta", metaBytes)
}

func (r *variableResource) updateVarSetVariable(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	// Read plan & state data
	var config, plan, state variableResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := schemas.VariableSetVariableRequest{}

	if !plan.Key.Equal(state.Key) {
		updateReq.Attributes.Key = value.Set(plan.Key.ValueString())
	}

	if !plan.Description.Equal(state.Description) {
		updateReq.Attributes.Description = value.SetPtr(plan.Description.ValueStringPointer())
	}

	if !plan.HCL.Equal(state.HCL) {
		updateReq.Attributes.Hcl = value.Set(plan.HCL.ValueBool())
	}

	if !plan.Sensitive.Equal(state.Sensitive) {
		updateReq.Attributes.Sensitive = value.Set(plan.Sensitive.ValueBool())
	}

	if !plan.Final.Equal(state.Final) {
		updateReq.Attributes.Final = value.Set(plan.Final.ValueBool())
	}

	metaBytes, diags := req.Private.GetKey(ctx, "meta")
	resp.Diagnostics.Append(diags...)

	var meta privateMeta
	if metaBytes != nil {
		_ = json.Unmarshal(metaBytes, &meta)
	}

	isWriteOnly := !config.ValueWO.IsNull() && !config.ValueWO.IsUnknown()
	isWriteOnlyChanged := isWriteOnly != meta.IsWriteOnly
	if isWriteOnly {
		// Only update write-only value if the version attribute has changed
		if !plan.ValueWOVersion.Equal(state.ValueWOVersion) || isWriteOnlyChanged {
			updateReq.Attributes.Value = value.SetPtr(config.ValueWO.ValueStringPointer())
		}
	} else if !plan.Value.Equal(state.Value) || isWriteOnlyChanged {
		updateReq.Attributes.Value = value.SetPtr(plan.Value.ValueStringPointer())
	}

	// Update existing resource
	variable, err := r.ClientV2.VariableSetVariable.UpdateVarSetVariable(ctx, plan.Id.ValueString(), &updateReq, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating variable", err.Error())
		return
	}

	result, diags := variableSetVariableResourceModelFromAPI(ctx, variable, &plan, isWriteOnly)
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
	metaBytes, err = json.Marshal(privateMeta{IsWriteOnly: isWriteOnly})
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal private metadata", err.Error())
		return
	}

	resp.Private.SetKey(ctx, "meta", metaBytes)
}

func (r *variableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if isVarSetVariable(ctx, req.State) {
		r.deleteVarSetVariable(ctx, req, resp)
	} else {
		r.deleteClassicVariable(ctx, req, resp)
	}
}

func (r *variableResource) deleteClassicVariable(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Get current state
	var state variableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.Variable.DeleteVariable(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting variable", err.Error())
		return
	}
}

func (r *variableResource) deleteVarSetVariable(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	// Get current state
	var state variableResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.VariableSetVariable.DeleteVarSetVariable(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting variable", err.Error())
		return
	}
}

func (r *variableResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *variableResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   variableResourceSchemaV0(),
			StateUpgrader: upgradeVariableResourceStateV0toV3(r.ClientV2),
		},
		1: {
			PriorSchema:   variableResourceSchemaV1(),
			StateUpgrader: upgradeVariableResourceStateV1toV3(r.ClientV2),
		},
		2: {
			PriorSchema:   variableResourceSchemaV2(),
			StateUpgrader: upgradeVariableResourceStateV2toV3(r.ClientV2),
		},
	}
}

package provider

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

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

func (r *variableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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

func (r *variableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

func (r *variableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

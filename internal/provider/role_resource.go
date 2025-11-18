package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
)

// Compile-time interface checks
var (
	_ resource.Resource                 = &roleResource{}
	_ resource.ResourceWithConfigure    = &roleResource{}
	_ resource.ResourceWithImportState  = &roleResource{}
	_ resource.ResourceWithUpgradeState = &roleResource{}
)

func newRoleResource() resource.Resource {
	return &roleResource{}
}

// roleResource defines the resource implementation.
type roleResource struct {
	framework.ResourceWithScalrClient
}

// roleResourceModel describes the resource data model.
type roleResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	AccountID   types.String `tfsdk:"account_id"`
	IsSystem    types.Bool   `tfsdk:"is_system"`
	Description types.String `tfsdk:"description"`
	Permissions types.Set    `tfsdk:"permissions"`
}

func roleResourceModelFromAPI(ctx context.Context, role *schemas.Role) (*roleResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &roleResourceModel{
		Id:          types.StringValue(role.ID),
		Name:        types.StringValue(role.Attributes.Name),
		AccountID:   types.StringNull(),
		IsSystem:    types.BoolValue(role.Attributes.IsSystem),
		Description: types.StringPointerValue(role.Attributes.Description),
		Permissions: types.SetNull(types.StringType),
	}

	if role.Relationships.Account != nil {
		model.AccountID = types.StringValue(role.Relationships.Account.ID)
	}

	permissions := make([]string, len(role.Relationships.Permissions))
	for i, permission := range role.Relationships.Permissions {
		permissions[i] = permission.ID
	}
	permissionsValue, d := types.SetValueFrom(ctx, types.StringType, permissions)
	diags.Append(d...)
	model.Permissions = permissionsValue

	return model, diags
}

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = *roleResourceSchema()
}

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.RoleRequest{
		Attributes: schemas.RoleAttributesRequest{
			Name:        value.Set(plan.Name.ValueString()),
			Description: framework.SetIfKnownString(plan.Description),
		},
	}

	if !plan.Permissions.IsUnknown() && !plan.Permissions.IsNull() {
		var permissionIDs []string
		resp.Diagnostics.Append(plan.Permissions.ElementsAs(ctx, &permissionIDs, false)...)

		permissions := make([]schemas.Permission, len(permissionIDs))
		for i, id := range permissionIDs {
			permissions[i] = schemas.Permission{ID: id}
		}

		opts.Relationships.Permissions = value.Set(permissions)
	}

	role, err := r.ClientV2.Role.CreateRole(ctx, &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating role", err.Error())
		return
	}

	result, d := roleResourceModelFromAPI(ctx, role)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	role, err := r.ClientV2.Role.GetRole(ctx, state.Id.ValueString(), nil)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving role", err.Error())
		return
	}

	result, d := roleResourceModelFromAPI(ctx, role)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state roleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.RoleRequest{}

	if !plan.Name.Equal(state.Name) {
		opts.Attributes.Name = value.Set(plan.Name.ValueString())
	}

	if !plan.Description.Equal(state.Description) {
		opts.Attributes.Description = framework.SetIfKnownString(plan.Description)
	}

	if !plan.Permissions.Equal(state.Permissions) && !plan.Permissions.IsUnknown() && !plan.Permissions.IsNull() {
		var permissionIDs []string
		resp.Diagnostics.Append(plan.Permissions.ElementsAs(ctx, &permissionIDs, false)...)

		permissions := make([]schemas.Permission, len(permissionIDs))
		for i, id := range permissionIDs {
			permissions[i] = schemas.Permission{ID: id}
		}

		opts.Relationships.Permissions = value.Set(permissions)
	}

	// Update existing resource
	role, err := r.ClientV2.Role.UpdateRole(ctx, plan.Id.ValueString(), &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating role", err.Error())
		return
	}

	result, d := roleResourceModelFromAPI(ctx, role)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.Role.DeleteRole(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting role", err.Error())
		return
	}
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Lookup the role before proceeding with import to ensure it is not a system role
	role, err := r.ClientV2.Role.GetRole(ctx, req.ID, nil)
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error retrieving role", err.Error())
		return
	}
	if err == nil && role.Attributes.IsSystem {
		resp.Diagnostics.AddError(
			"Cannot import system role",
			"This role is a read-only system role and cannot be imported as manageable resource."+
				" Use the scalr_role data source instead.",
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *roleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   roleResourceSchemaV0(),
			StateUpgrader: upgradeRoleResourceStateV0toV1,
		},
	}
}

package provider

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func upgradeRoleResourceStateV0toV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var permissions []string
	resp.Diagnostics.Append(state.Permissions.ElementsAs(ctx, &permissions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !slices.Contains(permissions, "accounts:set-quotas") &&
		slices.Contains(permissions, "global-scope:read") &&
		slices.Contains(permissions, "accounts:update") {
		permissions = append(permissions, "accounts:set-quotas")
		permissionsValue, d := types.SetValueFrom(ctx, types.StringType, permissions)
		resp.Diagnostics.Append(d...)
		state.Permissions = permissionsValue
	}
	d := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(d...)
}

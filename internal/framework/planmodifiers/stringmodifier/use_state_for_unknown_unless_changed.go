package stringmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseStateForUnknownUnlessChanged uses the state value for an unknown planned value,
// unless any of the given attributes is planned to change.
func UseStateForUnknownUnlessChanged(paths ...path.Path) planmodifier.String {
	return useStateForUnknownUnlessChangedModifier{paths: paths}
}

type useStateForUnknownUnlessChangedModifier struct {
	paths []path.Path
}

func (m useStateForUnknownUnlessChangedModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Use state value if planned value is unknown, unless any of %v changes.", m.paths)
}

func (m useStateForUnknownUnlessChangedModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateForUnknownUnlessChangedModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing on resource creation or destroy.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Do nothing if there is a known planned value or an unknown configuration value.
	if !req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() || req.StateValue.IsNull() {
		return
	}

	for _, p := range m.paths {
		var planValue, stateValue attr.Value
		resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, p, &planValue)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, p, &stateValue)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !planValue.Equal(stateValue) {
			return
		}
	}

	resp.PlanValue = req.StateValue
}

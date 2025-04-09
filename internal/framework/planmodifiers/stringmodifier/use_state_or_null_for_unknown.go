package stringmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func UseStateOrNullForUnknown() planmodifier.String {
	return useStateOrNullForUnknownModifier{}
}

type useStateOrNullForUnknownModifier struct{}

func (m useStateOrNullForUnknownModifier) Description(_ context.Context) string {
	return "Use state value if planned value is unknown, even if it is null."
}

func (m useStateOrNullForUnknownModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m useStateOrNullForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Do nothing on resource creation.
	if req.State.Raw.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value.
	if req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

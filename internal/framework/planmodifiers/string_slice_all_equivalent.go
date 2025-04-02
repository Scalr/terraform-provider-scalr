package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// stringSliceAllEquivalentModifier ensures that if a set of strings represents "all items"
// (either via a special "*" element or by containing all predefined allowed elements),
// it is treated as equivalent during planning to prevent unnecessary diffs.
// It prioritizes the state representation (usually from API) if config and state are equivalent.
type stringSliceAllEquivalentModifier struct {
	allowedValues []string // The full list of allowed non-"*" values
}

// StringSliceAllEquivalent creates a new plan modifier for sets.
// allowedValues: The definitive list of all possible non-"*" string values.
func StringSliceAllEquivalent(allowedValues []string) planmodifier.Set {
	allowed := make([]string, len(allowedValues))
	copy(allowed, allowedValues)
	return &stringSliceAllEquivalentModifier{
		allowedValues: allowed,
	}
}

func (m *stringSliceAllEquivalentModifier) Description(ctx context.Context) string {
	return fmt.Sprintf("Prevents persistent diffs when configuration and state are semantically equivalent (e.g., '*' vs full list of %d items).", len(m.allowedValues))
}

func (m *stringSliceAllEquivalentModifier) MarkdownDescription(ctx context.Context) string {
	return fmt.Sprintf("If the configuration specifies `set( \"*\" )` and the current state contains all %d allowed values (or vice versa), this modifier prevents Terraform from showing a difference in the plan by ensuring the planned value matches the state value.", len(m.allowedValues))
}

// PlanModifySet implements the planmodifier.Set interface.
func (m *stringSliceAllEquivalentModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	// Do nothing if plan, config or state is null/unknown
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() ||
		req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() ||
		req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	// Extract slices from config and state
	var configSlice, stateSlice []string
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &configSlice, false)...)
	resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &stateSlice, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine if config represents "all"
	isConfigStar := len(configSlice) == 1 && configSlice[0] == "*"
	isConfigExplicitAll := m.containsAll(configSlice)
	isConfigAll := isConfigStar || isConfigExplicitAll

	// Determine if state represents "all"
	isStateAll := m.containsAll(stateSlice)

	// --- Main Logic ---
	// Only intervene to suppress diffs if config and state ALREADY semantically match ("all")
	if isConfigAll && isStateAll {
		// If they already match, ensure the plan value is the same as the state value
		// to prevent Terraform from showing a diff.
		// We only modify the plan if it differs from the state in this specific case.
		if !req.PlanValue.Equal(req.StateValue) {
			resp.PlanValue = req.StateValue
		}
	}
	// Otherwise (if config and state don't both mean "all", or if they do but plan already matches state),
	// let Terraform's default plan proceed without modification.
	// This allows changes like Partial -> [*] or [*] -> Partial to be planned correctly based on config.
}

// containsAll checks if the given slice contains exactly all the allowed values.
// Note: This helper assumes the input slice does not contain duplicates if comparing against allowedValues.
// Sets inherently handle uniqueness, so we primarily care if all allowedValues are present.
func (m *stringSliceAllEquivalentModifier) containsAll(slice []string) bool {
	if len(slice) != len(m.allowedValues) {
		return false
	}
	seen := make(map[string]bool, len(slice))
	for _, s := range slice {
		seen[s] = true // Build a map of elements present in the slice
	}
	// Check if all allowed values were present in the slice's map
	for _, allowed := range m.allowedValues {
		if !seen[allowed] {
			return false
		}
	}
	return true // If lengths match and all allowed values were seen, the sets are equivalent
}

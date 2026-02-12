package stringvalidation

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Compile-time interface check
var _ validator.String = stringIsNamespaceNameValidator{}

type stringIsNamespaceNameValidator struct{}

func (v stringIsNamespaceNameValidator) Description(_ context.Context) string {
	return "must only contain letters, numbers, dashes, and underscores"
}

func (v stringIsNamespaceNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsNamespaceNameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	// Regex to match only letters, numbers, dashes, and underscores
	validNamespaceName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	if !validNamespaceName.MatchString(value) {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			v.Description(ctx),
			fmt.Sprintf("%q", value),
		))

		return
	}
}

func StringIsNamespaceName() validator.String {
	return stringIsNamespaceNameValidator{}
}

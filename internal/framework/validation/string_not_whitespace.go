package validation

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Compile-time interface check
var _ validator.String = stringIsNotWhiteSpaceValidator{}

type stringIsNotWhiteSpaceValidator struct{}

func (v stringIsNotWhiteSpaceValidator) Description(_ context.Context) string {
	return "must not be empty or consisting entirely of whitespace characters"
}

func (v stringIsNotWhiteSpaceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsNotWhiteSpaceValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	if strings.TrimSpace(value) == "" {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			v.Description(ctx),
			fmt.Sprintf("%q", value),
		))

		return
	}
}

func StringIsNotWhiteSpace() validator.String {
	return stringIsNotWhiteSpaceValidator{}
}

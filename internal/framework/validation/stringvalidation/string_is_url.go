package stringvalidation

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Compile-time interface check
var _ validator.String = stringIsValidURLValidator{}

type stringIsValidURLValidator struct{}

func (v stringIsValidURLValidator) Description(_ context.Context) string {
	return "must be a valid URL with a host and an https scheme"
}

func (v stringIsValidURLValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringIsValidURLValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	parsedURL, err := url.ParseRequestURI(value)
	if err != nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			v.Description(ctx),
			fmt.Sprintf("%q", value),
		))
		return
	}

	if parsedURL.Host == "" {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			"URL must include a host",
			fmt.Sprintf("%q", value),
		))
	}

	if strings.ToLower(parsedURL.Scheme) != "https" {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			req.Path,
			"URL scheme must be https",
			fmt.Sprintf("%q", value),
		))
	}
}

func StringIsValidURL() validator.String {
	return stringIsValidURLValidator{}
}

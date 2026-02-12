package stringvalidation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Compile-time interface check
var _ validator.String = preferWriteOnlyAttributeIf{}

type preferWriteOnlyAttributeIfFunc func(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) bool

type preferWriteOnlyAttributeIf struct {
	writeOnlyAttribute path.Expression
	ifFunc             preferWriteOnlyAttributeIfFunc
	validator          validator.String
}

func PreferWriteOnlyAttributeIf(writeOnlyAttribute path.Expression, ifFunc preferWriteOnlyAttributeIfFunc) validator.String {
	return preferWriteOnlyAttributeIf{
		writeOnlyAttribute: writeOnlyAttribute,
		ifFunc:             ifFunc,
		validator:          stringvalidator.PreferWriteOnlyAttribute(writeOnlyAttribute),
	}
}

func (v preferWriteOnlyAttributeIf) Description(ctx context.Context) string {
	return v.validator.Description(ctx)
}

func (v preferWriteOnlyAttributeIf) MarkdownDescription(ctx context.Context) string {
	return v.validator.MarkdownDescription(ctx)
}

func (v preferWriteOnlyAttributeIf) ValidateString(
	ctx context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if v.ifFunc(ctx, req, resp) {
		v.validator.ValidateString(ctx, req, resp)
	}
}

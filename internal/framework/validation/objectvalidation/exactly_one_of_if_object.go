package objectvalidation

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ExactlyOneOfIfObjectSet(expressions ...path.Expression) validator.Object {
	return &exactlyOneOfIfObjectSetValidator{
		pathExpressions: expressions,
	}
}

type exactlyOneOfIfObjectSetValidator struct {
	pathExpressions []path.Expression
}

func (v *exactlyOneOfIfObjectSetValidator) Description(_ context.Context) string {
	return "requires exactly one of the paths to be configured when the object is set"
}

func (v *exactlyOneOfIfObjectSetValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *exactlyOneOfIfObjectSetValidator) ValidateObject(
	ctx context.Context,
	req validator.ObjectRequest,
	resp *validator.ObjectResponse,
) {
	// If the object is null or unknown - do nothing
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	count := 0
	expressions := req.PathExpression.MergeExpressions(v.pathExpressions...)

	for _, expr := range expressions {
		matches, diags := req.Config.PathMatches(ctx, expr)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			continue
		}

		for _, m := range matches {
			if m.Equal(req.Path) {
				// Skip the current path if it was also specified in the expressions
				continue
			}

			var val attr.Value
			diags := req.Config.GetAttribute(ctx, m, &val)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			if val.IsUnknown() {
				// Delay the validation until all values are known
				return
			}

			if !val.IsNull() {
				count++
			}
		}
	}

	if count != 1 {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeCombinationDiagnostic(
			req.Path,
			fmt.Sprintf("Exactly one of these attributes must be configured when %s is set: %s.", req.Path, v.pathExpressions),
		))
	}
}

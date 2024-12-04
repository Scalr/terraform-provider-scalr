package defaults

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const CurrentAccountIDEnvVar = "SCALR_ACCOUNT_ID"

func GetDefaultScalrAccountID() (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	v := os.Getenv(CurrentAccountIDEnvVar)
	if v == "" {
		diags.AddError(
			"Cannot infer current account",
			fmt.Sprintf(
				"Default value for `account_id` could not be computed."+
					"\nIf you are using Scalr Provider for local runs, please set the attribute in resources explicitly,"+
					"\nor export `%s` environment variable prior the run.",
				CurrentAccountIDEnvVar,
			),
		)
	}
	return v, diags
}

// AccountIDRequired returns a default account id value handler.
//
// Use AccountIDRequired when a default value for account id must be set.
func AccountIDRequired() defaults.String {
	return accountIDRequiredDefault{}
}

// accountIDRequiredDefault implements defaults.String
type accountIDRequiredDefault struct{}

func (r accountIDRequiredDefault) Description(_ context.Context) string {
	return "value defaults to current Scalr account id"
}

func (r accountIDRequiredDefault) MarkdownDescription(ctx context.Context) string {
	return r.Description(ctx)
}

func (r accountIDRequiredDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	s, diags := GetDefaultScalrAccountID()
	resp.Diagnostics.Append(diags...)
	resp.PlanValue = types.StringValue(s)
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework/defaults"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

func variableResourceSchema() *schema.Schema {
	return &schema.Schema{
		MarkdownDescription: "Manages the state of variables in Scalr.",
		Version:             3,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Key of the variable.",
				Required:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							var sensitive types.Bool
							resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("sensitive"), &sensitive)...)
							resp.RequiresReplace = sensitive.ValueBool()
						},
						"Recreate the resource when changing the `key` value of a sensitive variable.",
						"Recreate the resource when changing the `key` value of a sensitive variable.",
					),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Variable value.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				Sensitive:           true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "Indicates if this is a Terraform or shell variable. Allowed values are `terraform` or `shell`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.CategoryEnv),
						string(scalr.CategoryTerraform),
						string(scalr.CategoryShell),
					),
				},
			},
			"hcl": schema.BoolAttribute{
				MarkdownDescription: "Set (true/false) to configure the variable as a string of HCL code. Has no effect for `category = \"shell\"` variables. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				Validators:          []validator.Bool{categoryHCLValidator{}},
			},
			"sensitive": schema.BoolAttribute{
				MarkdownDescription: "Set (true/false) to configure as sensitive. Sensitive variable values are not visible after being set. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
							resp.RequiresReplace = req.StateValue.ValueBool()
						},
						"Recreate the resource when changing the `sensitive` value from `true` to `false`.",
						"Recreate the resource when changing the `sensitive` value from `true` to `false`.",
					),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Variable verbose description, defaults to empty string.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"final": schema.BoolAttribute{
				MarkdownDescription: "Set (true/false) to configure as final. Indicates whether the variable can be overridden on a lower scope down the Scalr organizational model. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: "Set (true/false) to configure as force. Allows creating final variables on higher scope, even if the same variable exists on lower scope (lower is to be deleted). Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "The workspace that owns the variable, specified as an ID, in the format `ws-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "The environment that owns the variable, specified as an ID, in the format `env-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				Default:             defaults.AccountIDRequired(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Date/time the variable was updated.",
				Computed:            true,
			},
			"updated_by_email": schema.StringAttribute{
				MarkdownDescription: "Email of the user who updated the variable last time.",
				Computed:            true,
			},
			"updated_by": schema.ListAttribute{
				MarkdownDescription: "Details of the user that updated the variable last time.",
				ElementType:         userElementType,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Compile-time interface check
var _ validator.Bool = categoryHCLValidator{}

type categoryHCLValidator struct{}

func (v categoryHCLValidator) Description(_ context.Context) string {
	return "must not be empty or consisting entirely of whitespace characters"
}

func (v categoryHCLValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v categoryHCLValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var category types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("category"), &category)...)

	if category.ValueString() != string(scalr.CategoryTerraform) && req.ConfigValue.ValueBool() {
		resp.Diagnostics.Append(
			diag.NewAttributeWarningDiagnostic(
				req.Path,
				"HCL is not supported for shell variables",
				"Setting 'hcl' attribute to 'true' for shell variable is now deprecated.",
			),
		)
	}
}

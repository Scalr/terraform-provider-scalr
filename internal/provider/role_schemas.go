package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

func roleResourceSchema() *schema.Schema {
	return &schema.Schema{
		MarkdownDescription: "Manage the Scalr IAM roles. Create, update and destroy.",
		Version:             1,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the role.",
				Required:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Attribute `account_id` is deprecated, the account id is calculated from the API request context.",
			},
			"is_system": schema.BoolAttribute{
				MarkdownDescription: "Boolean indicates if the role can be edited. System roles are maintained by Scalr and cannot be changed.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Verbose description of the role.",
				Optional:            true,
			},
			"permissions": schema.SetAttribute{
				MarkdownDescription: "Array of permission names.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 128),
					setvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
			},
		},
	}
}

func roleResourceSchemaV0() *schema.Schema {
	return &schema.Schema{
		MarkdownDescription: "Manage the Scalr IAM roles. Create, update and destroy.",
		Version:             0,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the role.",
				Required:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Attribute `account_id` is deprecated, the account id is calculated from the API request context.",
			},
			"is_system": schema.BoolAttribute{
				MarkdownDescription: "Boolean indicates if the role can be edited. System roles are maintained by Scalr and cannot be changed.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Verbose description of the role.",
				Optional:            true,
			},
			"permissions": schema.SetAttribute{
				MarkdownDescription: "Array of permission names.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 128),
					setvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
			},
		},
	}
}

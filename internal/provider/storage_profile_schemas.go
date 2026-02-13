package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

func storageProfileResourceSchema(_ context.Context) *resourceSchema.Schema {
	backendChangedModifier := listplanmodifier.RequiresReplaceIf(
		func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
			// Requires replacement if a given backend block changed from null to non-null value, or vice versa.
			resp.RequiresReplace = req.PlanValue.IsNull() != req.StateValue.IsNull()
		},
		"Recreate the resource when the backend type has changed.",
		"Recreate the resource when the backend type has changed.",
	)

	return &resourceSchema.Schema{
		MarkdownDescription: "Manages the state of storage profiles in Scalr.",

		Attributes: map[string]resourceSchema.Attribute{
			"id": resourceSchema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": resourceSchema.StringAttribute{
				MarkdownDescription: "Name of the storage profile.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"default": resourceSchema.BoolAttribute{
				MarkdownDescription: "The default storage profile.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"created_at": resourceSchema.StringAttribute{
				MarkdownDescription: "The resource creation timestamp.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": resourceSchema.StringAttribute{
				MarkdownDescription: "The resource last update timestamp.",
				Computed:            true,
			},
			"error_message": resourceSchema.StringAttribute{
				MarkdownDescription: "The last error description, when these settings doesn't work properly.",
				Computed:            true,
			},
		},
		Blocks: map[string]resourceSchema.Block{
			"aws_s3": resourceSchema.ListNestedBlock{
				MarkdownDescription: "Settings for the AWS S3 storage profile.",
				NestedObject: resourceSchema.NestedBlockObject{
					Attributes: map[string]resourceSchema.Attribute{
						"audience": resourceSchema.StringAttribute{
							MarkdownDescription: "The value of the `aud` claim for the identity token.",
							Required:            true,
						},
						"bucket_name": resourceSchema.StringAttribute{
							MarkdownDescription: "AWS S3 Storage bucket name. Bucket must already exist.",
							Required:            true,
						},
						"region": resourceSchema.StringAttribute{
							MarkdownDescription: "AWS S3 bucket region.",
							Optional:            true,
						},
						"role_arn": resourceSchema.StringAttribute{
							MarkdownDescription: "Amazon Resource Name (ARN) of the IAM Role to assume.",
							Required:            true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{backendChangedModifier},
			},
			"azurerm": resourceSchema.ListNestedBlock{
				MarkdownDescription: "Settings for the AzureRM storage profile.",
				NestedObject: resourceSchema.NestedBlockObject{
					Attributes: map[string]resourceSchema.Attribute{
						"audience": resourceSchema.StringAttribute{
							MarkdownDescription: "Azure audience for authentication.",
							Required:            true,
						},
						"client_id": resourceSchema.StringAttribute{
							MarkdownDescription: "Azure client ID for authentication.",
							Required:            true,
						},
						"container_name": resourceSchema.StringAttribute{
							MarkdownDescription: "Azure storage container name.",
							Required:            true,
						},
						"storage_account": resourceSchema.StringAttribute{
							MarkdownDescription: "Azure storage account name.",
							Required:            true,
						},
						"tenant_id": resourceSchema.StringAttribute{
							MarkdownDescription: "Azure tenant ID for authentication.",
							Required:            true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{backendChangedModifier},
			},
			"google": resourceSchema.ListNestedBlock{
				MarkdownDescription: "Settings for the Google storage profile.",
				NestedObject: resourceSchema.NestedBlockObject{
					Attributes: map[string]resourceSchema.Attribute{
						"credentials": resourceSchema.StringAttribute{
							MarkdownDescription: "Service Account JSON key." +
								" Required IAM roles: `Storage Admin` assigned on a `google-storage-bucket` bucket." +
								" See: [use IAM with bucket](https://cloud.google.com/storage/docs/access-control/using-iam-permissions#bucket-iam).",
							Required:  true,
							Sensitive: true,
						},
						"encryption_key": resourceSchema.StringAttribute{
							MarkdownDescription: "Customer supplied encryption key. Must be exactly 32 bytes, encoded into base64.",
							Optional:            true,
							Sensitive:           true,
						},
						"project": resourceSchema.StringAttribute{
							MarkdownDescription: "Google Cloud project ID.",
							Optional:            true,
							Computed:            true,
						},
						"storage_bucket": resourceSchema.StringAttribute{
							MarkdownDescription: "Google Storage bucket name. Bucket must already exist.",
							Required:            true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{backendChangedModifier},
			},
		},
	}
}

func storageProfileDatasourceSchema(_ context.Context) *datasourceSchema.Schema {
	return &datasourceSchema.Schema{
		MarkdownDescription: "Retrieves information about storage profile.",

		Attributes: map[string]datasourceSchema.Attribute{
			"id": datasourceSchema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": datasourceSchema.StringAttribute{
				MarkdownDescription: "Name of the storage profile.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"default": datasourceSchema.BoolAttribute{
				MarkdownDescription: "The default storage profile.",
				Optional:            true,
				Computed:            true,
			},
			"created_at": datasourceSchema.StringAttribute{
				MarkdownDescription: "The resource creation timestamp.",
				Computed:            true,
			},
			"updated_at": datasourceSchema.StringAttribute{
				MarkdownDescription: "The resource last update timestamp.",
				Computed:            true,
			},
			"error_message": datasourceSchema.StringAttribute{
				MarkdownDescription: "The last error description, when these settings doesn't work properly.",
				Computed:            true,
			},
		},
		Blocks: map[string]datasourceSchema.Block{
			"aws_s3": datasourceSchema.ListNestedBlock{
				MarkdownDescription: "Settings for the AWS S3 storage profile.",
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"audience": datasourceSchema.StringAttribute{
							MarkdownDescription: "The value of the `aud` claim for the identity token.",
							Computed:            true,
						},
						"bucket_name": datasourceSchema.StringAttribute{
							MarkdownDescription: "AWS S3 Storage bucket name.",
							Computed:            true,
						},
						"region": datasourceSchema.StringAttribute{
							MarkdownDescription: "AWS S3 bucket region.",
							Computed:            true,
						},
						"role_arn": datasourceSchema.StringAttribute{
							MarkdownDescription: "Amazon Resource Name (ARN) of the IAM Role to assume.",
							Computed:            true,
						},
					},
				},
			},
			"azurerm": datasourceSchema.ListNestedBlock{
				MarkdownDescription: "Settings for the AzureRM storage profile.",
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"audience": datasourceSchema.StringAttribute{
							MarkdownDescription: "Azure audience for authentication.",
							Computed:            true,
						},
						"client_id": datasourceSchema.StringAttribute{
							MarkdownDescription: "Azure client ID for authentication.",
							Computed:            true,
						},
						"container_name": datasourceSchema.StringAttribute{
							MarkdownDescription: "Azure storage container name.",
							Computed:            true,
						},
						"storage_account": datasourceSchema.StringAttribute{
							MarkdownDescription: "Azure storage account name.",
							Computed:            true,
						},
						"tenant_id": datasourceSchema.StringAttribute{
							MarkdownDescription: "Azure tenant ID for authentication.",
							Computed:            true,
						},
					},
				},
			},
			"google": datasourceSchema.ListNestedBlock{
				MarkdownDescription: "Settings for the Google storage profile.",
				NestedObject: datasourceSchema.NestedBlockObject{
					Attributes: map[string]datasourceSchema.Attribute{
						"credentials": datasourceSchema.StringAttribute{
							MarkdownDescription: "Service Account JSON key.",
							Computed:            true,
							Sensitive:           true,
						},
						"encryption_key": datasourceSchema.StringAttribute{
							MarkdownDescription: "Customer supplied encryption key.",
							Computed:            true,
							Sensitive:           true,
						},
						"project": datasourceSchema.StringAttribute{
							MarkdownDescription: "Google Cloud project ID.",
							Computed:            true,
						},
						"storage_bucket": datasourceSchema.StringAttribute{
							MarkdownDescription: "Google Storage bucket name.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

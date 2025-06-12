package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"
)

var (
	awsS3StorageProfileSettingsElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"audience":    types.StringType,
			"bucket_name": types.StringType,
			"region":      types.StringType,
			"role_arn":    types.StringType,
		},
	}
	azureRMStorageProfileSettingsElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"audience":        types.StringType,
			"client_id":       types.StringType,
			"container_name":  types.StringType,
			"storage_account": types.StringType,
			"tenant_id":       types.StringType,
		},
	}
	googleStorageProfileSettingsElementType = types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"credentials":    types.StringType,
			"encryption_key": types.StringType,
			"project":        types.StringType,
			"storage_bucket": types.StringType,
		},
	}
)

type storageProfileResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Default      types.Bool   `tfsdk:"default"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	ErrorMessage types.String `tfsdk:"error_message"`
	AWSS3        types.List   `tfsdk:"aws_s3"`
	AzureRM      types.List   `tfsdk:"azurerm"`
	Google       types.List   `tfsdk:"google"`
}

type storageProfileDataSourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Default      types.Bool   `tfsdk:"default"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
	ErrorMessage types.String `tfsdk:"error_message"`
	AWSS3        types.List   `tfsdk:"aws_s3"`
	AzureRM      types.List   `tfsdk:"azurerm"`
	Google       types.List   `tfsdk:"google"`
}

type awsS3StorageProfileSettingsModel struct {
	Audience   types.String `tfsdk:"audience"`
	BucketName types.String `tfsdk:"bucket_name"`
	Region     types.String `tfsdk:"region"`
	RoleARN    types.String `tfsdk:"role_arn"`
}

type azureRMStorageProfileSettingsModel struct {
	Audience       types.String `tfsdk:"audience"`
	ClientID       types.String `tfsdk:"client_id"`
	ContainerName  types.String `tfsdk:"container_name"`
	StorageAccount types.String `tfsdk:"storage_account"`
	TenantID       types.String `tfsdk:"tenant_id"`
}

type googleStorageProfileSettingsModel struct {
	Credentials   types.String `tfsdk:"credentials"`
	EncryptionKey types.String `tfsdk:"encryption_key"`
	Project       types.String `tfsdk:"project"`
	StorageBucket types.String `tfsdk:"storage_bucket"`
}

func storageProfileResourceModelFromAPI(
	ctx context.Context, sp *scalr.StorageProfile, existing *storageProfileResourceModel,
) (*storageProfileResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &storageProfileResourceModel{
		Id:           types.StringValue(sp.ID),
		Name:         types.StringValue(sp.Name),
		Default:      types.BoolValue(sp.Default),
		CreatedAt:    types.StringValue(sp.CreatedAt),
		UpdatedAt:    types.StringNull(),
		ErrorMessage: types.StringNull(),
		AWSS3:        types.ListNull(awsS3StorageProfileSettingsElementType),
		AzureRM:      types.ListNull(azureRMStorageProfileSettingsElementType),
		Google:       types.ListNull(googleStorageProfileSettingsElementType),
	}

	if sp.UpdatedAt != nil {
		model.UpdatedAt = types.StringValue(sp.UpdatedAt.Format(time.RFC3339))
	}

	if sp.ErrorMessage != nil {
		model.ErrorMessage = types.StringValue(*sp.ErrorMessage)
	}

	switch sp.BackendType {
	case scalr.StorageProfileBackendTypeAWSS3:
		awsS3Settings := awsS3StorageProfileSettingsModel{
			Audience:   types.StringPointerValue(sp.AWSS3Audience),
			BucketName: types.StringPointerValue(sp.AWSS3BucketName),
			Region:     types.StringPointerValue(sp.AWSS3Region),
			RoleARN:    types.StringPointerValue(sp.AWSS3RoleArn),
		}
		awsS3SettingsValue, d := types.ListValueFrom(ctx, awsS3StorageProfileSettingsElementType, []awsS3StorageProfileSettingsModel{awsS3Settings})
		diags.Append(d...)
		model.AWSS3 = awsS3SettingsValue

	case scalr.StorageProfileBackendTypeAzureRM:
		azureRMSettings := azureRMStorageProfileSettingsModel{
			Audience:       types.StringPointerValue(sp.AzureRMAudience),
			ClientID:       types.StringPointerValue(sp.AzureRMClientID),
			ContainerName:  types.StringPointerValue(sp.AzureRMContainerName),
			StorageAccount: types.StringPointerValue(sp.AzureRMStorageAccount),
			TenantID:       types.StringPointerValue(sp.AzureRMTenantID),
		}
		azureRMSettingsValue, d := types.ListValueFrom(ctx, azureRMStorageProfileSettingsElementType, []azureRMStorageProfileSettingsModel{azureRMSettings})
		diags.Append(d...)
		model.AzureRM = azureRMSettingsValue

	case scalr.StorageProfileBackendTypeGoogle:
		googleSettings := googleStorageProfileSettingsModel{
			Project:       types.StringPointerValue(sp.GoogleProject),
			StorageBucket: types.StringPointerValue(sp.GoogleStorageBucket),
		}
		if existing != nil {
			var existingSettings []googleStorageProfileSettingsModel
			diags.Append(existing.Google.ElementsAs(ctx, &existingSettings, false)...)

			if len(existingSettings) > 0 {
				s := existingSettings[0]

				// Carry sensitive values from the plan or state as the API won't return them in response
				googleSettings.EncryptionKey = s.EncryptionKey
				googleSettings.Credentials = s.Credentials
			}
		}
		googleSettingsValue, d := types.ListValueFrom(ctx, googleStorageProfileSettingsElementType, []googleStorageProfileSettingsModel{googleSettings})
		diags.Append(d...)
		model.Google = googleSettingsValue

	default:
		diags.AddError(
			"Cannot infer storage backend.",
			fmt.Sprintf("Unsupported storage profile backend type %q received from the API.", sp.BackendType),
		)
	}

	return model, diags
}

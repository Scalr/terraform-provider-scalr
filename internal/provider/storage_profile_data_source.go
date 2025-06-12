package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &storageProfileDataSource{}
	_ datasource.DataSourceWithConfigure        = &storageProfileDataSource{}
	_ datasource.DataSourceWithConfigValidators = &storageProfileDataSource{}
)

func newStorageProfileDataSource() datasource.DataSource {
	return &storageProfileDataSource{}
}

// storageProfileDataSource defines the data source implementation.
type storageProfileDataSource struct {
	framework.DataSourceWithScalrClient
}

func (d *storageProfileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_profile"
}

func (d *storageProfileDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = *storageProfileDatasourceSchema(ctx)
}

func (d *storageProfileDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
			path.MatchRoot("default"),
		),
	}
}

func (d *storageProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg storageProfileDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.StorageProfileListOptions{
		Filter: &scalr.StorageProfileListFilter{
			ID:      cfg.Id.ValueStringPointer(),
			Name:    cfg.Name.ValueStringPointer(),
			Default: cfg.Default.ValueBoolPointer(),
		},
	}

	storageProfiles, err := d.Client.StorageProfiles.List(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving storage_profile", err.Error())
		return
	}

	if storageProfiles.TotalCount > 1 {
		resp.Diagnostics.AddError(
			"Error retrieving storage_profile",
			"Your query returned more than one result. Please try a more specific search criteria.",
		)
		return
	}

	if storageProfiles.TotalCount == 0 {
		resp.Diagnostics.AddError(
			"Error retrieving storage_profile",
			fmt.Sprintf("Could not find storage_profile with ID '%s', name '%s', default '%t'.", cfg.Id.ValueString(), cfg.Name.ValueString(), cfg.Default.ValueBool()),
		)
		return
	}

	sp := storageProfiles.Items[0]

	cfg.Id = types.StringValue(sp.ID)
	cfg.Name = types.StringValue(sp.Name)
	cfg.Default = types.BoolValue(sp.Default)
	cfg.CreatedAt = types.StringValue(sp.CreatedAt)

	if sp.UpdatedAt != nil {
		cfg.UpdatedAt = types.StringValue(sp.UpdatedAt.Format(time.RFC3339))
	}

	if sp.ErrorMessage != nil {
		cfg.ErrorMessage = types.StringValue(*sp.ErrorMessage)
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
		resp.Diagnostics.Append(d...)
		cfg.AWSS3 = awsS3SettingsValue

	case scalr.StorageProfileBackendTypeAzureRM:
		azureRMSettings := azureRMStorageProfileSettingsModel{
			Audience:       types.StringPointerValue(sp.AzureRMAudience),
			ClientID:       types.StringPointerValue(sp.AzureRMClientID),
			ContainerName:  types.StringPointerValue(sp.AzureRMContainerName),
			StorageAccount: types.StringPointerValue(sp.AzureRMStorageAccount),
			TenantID:       types.StringPointerValue(sp.AzureRMTenantID),
		}
		azureRMSettingsValue, d := types.ListValueFrom(ctx, azureRMStorageProfileSettingsElementType, []azureRMStorageProfileSettingsModel{azureRMSettings})
		resp.Diagnostics.Append(d...)
		cfg.AzureRM = azureRMSettingsValue

	case scalr.StorageProfileBackendTypeGoogle:
		googleSettings := googleStorageProfileSettingsModel{
			Project:       types.StringPointerValue(sp.GoogleProject),
			StorageBucket: types.StringPointerValue(sp.GoogleStorageBucket),
		}
		googleSettingsValue, d := types.ListValueFrom(ctx, googleStorageProfileSettingsElementType, []googleStorageProfileSettingsModel{googleSettings})
		resp.Diagnostics.Append(d...)
		cfg.Google = googleSettingsValue

	default:
		resp.Diagnostics.AddError(
			"Cannot infer storage backend.",
			fmt.Sprintf("Unsupported storage profile backend type %q received from the API.", sp.BackendType),
		)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

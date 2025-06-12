package provider

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &storageProfileResource{}
	_ resource.ResourceWithConfigure        = &storageProfileResource{}
	_ resource.ResourceWithConfigValidators = &storageProfileResource{}
	_ resource.ResourceWithImportState      = &storageProfileResource{}
)

func newStorageProfileResource() resource.Resource {
	return &storageProfileResource{}
}

// storageProfileResource defines the resource implementation.
type storageProfileResource struct {
	framework.ResourceWithScalrClient
}

func (r *storageProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_profile"
}

func (r *storageProfileResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = *storageProfileResourceSchema(ctx)
}

func (r *storageProfileResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("aws_s3"),
			path.MatchRoot("azurerm"),
			path.MatchRoot("google"),
		),
	}
}

func (r *storageProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageProfileResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.StorageProfileCreateOptions{
		Name:    plan.Name.ValueString(),
		Default: plan.Default.ValueBoolPointer(),
	}

	if !plan.AWSS3.IsUnknown() && !plan.AWSS3.IsNull() {
		var awsS3Settings []awsS3StorageProfileSettingsModel
		resp.Diagnostics.Append(plan.AWSS3.ElementsAs(ctx, &awsS3Settings, false)...)

		if len(awsS3Settings) > 0 {
			opts.BackendType = scalr.StorageProfileBackendTypeAWSS3
			settings := awsS3Settings[0]

			opts.AWSS3Audience = settings.Audience.ValueStringPointer()
			opts.AWSS3BucketName = settings.BucketName.ValueStringPointer()
			opts.AWSS3Region = settings.Region.ValueStringPointer()
			opts.AWSS3RoleArn = settings.RoleARN.ValueStringPointer()
		}
	}

	if !plan.AzureRM.IsUnknown() && !plan.AzureRM.IsNull() {
		var azureRMSettings []azureRMStorageProfileSettingsModel
		resp.Diagnostics.Append(plan.AzureRM.ElementsAs(ctx, &azureRMSettings, false)...)

		if len(azureRMSettings) > 0 {
			opts.BackendType = scalr.StorageProfileBackendTypeAzureRM
			settings := azureRMSettings[0]

			opts.AzureRMAudience = settings.Audience.ValueStringPointer()
			opts.AzureRMClientID = settings.ClientID.ValueStringPointer()
			opts.AzureRMContainerName = settings.ContainerName.ValueStringPointer()
			opts.AzureRMStorageAccount = settings.StorageAccount.ValueStringPointer()
			opts.AzureRMTenantID = settings.TenantID.ValueStringPointer()
		}
	}

	if !plan.Google.IsUnknown() && !plan.Google.IsNull() {
		var googleSettings []googleStorageProfileSettingsModel
		resp.Diagnostics.Append(plan.Google.ElementsAs(ctx, &googleSettings, false)...)

		if len(googleSettings) > 0 {
			opts.BackendType = scalr.StorageProfileBackendTypeGoogle
			settings := googleSettings[0]

			opts.GoogleCredentials = ptr(json.RawMessage(settings.Credentials.ValueString()))
			opts.GoogleEncryptionKey = settings.EncryptionKey.ValueStringPointer()
			opts.GoogleStorageBucket = settings.StorageBucket.ValueStringPointer()

			if !settings.Project.IsUnknown() { // Computed field
				opts.GoogleProject = settings.Project.ValueStringPointer()
			}
		}
	}

	storageProfile, err := r.Client.StorageProfiles.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage_profile", err.Error())
		return
	}

	result, d := storageProfileResourceModelFromAPI(ctx, storageProfile, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *storageProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state storageProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	storageProfile, err := r.Client.StorageProfiles.Read(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving storage_profile", err.Error())
		return
	}

	result, d := storageProfileResourceModelFromAPI(ctx, storageProfile, &state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *storageProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state storageProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.StorageProfileUpdateOptions{}

	if !plan.Name.Equal(state.Name) {
		opts.Name = plan.Name.ValueStringPointer()
	}

	if !plan.Default.Equal(state.Default) {
		opts.Default = plan.Default.ValueBoolPointer()
	}

	if !plan.AWSS3.Equal(state.AWSS3) && !plan.AWSS3.IsNull() {
		var awsS3Settings []awsS3StorageProfileSettingsModel
		resp.Diagnostics.Append(plan.AWSS3.ElementsAs(ctx, &awsS3Settings, false)...)

		if len(awsS3Settings) > 0 {
			opts.BackendType = ptr(scalr.StorageProfileBackendTypeAWSS3)
			settings := awsS3Settings[0]

			opts.AWSS3Audience = settings.Audience.ValueStringPointer()
			opts.AWSS3BucketName = settings.BucketName.ValueStringPointer()
			opts.AWSS3Region = settings.Region.ValueStringPointer()
			opts.AWSS3RoleArn = settings.RoleARN.ValueStringPointer()
		}
	}

	if !plan.AzureRM.Equal(state.AzureRM) && !plan.AzureRM.IsNull() {
		var azureRMSettings []azureRMStorageProfileSettingsModel
		resp.Diagnostics.Append(plan.AzureRM.ElementsAs(ctx, &azureRMSettings, false)...)

		if len(azureRMSettings) > 0 {
			opts.BackendType = ptr(scalr.StorageProfileBackendTypeAzureRM)
			settings := azureRMSettings[0]

			opts.AzureRMAudience = settings.Audience.ValueStringPointer()
			opts.AzureRMClientID = settings.ClientID.ValueStringPointer()
			opts.AzureRMContainerName = settings.ContainerName.ValueStringPointer()
			opts.AzureRMStorageAccount = settings.StorageAccount.ValueStringPointer()
			opts.AzureRMTenantID = settings.TenantID.ValueStringPointer()
		}
	}

	if !plan.Google.Equal(state.Google) && !plan.Google.IsNull() {
		var googleSettings []googleStorageProfileSettingsModel
		resp.Diagnostics.Append(plan.Google.ElementsAs(ctx, &googleSettings, false)...)

		if len(googleSettings) > 0 {
			opts.BackendType = ptr(scalr.StorageProfileBackendTypeGoogle)
			settings := googleSettings[0]

			opts.GoogleCredentials = ptr(json.RawMessage(settings.Credentials.ValueString()))
			opts.GoogleEncryptionKey = settings.EncryptionKey.ValueStringPointer()
			opts.GoogleStorageBucket = settings.StorageBucket.ValueStringPointer()

			if !settings.Project.IsUnknown() { // Computed field
				opts.GoogleProject = settings.Project.ValueStringPointer()
			}
		}
	}

	// Update existing resource
	storageProfile, err := r.Client.StorageProfiles.Update(ctx, plan.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating storage_profile", err.Error())
		return
	}

	result, d := storageProfileResourceModelFromAPI(ctx, storageProfile, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *storageProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state storageProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.StorageProfiles.Delete(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting storage_profile", err.Error())
		return
	}
}

func (r *storageProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

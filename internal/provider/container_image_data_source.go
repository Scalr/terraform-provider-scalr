package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	scalrV2 "github.com/scalr/go-scalr/v2/scalr"
	"github.com/scalr/go-scalr/v2/scalr/ops/container_image"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource                     = &containerImageDataSource{}
	_ datasource.DataSourceWithConfigure        = &containerImageDataSource{}
	_ datasource.DataSourceWithConfigValidators = &containerImageDataSource{}
)

func newContainerImageDataSource() datasource.DataSource {
	return &containerImageDataSource{}
}

// containerImageDataSource defines the data source implementation.
type containerImageDataSource struct {
	framework.DataSourceWithScalrClient
}

// containerImageDataSourceModel describes the data source data model.
type containerImageDataSourceModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Visibility      types.String `tfsdk:"visibility"`
	Description     types.String `tfsdk:"description"`
	RegistryURL     types.String `tfsdk:"registry_url"`
	Repository      types.String `tfsdk:"repository"`
	SyncStatus      types.String `tfsdk:"sync_status"`
	LatestVersionID types.String `tfsdk:"latest_version_id"`
}

func (d *containerImageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_image"
}

func (d *containerImageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a container image that can be used as a runner image." +
			" Requires the `container-images:read` permission on the account.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the container image, in the format `cimg-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The container image repository name without a tag, for example `scalr/runner`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "The image visibility: `system` for a Scalr-managed image or `public` for a custom image registered in the account.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(schemas.ContainerImageVisibilitySystem),
						string(schemas.ContainerImageVisibilityPublic),
					),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The image description.",
				Computed:            true,
			},
			"registry_url": schema.StringAttribute{
				MarkdownDescription: "The OCI registry URL the image is pulled from.",
				Computed:            true,
			},
			"repository": schema.StringAttribute{
				MarkdownDescription: "The repository path within the registry.",
				Computed:            true,
			},
			"sync_status": schema.StringAttribute{
				MarkdownDescription: "The tag synchronization state of the image: `pending`, `synced`, `failed` or `no_versions`.",
				Computed:            true,
			},
			"latest_version_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the latest version of the image.",
				Computed:            true,
			},
		},
	}
}

func (d *containerImageDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *containerImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg containerImageDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var image *schemas.ContainerImage
	var err error

	if !cfg.Id.IsNull() {
		image, err = d.ClientV2.ContainerImage.GetContainerImage(ctx, cfg.Id.ValueString(), nil)
	} else {
		image, err = getContainerImageByName(ctx, d.ClientV2, cfg.Name.ValueString(), cfg.Visibility.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving container image", err.Error())
		return
	}

	if (!cfg.Name.IsNull() && cfg.Name.ValueString() != image.Attributes.Name) ||
		(!cfg.Visibility.IsNull() && cfg.Visibility.ValueString() != string(image.Attributes.Visibility)) {
		resp.Diagnostics.AddError(
			"Error retrieving container image",
			fmt.Sprintf(
				"Could not find container image with ID '%s', name '%s' and visibility '%s'",
				cfg.Id.ValueString(), cfg.Name.ValueString(), cfg.Visibility.ValueString(),
			),
		)
		return
	}

	cfg.Id = types.StringValue(image.ID)
	cfg.Name = types.StringValue(image.Attributes.Name)
	cfg.Visibility = types.StringValue(string(image.Attributes.Visibility))
	cfg.Description = types.StringPointerValue(image.Attributes.Description)
	cfg.RegistryURL = types.StringValue(image.Attributes.RegistryUrl)
	cfg.Repository = types.StringValue(image.Attributes.Repository)
	cfg.SyncStatus = types.StringValue(string(image.Attributes.SyncStatus))
	cfg.LatestVersionID = types.StringNull()

	if image.Relationships.LatestVersion != nil {
		cfg.LatestVersionID = types.StringValue(image.Relationships.LatestVersion.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

// getContainerImageByName looks up a container image by its exact name.
// The API only supports a query search by name or description, so the results are matched locally.
func getContainerImageByName(
	ctx context.Context,
	client *scalrV2.Client,
	name string,
	visibility string,
) (*schemas.ContainerImage, error) {
	opts := container_image.ListContainerImagesOptions{
		Query:  name,
		Filter: map[string]string{},
	}
	if visibility != "" {
		opts.Filter["visibility"] = visibility
	}

	var matches []schemas.ContainerImage
	for image, err := range client.ContainerImage.ListContainerImagesIter(ctx, &opts) {
		if err != nil {
			return nil, err
		}
		if image.Attributes.Name == name {
			matches = append(matches, image)
		}
	}

	if len(matches) > 1 {
		return nil, fmt.Errorf(
			"Found %d container images named '%s'. Please set `visibility` or use `id` to narrow down the search.",
			len(matches), name,
		)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("Could not find container image with name '%s'", name)
	}

	return &matches[0], nil
}

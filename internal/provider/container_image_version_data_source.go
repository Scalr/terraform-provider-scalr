package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/ops/container_image_version"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ datasource.DataSource              = &containerImageVersionDataSource{}
	_ datasource.DataSourceWithConfigure = &containerImageVersionDataSource{}
)

func newContainerImageVersionDataSource() datasource.DataSource {
	return &containerImageVersionDataSource{}
}

// containerImageVersionDataSource defines the data source implementation.
type containerImageVersionDataSource struct {
	framework.DataSourceWithScalrClient
}

// containerImageVersionDataSourceModel describes the data source data model.
type containerImageVersionDataSourceModel struct {
	Id               types.String `tfsdk:"id"`
	ContainerImageID types.String `tfsdk:"container_image_id"`
	Version          types.String `tfsdk:"version"`
	Latest           types.Bool   `tfsdk:"latest"`
	Available        types.Bool   `tfsdk:"available"`
	CompressedSize   types.Int64  `tfsdk:"compressed_size"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

func (d *containerImageVersionDataSource) Metadata(
	_ context.Context,
	req datasource.MetadataRequest,
	resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_container_image_version"
}

func (d *containerImageVersionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a version of a container image." +
			" Requires the `container-images:read` permission on the account.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the container image version, in the format `cimgv-<RANDOM STRING>`.",
				Computed:            true,
			},
			"container_image_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the container image. Use the `scalr_container_image` data source to look it up.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The image tag, for example `0.4.0`. If omitted, the latest version of the image is returned.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"latest": schema.BoolAttribute{
				MarkdownDescription: "Whether this is the latest version of the container image.",
				Computed:            true,
			},
			"available": schema.BoolAttribute{
				MarkdownDescription: "Whether the version can be selected and used. Always `true` for system image versions.",
				Computed:            true,
			},
			"compressed_size": schema.Int64Attribute{
				MarkdownDescription: "The compressed linux/amd64 manifest size in bytes, including config and layers.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The time when the version was registered.",
				Computed:            true,
			},
		},
	}
}

func (d *containerImageVersionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg containerImageVersionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	imageID := cfg.ContainerImageID.ValueString()

	var version *schemas.ContainerImageVersion

	if !cfg.Version.IsNull() {
		versions, err := d.ClientV2.ContainerImageVersion.ListContainerImageVersions(
			ctx, &container_image_version.ListContainerImageVersionsOptions{
				Filter: map[string]string{
					"container-image": imageID,
					"version":         cfg.Version.ValueString(),
				},
			},
		)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving container image version", err.Error())
			return
		}
		if len(versions) == 0 {
			resp.Diagnostics.AddError(
				"Error retrieving container image version",
				fmt.Sprintf("Could not find version '%s' of container image '%s'", cfg.Version.ValueString(), imageID),
			)
			return
		}
		version = versions[0]
	} else {
		image, err := d.ClientV2.ContainerImage.GetContainerImage(ctx, imageID, nil)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving container image", err.Error())
			return
		}
		if image.Relationships.LatestVersion == nil {
			resp.Diagnostics.AddError(
				"Error retrieving container image version",
				fmt.Sprintf("Container image '%s' has no versions (sync status: %s)", imageID, image.Attributes.SyncStatus),
			)
			return
		}
		version, err = d.ClientV2.ContainerImageVersion.GetContainerImageVersion(
			ctx, image.Relationships.LatestVersion.ID, nil,
		)
		if err != nil {
			resp.Diagnostics.AddError("Error retrieving container image version", err.Error())
			return
		}
	}

	cfg.Id = types.StringValue(version.ID)
	cfg.Version = types.StringValue(version.Attributes.Version)
	cfg.Latest = types.BoolValue(version.Attributes.Latest)
	cfg.Available = types.BoolValue(version.Attributes.Available)
	cfg.CreatedAt = types.StringValue(version.Attributes.CreatedAt.Format(time.RFC3339))
	cfg.CompressedSize = types.Int64Null()

	if version.Attributes.CompressedSize != nil {
		cfg.CompressedSize = types.Int64Value(int64(*version.Attributes.CompressedSize))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

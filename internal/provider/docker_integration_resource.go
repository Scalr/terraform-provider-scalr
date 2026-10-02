package provider

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ resource.Resource                = &dockerIntegrationResource{}
	_ resource.ResourceWithConfigure   = &dockerIntegrationResource{}
	_ resource.ResourceWithImportState = &dockerIntegrationResource{}
)

func newDockerIntegrationResource() resource.Resource {
	return &dockerIntegrationResource{}
}

// dockerIntegrationResource defines the resource implementation.
type dockerIntegrationResource struct {
	framework.ResourceWithScalrClient
}

// dockerIntegrationResourceModel describes the resource data model.
type dockerIntegrationResourceModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	RegistryURL       types.String `tfsdk:"registry_url"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	ExportCredentials types.Bool   `tfsdk:"export_credentials"`
	Status            types.String `tfsdk:"status"`
}

func dockerIntegrationResourceModelFromAPI(
	di *schemas.DockerIntegration,
	password types.String,
	prior *dockerIntegrationResourceModel,
) *dockerIntegrationResourceModel {
	model := &dockerIntegrationResourceModel{
		Id:                types.StringValue(di.ID),
		Name:              types.StringValue(di.Attributes.Name),
		RegistryURL:       types.StringValue(di.Attributes.RegistryUrl),
		Username:          types.StringPointerValue(di.Attributes.Username),
		Password:          password,
		ExportCredentials: types.BoolValue(di.Attributes.ExportCredentials),
		Status:            types.StringValue(string(di.Attributes.Status)),
	}

	// The API stores the registry URL in a normalized form (e.g. `ghcr.io` becomes `https://ghcr.io`),
	// keep the configured spelling while it points to the same registry.
	if prior != nil && !prior.RegistryURL.IsNull() && !prior.RegistryURL.IsUnknown() &&
		normalizeRegistryURL(prior.RegistryURL.ValueString()) == di.Attributes.RegistryUrl {
		model.RegistryURL = prior.RegistryURL
	}

	return model
}

// normalizeRegistryURL mirrors the backend normalization of a Docker registry URL:
// `<scheme>://<lowercase host>[:<non-default port>]`, with `https` as the default scheme.
// Returns the input unchanged when it can't be parsed.
func normalizeRegistryURL(raw string) string {
	raw = strings.TrimSpace(raw)
	toParse := raw
	if !strings.Contains(raw, "://") {
		toParse = "//" + raw
	}

	u, err := url.Parse(toParse)
	if err != nil || u.Hostname() == "" {
		return raw
	}

	scheme := u.Scheme
	if scheme == "" {
		scheme = "https"
	}

	host := strings.ToLower(u.Hostname())
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}

	port := u.Port()
	if port != "" && !(scheme == "https" && port == "443") && !(scheme == "http" && port == "80") {
		host += ":" + port
	}

	return scheme + "://" + host
}

// warnOnDockerIntegrationFailure surfaces a backend-reported failure as a warning.
func warnOnDockerIntegrationFailure(di *schemas.DockerIntegration, diags *diag.Diagnostics) {
	var errMessage string
	if di.Attributes.ErrorMessage != nil {
		errMessage = strings.TrimSpace(*di.Attributes.ErrorMessage)
	}

	if di.Attributes.Status != schemas.DockerIntegrationStatusFailed && errMessage == "" {
		return
	}

	if errMessage == "" {
		errMessage = "Scalr reported the Docker integration status as failed."
	}

	diags.AddWarning("Issues detected", errMessage)
}

func (r *dockerIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_integration"
}

func (r *dockerIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of Docker (OCI) registry integrations in Scalr." +
			" The integration is used to publish OCI-sourced modules, see `scalr_module`.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Docker integration.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"registry_url": schema.StringAttribute{
				MarkdownDescription: "The registry URL, e.g. `https://ghcr.io`. Only one integration per registry host is allowed." +
					" Redirecting URLs are rejected.",
				Required: true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The registry username.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The registry password or personal access token.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"export_credentials": schema.BoolAttribute{
				MarkdownDescription: "Whether to injects the credentials into plan and apply runs for direct oci:// module access." +
					" Default `false`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the integration after the latest connection test.",
				Computed:            true,
			},
		},
	}
}

func (r *dockerIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dockerIntegrationResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.DockerIntegrationRequest{
		Attributes: schemas.DockerIntegrationAttributesRequest{
			Name:              value.Set(plan.Name.ValueString()),
			RegistryUrl:       value.Set(plan.RegistryURL.ValueString()),
			Username:          value.Set(plan.Username.ValueString()),
			Password:          value.Set(plan.Password.ValueString()),
			ExportCredentials: framework.SetIfKnownBool(plan.ExportCredentials),
		},
	}

	di, err := r.ClientV2.DockerIntegration.CreateDockerIntegration(ctx, &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Docker integration", err.Error())
		return
	}

	warnOnDockerIntegrationFailure(di, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, dockerIntegrationResourceModelFromAPI(di, plan.Password, &plan))...)
}

func (r *dockerIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state dockerIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	di, err := r.ClientV2.DockerIntegration.GetDockerIntegration(ctx, state.Id.ValueString(), nil)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving Docker integration", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, dockerIntegrationResourceModelFromAPI(di, state.Password, &state))...)
}

func (r *dockerIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read plan and state data
	var plan, state dockerIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.DockerIntegrationRequest{}

	if !plan.Name.Equal(state.Name) {
		opts.Attributes.Name = value.Set(plan.Name.ValueString())
	}

	if !plan.RegistryURL.Equal(state.RegistryURL) {
		opts.Attributes.RegistryUrl = value.Set(plan.RegistryURL.ValueString())
	}

	// The API validates username and password together, so send both when either changes.
	if !plan.Username.Equal(state.Username) || !plan.Password.Equal(state.Password) {
		opts.Attributes.Username = value.Set(plan.Username.ValueString())
		opts.Attributes.Password = value.Set(plan.Password.ValueString())
	}

	if !plan.ExportCredentials.Equal(state.ExportCredentials) {
		opts.Attributes.ExportCredentials = value.Set(plan.ExportCredentials.ValueBool())
	}

	di, err := r.ClientV2.DockerIntegration.UpdateDockerIntegration(ctx, plan.Id.ValueString(), &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Docker integration", err.Error())
		return
	}

	warnOnDockerIntegrationFailure(di, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, dockerIntegrationResourceModelFromAPI(di, plan.Password, &plan))...)
}

func (r *dockerIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state dockerIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.DockerIntegration.DeleteDockerIntegration(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting Docker integration", err.Error())
		return
	}
}

func (r *dockerIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

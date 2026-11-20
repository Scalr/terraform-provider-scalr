package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ resource.Resource                = &moduleNamespaceEnvironmentResource{}
	_ resource.ResourceWithConfigure   = &moduleNamespaceEnvironmentResource{}
	_ resource.ResourceWithImportState = &moduleNamespaceEnvironmentResource{}
)

func newModuleNamespaceEnvironmentResource() resource.Resource {
	return &moduleNamespaceEnvironmentResource{}
}

// moduleNamespaceEnvironmentResource defines the resource implementation.
type moduleNamespaceEnvironmentResource struct {
	framework.ResourceWithScalrClient
}

// moduleNamespaceEnvironmentResourceModel describes the resource data model.
type moduleNamespaceEnvironmentResourceModel struct {
	Id                types.String `tfsdk:"id"`
	ModuleNamespaceID types.String `tfsdk:"module_namespace_id"`
	EnvironmentID     types.String `tfsdk:"environment_id"`
}

func (r *moduleNamespaceEnvironmentResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_module_namespace_environment"
}

func (r *moduleNamespaceEnvironmentResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Gives a single environment access to the modules of a module namespace." +
			"\n\nEach link is a separate resource, so an environment can be attached to a shared" +
			" namespace from the configuration that creates the environment. Creating or destroying" +
			" a link leaves the other links of the namespace untouched." +
			"\n\n-> **Note** Manage a namespace's environments either with these resources or with" +
			" the `environments` attribute of the [`scalr_module_namespace`](module_namespace)" +
			" resource, not both. Omit `environments` there to leave the links to this resource." +
			"\n\nThe namespace must not be shared: a shared namespace is available in every" +
			" environment already.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource, in the format `<module_namespace_id>/<environment_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"module_namespace_id": schema.StringAttribute{
				MarkdownDescription: "ID of the module namespace, in the format `modns-<RANDOM STRING>`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
		},
	}
}

func (r *moduleNamespaceEnvironmentResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan moduleNamespaceEnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespaceID := plan.ModuleNamespaceID.ValueString()
	environmentID := plan.EnvironmentID.ValueString()

	err := r.ClientV2.ModuleNamespace.CreateModuleNamespaceEnvironments(
		ctx,
		namespaceID,
		[]schemas.Environment{{ID: environmentID}},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error linking environment to module namespace",
			fmt.Sprintf(
				"Failed to link environment %q to module namespace %q: %s",
				environmentID, namespaceID, err,
			),
		)
		return
	}

	plan.Id = types.StringValue(namespaceID + "/" + environmentID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *moduleNamespaceEnvironmentResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state moduleNamespaceEnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespaceID := state.ModuleNamespaceID.ValueString()
	environmentID := state.EnvironmentID.ValueString()

	namespace, err := r.ClientV2.ModuleNamespace.GetModuleNamespace(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving module namespace", err.Error())
		return
	}

	for _, env := range namespace.Relationships.Environments {
		if env != nil && env.ID == environmentID {
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *moduleNamespaceEnvironmentResource) Update(
	_ context.Context,
	_ resource.UpdateRequest,
	_ *resource.UpdateResponse,
) {
	// Not updatable: any attribute change forces recreate.
}

func (r *moduleNamespaceEnvironmentResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state moduleNamespaceEnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.ModuleNamespace.DeleteModuleNamespaceEnvironments(
		ctx,
		state.ModuleNamespaceID.ValueString(),
		state.EnvironmentID.ValueString(),
	)
	// Either the namespace or the environment may already be deleted.
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error unlinking environment from module namespace", err.Error())
		return
	}
}

func (r *moduleNamespaceEnvironmentResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf(
				"Expected import ID in the format <module_namespace_id>/<environment_id>, got: %q",
				req.ID,
			),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("module_namespace_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_id"), parts[1])...)
}

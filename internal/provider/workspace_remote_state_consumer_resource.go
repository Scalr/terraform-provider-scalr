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
	_ resource.Resource                   = &workspaceRemoteStateConsumerResource{}
	_ resource.ResourceWithConfigure      = &workspaceRemoteStateConsumerResource{}
	_ resource.ResourceWithValidateConfig = &workspaceRemoteStateConsumerResource{}
	_ resource.ResourceWithImportState    = &workspaceRemoteStateConsumerResource{}
)

func newWorkspaceRemoteStateConsumerResource() resource.Resource {
	return &workspaceRemoteStateConsumerResource{}
}

// workspaceRemoteStateConsumerResource defines the resource implementation.
type workspaceRemoteStateConsumerResource struct {
	framework.ResourceWithScalrClient
}

// workspaceRemoteStateConsumerResourceModel describes the resource data model.
type workspaceRemoteStateConsumerResourceModel struct {
	Id          types.String `tfsdk:"id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	ConsumerID  types.String `tfsdk:"consumer_id"`
}

func (r *workspaceRemoteStateConsumerResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_workspace_remote_state_consumer"
}

func (r *workspaceRemoteStateConsumerResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Allows a workspace to access the state of another workspace in Scalr." +
			"\n\nThe source workspace must have `remote_state_sharing` set to `false`." +
			" Do not use this resource together with the `remote_state_consumers` attribute" +
			" of the `scalr_workspace` resource for the same source workspace, as they will conflict.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource, in the format `<workspace_id>/<consumer_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "ID of the workspace whose state is shared.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"consumer_id": schema.StringAttribute{
				MarkdownDescription: "ID of the workspace that is allowed to access the state.",
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

func (r *workspaceRemoteStateConsumerResource) ValidateConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var cfg workspaceRemoteStateConsumerResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.WorkspaceID.IsUnknown() || cfg.ConsumerID.IsUnknown() {
		return
	}

	if cfg.WorkspaceID.ValueString() == cfg.ConsumerID.ValueString() {
		resp.Diagnostics.AddAttributeError(
			path.Root("consumer_id"),
			"Invalid consumer",
			"A workspace cannot be a remote state consumer of itself.",
		)
	}
}

func (r *workspaceRemoteStateConsumerResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan workspaceRemoteStateConsumerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := plan.WorkspaceID.ValueString()
	consumerID := plan.ConsumerID.ValueString()

	err := r.ClientV2.Workspace.AddRemoteStateConsumers(
		ctx,
		workspaceID,
		[]schemas.Workspace{{ID: consumerID}},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding remote state consumer",
			fmt.Sprintf(
				"Failed to add workspace %q as a remote state consumer of workspace %q: %s\n\n"+
					"Ensure the source workspace has `remote_state_sharing` set to `false`.",
				consumerID, workspaceID, err,
			),
		)
		return
	}

	plan.Id = types.StringValue(workspaceID + "/" + consumerID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workspaceRemoteStateConsumerResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state workspaceRemoteStateConsumerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := state.WorkspaceID.ValueString()
	consumerID := state.ConsumerID.ValueString()

	for c, err := range r.ClientV2.Workspace.ListRemoteStateConsumersIter(ctx, workspaceID, nil) {
		if err != nil {
			if errors.Is(err, client.ErrNotFound) {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error reading remote state consumers for workspace %s", workspaceID),
				err.Error(),
			)
			return
		}
		if c.ID == consumerID {
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	// Not found — remove from state
	resp.State.RemoveResource(ctx)
}

func (r *workspaceRemoteStateConsumerResource) Update(
	_ context.Context,
	_ resource.UpdateRequest,
	_ *resource.UpdateResponse,
) {
	// Not updatable - any attribute change forces recreate.
}

func (r *workspaceRemoteStateConsumerResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state workspaceRemoteStateConsumerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.Workspace.DeleteRemoteStateConsumers(
		ctx,
		state.WorkspaceID.ValueString(),
		[]schemas.Workspace{{ID: state.ConsumerID.ValueString()}},
	)
	// Either workspace may already be deleted.
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error removing remote state consumer", err.Error())
		return
	}
}

func (r *workspaceRemoteStateConsumerResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the format <workspace_id>/<consumer_id>, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("consumer_id"), parts[1])...)
}

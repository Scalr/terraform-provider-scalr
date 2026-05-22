package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/schemas"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ resource.Resource                = &workspaceVarSetResource{}
	_ resource.ResourceWithConfigure   = &workspaceVarSetResource{}
	_ resource.ResourceWithImportState = &workspaceVarSetResource{}
)

func newWorkspaceVarSetResource() resource.Resource {
	return &workspaceVarSetResource{}
}

// workspaceVarSetResource defines the resource implementation.
type workspaceVarSetResource struct {
	framework.ResourceWithScalrClient
}

// workspaceVarSetResourceModel describes the resource data model.
type workspaceVarSetResourceModel struct {
	Id          types.String `tfsdk:"id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	VarSetID    types.String `tfsdk:"var_set_id"`
}

func (r *workspaceVarSetResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_workspace_var_set"
}

func (r *workspaceVarSetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the link between a variable set and a workspace in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource, in the format `<workspace_id>/<var_set_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "ID of the workspace.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"var_set_id": schema.StringAttribute{
				MarkdownDescription: "ID of the variable set.",
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

func (r *workspaceVarSetResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan workspaceVarSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := plan.WorkspaceID.ValueString()
	varSetID := plan.VarSetID.ValueString()

	err := r.ClientV2.Workspace.AddWorkspaceVariableSets(
		ctx,
		workspaceID,
		[]schemas.VariableSet{{ID: varSetID}},
	)
	if err != nil {
		resp.Diagnostics.AddError("Error linking variable set to workspace", err.Error())
		return
	}

	plan.Id = types.StringValue(workspaceID + "/" + varSetID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workspaceVarSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workspaceVarSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workspaceID := state.WorkspaceID.ValueString()
	varSetID := state.VarSetID.ValueString()

	for vs, err := range r.ClientV2.Workspace.ListWorkspaceVariableSetsIter(ctx, workspaceID, nil) {
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error reading variable sets for workspace %s", workspaceID),
				err.Error(),
			)
			return
		}
		if vs.ID == varSetID {
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	// Not found — remove from state
	resp.State.RemoveResource(ctx)
}

func (r *workspaceVarSetResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Not updatable - any attribute change forces recreate.
}

func (r *workspaceVarSetResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state workspaceVarSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.Workspace.DeleteWorkspaceVariableSets(
		ctx,
		state.WorkspaceID.ValueString(),
		[]schemas.VariableSet{{ID: state.VarSetID.ValueString()}},
	)
	if err != nil {
		resp.Diagnostics.AddError("Error unlinking variable set from workspace", err.Error())
		return
	}
}

func (r *workspaceVarSetResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the format <workspace_id>/<var_set_id>, got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("var_set_id"), parts[1])...)
}

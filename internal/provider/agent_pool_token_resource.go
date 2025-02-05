package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
)

// Compile-time interface checks
var (
	_ resource.Resource              = &agentPoolTokenResource{}
	_ resource.ResourceWithConfigure = &agentPoolTokenResource{}
)

func newAgentPoolTokenResource() resource.Resource {
	return &agentPoolTokenResource{}
}

// agentPoolTokenResource defines the resource implementation.
type agentPoolTokenResource struct {
	framework.ResourceWithScalrClient
}

// agentPoolTokenResourceModel describes the resource data model.
type agentPoolTokenResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	AgentPoolID types.String `tfsdk:"agent_pool_id"`
	Token       types.String `tfsdk:"token"`
}

func (r *agentPoolTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_pool_token"
}

func (r *agentPoolTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage the state of agent pool's tokens in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the token.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"agent_pool_id": schema.StringAttribute{
				MarkdownDescription: "ID of the agent pool.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The token of the agent pool.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *agentPoolTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan agentPoolTokenResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.AccessTokenCreateOptions{
		Description: plan.Description.ValueStringPointer(),
	}

	agentPoolToken, err := r.Client.AgentPoolTokens.Create(ctx, plan.AgentPoolID.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating agent_pool_token", err.Error())
		return
	}

	plan.Id = types.StringValue(agentPoolToken.ID)
	plan.Description = types.StringValue(agentPoolToken.Description)
	plan.Token = types.StringValue(agentPoolToken.Token)

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *agentPoolTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state agentPoolTokenResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.AccessTokenListOptions{}

	for {
		tokensList, err := r.Client.AgentPoolTokens.List(ctx, state.AgentPoolID.ValueString(), opts)

		if err != nil {
			if errors.Is(err, scalr.ErrResourceNotFound) {
				resp.State.RemoveResource(ctx)
				return
			}
			resp.Diagnostics.AddError("Error retrieving agent_pool_token", err.Error())
			return
		}

		for _, t := range tokensList.Items {
			if t.ID == state.Id.ValueString() {
				state.Description = types.StringValue(t.Description)

				// Set refreshed state
				resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
				return
			}
		}

		// Exit the loop when we've seen all pages.
		if tokensList.CurrentPage >= tokensList.TotalPages {
			break
		}

		// Update the page number to get the next page.
		opts.PageNumber = tokensList.NextPage
	}

	// The token has been deleted
	resp.State.RemoveResource(ctx)
}

func (r *agentPoolTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan agentPoolTokenResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.AccessTokenUpdateOptions{
		Description: plan.Description.ValueStringPointer(),
	}

	// Update existing resource
	agentPoolToken, err := r.Client.AccessTokens.Update(ctx, plan.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating agent_pool_token", err.Error())
		return
	}

	// Overwrite attributes with refreshed values
	plan.Description = types.StringValue(agentPoolToken.Description)

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *agentPoolTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state agentPoolTokenResourceModel

	// Get current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.AccessTokens.Delete(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting agent_pool_token", err.Error())
		return
	}
}

package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &environmentHookResource{}
	_ resource.ResourceWithConfigure        = &environmentHookResource{}
	_ resource.ResourceWithConfigValidators = &environmentHookResource{}
	_ resource.ResourceWithImportState      = &environmentHookResource{}
)

func newEnvironmentHookResource() resource.Resource {
	return &environmentHookResource{}
}

// The list of allowed hook events
var allowedHookEvents = []string{
	"pre-init", "pre-plan", "post-plan", "pre-apply", "post-apply",
}

// environmentHookResource defines the resource implementation.
type environmentHookResource struct {
	framework.ResourceWithScalrClient
}

// environmentHookResourceModel describes the resource data model.
type environmentHookResourceModel struct {
	Id            types.String `tfsdk:"id"`
	HookId        types.String `tfsdk:"hook_id"`
	EnvironmentId types.String `tfsdk:"environment_id"`
	Events        types.List   `tfsdk:"events"`
}

func (r *environmentHookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_hook"
}

func (r *environmentHookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the link between a hook and an environment in Scalr. This allows you to attach hooks to specific environments for execution during the Terraform workflow.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this link resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hook_id": schema.StringAttribute{
				MarkdownDescription: "ID of the hook, in the format `hook-<RANDOM STRING>`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"events": schema.ListAttribute{
				MarkdownDescription: "List of events that trigger the hook execution. Valid values include: `pre-init`, `pre-plan`, `post-plan`, `pre-apply`, `post-apply`. Use `[\"*\"]` to select all events. Each event can only be specified once.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf(append([]string{"*"}, allowedHookEvents...)...),
					),
					// Ensure events are unique
					listvalidator.UniqueValues(),
				},
			},
		},
	}
}

func (r *environmentHookResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

// Helper function to deduplicate events
func deduplicateEvents(events []string) []string {
	seen := make(map[string]struct{}, len(events))
	result := make([]string, 0, len(events))

	for _, event := range events {
		if _, exists := seen[event]; !exists {
			seen[event] = struct{}{}
			result = append(result, event)
		}
	}

	return result
}

func (r *environmentHookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan environmentHookResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var eventsSlice []string
	resp.Diagnostics.Append(plan.Events.ElementsAs(ctx, &eventsSlice, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	opts := scalr.EnvironmentHookCreateOptions{
		Hook:        &scalr.Hook{ID: plan.HookId.ValueString()},
		Environment: &scalr.Environment{ID: plan.EnvironmentId.ValueString()},
	}

	if len(eventsSlice) == 1 && eventsSlice[0] == "*" {
		opts.Events = allowedHookEvents
	} else {
		opts.Events = eventsSlice
	}

	link, err := r.Client.EnvironmentHooks.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating environment hook", err.Error())
		return
	}

	plan.Id = types.StringValue(link.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *environmentHookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state environmentHookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	link, err := r.Client.EnvironmentHooks.Read(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading environment-hook link", err.Error())
		return
	}

	if link.Hook != nil {
		state.HookId = types.StringValue(link.Hook.ID)
	}

	if link.Environment != nil {
		state.EnvironmentId = types.StringValue(link.Environment.ID)
	}

	// Check if API returned all possible events, and convert back to "*" if needed
	if containsAllEvents(link.Events) {
		// Get the events from state to see if "*" was originally configured
		var eventsFromState []string
		resp.Diagnostics.Append(state.Events.ElementsAs(ctx, &eventsFromState, false)...)

		// If there was a "*" in the state or we have all events returned from API, use "*"
		if len(eventsFromState) == 1 && eventsFromState[0] == "*" {
			state.Events, _ = types.ListValueFrom(ctx, types.StringType, []string{"*"})
		} else {
			// Otherwise, use the actual list from API
			eventsList, diags := types.ListValueFrom(ctx, types.StringType, link.Events)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			state.Events = eventsList
		}
	} else {
		// Not all events, just use the list from API
		eventsList, diags := types.ListValueFrom(ctx, types.StringType, link.Events)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Events = eventsList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Helper function to check if the events slice contains all allowed events
func containsAllEvents(events []string) bool {
	if len(events) != len(allowedHookEvents) {
		return false
	}

	eventMap := make(map[string]bool)
	for _, event := range events {
		eventMap[event] = true
	}

	for _, allowedEvent := range allowedHookEvents {
		if !eventMap[allowedEvent] {
			return false
		}
	}

	return true
}

func (r *environmentHookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan environmentHookResourceModel
	var state environmentHookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Events.Equal(state.Events) {
		var eventsSlice []string
		resp.Diagnostics.Append(plan.Events.ElementsAs(ctx, &eventsSlice, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateOpts := scalr.EnvironmentHookUpdateOptions{}

		if len(eventsSlice) == 1 && eventsSlice[0] == "*" {
			updateOpts.Events = &allowedHookEvents
		} else {
			eventsSlice = deduplicateEvents(eventsSlice)
			updateOpts.Events = &eventsSlice
		}

		_, err := r.Client.EnvironmentHooks.Update(ctx, state.Id.ValueString(), updateOpts)
		if err != nil {
			resp.Diagnostics.AddError("Error updating environment hook", err.Error())
			return
		}
	}

	plan.Id = state.Id
	plan.HookId = state.HookId
	plan.EnvironmentId = state.EnvironmentId

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *environmentHookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state environmentHookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.EnvironmentHooks.Delete(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting environment-hook link", err.Error())
		return
	}
}

func (r *environmentHookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

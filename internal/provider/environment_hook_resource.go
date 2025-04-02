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

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/planmodifiers"
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
	Events        types.Set    `tfsdk:"events"`
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
			"events": schema.SetAttribute{
				MarkdownDescription: "Set of events that trigger the hook execution. Valid values include: `pre-init`, `pre-plan`, `post-plan`, `pre-apply`, `post-apply`. Use `set( [\"*\"] )` to select all events.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(append([]string{"*"}, allowedHookEvents...)...),
					),
				},
				PlanModifiers: []planmodifier.Set{
					planmodifiers.StringSliceAllEquivalent(allowedHookEvents),
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

	eventsSet, diags := types.SetValueFrom(ctx, types.StringType, link.Events)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Events = eventsSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
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

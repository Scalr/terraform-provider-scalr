package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

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

var asteriskSetValue = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("*")})

// environmentHookResource defines the resource implementation.
type environmentHookResource struct {
	framework.ResourceWithScalrClient
}

// environmentHookResourceModel describes the resource data model.
type environmentHookResourceModel struct {
	Id            types.String `tfsdk:"id"`
	HookID        types.String `tfsdk:"hook_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
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
				MarkdownDescription: "The ID of this resource.",
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
				MarkdownDescription: "Set of events that trigger the hook execution. Valid values include: `pre-init`, `pre-plan`, `post-plan`, `pre-apply`, `post-apply`. Use `[\"*\"]` to select all events.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(append(allowedHookEvents, "*")...),
					),
				},
			},
		},
	}
}

func (r *environmentHookResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *environmentHookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var cfg environmentHookResourceModel
	var plan environmentHookResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.EnvironmentHookCreateOptions{
		Hook:        &scalr.Hook{ID: plan.HookID.ValueString()},
		Environment: &scalr.Environment{ID: plan.EnvironmentID.ValueString()},
	}

	var events []string
	resp.Diagnostics.Append(plan.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(events) == 1 && events[0] == "*" {
		// Expand the "*" to all allowed events for the API call
		opts.Events = allowedHookEvents
	} else {
		opts.Events = events
	}

	envHook, err := r.Client.EnvironmentHooks.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating environment hook", err.Error())
		return
	}

	result, d := environmentHookResourceModelFromAPI(ctx, envHook, cfg.Events)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
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

	envHook, err := r.Client.EnvironmentHooks.Read(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading environment hook", err.Error())
		return
	}

	result, d := environmentHookResourceModelFromAPI(ctx, envHook, state.Events)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *environmentHookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var cfg environmentHookResourceModel
	var plan environmentHookResourceModel
	var state environmentHookResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.EnvironmentHookUpdateOptions{}

	var events []string
	resp.Diagnostics.Append(plan.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(events) == 1 && events[0] == "*" {
		// Expand the "*" to all allowed events for the API call
		opts.Events = &allowedHookEvents
	} else {
		opts.Events = &events
	}

	envHook, err := r.Client.EnvironmentHooks.Update(ctx, state.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating environment hook", err.Error())
		return
	}

	result, d := environmentHookResourceModelFromAPI(ctx, envHook, cfg.Events)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
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
		resp.Diagnostics.AddError("Error deleting environment hook", err.Error())
		return
	}
}

func (r *environmentHookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func environmentHookResourceModelFromAPI(
	ctx context.Context,
	eh *scalr.EnvironmentHook,
	cfgEventsValue types.Set,
) (*environmentHookResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &environmentHookResourceModel{
		Id:            types.StringValue(eh.ID),
		HookID:        types.StringNull(),
		EnvironmentID: types.StringNull(),
		Events:        types.SetNull(types.StringType),
	}

	if eh.Hook != nil {
		model.HookID = types.StringValue(eh.Hook.ID)
	}
	if eh.Environment != nil {
		model.EnvironmentID = types.StringValue(eh.Environment.ID)
	}

	// If all events are selected, collapse the value to "*",
	// but only if the value in config is set to "*" too.
	if setsEqual(eh.Events, allowedHookEvents) && cfgEventsValue.Equal(asteriskSetValue) {
		model.Events = asteriskSetValue
	} else {
		// Otherwise set the values as seen in the API response
		eventValues, d := types.SetValueFrom(ctx, types.StringType, eh.Events)
		diags.Append(d...)
		model.Events = eventValues
	}

	return model, diags
}

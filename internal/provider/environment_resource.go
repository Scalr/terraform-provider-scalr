package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/defaults"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Compile-time interface checks
var (
	_ resource.Resource                = &environmentResource{}
	_ resource.ResourceWithConfigure   = &environmentResource{}
	_ resource.ResourceWithImportState = &environmentResource{}
)

func newEnvironmentResource() resource.Resource {
	return &environmentResource{}
}

// environmentResource defines the resource implementation.
type environmentResource struct {
	framework.ResourceWithScalrClient
}

// environmentResourceModel describes the resource data model.
type environmentResourceModel struct {
	Id                            types.String `tfsdk:"id"`
	Name                          types.String `tfsdk:"name"`
	Status                        types.String `tfsdk:"status"`
	CreatedBy                     types.List   `tfsdk:"created_by"`
	PolicyGroups                  types.List   `tfsdk:"policy_groups"`
	DefaultProviderConfigurations types.Set    `tfsdk:"default_provider_configurations"`
	TagIDs                        types.Set    `tfsdk:"tag_ids"`
	RemoteBackend                 types.Bool   `tfsdk:"remote_backend"`
	MaskSensitiveOutput           types.Bool   `tfsdk:"mask_sensitive_output"`
	AccountID                     types.String `tfsdk:"account_id"`
}

func environmentResourceModelFromAPI(ctx context.Context, env *scalr.Environment) (*environmentResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &environmentResourceModel{
		Id:                            types.StringValue(env.ID),
		Name:                          types.StringValue(env.Name),
		Status:                        types.StringValue(string(env.Status)),
		CreatedBy:                     types.ListNull(userElementType),
		PolicyGroups:                  types.ListNull(types.StringType),
		DefaultProviderConfigurations: types.SetNull(types.StringType),
		TagIDs:                        types.SetNull(types.StringType),
		RemoteBackend:                 types.BoolValue(env.RemoteBackend),
		MaskSensitiveOutput:           types.BoolValue(env.MaskSensitiveOutput),
		AccountID:                     types.StringValue(env.Account.ID),
	}

	if env.CreatedBy != nil {
		createdBy := []userModel{*userModelFromAPI(env.CreatedBy)}
		createdByValue, d := types.ListValueFrom(ctx, userElementType, createdBy)
		diags.Append(d...)
		model.CreatedBy = createdByValue
	}

	policyGroups := make([]string, len(env.PolicyGroups))
	for i, group := range env.PolicyGroups {
		policyGroups[i] = group.ID
	}
	policyGroupsValue, d := types.ListValueFrom(ctx, types.StringType, policyGroups)
	diags.Append(d...)
	model.PolicyGroups = policyGroupsValue

	defaultPcfgs := make([]string, len(env.DefaultProviderConfigurations))
	for i, pcfg := range env.DefaultProviderConfigurations {
		defaultPcfgs[i] = pcfg.ID
	}
	defaultPcfgsValue, d := types.SetValueFrom(ctx, types.StringType, defaultPcfgs)
	diags.Append(d...)
	model.DefaultProviderConfigurations = defaultPcfgsValue

	tags := make([]string, len(env.Tags))
	for i, tag := range env.Tags {
		tags[i] = tag.ID
	}
	tagsValue, d := types.SetValueFrom(ctx, types.StringType, tags)
	diags.Append(d...)
	model.TagIDs = tagsValue

	return model, diags
}

func (r *environmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *environmentResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	emptyStringSet, _ := types.SetValueFrom(ctx, types.StringType, []string{})

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of environments in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the environment.",
				Required:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the environment.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.ListAttribute{
				MarkdownDescription: "Details of the user that created the environment.",
				ElementType:         userElementType,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_groups": schema.ListAttribute{
				MarkdownDescription: "List of the environment policy-groups IDs, in the format `pgrp-<RANDOM STRING>`.",
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"default_provider_configurations": schema.SetAttribute{
				MarkdownDescription: "List of IDs of provider configurations, used in the environment workspaces by default.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(emptyStringSet),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
			},
			"tag_ids": schema.SetAttribute{
				MarkdownDescription: "List of tag IDs associated with the environment.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(emptyStringSet),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
			},
			"remote_backend": schema.BoolAttribute{
				MarkdownDescription: "If Scalr exports the remote backend configuration and state storage for your infrastructure management." +
					" Disabling this feature will also prevent the ability to perform state locking, which ensures that concurrent operations do not conflict." +
					" Additionally, it will disable the capability to initiate CLI-driven runs through Scalr.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"mask_sensitive_output": schema.BoolAttribute{
				MarkdownDescription: "Enable masking of the sensitive console output. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "ID of the account, in the format `acc-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				Default:             defaults.AccountIDRequired(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *environmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan environmentResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.EnvironmentCreateOptions{
		Name:                plan.Name.ValueStringPointer(),
		Account:             &scalr.Account{ID: plan.AccountID.ValueString()},
		MaskSensitiveOutput: plan.MaskSensitiveOutput.ValueBoolPointer(),
	}

	if !plan.RemoteBackend.IsUnknown() && !plan.RemoteBackend.IsNull() {
		opts.RemoteBackend = plan.RemoteBackend.ValueBoolPointer()
	}

	if !plan.DefaultProviderConfigurations.IsUnknown() && !plan.DefaultProviderConfigurations.IsNull() {
		var defaultPcfgIDs []string
		resp.Diagnostics.Append(plan.DefaultProviderConfigurations.ElementsAs(ctx, &defaultPcfgIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		defaultPcfgs := make([]*scalr.ProviderConfiguration, len(defaultPcfgIDs))
		for i, pcfgID := range defaultPcfgIDs {
			defaultPcfgs[i] = &scalr.ProviderConfiguration{ID: pcfgID}
		}

		opts.DefaultProviderConfigurations = defaultPcfgs
	}

	if !plan.TagIDs.IsUnknown() && !plan.TagIDs.IsNull() {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagIDs.ElementsAs(ctx, &tagIDs, false)...)

		tags := make([]*scalr.Tag, len(tagIDs))
		for i, tagID := range tagIDs {
			tags[i] = &scalr.Tag{ID: tagID}
		}

		opts.Tags = tags
	}

	environment, err := r.Client.Environments.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating environment", err.Error())
		return
	}

	// Get refreshed resource state from API
	environment, err = r.Client.Environments.Read(ctx, environment.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving environment", err.Error())
		return
	}

	result, diags := environmentResourceModelFromAPI(ctx, environment)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *environmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state environmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	environment, err := r.Client.Environments.Read(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving environment", err.Error())
		return
	}

	result, diags := environmentResourceModelFromAPI(ctx, environment)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *environmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state environmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := scalr.EnvironmentUpdateOptions{}

	if !plan.Name.Equal(state.Name) {
		opts.Name = plan.Name.ValueStringPointer()
	}

	if !plan.MaskSensitiveOutput.Equal(state.MaskSensitiveOutput) {
		opts.MaskSensitiveOutput = plan.MaskSensitiveOutput.ValueBoolPointer()
	}

	if !plan.DefaultProviderConfigurations.IsNull() {
		var defaultPcfgIDs []string
		resp.Diagnostics.Append(plan.DefaultProviderConfigurations.ElementsAs(ctx, &defaultPcfgIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		defaultPcfgs := make([]*scalr.ProviderConfiguration, len(defaultPcfgIDs))
		for i, pcfgID := range defaultPcfgIDs {
			defaultPcfgs[i] = &scalr.ProviderConfiguration{ID: pcfgID}
		}

		opts.DefaultProviderConfigurations = defaultPcfgs
	}

	// Update existing resource
	_, err := r.Client.Environments.Update(ctx, plan.Id.ValueString(), opts)
	if err != nil {
		resp.Diagnostics.AddError("Error updating environment", err.Error())
		return
	}

	if !plan.TagIDs.Equal(state.TagIDs) {
		var planTags []string
		var stateTags []string
		resp.Diagnostics.Append(plan.TagIDs.ElementsAs(ctx, &planTags, false)...)
		resp.Diagnostics.Append(state.TagIDs.ElementsAs(ctx, &stateTags, false)...)

		tagsToAdd, tagsToRemove := diff(stateTags, planTags)

		if len(tagsToAdd) > 0 {
			tagRelations := make([]*scalr.TagRelation, len(tagsToAdd))
			for i, tag := range tagsToAdd {
				tagRelations[i] = &scalr.TagRelation{ID: tag}
			}
			err = r.Client.EnvironmentTags.Add(ctx, plan.Id.ValueString(), tagRelations)
			if err != nil {
				resp.Diagnostics.AddError("Error adding tags to environment", err.Error())
			}
		}

		if len(tagsToRemove) > 0 {
			tagRelations := make([]*scalr.TagRelation, len(tagsToRemove))
			for i, tag := range tagsToRemove {
				tagRelations[i] = &scalr.TagRelation{ID: tag}
			}
			err = r.Client.EnvironmentTags.Delete(ctx, plan.Id.ValueString(), tagRelations)
			if err != nil {
				resp.Diagnostics.AddError("Error removing tags from environment", err.Error())
			}
		}
	}

	// Get refreshed resource state from API
	environment, err := r.Client.Environments.Read(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving environment", err.Error())
		return
	}

	result, diags := environmentResourceModelFromAPI(ctx, environment)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *environmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state environmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.Environments.Delete(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting environment", err.Error())
		return
	}
}

func (r *environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

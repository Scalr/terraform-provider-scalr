package provider

import (
	"context"
	"errors"
	"sort"

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
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
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
	RemoteBackendOverridable      types.Bool   `tfsdk:"remote_backend_overridable"`
	MaskSensitiveOutput           types.Bool   `tfsdk:"mask_sensitive_output"`
	FederatedEnvironments         types.Set    `tfsdk:"federated_environments"`
	AccountID                     types.String `tfsdk:"account_id"`
	StorageProfileID              types.String `tfsdk:"storage_profile_id"`
	DefaultWorkspaceAgentPoolID   types.String `tfsdk:"default_workspace_agent_pool_id"`
}

func environmentResourceModelFromAPI(ctx context.Context, env *scalr.Environment, federatedEnvironments []string) (*environmentResourceModel, diag.Diagnostics) {
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
		RemoteBackendOverridable:      types.BoolValue(env.RemoteBackendOverridable),
		MaskSensitiveOutput:           types.BoolValue(env.MaskSensitiveOutput),
		FederatedEnvironments:         types.SetNull(types.StringType),
		AccountID:                     types.StringValue(env.Account.ID),
		StorageProfileID:              types.StringNull(),
		DefaultWorkspaceAgentPoolID:   types.StringNull(),
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
	sort.Strings(policyGroups)
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

	if env.IsFederatedToAccount {
		federatedEnvironments = []string{"*"}
	} else if federatedEnvironments == nil {
		federatedEnvironments = []string{}
	}
	federatedValue, d := types.SetValueFrom(ctx, types.StringType, federatedEnvironments)
	diags.Append(d...)
	model.FederatedEnvironments = federatedValue

	if env.StorageProfile != nil {
		model.StorageProfileID = types.StringValue(env.StorageProfile.ID)
	}

	if env.DefaultWorkspaceAgentPool != nil {
		model.DefaultWorkspaceAgentPoolID = types.StringValue(env.DefaultWorkspaceAgentPool.ID)
	}

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
					stringvalidation.StringIsNotWhiteSpace(),
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
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidation.StringIsNotWhiteSpace()),
				},
			},
			"tag_ids": schema.SetAttribute{
				MarkdownDescription: "List of tag IDs associated with the environment.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(emptyStringSet),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidation.StringIsNotWhiteSpace()),
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
			"remote_backend_overridable": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the remote backend configuration can be overridden on the workspace level.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"mask_sensitive_output": schema.BoolAttribute{
				MarkdownDescription: "Enable masking of the sensitive console output. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"federated_environments": schema.SetAttribute{
				MarkdownDescription: "The list of environment identifiers that are allowed to access this environment. Use `[\"*\"]` to share with all environments.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Use the scalr_federated_environments resource instead. This attribute will be removed in the future.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidation.StringIsNotWhiteSpace()),
				},
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
			"storage_profile_id": schema.StringAttribute{
				MarkdownDescription: "The storage profile for this environment. If not set, the account's default storage profile will be used.",
				Optional:            true,
			},
			"default_workspace_agent_pool_id": schema.StringAttribute{
				MarkdownDescription: "Default agent pool that will be set for the entire environment. It will be used by a workspace if no other pool is explicitly linked.",
				Optional:            true,
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
		Name:                     plan.Name.ValueStringPointer(),
		Account:                  &scalr.Account{ID: plan.AccountID.ValueString()},
		MaskSensitiveOutput:      plan.MaskSensitiveOutput.ValueBoolPointer(),
		RemoteBackendOverridable: plan.RemoteBackendOverridable.ValueBoolPointer(),
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

	federatedEnvironments := make([]*scalr.EnvironmentRelation, 0)
	if !plan.FederatedEnvironments.IsUnknown() && !plan.FederatedEnvironments.IsNull() {
		opts.IsFederatedToAccount = ptr(false)
		var federatedIDs []string
		resp.Diagnostics.Append(plan.FederatedEnvironments.ElementsAs(ctx, &federatedIDs, false)...)

		if (len(federatedIDs) == 1) && (federatedIDs[0] == "*") {
			opts.IsFederatedToAccount = ptr(true)
		} else if len(federatedIDs) > 0 {
			for _, envID := range federatedIDs {
				federatedEnvironments = append(federatedEnvironments, &scalr.EnvironmentRelation{ID: envID})
			}
		}
	}

	if !plan.StorageProfileID.IsUnknown() && !plan.StorageProfileID.IsNull() {
		opts.StorageProfile = &scalr.StorageProfile{ID: plan.StorageProfileID.ValueString()}
	}

	if !plan.DefaultWorkspaceAgentPoolID.IsUnknown() && !plan.DefaultWorkspaceAgentPoolID.IsNull() {
		opts.DefaultWorkspaceAgentPool = &scalr.AgentPool{ID: plan.DefaultWorkspaceAgentPoolID.ValueString()}
	}

	environment, err := r.Client.Environments.Create(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating environment", err.Error())
		return
	}

	if len(federatedEnvironments) > 0 {
		err = r.Client.FederatedEnvironments.Add(ctx, environment.ID, federatedEnvironments)
		if err != nil {
			resp.Diagnostics.AddError("Error adding federated environments", err.Error())
			return
		}
	}

	// Get refreshed resource state from API
	environment, err = r.Client.Environments.Read(ctx, environment.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving environment", err.Error())
		return
	}

	federated, err := getFederatedEnvironments(ctx, r.Client, environment.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving federated environments", err.Error())
	}

	result, diags := environmentResourceModelFromAPI(ctx, environment, federated)
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

	federated, err := getFederatedEnvironments(ctx, r.Client, environment.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving federated environments", err.Error())
	}

	result, diags := environmentResourceModelFromAPI(ctx, environment, federated)
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

	if !plan.RemoteBackendOverridable.Equal(state.RemoteBackendOverridable) {
		opts.RemoteBackendOverridable = plan.RemoteBackendOverridable.ValueBoolPointer()
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

	var federatedToAdd, federatedToRemove []string
	if !plan.FederatedEnvironments.Equal(state.FederatedEnvironments) {
		var planFederated []string
		var stateFederated []string

		if !plan.FederatedEnvironments.IsUnknown() && !plan.FederatedEnvironments.IsNull() {
			resp.Diagnostics.Append(plan.FederatedEnvironments.ElementsAs(ctx, &planFederated, false)...)
		}
		if !state.FederatedEnvironments.IsUnknown() && !state.FederatedEnvironments.IsNull() {
			resp.Diagnostics.Append(state.FederatedEnvironments.ElementsAs(ctx, &stateFederated, false)...)
		}

		opts.IsFederatedToAccount = ptr(false)

		if len(planFederated) == 1 && planFederated[0] == "*" {
			opts.IsFederatedToAccount = ptr(true)
			planFederated = []string{}
		}
		if len(stateFederated) == 1 && stateFederated[0] == "*" {
			stateFederated = []string{}
		}

		federatedToAdd, federatedToRemove = diff(stateFederated, planFederated)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.StorageProfileID.IsUnknown() && !plan.StorageProfileID.IsNull() {
		opts.StorageProfile = &scalr.StorageProfile{ID: plan.StorageProfileID.ValueString()}
	}

	if !plan.DefaultWorkspaceAgentPoolID.IsUnknown() && !plan.DefaultWorkspaceAgentPoolID.IsNull() {
		opts.DefaultWorkspaceAgentPool = &scalr.AgentPool{ID: plan.DefaultWorkspaceAgentPoolID.ValueString()}
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

	if len(federatedToAdd) > 0 {
		e := make([]*scalr.EnvironmentRelation, len(federatedToAdd))
		for i, env := range federatedToAdd {
			e[i] = &scalr.EnvironmentRelation{ID: env}
		}
		err = r.Client.FederatedEnvironments.Add(ctx, plan.Id.ValueString(), e)
		if err != nil {
			resp.Diagnostics.AddError("Error adding federated environments", err.Error())
		}
	}
	if len(federatedToRemove) > 0 {
		e := make([]*scalr.EnvironmentRelation, len(federatedToRemove))
		for i, env := range federatedToRemove {
			e[i] = &scalr.EnvironmentRelation{ID: env}
		}
		err = r.Client.FederatedEnvironments.Delete(ctx, plan.Id.ValueString(), e)
		if err != nil {
			resp.Diagnostics.AddError("Error removing federated environments", err.Error())
		}
	}

	// Get refreshed resource state from API
	environment, err := r.Client.Environments.Read(ctx, plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving environment", err.Error())
		return
	}

	federated, err := getFederatedEnvironments(ctx, r.Client, environment.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving federated environments", err.Error())
	}

	result, diags := environmentResourceModelFromAPI(ctx, environment, federated)
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

func getFederatedEnvironments(ctx context.Context, scalrClient *scalr.Client, envID string) (envs []string, err error) {
	listOpts := scalr.ListOptions{}
	for {
		el, err := scalrClient.FederatedEnvironments.List(ctx, envID, listOpts)
		if err != nil {
			return nil, err
		}

		for _, e := range el.Items {
			envs = append(envs, e.ID)
		}

		if el.CurrentPage >= el.TotalPages {
			break
		}
		listOpts.PageNumber = el.NextPage
	}
	return
}

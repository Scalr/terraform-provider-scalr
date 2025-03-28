package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &hookResource{}
	_ resource.ResourceWithConfigure        = &hookResource{}
	_ resource.ResourceWithConfigValidators = &hookResource{}
	_ resource.ResourceWithImportState      = &hookResource{}
	_ resource.ResourceWithModifyPlan       = &hookResource{}
)

func newHookResource() resource.Resource {
	return &hookResource{}
}

// hookResource defines the resource implementation.
type hookResource struct {
	framework.ResourceWithScalrClient
}

// hookResourceModel describes the resource data model.
type hookResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Interpreter    types.String `tfsdk:"interpreter"`
	ScriptfilePath types.String `tfsdk:"scriptfile_path"`
	VcsProviderId  types.String `tfsdk:"vcs_provider_id"`
	VcsRepo        types.List   `tfsdk:"vcs_repo"`
}

// hookResourceVcsRepoModel maps the vcs_repo nested schema data.
type hookResourceVcsRepoModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Branch     types.String `tfsdk:"branch"`
}

func (r *hookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hook"
}

func (r *hookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a hook in Scalr. Hooks allow you to execute custom scripts at different stages of the Terraform workflow.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the hook.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the hook.",
				Required:            true,
				Validators: []validator.String{
					validation.StringIsNotWhiteSpace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the hook.",
				Optional:            true,
			},
			"interpreter": schema.StringAttribute{
				MarkdownDescription: "The interpreter to execute the hook script, such as 'bash', 'python3', etc.",
				Required:            true,
			},
			"scriptfile_path": schema.StringAttribute{
				MarkdownDescription: "Path to the script file in the repository.",
				Required:            true,
			},
			"vcs_provider_id": schema.StringAttribute{
				MarkdownDescription: "ID of the VCS provider in the format `vcs-<RANDOM STRING>`.",
				Required:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"vcs_repo": schema.ListNestedBlock{
				MarkdownDescription: "Source configuration of a VCS repository.",
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							MarkdownDescription: "The identifier of a VCS repository in the format `:org/:repo`.",
							Required:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"branch": schema.StringAttribute{
							MarkdownDescription: "Repository branch name.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *hookResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{}
}

func (r *hookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := scalr.HookCreateOptions{
		Name:           plan.Name.ValueString(),
		Interpreter:    plan.Interpreter.ValueString(),
		ScriptfilePath: plan.ScriptfilePath.ValueString(),
		VcsProvider:    &scalr.VcsProvider{ID: plan.VcsProviderId.ValueString()},
	}

	if !plan.VcsRepo.IsNull() {
		var vcsRepo []hookResourceVcsRepoModel
		resp.Diagnostics.Append(plan.VcsRepo.ElementsAs(ctx, &vcsRepo, false)...)

		if len(vcsRepo) > 0 {
			repo := vcsRepo[0]
			vcsRepoOptions := &scalr.HookVcsRepo{
				Identifier: repo.Identifier.ValueString(),
			}

			if !repo.Branch.IsNull() && !repo.Branch.IsUnknown() {
				vcsRepoOptions.Branch = repo.Branch.ValueString()
			}

			options.VcsRepo = vcsRepoOptions
		}
	}

	if !plan.Description.IsNull() {
		options.Description = scalr.String(plan.Description.ValueString())
	}

	hook, err := r.Client.Hooks.Create(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError("Error creating hook", err.Error())
		return
	}

	plan.Id = types.StringValue(hook.ID)
	plan.Name = types.StringValue(hook.Name)

	if hook.Description != "" {
		plan.Description = types.StringValue(hook.Description)
	} else {
		plan.Description = types.StringNull()
	}

	plan.Interpreter = types.StringValue(hook.Interpreter)
	plan.ScriptfilePath = types.StringValue(hook.ScriptfilePath)

	if hook.VcsProvider != nil {
		plan.VcsProviderId = types.StringValue(hook.VcsProvider.ID)
	}

	if hook.VcsRepo != nil {
		vcsRepoModel := hookResourceVcsRepoModel{
			Identifier: types.StringValue(hook.VcsRepo.Identifier),
			Branch:     types.StringValue(hook.VcsRepo.Branch),
		}

		vcsRepos := []hookResourceVcsRepoModel{vcsRepoModel}

		vcsRepoList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"identifier": types.StringType,
				"branch":     types.StringType,
			},
		}, vcsRepos)

		if diags.HasError() {
			resp.Diagnostics.AddError("Error creating VCS repo list", diags.Errors()[0].Summary())
			return
		}

		plan.VcsRepo = vcsRepoList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *hookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hook, err := r.Client.Hooks.Read(ctx, state.Id.ValueString())
	if err != nil {
		if errors.Is(err, scalr.ErrResourceNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving hook", err.Error())
		return
	}

	state.Name = types.StringValue(hook.Name)

	if hook.Description != "" {
		state.Description = types.StringValue(hook.Description)
	} else {
		state.Description = types.StringNull()
	}

	state.Interpreter = types.StringValue(hook.Interpreter)
	state.ScriptfilePath = types.StringValue(hook.ScriptfilePath)

	if hook.VcsProvider != nil {
		state.VcsProviderId = types.StringValue(hook.VcsProvider.ID)
	}

	if hook.VcsRepo != nil {
		vcsRepoModel := hookResourceVcsRepoModel{
			Identifier: types.StringValue(hook.VcsRepo.Identifier),
			Branch:     types.StringValue(hook.VcsRepo.Branch),
		}

		vcsRepos := []hookResourceVcsRepoModel{vcsRepoModel}

		vcsRepoList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"identifier": types.StringType,
				"branch":     types.StringType,
			},
		}, vcsRepos)

		if diags.HasError() {
			resp.Diagnostics.AddError("Error converting VCS repo list", diags.Errors()[0].Summary())
			return
		}

		state.VcsRepo = vcsRepoList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hookResourceModel
	var state hookResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := scalr.HookUpdateOptions{
		Name:           scalr.String(plan.Name.ValueString()),
		Interpreter:    scalr.String(plan.Interpreter.ValueString()),
		ScriptfilePath: scalr.String(plan.ScriptfilePath.ValueString()),
		VcsProvider:    &scalr.VcsProvider{ID: plan.VcsProviderId.ValueString()},
	}

	if !plan.VcsRepo.IsNull() {
		var planVcsRepo []hookResourceVcsRepoModel
		var stateVcsRepo []hookResourceVcsRepoModel

		resp.Diagnostics.Append(plan.VcsRepo.ElementsAs(ctx, &planVcsRepo, false)...)
		resp.Diagnostics.Append(state.VcsRepo.ElementsAs(ctx, &stateVcsRepo, false)...)

		if len(planVcsRepo) > 0 {
			planRepo := planVcsRepo[0]

			vcsRepoOptions := &scalr.HookVcsRepo{
				Identifier: planRepo.Identifier.ValueString(),
			}

			if !planRepo.Branch.IsNull() && !planRepo.Branch.IsUnknown() {
				vcsRepoOptions.Branch = planRepo.Branch.ValueString()
			} else if len(stateVcsRepo) > 0 {
				vcsRepoOptions.Branch = stateVcsRepo[0].Branch.ValueString()
			}

			options.VcsRepo = vcsRepoOptions
		}
	}

	if !plan.Description.IsNull() {
		options.Description = scalr.String(plan.Description.ValueString())
	} else {
		options.Description = scalr.String("")
	}

	hook, err := r.Client.Hooks.Update(ctx, plan.Id.ValueString(), options)
	if err != nil {
		resp.Diagnostics.AddError("Error updating hook", err.Error())
		return
	}

	plan.Name = types.StringValue(hook.Name)

	if hook.Description != "" {
		plan.Description = types.StringValue(hook.Description)
	} else {
		plan.Description = types.StringNull()
	}

	plan.Interpreter = types.StringValue(hook.Interpreter)
	plan.ScriptfilePath = types.StringValue(hook.ScriptfilePath)

	if hook.VcsProvider != nil {
		plan.VcsProviderId = types.StringValue(hook.VcsProvider.ID)
	}

	if hook.VcsRepo != nil {
		vcsRepoModel := hookResourceVcsRepoModel{
			Identifier: types.StringValue(hook.VcsRepo.Identifier),
			Branch:     types.StringValue(hook.VcsRepo.Branch),
		}

		vcsRepos := []hookResourceVcsRepoModel{vcsRepoModel}

		vcsRepoList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"identifier": types.StringType,
				"branch":     types.StringType,
			},
		}, vcsRepos)

		if diags.HasError() {
			resp.Diagnostics.AddError("Error converting VCS repo list", diags.Errors()[0].Summary())
			return
		}

		plan.VcsRepo = vcsRepoList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *hookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Client.Hooks.Delete(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, scalr.ErrResourceNotFound) {
		resp.Diagnostics.AddError("Error deleting hook", err.Error())
	}
}

func (r *hookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *hookResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan hookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.VcsRepo.IsNull() || len(plan.VcsRepo.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("vcs_repo"),
			"Missing required block",
			"The vcs_repo block is required for scalr_hook resource",
		)
		return
	}
}

package provider

import (
	"context"
	"errors"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
)

// Compile-time interface checks
var (
	_ resource.Resource                     = &moduleResource{}
	_ resource.ResourceWithConfigure        = &moduleResource{}
	_ resource.ResourceWithConfigValidators = &moduleResource{}
	_ resource.ResourceWithValidateConfig   = &moduleResource{}
	_ resource.ResourceWithImportState      = &moduleResource{}
)

func newModuleResource() resource.Resource {
	return &moduleResource{}
}

// moduleResource defines the resource implementation.
type moduleResource struct {
	framework.ResourceWithScalrClient
}

// moduleResourceModel describes the resource data model.
type moduleResourceModel struct {
	Id                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	ModuleProvider      types.String `tfsdk:"module_provider"`
	Status              types.String `tfsdk:"status"`
	Source              types.String `tfsdk:"source"`
	SourceType          types.String `tfsdk:"source_type"`
	VcsRepo             types.List   `tfsdk:"vcs_repo"`
	VcsProviderID       types.String `tfsdk:"vcs_provider_id"`
	DockerImage         types.String `tfsdk:"docker_image"`
	DockerIntegrationID types.String `tfsdk:"docker_integration_id"`
	AccountID           types.String `tfsdk:"account_id"`
	EnvironmentID       types.String `tfsdk:"environment_id"`
	NamespaceID         types.String `tfsdk:"namespace_id"`
}

// moduleVcsRepoModel maps the vcs_repo nested schema data.
type moduleVcsRepoModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Path       types.String `tfsdk:"path"`
	TagPrefix  types.String `tfsdk:"tag_prefix"`
}

func moduleVcsRepoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"identifier": types.StringType,
		"path":       types.StringType,
		"tag_prefix": types.StringType,
	}
}

// stringValueOrNull treats an empty string as an absent value.
func moduleStringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

func moduleResourceModelFromAPI(ctx context.Context, m *schemas.Module) (*moduleResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &moduleResourceModel{
		Id:                  types.StringValue(m.ID),
		Name:                types.StringValue(m.Attributes.Name),
		ModuleProvider:      types.StringValue(m.Attributes.Provider),
		Status:              types.StringValue(string(m.Attributes.Status)),
		Source:              types.StringValue(m.Attributes.Source),
		SourceType:          types.StringValue(string(m.Attributes.SourceType)),
		VcsRepo:             types.ListNull(types.ObjectType{AttrTypes: moduleVcsRepoAttrTypes()}),
		VcsProviderID:       types.StringNull(),
		DockerImage:         types.StringPointerValue(m.Attributes.DockerImage),
		DockerIntegrationID: types.StringNull(),
		AccountID:           types.StringNull(),
		EnvironmentID:       types.StringNull(),
		NamespaceID:         types.StringNull(),
	}

	if m.Attributes.VcsRepo != nil {
		vcsRepo := []moduleVcsRepoModel{{
			Identifier: types.StringValue(m.Attributes.VcsRepo.Identifier),
			Path:       moduleStringOrNull(m.Attributes.VcsRepo.Path),
			TagPrefix:  moduleStringOrNull(m.Attributes.VcsRepo.TagPrefix),
		}}
		vcsRepoValue, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: moduleVcsRepoAttrTypes()}, vcsRepo)
		diags.Append(d...)
		model.VcsRepo = vcsRepoValue
	}

	if m.Relationships.VcsProvider != nil {
		model.VcsProviderID = types.StringValue(m.Relationships.VcsProvider.ID)
	}
	if m.Relationships.DockerIntegration != nil {
		model.DockerIntegrationID = types.StringValue(m.Relationships.DockerIntegration.ID)
	}
	if m.Relationships.Account != nil {
		model.AccountID = types.StringValue(m.Relationships.Account.ID)
	}
	if m.Relationships.Environment != nil {
		model.EnvironmentID = types.StringValue(m.Relationships.Environment.ID)
	}
	if m.Relationships.Namespace != nil {
		model.NamespaceID = types.StringValue(m.Relationships.Namespace.ID)
	}

	return model, diags
}

func (r *moduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module"
}

func (r *moduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of a module in the Private Modules Registry. Create and destroy operations are available only." +
			" A module is sourced either from a VCS repository (`vcs_repo` and `vcs_provider_id`)" +
			" or from an OCI registry (`docker_image` and `docker_integration_id`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the module.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the module, e.g. `rds`, `compute`, `kubernetes-engine`." +
					" Required for OCI-sourced modules.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9]([A-Za-z0-9_-]{0,62}[a-zA-Z0-9])?$`),
						"must start and end with a letter or digit, contain only letters, digits, underscores, and hyphens, and be at most 64 characters",
					),
					stringvalidator.AlsoRequires(path.MatchRoot("module_provider")),
				},
			},
			"module_provider": schema.StringAttribute{
				MarkdownDescription: "Module provider name, e.g `aws`, `azurerm`, `google`, etc." +
					" Required for OCI-sourced modules.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-z]{1,64}$`),
						"must be 1-64 characters of lowercase letters and digits only",
					),
					stringvalidator.AlsoRequires(path.MatchRoot("name")),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "A system status of the Module.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "The source of a remote module in the private registry, e.g `env-xxxx/aws/vpc`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"source_type": schema.StringAttribute{
				MarkdownDescription: "The upstream source type of the module: `vcs` or `docker`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vcs_provider_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of a VCS provider in the format `vcs-<RANDOM STRING>`." +
					" Required for VCS-sourced modules.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"docker_image": schema.StringAttribute{
				MarkdownDescription: "The OCI repository path of the module inside the registry configured in `docker_integration_id`," +
					" e.g. `scalr/terraform-aws-network`. Conflicts with `vcs_repo`.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"docker_integration_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of a Docker registry integration the module is pulled from." +
					" Required for OCI-sourced modules.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the account in the format `acc-<RANDOM STRING>`." +
					" If it is not specified the module will be registered globally and available across the whole installation." +
					" **Deprecated:** Use `namespace_id` instead.",
				DeprecationMessage: "Use namespace_id instead",
				Optional:           true,
				Computed:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of an environment in the format `env-<RANDOM STRING>`." +
					" If it is not specified the module will be registered at the account level and available across all environments" +
					" within the account specified in `account_id` attribute. **Deprecated:** Use `namespace_id` instead.",
				DeprecationMessage: "Use namespace_id instead",
				Optional:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of a module namespace in the format `modns-<RANDOM STRING>`." +
					" If specified, the module will be registered in this namespace. Conflicts with `environment_id`.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"vcs_repo": schema.ListNestedBlock{
				MarkdownDescription: "Source configuration of a VCS repository. Conflicts with `docker_image`.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							MarkdownDescription: "The identifier of a VCS repository in the format `:org/:repo` (`:org/:project/:name` is used for Azure DevOps). It refers to an organization and a repository name in a VCS provider.",
							Required:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "The path to the root module folder. It is expected to have the format `<path>/terraform-<provider_name>-<module_name>`, where `<path>` stands for any folder within the repository inclusively a repository root.",
							Optional:            true,
						},
						"tag_prefix": schema.StringAttribute{
							MarkdownDescription: "Registry ignores tags which do not match specified prefix, e.g. `aws/`.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *moduleResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("vcs_repo"),
			path.MatchRoot("docker_image"),
		),
		resourcevalidator.RequiredTogether(
			path.MatchRoot("vcs_repo"),
			path.MatchRoot("vcs_provider_id"),
		),
		resourcevalidator.RequiredTogether(
			path.MatchRoot("docker_image"),
			path.MatchRoot("docker_integration_id"),
		),
		resourcevalidator.Conflicting(
			path.MatchRoot("environment_id"),
			path.MatchRoot("namespace_id"),
		),
	}
}

func (r *moduleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg moduleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.DockerImage.IsNull() || cfg.DockerImage.IsUnknown() {
		return
	}
	for attrName, v := range map[string]types.String{"name": cfg.Name, "module_provider": cfg.ModuleProvider} {
		if v.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root(attrName),
				"Missing required argument",
				"The argument \""+attrName+"\" is required for OCI-sourced modules (when \"docker_image\" is set).",
			)
		}
	}
}

func (r *moduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan moduleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.ModuleRequest{
		Attributes: schemas.ModuleAttributesRequest{
			Name:     framework.SetIfKnownString(plan.Name),
			Provider: framework.SetIfKnownString(plan.ModuleProvider),
		},
	}

	if !plan.DockerImage.IsNull() {
		opts.Attributes.SourceType = value.Set(schemas.ModuleSourceTypeDocker)
		opts.Attributes.DockerImage = value.Set(plan.DockerImage.ValueString())
		opts.Relationships.DockerIntegration = value.Set(
			schemas.DockerIntegration{ID: plan.DockerIntegrationID.ValueString()},
		)
	} else {
		var vcsRepo []moduleVcsRepoModel
		resp.Diagnostics.Append(plan.VcsRepo.ElementsAs(ctx, &vcsRepo, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(vcsRepo) == 0 {
			resp.Diagnostics.AddError("Error creating module", "The vcs_repo block is required for VCS-sourced modules.")
			return
		}

		opts.Attributes.SourceType = value.Set(schemas.ModuleSourceTypeVcs)
		opts.Attributes.VcsRepo = value.Set(schemas.ModuleVcsRepoRequest{
			Identifier: value.Set(vcsRepo[0].Identifier.ValueString()),
			Path:       framework.SetIfKnownString(vcsRepo[0].Path),
			TagPrefix:  framework.SetIfKnownString(vcsRepo[0].TagPrefix),
		})
		opts.Relationships.VcsProvider = value.Set(schemas.VcsProvider{ID: plan.VcsProviderID.ValueString()})
	}

	if !plan.EnvironmentID.IsNull() {
		opts.Relationships.Environment = value.Set(schemas.Environment{ID: plan.EnvironmentID.ValueString()})
	}
	if !plan.NamespaceID.IsUnknown() && !plan.NamespaceID.IsNull() {
		opts.Relationships.Namespace = value.Set(schemas.ModuleNamespace{ID: plan.NamespaceID.ValueString()})
	}

	module, err := r.ClientV2.Module.CreateModule(ctx, &opts)
	if err != nil {
		resp.Diagnostics.AddError("Error creating module", err.Error())
		return
	}

	result, d := moduleResourceModelFromAPI(ctx, module)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *moduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state moduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	module, err := r.ClientV2.Module.GetModule(ctx, state.Id.ValueString(), nil)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving module", err.Error())
		return
	}

	result, d := moduleResourceModelFromAPI(ctx, module)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Update is a no-op: every configurable attribute forces replacement, except the deprecated computed `account_id`.
func (r *moduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan moduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *moduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state moduleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.Module.DeleteModule(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting module", err.Error())
		return
	}
}

func (r *moduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

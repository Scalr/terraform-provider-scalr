package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scalr/go-scalr"

	"github.com/scalr/terraform-provider-scalr/internal/framework/validation"
)

func workspaceResourceSchema(ctx context.Context) *schema.Schema {
	emptyStringList, _ := types.ListValueFrom(ctx, types.StringType, []string{})
	emptyStringSet, _ := types.SetValueFrom(ctx, types.StringType, []string{})
	asteriskStringSet, _ := types.SetValueFrom(ctx, types.StringType, []string{"*"})

	return &schema.Schema{
		MarkdownDescription: "Manages the state of workspaces in Scalr.",
		Version:             4,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the workspace.",
				Required:            true,
			},
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "ID of the environment, in the format `env-<RANDOM STRING>`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vcs_provider_id": schema.StringAttribute{
				MarkdownDescription: "ID of VCS provider - required if vcs-repo present and vice versa, in the format `vcs-<RANDOM STRING>`.",
				Optional:            true,
			},
			"module_version_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of a module version in the format `modver-<RANDOM STRING>`. This attribute conflicts with `vcs_provider_id` and `vcs_repo` attributes.",
				Optional:            true,
			},
			"agent_pool_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of an agent pool in the format `apool-<RANDOM STRING>`.",
				Optional:            true,
			},
			"auto_apply": schema.BoolAttribute{
				MarkdownDescription: "Set (true/false) to configure if `terraform apply` should automatically run when `terraform plan` ends without error. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"force_latest_run": schema.BoolAttribute{
				MarkdownDescription: "Set (true/false) to configure if latest new run will be automatically raised in priority. Default `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"deletion_protection_enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the workspace has the protection from an accidental state lost. If enabled and the workspace has resource, the deletion will not be allowed. Default `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"var_files": schema.ListAttribute{
				MarkdownDescription: "A list of paths to the `.tfvars` file(s) to be used as part of the workspace configuration.",
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(emptyStringList),
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
			},
			"operations": schema.BoolAttribute{
				MarkdownDescription: "Set (true/false) to configure workspace remote execution. When `false` workspace is only used to store state. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				DeprecationMessage:  "The attribute `operations` is deprecated. Use `execution_mode` instead",
			},
			"execution_mode": schema.StringAttribute{
				MarkdownDescription: "Which execution mode to use. Valid values are `remote` and `local`. When set to `local`, the workspace will be used for state storage only. Defaults to `remote`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(scalr.WorkspaceExecutionModeRemote)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.WorkspaceExecutionModeRemote),
						string(scalr.WorkspaceExecutionModeLocal),
					),
				},
			},
			"terraform_version": schema.StringAttribute{
				MarkdownDescription: "The version of Terraform to use for this workspace. Defaults to the latest available version.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"iac_platform": schema.StringAttribute{
				MarkdownDescription: "The IaC platform to use for this workspace. Valid values are `terraform` and `opentofu`. Defaults to `terraform`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(scalr.WorkspaceIaCPlatformTerraform)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.WorkspaceIaCPlatformTerraform),
						string(scalr.WorkspaceIaCPlatformOpenTofu),
					),
				},
			},
			"working_directory": schema.StringAttribute{
				MarkdownDescription: "A relative path that Terraform will be run in. Defaults to the root of the repository `\"\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"has_resources": schema.BoolAttribute{
				MarkdownDescription: "The presence of active terraform resources in the current state version.",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_queue_runs": schema.StringAttribute{
				MarkdownDescription: "Indicates if runs have to be queued automatically when a new configuration version is uploaded. Supported values are `skip_first`, `always`, `never`:" +
					"\n  * `skip_first` - after the very first configuration version is uploaded into the workspace the run will not be triggered. But the following configurations will do. This is the default behavior." +
					"\n  * `always` - runs will be triggered automatically on every upload of the configuration version." +
					"\n  * `never` - configuration versions are uploaded into the workspace, but runs will not be triggered.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(scalr.AutoQueueRunsModeSkipFirst)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.AutoQueueRunsModeSkipFirst),
						string(scalr.AutoQueueRunsModeAlways),
						string(scalr.AutoQueueRunsModeNever),
					),
				},
			},
			"created_by": schema.ListAttribute{
				MarkdownDescription: "Details of the user that created the workspace.",
				ElementType:         userElementType,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"run_operation_timeout": schema.Int32Attribute{
				MarkdownDescription: "The number of minutes run operation can be executed before termination.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the Scalr Workspace environment, available options: `production`, `staging`, `testing`, `development`, `unmapped`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(scalr.WorkspaceEnvironmentTypeUnmapped)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(scalr.WorkspaceEnvironmentTypeProduction),
						string(scalr.WorkspaceEnvironmentTypeStaging),
						string(scalr.WorkspaceEnvironmentTypeTesting),
						string(scalr.WorkspaceEnvironmentTypeDevelopment),
						string(scalr.WorkspaceEnvironmentTypeUnmapped),
					),
				},
			},
			"ssh_key_id": schema.StringAttribute{
				MarkdownDescription: "The identifier of the SSH key to use for the workspace.",
				Optional:            true,
			},
			"tag_ids": schema.SetAttribute{
				MarkdownDescription: "List of tag IDs associated with the workspace.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(emptyStringSet),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
			},
			"remote_state_consumers": schema.SetAttribute{
				MarkdownDescription: "The list of workspace identifiers that are allowed to access the state of this workspace. Use `[\"*\"]` to share the state with all the workspaces within the environment (default).",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(asteriskStringSet),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validation.StringIsNotWhiteSpace()),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"hooks": schema.ListNestedBlock{
				MarkdownDescription: "Settings for the workspaces custom hooks.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"pre_init": schema.StringAttribute{
							MarkdownDescription: "Action that will be called before the init phase.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"pre_plan": schema.StringAttribute{
							MarkdownDescription: "Action that will be called before the plan phase.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"post_plan": schema.StringAttribute{
							MarkdownDescription: "Action that will be called after plan phase.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"pre_apply": schema.StringAttribute{
							MarkdownDescription: "Action that will be called before apply phase.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
						"post_apply": schema.StringAttribute{
							MarkdownDescription: "Action that will be called after apply phase.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"vcs_repo": schema.ListNestedBlock{
				MarkdownDescription: "Settings for the workspace's VCS repository.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							MarkdownDescription: "A reference to your VCS repository in the format `:org/:repo`, it refers to the organization and repository in your VCS provider.",
							Required:            true,
						},
						"branch": schema.StringAttribute{
							MarkdownDescription: "The repository branch where Terraform will be run from. If omitted, the repository default branch will be used.",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "The repository subdirectory that Terraform will execute from. If omitted or submitted as an empty string, this defaults to the repository's root.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
							DeprecationMessage:  "The attribute `vcs-repo.path` is deprecated. Use working-directory and trigger-prefixes instead.",
						},
						"trigger_prefixes": schema.ListAttribute{
							MarkdownDescription: "List of paths (relative to `path`), whose changes will trigger a run for the workspace using this binding when the CV is created. Conflicts with `trigger_patterns`. If `trigger_prefixes` and `trigger_patterns` are omitted, any change in `path` will trigger a new run.",
							ElementType:         types.StringType,
							Optional:            true,
							Computed:            true,
							Default:             listdefault.StaticValue(emptyStringList),
						},
						"trigger_patterns": schema.StringAttribute{
							MarkdownDescription: "The gitignore-style patterns for files, whose changes will trigger a run for the workspace using this binding when the CV is created. Conflicts with `trigger_prefixes`. If `trigger_prefixes` and `trigger_patterns` are omitted, any change in `path` will trigger a new run.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("trigger_prefixes")),
							},
						},
						"dry_runs_enabled": schema.BoolAttribute{
							MarkdownDescription: "Set (true/false) to configure the VCS driven dry runs should run when pull request to configuration versions branch created. Default `true`.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(true),
						},
						"ingress_submodules": schema.BoolAttribute{
							MarkdownDescription: "Designates whether to clone git submodules of the VCS repository.",
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"provider_configuration": schema.SetNestedBlock{
				MarkdownDescription: "Provider configurations used in workspace runs.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of provider configuration.",
							Required:            true,
						},
						"alias": schema.StringAttribute{
							MarkdownDescription: "The alias of provider configuration.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
					},
				},
			},
			"terragrunt": schema.ListNestedBlock{
				MarkdownDescription: "Settings for the workspace's Terragrunt configuration.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"version": schema.StringAttribute{
							MarkdownDescription: "The version of Terragrunt the workspace performs runs on.",
							Required:            true,
						},
						"use_run_all": schema.BoolAttribute{
							MarkdownDescription: "Indicates whether the workspace uses `terragrunt run-all`.",
							Default:             booldefault.StaticBool(false),
							Optional:            true,
							Computed:            true,
						},
						"include_external_dependencies": schema.BoolAttribute{
							MarkdownDescription: "Indicates whether the workspace includes external dependencies.",
							Default:             booldefault.StaticBool(false),
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func workspaceResourceSchemaV3(ctx context.Context) *schema.Schema {
	emptyStringList, _ := types.ListValueFrom(ctx, types.StringType, []string{})

	return &schema.Schema{
		Version: 3,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vcs_provider_id": schema.StringAttribute{
				Optional: true,
			},
			"module_version_id": schema.StringAttribute{
				Optional: true,
			},
			"agent_pool_id": schema.StringAttribute{
				Optional: true,
			},
			"auto_apply": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"var_files": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(emptyStringList),
				ElementType: types.StringType,
			},
			"operations": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"terraform_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"working_directory": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"has_resources": schema.BoolAttribute{
				Computed: true,
			},
			"created_by": schema.ListAttribute{
				ElementType: userElementType,
				Computed:    true,
			},
			"run_operation_timeout": schema.Int32Attribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"hooks": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"pre_init": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
						"pre_plan": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
						"post_plan": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
						"pre_apply": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
						"post_apply": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
					},
				},
			},
			"vcs_repo": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Required: true,
						},
						"branch": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"path": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
						"trigger_prefixes": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
						"dry_runs_enabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
						"ingress_submodules": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
					},
				},
			},
			"provider_configuration": schema.SetNestedBlock{
				MarkdownDescription: "Provider configurations used in workspace runs.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The identifier of provider configuration.",
							Required:            true,
						},
						"alias": schema.StringAttribute{
							MarkdownDescription: "The alias of provider configuration.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
						},
					},
				},
			},
		},
	}
}

func workspaceResourceSchemaV2(_ context.Context) *schema.Schema {
	return &schema.Schema{
		Version: 2,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vcs_provider_id": schema.StringAttribute{
				Optional: true,
			},
			"auto_apply": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"operations": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"terraform_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"working_directory": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"created_by": schema.ListAttribute{
				ElementType: userElementType,
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"vcs_repo": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Required: true,
						},
						"branch": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"path": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
					},
				},
			},
		},
	}
}

func workspaceResourceSchemaV1(_ context.Context) *schema.Schema {
	return &schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auto_apply": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"operations": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"queue_all_runs": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"terraform_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"working_directory": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"created_by": schema.ListAttribute{
				ElementType: userElementType,
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"vcs_repo": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Required: true,
						},
						"branch": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},
						"oauth_token_id": schema.StringAttribute{
							Required: true,
						},
						"path": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
					},
				},
			},
		},
	}
}

func workspaceResourceSchemaV0(_ context.Context) *schema.Schema {
	return &schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"organization": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"auto_apply": schema.BoolAttribute{
				Optional: true,
				Default:  booldefault.StaticBool(false),
			},
			"operations": schema.BoolAttribute{
				Optional: true,
				Default:  booldefault.StaticBool(true),
			},
			"ssh_key_id": schema.StringAttribute{
				Optional: true,
				Default:  stringdefault.StaticString(""),
			},
			"terraform_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"working_directory": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"created_by": schema.ListAttribute{
				ElementType: userElementType,
				Computed:    true,
			},
			"external_id": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"vcs_repo": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Required: true,
						},
						"branch": schema.StringAttribute{
							Optional: true,
						},
						"ingress_submodules": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"oauth_token_id": schema.StringAttribute{
							Optional: true,
						},
						"path": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

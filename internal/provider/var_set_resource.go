package provider

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ resource.Resource                = &varSetResource{}
	_ resource.ResourceWithConfigure   = &varSetResource{}
	_ resource.ResourceWithImportState = &varSetResource{}
)

func newVarSetResource() resource.Resource {
	return &varSetResource{}
}

// varSetResource defines the resource implementation.
type varSetResource struct {
	framework.ResourceWithScalrClient
}

// varSetResourceModel describes the resource data model.
type varSetResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Environments   types.Set    `tfsdk:"environments"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
	UpdatedByEmail types.String `tfsdk:"updated_by_email"`
	AccountID      types.String `tfsdk:"account_id"`
	Owners         types.Set    `tfsdk:"owners"`
}

func varSetResourceModelFromAPI(
	ctx context.Context,
	vs *schemas.VariableSet,
) (*varSetResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &varSetResourceModel{
		Id:             types.StringValue(vs.ID),
		Name:           types.StringValue(vs.Attributes.Name),
		Description:    types.StringPointerValue(vs.Attributes.Description),
		UpdatedAt:      types.StringValue(vs.Attributes.UpdatedAt.Format(time.RFC3339)),
		UpdatedByEmail: types.StringPointerValue(vs.Attributes.UpdatedByEmail),
		AccountID:      types.StringNull(),
		Environments:   types.SetNull(types.StringType),
		Owners:         types.SetNull(types.StringType),
	}

	if vs.Relationships.Account != nil {
		model.AccountID = types.StringValue(vs.Relationships.Account.ID)
	}

	owners, d := framework.FlattenRelationshipIDsSet(
		ctx,
		vs.Relationships.Owners,
		func(t *schemas.Team) string { return t.ID },
		nil,
	)
	diags.Append(d...)
	model.Owners = owners

	var envs types.Set
	if vs.Attributes.IsShared {
		envs, d = types.SetValueFrom(ctx, types.StringType, []string{"*"})
	} else {
		envs, d = framework.FlattenRelationshipIDsSet(
			ctx,
			vs.Relationships.Environments,
			func(e *schemas.Environment) string { return e.ID },
			nil,
		)
	}
	diags.Append(d...)
	model.Environments = envs

	return model, diags
}

func (r *varSetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_var_set"
}

func (r *varSetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of variable sets in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the variable set.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the variable set.",
				Optional:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "UTC timestamp of the last update to this variable set.",
				Computed:            true,
			},
			"updated_by_email": schema.StringAttribute{
				MarkdownDescription: "Email of the user who last updated this variable set.",
				Computed:            true,
			},
			"account_id": schema.StringAttribute{
				MarkdownDescription: "ID of the account this variable set belongs to.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "List of environment IDs that this variable set is shared to. Use `[\"*\"]` to share with all environments.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidation.StringIsNotWhiteSpace()),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"owners": schema.SetAttribute{
				MarkdownDescription: "List of team IDs this variable set belongs to.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidation.StringIsNotWhiteSpace()),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *varSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan varSetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.VariableSetRequest{
		Attributes: schemas.VariableSetAttributesRequest{
			Name:        value.Set(plan.Name.ValueString()),
			Description: value.SetPtrMaybe(plan.Description.ValueStringPointer()),
		},
	}

	if !plan.Owners.IsUnknown() && !plan.Owners.IsNull() {
		owners, diags := framework.ExpandRelationshipIDsSet(
			ctx, plan.Owners, func(id string) schemas.Team {
				return schemas.Team{ID: id}
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		opts.Relationships.Owners = value.Set(owners)
	}

	if !plan.Environments.IsUnknown() && !plan.Environments.IsNull() {
		envs, diags := framework.ExpandRelationshipIDsSet(
			ctx, plan.Environments, func(id string) schemas.Environment {
				return schemas.Environment{ID: id}
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(envs) == 1 && envs[0].ID == "*" {
			opts.Attributes.IsShared = value.Set(true)
		} else {
			opts.Relationships.Environments = value.Set(envs)
		}
	}

	vs, err := r.ClientV2.VariableSet.CreateVarSet(ctx, &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating var_set", err.Error())
		return
	}

	result, diags := varSetResourceModelFromAPI(ctx, vs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *varSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state varSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vs, err := r.ClientV2.VariableSet.GetVarSet(ctx, state.Id.ValueString(), nil)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving var_set", err.Error())
		return
	}

	result, diags := varSetResourceModelFromAPI(ctx, vs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *varSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state varSetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.VariableSetRequest{}

	if !plan.Name.Equal(state.Name) {
		opts.Attributes.Name = value.Set(plan.Name.ValueString())
	}

	if !plan.Description.Equal(state.Description) {
		opts.Attributes.Description = value.SetPtr(plan.Description.ValueStringPointer())
	}

	if !plan.Owners.Equal(state.Owners) && !plan.Owners.IsUnknown() && !plan.Owners.IsNull() {
		owners, diags := framework.ExpandRelationshipIDsSet(
			ctx, plan.Owners, func(id string) schemas.Team {
				return schemas.Team{ID: id}
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		opts.Relationships.Owners = value.Set(owners)
	}

	if !plan.Environments.Equal(state.Environments) && !plan.Environments.IsUnknown() && !plan.Environments.IsNull() {
		envs, diags := framework.ExpandRelationshipIDsSet(
			ctx, plan.Environments, func(id string) schemas.Environment {
				return schemas.Environment{ID: id}
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if len(envs) == 1 && envs[0].ID == "*" {
			opts.Attributes.IsShared = value.Set(true)
			opts.Relationships.Environments = value.Set([]schemas.Environment{})
		} else {
			opts.Attributes.IsShared = value.Set(false)
			opts.Relationships.Environments = value.Set(envs)
		}
	}

	vs, err := r.ClientV2.VariableSet.UpdateVarSet(ctx, plan.Id.ValueString(), &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating var_set", err.Error())
		return
	}

	result, diags := varSetResourceModelFromAPI(ctx, vs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *varSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state varSetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.VariableSet.DeleteVarSet(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting var_set", err.Error())
		return
	}
}

func (r *varSetResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

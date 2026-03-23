package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scalr/go-scalr/v2/scalr/client"
	"github.com/scalr/go-scalr/v2/scalr/ops/account"
	"github.com/scalr/go-scalr/v2/scalr/ops/team"
	"github.com/scalr/go-scalr/v2/scalr/schemas"
	"github.com/scalr/go-scalr/v2/scalr/value"

	"github.com/scalr/terraform-provider-scalr/internal/framework"
	"github.com/scalr/terraform-provider-scalr/internal/framework/defaults"
	"github.com/scalr/terraform-provider-scalr/internal/framework/validation/stringvalidation"
)

// Compile-time interface checks
var (
	_ resource.Resource                = &iamTeamResource{}
	_ resource.ResourceWithConfigure   = &iamTeamResource{}
	_ resource.ResourceWithModifyPlan  = &iamTeamResource{}
	_ resource.ResourceWithImportState = &iamTeamResource{}
)

func newIamTeamResource() resource.Resource {
	return &iamTeamResource{}
}

// iamTeamResource defines the resource implementation.
type iamTeamResource struct {
	framework.ResourceWithScalrClient
}

// iamTeamResourceModel describes the resource data model.
type iamTeamResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	AccountID          types.String `tfsdk:"account_id"`
	IdentityProviderID types.String `tfsdk:"identity_provider_id"`
	Users              types.Set    `tfsdk:"users"`
}

func iamTeamResourceModelFromAPI(ctx context.Context, team *schemas.Team, priorUsers *types.Set) (
	*iamTeamResourceModel,
	diag.Diagnostics,
) {
	var diags diag.Diagnostics

	model := &iamTeamResourceModel{
		Id:                 types.StringValue(team.ID),
		Name:               types.StringValue(team.Attributes.Name),
		Description:        types.StringPointerValue(team.Attributes.Description),
		AccountID:          types.StringNull(),
		IdentityProviderID: types.StringNull(),
		Users:              types.SetNull(types.StringType),
	}

	if team.Relationships.Account != nil {
		model.AccountID = types.StringValue(team.Relationships.Account.ID)
	}

	if team.Relationships.IdentityProvider != nil {
		model.IdentityProviderID = types.StringValue(team.Relationships.IdentityProvider.ID)
	}

	if priorUsers != nil {
		// Until the next major release we should keep the old behavior,
		// where the state keeps users from config, not from the API response.
		// This should only be used for the teams with an external IDP.
		model.Users = *priorUsers
	} else {
		users, d := framework.FlattenRelationshipIDsSet(
			ctx,
			team.Relationships.Users,
			func(user *schemas.User) string { return user.ID },
			nil,
		)
		diags.Append(d...)
		model.Users = users
	}

	return model, diags
}

func (r *iamTeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_team"
}

func (r *iamTeamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the Scalr IAM teams: performs create, update and destroy actions.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A name of the team.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A verbose description of the team.",
				Optional:            true,
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
			"identity_provider_id": schema.StringAttribute{
				MarkdownDescription: "An identifier of the login identity provider, in the format `idp-<RANDOM STRING>`.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage: "Setting this attribute is deprecated." +
					" It is no longer in use and will become read-only in the next major version of the provider.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"users": schema.SetAttribute{
				MarkdownDescription: "A list of the user identifiers to add to the team." +
					" This attribute should not be used when the account's identity provider is not of type `scalr`," +
					" as team membership is managed externally in these cases.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidation.StringIsNotWhiteSpace()),
				},
			},
		},
	}
}

func (r *iamTeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan iamTeamResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.TeamRequest{
		Attributes: schemas.TeamAttributesRequest{
			Name:        value.Set(plan.Name.ValueString()),
			Description: framework.SetIfKnownString(plan.Description),
		},
	}

	if !plan.Users.IsUnknown() && !plan.Users.IsNull() {
		users, diags := framework.ExpandRelationshipIDsSet(
			ctx, plan.Users, func(id string) schemas.User {
				return schemas.User{ID: id}
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		opts.Relationships.Users = value.Set(users)
	}

	iamTeam, err := r.ClientV2.Team.CreateTeam(
		ctx, &opts, &team.CreateTeamOptions{
			Include: []string{"identity-provider"},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating team", err.Error())
		return
	}

	var priorUsers *types.Set
	if iamTeam.Relationships.IdentityProvider != nil &&
		iamTeam.Relationships.IdentityProvider.Attributes.IdpType != schemas.IdentityProviderIdpTypeScalr {
		priorUsers = &plan.Users
	}

	result, diags := iamTeamResourceModelFromAPI(ctx, iamTeam, priorUsers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *iamTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iamTeamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	iamTeam, err := r.ClientV2.Team.GetTeam(
		ctx, state.Id.ValueString(), &team.GetTeamOptions{
			Include: []string{"identity-provider"},
		},
	)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving team", err.Error())
		return
	}

	var priorUsers *types.Set
	if iamTeam.Relationships.IdentityProvider != nil &&
		iamTeam.Relationships.IdentityProvider.Attributes.IdpType != schemas.IdentityProviderIdpTypeScalr {
		priorUsers = &state.Users
	}

	result, diags := iamTeamResourceModelFromAPI(ctx, iamTeam, priorUsers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *iamTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state iamTeamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.TeamRequest{}

	if !plan.Name.Equal(state.Name) {
		opts.Attributes.Name = value.Set(plan.Name.ValueString())
	}

	if !plan.Description.Equal(state.Description) {
		opts.Attributes.Description = value.SetPtr(plan.Description.ValueStringPointer())
	}

	if !plan.Users.Equal(state.Users) && !plan.Users.IsUnknown() && !plan.Users.IsNull() {
		users, diags := framework.ExpandRelationshipIDsSet(
			ctx, plan.Users, func(id string) schemas.User {
				return schemas.User{ID: id}
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		opts.Relationships.Users = value.Set(users)
	}

	iamTeam, err := r.ClientV2.Team.UpdateTeam(
		ctx, plan.Id.ValueString(), &opts, &team.UpdateTeamOptions{
			Include: []string{"identity-provider"},
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating team", err.Error())
		return
	}

	var priorUsers *types.Set
	if iamTeam.Relationships.IdentityProvider != nil &&
		iamTeam.Relationships.IdentityProvider.Attributes.IdpType != schemas.IdentityProviderIdpTypeScalr {
		priorUsers = &plan.Users
	}

	result, diags := iamTeamResourceModelFromAPI(ctx, iamTeam, priorUsers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *iamTeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iamTeamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.Team.DeleteTeam(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting team", err.Error())
		return
	}
}

func (r *iamTeamResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	// Fetch the account with its identity provider;
	// issue a warning if it's an external IDP and the `users` attribute is set.
	var users types.Set
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("users"), &users)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if users.IsUnknown() || users.IsNull() {
		return
	}

	var accountID types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("account_id"), &accountID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if accountID.IsUnknown() || accountID.IsNull() {
		// Fool-proof, account_id is required in the schema so far.
		resp.Diagnostics.AddError("Error retrieving account", "Account ID unknown")
		return
	}

	acc, err := r.ClientV2.Account.GetAccount(
		ctx,
		accountID.ValueString(),
		&account.GetAccountOptions{Include: []string{"identity-provider"}},
	)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving account", err.Error())
		return
	}
	if acc == nil {
		resp.Diagnostics.AddError("Error retrieving account", "Account not found")
		return
	}
	if acc.Relationships.IdentityProvider == nil {
		resp.Diagnostics.AddError("Error retrieving account", "Account identity provider not found")
		return
	}

	if acc.Relationships.IdentityProvider.Attributes.IdpType != schemas.IdentityProviderIdpTypeScalr {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("users"),
			"Team users membership is managed externally.",
			"The account uses an external identity provider, that is responsible for managing team memberships."+
				"\nRemove the 'users' attribute from your configuration."+
				"\nThis will become a hard error in the next major version of the provider.",
		)
	}
}

func (r *iamTeamResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

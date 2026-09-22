package provider

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
	_ resource.Resource                = &wizIntegrationResource{}
	_ resource.ResourceWithConfigure   = &wizIntegrationResource{}
	_ resource.ResourceWithImportState = &wizIntegrationResource{}
)

func newWizIntegrationResource() resource.Resource {
	return &wizIntegrationResource{}
}

// wizIntegrationResource defines the resource implementation.
type wizIntegrationResource struct {
	framework.ResourceWithScalrClient
}

// wizIntegrationResourceModel describes the resource data model.
type wizIntegrationResourceModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ClientID        types.String `tfsdk:"client_id"`
	ClientSecret    types.String `tfsdk:"client_secret"`
	DefaultPolicies types.Set    `tfsdk:"default_policies"`
	DefaultScanName types.String `tfsdk:"default_scan_name"`
	Autofail        types.Bool   `tfsdk:"autofail"`
	EndpointMode    types.String `tfsdk:"endpoint_mode"`
	Environments    types.Set    `tfsdk:"environments"`
}

func wizIntegrationResourceModelFromAPI(
	ctx context.Context,
	wi *schemas.WizIntegration,
	clientSecret types.String,
	prior *wizIntegrationResourceModel,
) (*wizIntegrationResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &wizIntegrationResourceModel{
		Id:              types.StringValue(wi.ID),
		Name:            types.StringValue(wi.Attributes.Name),
		ClientID:        types.StringValue(wi.Attributes.ClientId),
		ClientSecret:    clientSecret,
		DefaultScanName: types.StringPointerValue(wi.Attributes.DefaultScanName),
		Autofail:        types.BoolValue(wi.Attributes.Autofail),
		EndpointMode:    types.StringValue(string(wi.Attributes.EndpointMode)),
		DefaultPolicies: types.SetNull(types.StringType),
		Environments:    types.SetNull(types.StringType),
	}

	var priorPolicies *types.Set
	if prior != nil {
		priorPolicies = &prior.DefaultPolicies
	}
	policies, d := flattenStringSet(ctx, wi.Attributes.DefaultPolicies, priorPolicies)
	diags.Append(d...)
	model.DefaultPolicies = policies

	var envs types.Set
	if wi.Attributes.IsShared {
		envs, d = types.SetValueFrom(ctx, types.StringType, []string{"*"})
	} else {
		var priorEnvs *types.Set
		if prior != nil {
			priorEnvs = &prior.Environments
		}
		envs, d = framework.FlattenRelationshipIDsSet(
			ctx,
			wi.Relationships.Environments,
			func(e *schemas.Environment) string { return e.ID },
			priorEnvs,
		)
	}
	diags.Append(d...)
	model.Environments = envs

	return model, diags
}

// flattenStringSet converts an optional API string list into a set, preserving a
// null prior value when the API reports nothing.
func flattenStringSet(ctx context.Context, apiValue *[]string, prior *types.Set) (types.Set, diag.Diagnostics) {
	if apiValue != nil && len(*apiValue) > 0 {
		return types.SetValueFrom(ctx, types.StringType, *apiValue)
	}

	if prior != nil && prior.IsNull() {
		return types.SetNull(types.StringType), nil
	}

	// preserve explicit empty set
	return types.SetValueFrom(ctx, types.StringType, []string{})
}

// warnOnIntegrationFailure surfaces a backend-reported failure as a warning
func warnOnIntegrationFailure(wi *schemas.WizIntegration, diags *diag.Diagnostics) {
	var errMessage string
	if wi.Attributes.ErrMessage != nil {
		errMessage = strings.TrimSpace(*wi.Attributes.ErrMessage)
	}

	if wi.Attributes.Status != schemas.WizIntegrationStatusFailed && errMessage == "" {
		return
	}

	if errMessage == "" {
		errMessage = "Scalr reported the Wiz integration status as failed."
	}

	diags.AddWarning("Issues detected", errMessage)
}

func (r *wizIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wiz_integration"
}

func (r *wizIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the state of Wiz integrations in Scalr.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the Wiz integration.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthAtMost(128),
				},
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Wiz service-account client ID.",
				Required:            true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthAtMost(255),
				},
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Wiz service-account client secret.",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthBetween(1, 512),
				},
			},
			"default_policies": schema.SetAttribute{
				MarkdownDescription: "Wiz CI/CD policy names passed to `wizcli` (`--policies`), one list item per policy.",
				ElementType:         types.StringType,
				Optional:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidation.StringIsNotWhiteSpace()),
				},
			},
			"default_scan_name": schema.StringAttribute{
				MarkdownDescription: "Default scan name passed to `wizcli` (`--name`)." +
					" When unset, scans are named `{workspace}-{run}`.",
				Optional: true,
				Validators: []validator.String{
					stringvalidation.StringIsNotWhiteSpace(),
					stringvalidator.LengthAtMost(255),
				},
			},
			"autofail": schema.BoolAttribute{
				MarkdownDescription: "Block the run when the Wiz scan fails or cannot run." +
					" When disabled, Wiz failures are represented in the following Scalr policy check" +
					" and do not block the Wiz stage. Default `false`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"endpoint_mode": schema.StringAttribute{
				MarkdownDescription: "Wiz tenant region: the commercial, GovCloud, or FedRAMP auth/API host." +
					" Valid values are `commercial`, `govcloud`, `fedramp`. Default `commercial`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(string(schemas.WizIntegrationEndpointModeCommercial)),
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(schemas.WizIntegrationEndpointModeCommercial),
						string(schemas.WizIntegrationEndpointModeGovcloud),
						string(schemas.WizIntegrationEndpointModeFedramp),
					),
				},
			},
			"environments": schema.SetAttribute{
				MarkdownDescription: "List of environments this integration is linked to." +
					" Use `[\"*\"]` to allow in all current and future environments.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
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

func (r *wizIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan wizIntegrationResourceModel

	// Read plan data
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.WizIntegrationRequest{
		Attributes: schemas.WizIntegrationAttributesRequest{
			Name:         value.Set(plan.Name.ValueString()),
			ClientId:     value.Set(plan.ClientID.ValueString()),
			ClientSecret: value.Set(plan.ClientSecret.ValueString()),
			Autofail:     framework.SetIfKnownBool(plan.Autofail),
			EndpointMode: value.Set(schemas.WizIntegrationEndpointMode(plan.EndpointMode.ValueString())),
			IsShared:     value.Set(false),
		},
	}

	if !plan.DefaultScanName.IsNull() {
		opts.Attributes.DefaultScanName = value.Set(plan.DefaultScanName.ValueString())
	}

	if !plan.DefaultPolicies.IsUnknown() && !plan.DefaultPolicies.IsNull() {
		var policies []string
		resp.Diagnostics.Append(plan.DefaultPolicies.ElementsAs(ctx, &policies, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		opts.Attributes.DefaultPolicies = value.Set(policies)
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

	wi, err := r.ClientV2.WizIntegration.CreateWizIntegration(ctx, &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Wiz integration", err.Error())
		return
	}

	result, diags := wizIntegrationResourceModelFromAPI(ctx, wi, plan.ClientSecret, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	warnOnIntegrationFailure(wi, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *wizIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state wizIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed resource state from API
	wi, err := r.ClientV2.WizIntegration.GetWizIntegration(ctx, state.Id.ValueString(), nil)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error retrieving Wiz integration", err.Error())
		return
	}

	result, diags := wizIntegrationResourceModelFromAPI(ctx, wi, state.ClientSecret, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *wizIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read plan and state data
	var plan, state wizIntegrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := schemas.WizIntegrationRequest{}

	if !plan.Name.Equal(state.Name) {
		opts.Attributes.Name = value.Set(plan.Name.ValueString())
	}

	if !plan.ClientID.Equal(state.ClientID) {
		opts.Attributes.ClientId = value.Set(plan.ClientID.ValueString())
	}

	// The API never returns the secret, so state can't be trusted to detect a
	// change: always send it to keep Scalr in sync with the configuration.
	opts.Attributes.ClientSecret = value.Set(plan.ClientSecret.ValueString())

	if !plan.DefaultScanName.Equal(state.DefaultScanName) {
		opts.Attributes.DefaultScanName = value.SetPtr(plan.DefaultScanName.ValueStringPointer())
	}

	if !plan.Autofail.Equal(state.Autofail) {
		opts.Attributes.Autofail = value.Set(plan.Autofail.ValueBool())
	}

	if !plan.EndpointMode.Equal(state.EndpointMode) {
		opts.Attributes.EndpointMode = value.Set(
			schemas.WizIntegrationEndpointMode(plan.EndpointMode.ValueString()),
		)
	}

	if !plan.DefaultPolicies.Equal(state.DefaultPolicies) {
		if plan.DefaultPolicies.IsNull() {
			opts.Attributes.DefaultPolicies = value.Null[[]string]()
		} else if !plan.DefaultPolicies.IsUnknown() {
			var policies []string
			resp.Diagnostics.Append(plan.DefaultPolicies.ElementsAs(ctx, &policies, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			opts.Attributes.DefaultPolicies = value.Set(policies)
		}
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

	wi, err := r.ClientV2.WizIntegration.UpdateWizIntegration(ctx, plan.Id.ValueString(), &opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Wiz integration", err.Error())
		return
	}

	result, diags := wizIntegrationResourceModelFromAPI(ctx, wi, plan.ClientSecret, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	warnOnIntegrationFailure(wi, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *wizIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state wizIntegrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.ClientV2.WizIntegration.DeleteWizIntegration(ctx, state.Id.ValueString())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Error deleting Wiz integration", err.Error())
		return
	}
}

func (r *wizIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

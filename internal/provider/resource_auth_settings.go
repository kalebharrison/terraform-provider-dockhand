package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*authSettingsResource)(nil)
	_ resource.ResourceWithConfigure   = (*authSettingsResource)(nil)
	_ resource.ResourceWithImportState = (*authSettingsResource)(nil)
)

func NewAuthSettingsResource() resource.Resource {
	return &authSettingsResource{}
}

type authSettingsResource struct {
	client *Client
}

type authSettingsModel struct {
	ID types.String `tfsdk:"id"`

	AuthEnabled     types.Bool   `tfsdk:"auth_enabled"`
	DefaultProvider types.String `tfsdk:"default_provider"`
	SessionTimeout  types.Int64  `tfsdk:"session_timeout"`

	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *authSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_settings"
}

func (r *authSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Dockhand authentication settings via `GET/PUT /api/auth/settings`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Singleton ID. Always `auth`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auth_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"default_provider": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"session_timeout": schema.Int64Attribute{
				MarkdownDescription: "Session timeout in seconds.",
				Optional:            true,
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *authSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *authSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan authSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, diags := r.applyPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *authSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	current, status, err := r.client.GetAuthSettings(ctx)
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand auth settings", err.Error())
		return
	}

	state := modelFromAuthSettings(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *authSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan authSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, diags := r.applyPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *authSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Dockhand does not expose a delete/reset endpoint for auth settings. Delete is a no-op.
}

func (r *authSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "auth")...)
}

func (r *authSettingsResource) applyPlan(ctx context.Context, plan authSettingsModel) (authSettingsModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if r.client == nil {
		diags.AddError("Unconfigured client", "The provider client was not configured.")
		return authSettingsModel{}, diags
	}

	current, _, err := r.client.GetAuthSettings(ctx)
	if err != nil {
		diags.AddError("Error reading Dockhand auth settings", err.Error())
		return authSettingsModel{}, diags
	}

	payload := authSettingsPayload{
		AuthEnabled:     current.AuthEnabled,
		DefaultProvider: current.DefaultProvider,
		SessionTimeout:  current.SessionTimeout,
	}

	if !plan.AuthEnabled.IsNull() && !plan.AuthEnabled.IsUnknown() {
		payload.AuthEnabled = plan.AuthEnabled.ValueBool()
	}
	if !plan.DefaultProvider.IsNull() && !plan.DefaultProvider.IsUnknown() {
		payload.DefaultProvider = plan.DefaultProvider.ValueString()
	}
	if !plan.SessionTimeout.IsNull() && !plan.SessionTimeout.IsUnknown() {
		payload.SessionTimeout = plan.SessionTimeout.ValueInt64()
	}

	updated, _, err := r.client.UpdateAuthSettings(ctx, payload)
	if err != nil {
		diags.AddError("Error updating Dockhand auth settings", err.Error())
		return authSettingsModel{}, diags
	}

	return modelFromAuthSettings(updated), diags
}

func modelFromAuthSettings(in *authSettingsResponse) authSettingsModel {
	out := authSettingsModel{
		ID:              types.StringValue("auth"),
		AuthEnabled:     types.BoolValue(in.AuthEnabled),
		DefaultProvider: types.StringValue(in.DefaultProvider),
		SessionTimeout:  types.Int64Value(in.SessionTimeout),
	}

	if in.CreatedAt != nil {
		out.CreatedAt = types.StringValue(*in.CreatedAt)
	} else {
		out.CreatedAt = types.StringNull()
	}
	if in.UpdatedAt != nil {
		out.UpdatedAt = types.StringValue(*in.UpdatedAt)
	} else {
		out.UpdatedAt = types.StringNull()
	}

	return out
}

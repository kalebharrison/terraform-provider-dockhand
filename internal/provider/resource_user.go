package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*userResource)(nil)
	_ resource.ResourceWithConfigure   = (*userResource)(nil)
	_ resource.ResourceWithImportState = (*userResource)(nil)
)

func NewUserResource() resource.Resource {
	return &userResource{}
}

type userResource struct {
	client *Client
}

type userResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	Email       types.String `tfsdk:"email"`
	DisplayName types.String `tfsdk:"display_name"`
	IsAdmin     types.Bool   `tfsdk:"is_admin"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	MFAEnabled  types.Bool   `tfsdk:"mfa_enabled"`
	LastLogin   types.String `tfsdk:"last_login"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Numeric Dockhand user ID.",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "User password. Write-only.",
				Optional:            true,
				Sensitive:           true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "User email address.",
				Optional:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Display name.",
				Optional:            true,
			},
			"is_admin": schema.BoolAttribute{
				MarkdownDescription: "Whether this user is an admin.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"is_active": schema.BoolAttribute{
				MarkdownDescription: "Whether this user is active.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"mfa_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether MFA is enabled for this user.",
				Computed:            true,
			},
			"last_login": schema.StringAttribute{
				MarkdownDescription: "Last login timestamp.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Last updated timestamp.",
				Computed:            true,
			},
		},
	}
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := buildUserPayload(plan)
	created, err := r.client.CreateUser(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand user", err.Error())
		return
	}

	// Dockhand currently creates users as admin regardless of submitted isAdmin.
	// Reconcile desired state immediately after create.
	reconciled := created
	if userNeedsReconcile(plan, created) {
		updated, err := r.client.UpdateUser(ctx, fmt.Sprintf("%d", created.ID), buildUserPayload(plan))
		if err != nil {
			resp.Diagnostics.AddError("Error reconciling Dockhand user after create", err.Error())
			return
		}
		reconciled = updated
	}

	state := modelFromUserResponse(plan.Password, reconciled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, status, err := r.client.GetUser(ctx, state.ID.ValueString())
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand user", err.Error())
		return
	}

	newState := modelFromUserResponse(state.Password, user)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan userResourceModel
	var state userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Missing user ID", "Cannot update Dockhand user because ID is unknown.")
		return
	}

	updated, err := r.client.UpdateUser(ctx, id, buildUserPayload(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand user", err.Error())
		return
	}

	newState := modelFromUserResponse(plan.Password, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteUser(ctx, state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand user", err.Error())
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildUserPayload(model userResourceModel) userPayload {
	payload := userPayload{
		Username: model.Username.ValueString(),
		IsAdmin:  model.IsAdmin.ValueBool(),
		IsActive: model.IsActive.ValueBool(),
	}
	if !model.Password.IsNull() && !model.Password.IsUnknown() && model.Password.ValueString() != "" {
		value := model.Password.ValueString()
		payload.Password = &value
	}
	if !model.Email.IsNull() && !model.Email.IsUnknown() {
		value := model.Email.ValueString()
		payload.Email = &value
	}
	if !model.DisplayName.IsNull() && !model.DisplayName.IsUnknown() {
		value := model.DisplayName.ValueString()
		payload.DisplayName = &value
	}
	return payload
}

func modelFromUserResponse(password types.String, user *userResponse) userResourceModel {
	out := userResourceModel{
		ID:       types.StringValue(fmt.Sprintf("%d", user.ID)),
		Username: types.StringValue(user.Username),
		Password: password,
		IsAdmin:  types.BoolValue(user.IsAdmin),
		IsActive: types.BoolValue(user.IsActive),
	}

	if user.Email != nil {
		out.Email = types.StringValue(*user.Email)
	} else {
		out.Email = types.StringNull()
	}
	if user.DisplayName != nil {
		out.DisplayName = types.StringValue(*user.DisplayName)
	} else {
		out.DisplayName = types.StringNull()
	}
	out.MFAEnabled = types.BoolValue(user.MFAEnabled)
	if user.LastLogin != nil {
		out.LastLogin = types.StringValue(*user.LastLogin)
	} else {
		out.LastLogin = types.StringNull()
	}
	if user.CreatedAt != nil {
		out.CreatedAt = types.StringValue(*user.CreatedAt)
	} else {
		out.CreatedAt = types.StringNull()
	}
	if user.UpdatedAt != nil {
		out.UpdatedAt = types.StringValue(*user.UpdatedAt)
	} else {
		out.UpdatedAt = types.StringNull()
	}

	return out
}

func userNeedsReconcile(plan userResourceModel, created *userResponse) bool {
	if created == nil {
		return false
	}

	if created.IsAdmin != plan.IsAdmin.ValueBool() {
		return true
	}
	if created.IsActive != plan.IsActive.ValueBool() {
		return true
	}

	if !plan.Email.IsNull() && !plan.Email.IsUnknown() {
		want := plan.Email.ValueString()
		have := ""
		if created.Email != nil {
			have = *created.Email
		}
		if have != want {
			return true
		}
	}

	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		want := plan.DisplayName.ValueString()
		have := ""
		if created.DisplayName != nil {
			have = *created.DisplayName
		}
		if have != want {
			return true
		}
	}

	return false
}

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*registryResource)(nil)
	_ resource.ResourceWithConfigure   = (*registryResource)(nil)
	_ resource.ResourceWithImportState = (*registryResource)(nil)
)

func NewRegistryResource() resource.Resource {
	return &registryResource{}
}

type registryResource struct {
	client *Client
}

type registryModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	URL            types.String `tfsdk:"url"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	HasCredentials types.Bool   `tfsdk:"has_credentials"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func (r *registryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry"
}

func (r *registryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand image registry via `/api/registries`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Numeric Dockhand registry ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Registry display name.",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Registry base URL.",
				Required:            true,
			},
			"is_default": schema.BoolAttribute{
				MarkdownDescription: "Whether this registry is the default. Dockhand expects at most one default registry.",
				Optional:            true,
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Registry username. Set to an empty string along with `password` empty string to clear stored credentials.",
				Optional:            true,
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Registry password. Write-only. To clear stored credentials, set `username = \"\"` and `password = \"\"`.",
				Optional:            true,
				Sensitive:           true,
			},
			"has_credentials": schema.BoolAttribute{
				MarkdownDescription: "Whether this registry currently has stored credentials in Dockhand.",
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

func (r *registryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *registryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan registryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildRegistryPayload(plan, registryModel{})
	if err != nil {
		resp.Diagnostics.AddError("Invalid registry configuration", err.Error())
		return
	}

	created, _, err := r.client.CreateRegistry(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand registry", err.Error())
		return
	}

	state := modelFromRegistryResponse(plan.Password, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *registryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state registryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reg, status, err := r.client.GetRegistry(ctx, state.ID.ValueString())
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand registry", err.Error())
		return
	}

	newState := modelFromRegistryResponse(state.Password, reg)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *registryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan registryModel
	var state registryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Missing registry ID", "Cannot update Dockhand registry because ID is unknown.")
		return
	}

	payload, err := buildRegistryPayload(plan, state)
	if err != nil {
		resp.Diagnostics.AddError("Invalid registry configuration", err.Error())
		return
	}

	updated, _, err := r.client.UpdateRegistry(ctx, id, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand registry", err.Error())
		return
	}

	newState := modelFromRegistryResponse(plan.Password, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *registryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state registryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteRegistry(ctx, state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand registry", err.Error())
		return
	}
}

func (r *registryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildRegistryPayload(plan registryModel, prior registryModel) (map[string]any, error) {
	name := plan.Name.ValueString()
	url := plan.URL.ValueString()
	if name == "" || url == "" {
		return nil, fmt.Errorf("name and url are required")
	}

	payload := map[string]any{
		"name": name,
		"url":  url,
	}

	if !plan.IsDefault.IsUnknown() && !plan.IsDefault.IsNull() {
		payload["isDefault"] = plan.IsDefault.ValueBool()
	} else if !prior.IsDefault.IsUnknown() && !prior.IsDefault.IsNull() {
		payload["isDefault"] = prior.IsDefault.ValueBool()
	} else {
		payload["isDefault"] = false
	}

	usernameSet := !plan.Username.IsUnknown() && !plan.Username.IsNull()
	passwordSet := !plan.Password.IsUnknown() && !plan.Password.IsNull()

	if usernameSet {
		payload["username"] = plan.Username.ValueString()
	}
	if passwordSet {
		payload["password"] = plan.Password.ValueString()
	}

	// Clearing credentials requires both empty strings, otherwise we consider it ambiguous.
	if usernameSet && plan.Username.ValueString() == "" && !passwordSet {
		return nil, fmt.Errorf("to clear credentials, set both username and password to empty strings")
	}
	if passwordSet && plan.Password.ValueString() == "" && !usernameSet {
		return nil, fmt.Errorf("to clear credentials, set both username and password to empty strings")
	}

	return payload, nil
}

func modelFromRegistryResponse(password types.String, reg *registryResponse) registryModel {
	out := registryModel{
		ID:             types.StringValue(fmt.Sprintf("%d", reg.ID)),
		Name:           types.StringValue(reg.Name),
		URL:            types.StringValue(reg.URL),
		IsDefault:      types.BoolValue(reg.IsDefault),
		Password:       password, // write-only passthrough
		HasCredentials: types.BoolValue(reg.HasCredentials),
	}

	if reg.Username != nil {
		out.Username = types.StringValue(*reg.Username)
	} else {
		out.Username = types.StringNull()
	}
	if reg.CreatedAt != nil {
		out.CreatedAt = types.StringValue(*reg.CreatedAt)
	} else {
		out.CreatedAt = types.StringNull()
	}
	if reg.UpdatedAt != nil {
		out.UpdatedAt = types.StringValue(*reg.UpdatedAt)
	} else {
		out.UpdatedAt = types.StringNull()
	}

	return out
}

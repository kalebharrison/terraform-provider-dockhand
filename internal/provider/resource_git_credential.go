package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*gitCredentialResource)(nil)
	_ resource.ResourceWithConfigure   = (*gitCredentialResource)(nil)
	_ resource.ResourceWithImportState = (*gitCredentialResource)(nil)
)

func NewGitCredentialResource() resource.Resource {
	return &gitCredentialResource{}
}

type gitCredentialResource struct {
	client *Client
}

type gitCredentialModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	AuthType    types.String `tfsdk:"auth_type"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	SSHKey      types.String `tfsdk:"ssh_key"`
	HasPassword types.Bool   `tfsdk:"has_password"`
	HasSSHKey   types.Bool   `tfsdk:"has_ssh_key"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *gitCredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_credential"
}

func (r *gitCredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand Git credential via `/api/git/credentials`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Numeric Dockhand Git credential ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Credential display name.",
				Required:            true,
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "Authentication type. Known values observed: `password`.",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username used for `auth_type = \"password\"`.",
				Optional:            true,
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password/token used for `auth_type = \"password\"`. Write-only.",
				Optional:            true,
				Sensitive:           true,
			},
			"ssh_key": schema.StringAttribute{
				MarkdownDescription: "SSH private key used for `auth_type = \"ssh\"`. Write-only.",
				Optional:            true,
				Sensitive:           true,
			},
			"has_password": schema.BoolAttribute{
				MarkdownDescription: "Whether Dockhand has a password/token stored for this credential.",
				Computed:            true,
			},
			"has_ssh_key": schema.BoolAttribute{
				MarkdownDescription: "Whether Dockhand has an SSH key stored for this credential.",
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

func (r *gitCredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gitCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan gitCredentialModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildGitCredentialPayload(plan, gitCredentialModel{}, true)
	if err != nil {
		resp.Diagnostics.AddError("Invalid git credential configuration", err.Error())
		return
	}

	created, _, err := r.client.CreateGitCredential(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand git credential", err.Error())
		return
	}

	state := modelFromGitCredentialResponse(plan.Password, plan.SSHKey, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gitCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state gitCredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cred, status, err := r.client.GetGitCredential(ctx, state.ID.ValueString())
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand git credential", err.Error())
		return
	}

	newState := modelFromGitCredentialResponse(state.Password, state.SSHKey, cred)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *gitCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan gitCredentialModel
	var state gitCredentialModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Missing credential ID", "Cannot update Dockhand git credential because ID is unknown.")
		return
	}

	payload, err := buildGitCredentialPayload(plan, state, false)
	if err != nil {
		resp.Diagnostics.AddError("Invalid git credential configuration", err.Error())
		return
	}

	updated, _, err := r.client.UpdateGitCredential(ctx, id, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand git credential", err.Error())
		return
	}

	newState := modelFromGitCredentialResponse(plan.Password, plan.SSHKey, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *gitCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state gitCredentialModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteGitCredential(ctx, state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand git credential", err.Error())
		return
	}
}

func (r *gitCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildGitCredentialPayload(plan gitCredentialModel, prior gitCredentialModel, requireSecrets bool) (gitCredentialPayload, error) {
	name := plan.Name.ValueString()
	authType := strings.TrimSpace(plan.AuthType.ValueString())

	if name == "" {
		return gitCredentialPayload{}, fmt.Errorf("name is required")
	}
	if authType == "" {
		return gitCredentialPayload{}, fmt.Errorf("auth_type is required")
	}

	payload := gitCredentialPayload{
		Name:     name,
		AuthType: authType,
	}

	if !plan.Username.IsNull() && !plan.Username.IsUnknown() {
		v := plan.Username.ValueString()
		payload.Username = &v
	} else if !prior.Username.IsNull() && !prior.Username.IsUnknown() {
		v := prior.Username.ValueString()
		if v != "" {
			payload.Username = &v
		}
	}

	if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		v := plan.Password.ValueString()
		payload.Password = &v
	}

	if !plan.SSHKey.IsNull() && !plan.SSHKey.IsUnknown() {
		v := plan.SSHKey.ValueString()
		payload.SSHKey = &v
	}

	switch authType {
	case "password":
		if payload.Username == nil || *payload.Username == "" {
			return gitCredentialPayload{}, fmt.Errorf("username is required when auth_type is \"password\"")
		}
		if requireSecrets && (payload.Password == nil || *payload.Password == "") {
			return gitCredentialPayload{}, fmt.Errorf("password is required when auth_type is \"password\"")
		}
	case "ssh":
		if requireSecrets && (payload.SSHKey == nil || *payload.SSHKey == "") {
			return gitCredentialPayload{}, fmt.Errorf("ssh_key is required when auth_type is \"ssh\"")
		}
	default:
		// We don't hard-fail on unknown types, but we do validate required fields for the known ones.
	}

	return payload, nil
}

func modelFromGitCredentialResponse(password types.String, sshKey types.String, in *gitCredentialResponse) gitCredentialModel {
	out := gitCredentialModel{
		ID:          types.StringValue(fmt.Sprintf("%d", in.ID)),
		Name:        types.StringValue(in.Name),
		AuthType:    types.StringValue(in.AuthType),
		Password:    password,
		SSHKey:      sshKey,
		HasPassword: types.BoolValue(in.HasPassword),
		HasSSHKey:   types.BoolValue(in.HasSSHKey),
	}

	if in.Username != nil {
		out.Username = types.StringValue(*in.Username)
	} else {
		out.Username = types.StringNull()
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

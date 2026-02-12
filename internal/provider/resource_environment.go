package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*environmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*environmentResource)(nil)
	_ resource.ResourceWithImportState = (*environmentResource)(nil)
)

func NewEnvironmentResource() resource.Resource {
	return &environmentResource{}
}

type environmentResource struct {
	client *Client
}

type environmentModel struct {
	ID types.String `tfsdk:"id"`

	Name           types.String `tfsdk:"name"`
	ConnectionType types.String `tfsdk:"connection_type"`
	Host           types.String `tfsdk:"host"`
	Port           types.Int64  `tfsdk:"port"`
	Protocol       types.String `tfsdk:"protocol"`
	SocketPath     types.String `tfsdk:"socket_path"`
	TLSSkipVerify  types.Bool   `tfsdk:"tls_skip_verify"`
	Icon           types.String `tfsdk:"icon"`

	CollectActivity  types.Bool `tfsdk:"collect_activity"`
	CollectMetrics   types.Bool `tfsdk:"collect_metrics"`
	HighlightChanges types.Bool `tfsdk:"highlight_changes"`

	Timezone types.String `tfsdk:"timezone"`

	UpdateCheckEnabled    types.Bool `tfsdk:"update_check_enabled"`
	UpdateCheckAutoUpdate types.Bool `tfsdk:"update_check_auto_update"`
	ImagePruneEnabled     types.Bool `tfsdk:"image_prune_enabled"`

	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *environmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *environmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand environment via `/api/environments`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Numeric Dockhand environment ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"connection_type": schema.StringAttribute{
				MarkdownDescription: "Environment connection type. Example: `socket`.",
				Optional:            true,
				Computed:            true,
			},
			"host": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"protocol": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"socket_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"tls_skip_verify": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"icon": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"collect_activity": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"collect_metrics": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"highlight_changes": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"timezone": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"update_check_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"update_check_auto_update": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"image_prune_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
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

func (r *environmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *environmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan environmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildEnvironmentPayload(plan, environmentModel{})
	if err != nil {
		resp.Diagnostics.AddError("Invalid environment configuration", err.Error())
		return
	}

	created, _, err := r.client.CreateEnvironment(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand environment", err.Error())
		return
	}

	state := modelFromEnvironmentResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *environmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state environmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, status, err := r.client.GetEnvironment(ctx, state.ID.ValueString())
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand environment", err.Error())
		return
	}

	newState := modelFromEnvironmentResponse(env)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *environmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan environmentModel
	var state environmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Missing environment ID", "Cannot update Dockhand environment because ID is unknown.")
		return
	}

	payload, err := buildEnvironmentPayload(plan, state)
	if err != nil {
		resp.Diagnostics.AddError("Invalid environment configuration", err.Error())
		return
	}

	updated, _, err := r.client.UpdateEnvironment(ctx, id, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand environment", err.Error())
		return
	}

	newState := modelFromEnvironmentResponse(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *environmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state environmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteEnvironment(ctx, state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand environment", err.Error())
		return
	}
}

func (r *environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildEnvironmentPayload(plan environmentModel, prior environmentModel) (environmentPayload, error) {
	name := strings.TrimSpace(plan.Name.ValueString())
	if name == "" {
		return environmentPayload{}, fmt.Errorf("name is required")
	}

	connectionType := firstKnownString(plan.ConnectionType, prior.ConnectionType)
	if connectionType == "" {
		connectionType = "socket"
	}

	payload := environmentPayload{
		Name:           name,
		ConnectionType: connectionType,
	}

	if v := firstKnownString(plan.Host, prior.Host); v != "" {
		payload.Host = &v
	}
	if v := firstKnownInt64(plan.Port, prior.Port); v != 0 {
		payload.Port = &v
	}
	if v := firstKnownString(plan.Protocol, prior.Protocol); v != "" {
		payload.Protocol = &v
	}
	if v := firstKnownString(plan.SocketPath, prior.SocketPath); v != "" {
		payload.SocketPath = &v
	}
	if v, ok := firstKnownBoolPtr(plan.TLSSkipVerify, prior.TLSSkipVerify); ok {
		payload.TLSSkipVerify = &v
	}
	if v := firstKnownString(plan.Icon, prior.Icon); v != "" {
		payload.Icon = &v
	}
	if v, ok := firstKnownBoolPtr(plan.CollectActivity, prior.CollectActivity); ok {
		payload.CollectActivity = &v
	}
	if v, ok := firstKnownBoolPtr(plan.CollectMetrics, prior.CollectMetrics); ok {
		payload.CollectMetrics = &v
	}
	if v, ok := firstKnownBoolPtr(plan.HighlightChanges, prior.HighlightChanges); ok {
		payload.HighlightChanges = &v
	}
	if v := firstKnownString(plan.Timezone, prior.Timezone); v != "" {
		payload.Timezone = &v
	}
	if v, ok := firstKnownBoolPtr(plan.UpdateCheckEnabled, prior.UpdateCheckEnabled); ok {
		payload.UpdateCheckEnabled = &v
	}
	if v, ok := firstKnownBoolPtr(plan.UpdateCheckAutoUpdate, prior.UpdateCheckAutoUpdate); ok {
		payload.UpdateCheckAutoUpdate = &v
	}
	if v, ok := firstKnownBoolPtr(plan.ImagePruneEnabled, prior.ImagePruneEnabled); ok {
		payload.ImagePruneEnabled = &v
	}

	if connectionType == "socket" && payload.SocketPath == nil {
		return environmentPayload{}, fmt.Errorf("socket_path is required when connection_type is \"socket\"")
	}

	return payload, nil
}

func modelFromEnvironmentResponse(in *environmentResponse) environmentModel {
	out := environmentModel{
		ID:                    types.StringValue(strconv.FormatInt(in.ID, 10)),
		Name:                  types.StringValue(in.Name),
		ConnectionType:        types.StringValue(in.ConnectionType),
		Port:                  types.Int64Value(in.Port),
		Protocol:              types.StringValue(in.Protocol),
		TLSSkipVerify:         types.BoolValue(in.TLSSkipVerify),
		Icon:                  types.StringValue(in.Icon),
		CollectActivity:       types.BoolValue(in.CollectActivity),
		CollectMetrics:        types.BoolValue(in.CollectMetrics),
		HighlightChanges:      types.BoolValue(in.HighlightChanges),
		UpdateCheckEnabled:    types.BoolValue(in.UpdateCheckEnabled),
		UpdateCheckAutoUpdate: types.BoolValue(in.UpdateCheckAutoUpdate),
		ImagePruneEnabled:     types.BoolValue(in.ImagePruneEnabled),
	}

	if in.Host != nil {
		out.Host = types.StringValue(*in.Host)
	} else {
		out.Host = types.StringNull()
	}
	if in.SocketPath != nil {
		out.SocketPath = types.StringValue(*in.SocketPath)
	} else {
		out.SocketPath = types.StringNull()
	}
	if in.Timezone != nil {
		out.Timezone = types.StringValue(*in.Timezone)
	} else {
		out.Timezone = types.StringNull()
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

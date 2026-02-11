package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*configSetResource)(nil)
	_ resource.ResourceWithConfigure   = (*configSetResource)(nil)
	_ resource.ResourceWithImportState = (*configSetResource)(nil)
)

func NewConfigSetResource() resource.Resource {
	return &configSetResource{}
}

type configSetResource struct {
	client *Client
}

type configSetPortModel struct {
	ContainerPort types.Int64  `tfsdk:"container_port"`
	HostPort      types.Int64  `tfsdk:"host_port"`
	Protocol      types.String `tfsdk:"protocol"`
}

type configSetVolumeModel struct {
	Source   types.String `tfsdk:"source"`
	Target   types.String `tfsdk:"target"`
	Type     types.String `tfsdk:"type"`
	ReadOnly types.Bool   `tfsdk:"read_only"`
}

type configSetModel struct {
	ID            types.String           `tfsdk:"id"`
	Name          types.String           `tfsdk:"name"`
	Description   types.String           `tfsdk:"description"`
	EnvVars       types.Map              `tfsdk:"env_vars"`
	Labels        types.Map              `tfsdk:"labels"`
	Ports         []configSetPortModel   `tfsdk:"ports"`
	Volumes       []configSetVolumeModel `tfsdk:"volumes"`
	NetworkMode   types.String           `tfsdk:"network_mode"`
	RestartPolicy types.String           `tfsdk:"restart_policy"`
	CreatedAt     types.String           `tfsdk:"created_at"`
	UpdatedAt     types.String           `tfsdk:"updated_at"`
}

func (r *configSetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_set"
}

func (r *configSetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand Config Set via `/api/config-sets`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Numeric Dockhand config set ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Config set name.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"env_vars": schema.MapAttribute{
				MarkdownDescription: "Environment variables applied by the config set.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Labels applied by the config set.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"ports": schema.ListNestedAttribute{
				MarkdownDescription: "Port mappings applied by the config set.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"container_port": schema.Int64Attribute{Required: true},
						"host_port":      schema.Int64Attribute{Required: true},
						"protocol":       schema.StringAttribute{Optional: true, Computed: true},
					},
				},
			},
			"volumes": schema.ListNestedAttribute{
				MarkdownDescription: "Volume mounts applied by the config set.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"source":    schema.StringAttribute{Required: true},
						"target":    schema.StringAttribute{Required: true},
						"type":      schema.StringAttribute{Optional: true, Computed: true},
						"read_only": schema.BoolAttribute{Optional: true, Computed: true},
					},
				},
			},
			"network_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"restart_policy": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

func (r *configSetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *configSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan configSetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := buildConfigSetPayload(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, _, err := r.client.CreateConfigSet(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand config set", err.Error())
		return
	}

	state, diags := modelFromConfigSetResponse(ctx, plan, created)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *configSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state configSetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cs, status, err := r.client.GetConfigSet(ctx, state.ID.ValueString())
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand config set", err.Error())
		return
	}

	newState, diags := modelFromConfigSetResponse(ctx, state, cs)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *configSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan configSetModel
	var state configSetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Missing config set ID", "Cannot update Dockhand config set because ID is unknown.")
		return
	}

	payload, diags := buildConfigSetPayload(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, _, err := r.client.UpdateConfigSet(ctx, id, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand config set", err.Error())
		return
	}

	newState, diags := modelFromConfigSetResponse(ctx, plan, updated)
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *configSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state configSetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteConfigSet(ctx, state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand config set", err.Error())
		return
	}
}

func (r *configSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildConfigSetPayload(ctx context.Context, plan configSetModel) (configSetPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := plan.Name.ValueString()
	if name == "" {
		diags.AddError("Invalid config set configuration", "name is required")
		return configSetPayload{}, diags
	}

	var description *string
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		v := plan.Description.ValueString()
		description = &v
	}

	payload := configSetPayload{
		Name:        name,
		Description: description,
	}

	if !plan.NetworkMode.IsNull() && !plan.NetworkMode.IsUnknown() {
		v := plan.NetworkMode.ValueString()
		payload.NetworkMode = &v
	}
	if !plan.RestartPolicy.IsNull() && !plan.RestartPolicy.IsUnknown() {
		v := plan.RestartPolicy.ValueString()
		payload.RestartPolicy = &v
	}

	if !plan.EnvVars.IsNull() && !plan.EnvVars.IsUnknown() {
		var envMap map[string]string
		diags.Append(plan.EnvVars.ElementsAs(ctx, &envMap, false)...)
		if !diags.HasError() {
			payload.EnvVars = kvsFromMap(envMap)
		}
	}
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		var labelMap map[string]string
		diags.Append(plan.Labels.ElementsAs(ctx, &labelMap, false)...)
		if !diags.HasError() {
			payload.Labels = kvsFromMap(labelMap)
		}
	}

	if plan.Ports != nil {
		payload.Ports = make([]configSetPort, 0, len(plan.Ports))
		for _, p := range plan.Ports {
			proto := "tcp"
			if !p.Protocol.IsNull() && !p.Protocol.IsUnknown() && p.Protocol.ValueString() != "" {
				proto = p.Protocol.ValueString()
			}
			payload.Ports = append(payload.Ports, configSetPort{
				ContainerPort: p.ContainerPort.ValueInt64(),
				HostPort:      p.HostPort.ValueInt64(),
				Protocol:      proto,
			})
		}
	}

	if plan.Volumes != nil {
		payload.Volumes = make([]configSetVolume, 0, len(plan.Volumes))
		for _, v := range plan.Volumes {
			t := "bind"
			if !v.Type.IsNull() && !v.Type.IsUnknown() && v.Type.ValueString() != "" {
				t = v.Type.ValueString()
			}
			ro := false
			if !v.ReadOnly.IsNull() && !v.ReadOnly.IsUnknown() {
				ro = v.ReadOnly.ValueBool()
			}
			payload.Volumes = append(payload.Volumes, configSetVolume{
				Source:   v.Source.ValueString(),
				Target:   v.Target.ValueString(),
				Type:     t,
				ReadOnly: ro,
			})
		}
	}

	return payload, diags
}

func modelFromConfigSetResponse(ctx context.Context, prior configSetModel, in *configSetResponse) (configSetModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	out := configSetModel{
		ID:            types.StringValue(fmt.Sprintf("%d", in.ID)),
		Name:          types.StringValue(in.Name),
		NetworkMode:   types.StringValue(in.NetworkMode),
		RestartPolicy: types.StringValue(in.RestartPolicy),
	}

	if in.Description != nil {
		out.Description = types.StringValue(*in.Description)
	} else {
		out.Description = types.StringNull()
	}

	envMap := mapFromKVs(in.EnvVars)
	labelMap := mapFromKVs(in.Labels)

	envVal, envDiag := types.MapValueFrom(ctx, types.StringType, envMap)
	diags.Append(envDiag...)
	out.EnvVars = envVal

	labelVal, labelDiag := types.MapValueFrom(ctx, types.StringType, labelMap)
	diags.Append(labelDiag...)
	out.Labels = labelVal

	out.Ports = make([]configSetPortModel, 0, len(in.Ports))
	for _, p := range in.Ports {
		out.Ports = append(out.Ports, configSetPortModel{
			ContainerPort: types.Int64Value(p.ContainerPort),
			HostPort:      types.Int64Value(p.HostPort),
			Protocol:      types.StringValue(p.Protocol),
		})
	}

	out.Volumes = make([]configSetVolumeModel, 0, len(in.Volumes))
	for _, v := range in.Volumes {
		out.Volumes = append(out.Volumes, configSetVolumeModel{
			Source:   types.StringValue(v.Source),
			Target:   types.StringValue(v.Target),
			Type:     types.StringValue(v.Type),
			ReadOnly: types.BoolValue(v.ReadOnly),
		})
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

	// Preserve unknown id if needed.
	if !prior.ID.IsNull() && !prior.ID.IsUnknown() && prior.ID.ValueString() != "" {
		out.ID = prior.ID
	}

	return out, diags
}

func kvsFromMap(in map[string]string) []configSetKV {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]configSetKV, 0, len(keys))
	for _, k := range keys {
		out = append(out, configSetKV{Key: k, Value: in[k]})
	}
	return out
}

func mapFromKVs(in []configSetKV) map[string]string {
	out := make(map[string]string, len(in))
	for _, kv := range in {
		if kv.Key == "" {
			continue
		}
		out[kv.Key] = kv.Value
	}
	return out
}

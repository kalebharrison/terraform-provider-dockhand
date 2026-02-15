package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*containerResource)(nil)
	_ resource.ResourceWithConfigure   = (*containerResource)(nil)
	_ resource.ResourceWithImportState = (*containerResource)(nil)
)

func NewContainerResource() resource.Resource {
	return &containerResource{}
}

type containerResource struct {
	client *Client
}

type containerPortModel struct {
	ContainerPort types.Int64  `tfsdk:"container_port"`
	HostPort      types.String `tfsdk:"host_port"`
	Protocol      types.String `tfsdk:"protocol"`
}

type containerResourceModel struct {
	ID            types.String         `tfsdk:"id"`
	Name          types.String         `tfsdk:"name"`
	Env           types.String         `tfsdk:"env"`
	Image         types.String         `tfsdk:"image"`
	Command       types.String         `tfsdk:"command"`
	Enabled       types.Bool           `tfsdk:"enabled"`
	NetworkMode   types.String         `tfsdk:"network_mode"`
	RestartPolicy types.String         `tfsdk:"restart_policy"`
	Privileged    types.Bool           `tfsdk:"privileged"`
	TTY           types.Bool           `tfsdk:"tty"`
	MemoryBytes   types.Int64          `tfsdk:"memory_bytes"`
	NanoCPUs      types.Int64          `tfsdk:"nano_cpus"`
	CapAdd        types.List           `tfsdk:"cap_add"`
	EnvVars       types.Map            `tfsdk:"env_vars"`
	Labels        types.Map            `tfsdk:"labels"`
	Ports         []containerPortModel `tfsdk:"ports"`
	UpdatePayload types.String         `tfsdk:"update_payload_json"`
	State         types.String         `tfsdk:"state"`
	Status        types.String         `tfsdk:"status"`
	Health        types.String         `tfsdk:"health"`
	RestartCount  types.Int64          `tfsdk:"restart_count"`
}

func (r *containerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container"
}

func (r *containerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand container using `/api/containers` endpoints.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Container ID.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Container name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env": schema.StringAttribute{
				MarkdownDescription: "Dockhand environment ID sent as the `env` query parameter.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "Container image reference.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"command": schema.StringAttribute{
				MarkdownDescription: "Optional command string sent at container create time.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Desired runtime state. `true` starts container, `false` stops it.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"network_mode": schema.StringAttribute{
				MarkdownDescription: "Network mode for create request (for example `bridge`, `host`).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"restart_policy": schema.StringAttribute{
				MarkdownDescription: "Restart policy for create request (for example `no`, `unless-stopped`).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"privileged": schema.BoolAttribute{
				MarkdownDescription: "Whether to create the container in privileged mode.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"tty": schema.BoolAttribute{
				MarkdownDescription: "Whether to allocate a TTY at create time.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"memory_bytes": schema.Int64Attribute{
				MarkdownDescription: "Container memory limit in bytes.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"nano_cpus": schema.Int64Attribute{
				MarkdownDescription: "CPU quota in NanoCPUs (for example `500000000` = 0.5 CPU).",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"cap_add": schema.ListAttribute{
				MarkdownDescription: "Linux capabilities to add at container create time.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"env_vars": schema.MapAttribute{
				MarkdownDescription: "Environment variables map for create request.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Labels map for create request.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"ports": schema.ListNestedAttribute{
				MarkdownDescription: "Port mappings for create request.",
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"container_port": schema.Int64Attribute{
							Required: true,
						},
						"host_port": schema.StringAttribute{
							Required: true,
						},
						"protocol": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"update_payload_json": schema.StringAttribute{
				MarkdownDescription: "Optional raw JSON object sent to `/api/containers/{id}/update` after create and on updates. Use this to access advanced Dockhand update fields not yet modeled as first-class attributes.",
				Optional:            true,
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Current container state from Dockhand.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current container status from Dockhand.",
				Computed:            true,
			},
			"health": schema.StringAttribute{
				MarkdownDescription: "Current container health from Dockhand.",
				Computed:            true,
			},
			"restart_count": schema.Int64Attribute{
				MarkdownDescription: "Current container restart count.",
				Computed:            true,
			},
		},
	}
}

func (r *containerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *containerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan containerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := containerPayload{
		Name:   plan.Name.ValueString(),
		Image:  plan.Image.ValueString(),
		Env:    flattenEnvVars(ctx, plan.EnvVars),
		Labels: flattenStringMap(ctx, plan.Labels),
		Ports:  flattenContainerPorts(plan.Ports),
	}
	if v := stringPtrFromStringValue(plan.Command); v != nil {
		payload.Command = v
	}
	if v := stringPtrFromStringValue(plan.NetworkMode); v != nil {
		payload.NetworkMode = v
	}
	if v := stringPtrFromStringValue(plan.RestartPolicy); v != nil {
		payload.RestartPolicy = v
	}
	if v := boolPtrFromBoolValue(plan.Privileged); v != nil {
		payload.Privileged = v
	}
	if v := boolPtrFromBoolValue(plan.TTY); v != nil {
		payload.TTY = v
	}
	if v := int64PtrFromInt64Value(plan.MemoryBytes); v != nil {
		payload.Memory = v
	}
	if v := int64PtrFromInt64Value(plan.NanoCPUs); v != nil {
		payload.NanoCPUs = v
	}
	if v := flattenStringList(ctx, plan.CapAdd); len(v) > 0 {
		payload.CapAdd = v
	}

	created, _, err := r.client.CreateContainer(ctx, plan.Env.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand container", err.Error())
		return
	}
	if created == nil || strings.TrimSpace(created.ID) == "" {
		resp.Diagnostics.AddError("Error creating Dockhand container", "Dockhand create endpoint did not return a container ID.")
		return
	}

	plan.ID = types.StringValue(created.ID)

	if payloadRaw := strings.TrimSpace(plan.UpdatePayload.ValueString()); payloadRaw != "" {
		updatePayload, parseErr := parseContainerUpdatePayload(payloadRaw)
		if parseErr != nil {
			resp.Diagnostics.AddError("Invalid `update_payload_json`", parseErr.Error())
			return
		}
		if _, status, err := r.client.UpdateContainer(ctx, plan.Env.ValueString(), created.ID, updatePayload); err != nil {
			resp.Diagnostics.AddError("Error applying Dockhand container update payload", err.Error())
			return
		} else if status < 200 || status > 299 {
			resp.Diagnostics.AddError("Error applying Dockhand container update payload", fmt.Sprintf("Dockhand returned status %d", status))
			return
		}
		plan.UpdatePayload = types.StringValue(payloadRaw)
	} else {
		plan.UpdatePayload = types.StringNull()
	}

	if !plan.Enabled.ValueBool() {
		if _, err := r.client.StopContainer(ctx, plan.Env.ValueString(), created.ID); err != nil {
			resp.Diagnostics.AddError("Error stopping Dockhand container after create", err.Error())
			return
		}
	}

	container, found, err := r.client.GetContainerByID(ctx, plan.Env.ValueString(), created.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand container after create", err.Error())
		return
	}
	if found {
		applyContainerRuntimeToState(&plan, container)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state containerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	container, found, err := r.client.GetContainerByID(ctx, state.Env.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand container", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	applyContainerRuntimeToState(&state, container)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *containerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan containerResourceModel
	var state containerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	env := plan.Env.ValueString()
	if plan.Enabled.ValueBool() != state.Enabled.ValueBool() {
		var err error
		if plan.Enabled.ValueBool() {
			_, err = r.client.StartContainer(ctx, env, id)
		} else {
			_, err = r.client.StopContainer(ctx, env, id)
		}
		if err != nil {
			resp.Diagnostics.AddError("Error updating Dockhand container runtime state", err.Error())
			return
		}
	}

	payloadRaw := strings.TrimSpace(plan.UpdatePayload.ValueString())
	if payloadRaw != "" {
		updatePayload, parseErr := parseContainerUpdatePayload(payloadRaw)
		if parseErr != nil {
			resp.Diagnostics.AddError("Invalid `update_payload_json`", parseErr.Error())
			return
		}
		if payloadRaw != strings.TrimSpace(state.UpdatePayload.ValueString()) {
			if _, status, err := r.client.UpdateContainer(ctx, env, id, updatePayload); err != nil {
				resp.Diagnostics.AddError("Error applying Dockhand container update payload", err.Error())
				return
			} else if status < 200 || status > 299 {
				resp.Diagnostics.AddError("Error applying Dockhand container update payload", fmt.Sprintf("Dockhand returned status %d", status))
				return
			}
		}
		plan.UpdatePayload = types.StringValue(payloadRaw)
	} else {
		plan.UpdatePayload = types.StringNull()
	}

	container, found, err := r.client.GetContainerByID(ctx, env, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand container after update", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	plan.ID = state.ID
	applyContainerRuntimeToState(&plan, container)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state containerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteContainer(ctx, state.Env.ValueString(), state.ID.ValueString())
	if err != nil && status != 404 {
		// Some Dockhand builds return 500 "Failed to remove container" when the
		// container is already gone. Re-check existence before failing destroy.
		_, found, readErr := r.client.GetContainerByID(ctx, state.Env.ValueString(), state.ID.ValueString())
		if readErr != nil {
			resp.Diagnostics.AddError("Error confirming Dockhand container deletion", readErr.Error())
			return
		}
		if found {
			resp.Diagnostics.AddError("Error deleting Dockhand container", err.Error())
			return
		}
	}

	// Dockhand may briefly report containers as "marked for removal". Wait until
	// the container no longer appears before allowing dependent deletes (images).
	for range 30 {
		_, found, readErr := r.client.GetContainerByID(ctx, state.Env.ValueString(), state.ID.ValueString())
		if readErr != nil {
			resp.Diagnostics.AddError("Error confirming Dockhand container deletion", readErr.Error())
			return
		}
		if !found {
			return
		}
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Error confirming Dockhand container deletion", ctx.Err().Error())
			return
		case <-time.After(1 * time.Second):
		}
	}

	resp.Diagnostics.AddError(
		"Timed out waiting for Dockhand container deletion",
		fmt.Sprintf("Container %q still appears after delete; retry destroy.", state.ID.ValueString()),
	)
}

func (r *containerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<id>` or `<env>:<id>`.")
		return
	}

	env := ""
	id := raw
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) == 2 {
		env = strings.TrimSpace(parts[0])
		id = strings.TrimSpace(parts[1])
	}
	if id == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Container ID cannot be empty.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	if env != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("env"), env)...)
	}
}

func applyContainerRuntimeToState(state *containerResourceModel, container *containerResponse) {
	if container == nil {
		return
	}
	state.State = types.StringValue(container.State)
	state.Status = types.StringValue(container.Status)
	state.Health = types.StringValue(container.Health)
	state.RestartCount = types.Int64Value(container.RestartCount)
}

func parseContainerUpdatePayload(raw string) (map[string]any, error) {
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("must be valid JSON object: %w", err)
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return payload, nil
}

func flattenEnvVars(ctx context.Context, value types.Map) []string {
	values := flattenStringMap(ctx, value)
	if len(values) == 0 {
		return nil
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(values))
	for _, key := range keys {
		out = append(out, fmt.Sprintf("%s=%s", key, values[key]))
	}
	return out
}

func flattenStringMap(ctx context.Context, value types.Map) map[string]string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	out := map[string]string{}
	diags := value.ElementsAs(ctx, &out, false)
	if diags.HasError() {
		return nil
	}
	return out
}

func flattenContainerPorts(values []containerPortModel) []containerPortPayload {
	if len(values) == 0 {
		return nil
	}

	out := make([]containerPortPayload, 0, len(values))
	for _, p := range values {
		protocol := strings.TrimSpace(p.Protocol.ValueString())
		if protocol == "" {
			protocol = "tcp"
		}
		out = append(out, containerPortPayload{
			ContainerPort: p.ContainerPort.ValueInt64(),
			HostPort:      p.HostPort.ValueString(),
			Protocol:      protocol,
		})
	}
	return out
}

func stringPtrFromStringValue(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := strings.TrimSpace(value.ValueString())
	if v == "" {
		return nil
	}
	return &v
}

func boolPtrFromBoolValue(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueBool()
	return &v
}

func int64PtrFromInt64Value(value types.Int64) *int64 {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueInt64()
	return &v
}

func flattenStringList(ctx context.Context, value types.List) []string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	out := []string{}
	diags := value.ElementsAs(ctx, &out, false)
	if diags.HasError() {
		return nil
	}
	return out
}

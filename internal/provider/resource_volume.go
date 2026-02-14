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
	_ resource.Resource                = (*volumeResource)(nil)
	_ resource.ResourceWithConfigure   = (*volumeResource)(nil)
	_ resource.ResourceWithImportState = (*volumeResource)(nil)
)

func NewVolumeResource() resource.Resource {
	return &volumeResource{}
}

type volumeResource struct {
	client *Client
}

type volumeModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Driver        types.String `tfsdk:"driver"`
	Env           types.String `tfsdk:"env"`
	DriverOptions types.Map    `tfsdk:"driver_options"`
	Labels        types.Map    `tfsdk:"labels"`
	Mountpoint    types.String `tfsdk:"mountpoint"`
	Scope         types.String `tfsdk:"scope"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func (r *volumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *volumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand volume via `/api/volumes`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Volume name (same as `name`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Volume name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"driver": schema.StringAttribute{
				MarkdownDescription: "Volume driver. Defaults to `local`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env": schema.StringAttribute{
				MarkdownDescription: "Optional environment ID. If omitted, provider `default_env` is used.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"driver_options": schema.MapAttribute{
				MarkdownDescription: "Driver options for the volume.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Labels for the volume.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"mountpoint": schema.StringAttribute{
				Computed: true,
			},
			"scope": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *volumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan volumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(plan.Name.ValueString())
	if name == "" {
		resp.Diagnostics.AddError("Missing volume name", "`name` must be set.")
		return
	}

	driver := strings.TrimSpace(plan.Driver.ValueString())
	if driver == "" {
		driver = "local"
	}
	driverOptions := map[string]string{}
	if !plan.DriverOptions.IsNull() && !plan.DriverOptions.IsUnknown() {
		resp.Diagnostics.Append(plan.DriverOptions.ElementsAs(ctx, &driverOptions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	labels := map[string]string{}
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	env := strings.TrimSpace(plan.Env.ValueString())
	resolvedEnv := r.client.resolveEnv(env)

	_, _, err := r.client.CreateVolume(ctx, env, volumePayload{
		Name:       name,
		Driver:     driver,
		DriverOpts: driverOptions,
		Labels:     labels,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand volume", err.Error())
		return
	}

	vol, status, err := r.client.GetVolumeInspect(ctx, env, name)
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error reading created Dockhand volume", err.Error())
		return
	}

	state := volumeModel{
		ID:            types.StringValue(name),
		Name:          types.StringValue(name),
		Driver:        types.StringValue(driver),
		Env:           types.StringNull(),
		DriverOptions: plan.DriverOptions,
		Labels:        plan.Labels,
		Mountpoint:    types.StringNull(),
		Scope:         types.StringNull(),
		CreatedAt:     types.StringNull(),
	}
	if resolvedEnv != "" {
		state.Env = types.StringValue(resolvedEnv)
	}
	if vol != nil {
		state = modelFromVolumeResponse(state.Env, vol)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vol, status, err := r.client.GetVolumeInspect(ctx, strings.TrimSpace(state.Env.ValueString()), state.Name.ValueString())
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand volume", err.Error())
		return
	}

	newState := modelFromVolumeResponse(state.Env, vol)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *volumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable attributes require replacement.
	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state volumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteVolume(ctx, strings.TrimSpace(state.Env.ValueString()), state.Name.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand volume", err.Error())
		return
	}
}

func (r *volumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func modelFromVolumeResponse(env types.String, vol *volumeResponse) volumeModel {
	out := volumeModel{
		ID:         types.StringValue(vol.Name),
		Name:       types.StringValue(vol.Name),
		Env:        env,
		Driver:     types.StringValue(vol.Driver),
		Mountpoint: types.StringNull(),
		Scope:      types.StringNull(),
		CreatedAt:  types.StringNull(),
	}
	if vol.Mountpoint != nil {
		out.Mountpoint = types.StringValue(*vol.Mountpoint)
	} else {
		out.Mountpoint = types.StringNull()
	}
	if vol.Scope != nil {
		out.Scope = types.StringValue(*vol.Scope)
	} else {
		out.Scope = types.StringNull()
	}
	if vol.CreatedAt != nil {
		out.CreatedAt = types.StringValue(*vol.CreatedAt)
	} else {
		out.CreatedAt = types.StringNull()
	}
	driverOpts := map[string]string{}
	for k, v := range vol.Options {
		switch val := v.(type) {
		case string:
			driverOpts[k] = val
		}
	}
	if len(driverOpts) > 0 {
		mv, diags := types.MapValueFrom(context.Background(), types.StringType, driverOpts)
		_ = diags
		out.DriverOptions = mv
	} else {
		out.DriverOptions = types.MapNull(types.StringType)
	}
	if vol.Labels != nil {
		mv, diags := types.MapValueFrom(context.Background(), types.StringType, vol.Labels)
		_ = diags
		out.Labels = mv
	} else {
		out.Labels = types.MapNull(types.StringType)
	}
	return out
}

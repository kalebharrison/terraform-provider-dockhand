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
	_ resource.Resource                = (*networkResource)(nil)
	_ resource.ResourceWithConfigure   = (*networkResource)(nil)
	_ resource.ResourceWithImportState = (*networkResource)(nil)
)

func NewNetworkResource() resource.Resource {
	return &networkResource{}
}

type networkResource struct {
	client *Client
}

type networkModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Driver     types.String `tfsdk:"driver"`
	Env        types.String `tfsdk:"env"`
	Internal   types.Bool   `tfsdk:"internal"`
	Attachable types.Bool   `tfsdk:"attachable"`
	Options    types.Map    `tfsdk:"options"`
	Labels     types.Map    `tfsdk:"labels"`
	Scope      types.String `tfsdk:"scope"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

func (r *networkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand network via `/api/networks`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Network ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Network name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"driver": schema.StringAttribute{
				MarkdownDescription: "Network driver. Defaults to `bridge`.",
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
			"internal": schema.BoolAttribute{
				MarkdownDescription: "Whether the network is internal.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolRequiresReplace{},
				},
			},
			"attachable": schema.BoolAttribute{
				MarkdownDescription: "Whether the network is attachable.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolRequiresReplace{},
				},
			},
			"options": schema.MapAttribute{
				MarkdownDescription: "Driver options map used when creating the network.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"labels": schema.MapAttribute{
				MarkdownDescription: "Network labels returned by Dockhand inspect.",
				Computed:            true,
				ElementType:         types.StringType,
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

func (r *networkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan networkModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(plan.Name.ValueString())
	if name == "" {
		resp.Diagnostics.AddError("Missing network name", "`name` must be set.")
		return
	}

	driver := strings.TrimSpace(plan.Driver.ValueString())
	if driver == "" {
		driver = "bridge"
	}
	internal := false
	if !plan.Internal.IsNull() && !plan.Internal.IsUnknown() {
		internal = plan.Internal.ValueBool()
	}
	attachable := true
	if !plan.Attachable.IsNull() && !plan.Attachable.IsUnknown() {
		attachable = plan.Attachable.ValueBool()
	}
	options := map[string]string{}
	if !plan.Options.IsNull() && !plan.Options.IsUnknown() {
		resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &options, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	created, _, err := r.client.CreateNetwork(ctx, strings.TrimSpace(plan.Env.ValueString()), networkPayload{
		Name:       name,
		Driver:     driver,
		Internal:   internal,
		Attachable: attachable,
		Options:    options,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand network", err.Error())
		return
	}

	state := networkModel{
		ID:         types.StringValue(created.ID),
		Name:       types.StringValue(created.Name),
		Driver:     types.StringValue(created.Driver),
		Env:        plan.Env,
		Internal:   types.BoolValue(created.Internal),
		Attachable: types.BoolValue(created.Attachable),
		Scope:      types.StringNull(),
		CreatedAt:  types.StringNull(),
		Options:    types.MapNull(types.StringType),
		Labels:     types.MapNull(types.StringType),
	}
	inspected, _, inspectErr := r.client.GetNetworkInspect(ctx, strings.TrimSpace(plan.Env.ValueString()), state.ID.ValueString())
	if inspectErr == nil && inspected != nil {
		applyNetworkInspectToState(ctx, &state, inspected, nil)
	} else {
		if created.Scope != nil {
			state.Scope = types.StringValue(*created.Scope)
		}
		if created.CreatedAt != nil {
			state.CreatedAt = types.StringValue(*created.CreatedAt)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	networks, _, err := r.client.ListNetworks(ctx, strings.TrimSpace(state.Env.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand network", err.Error())
		return
	}

	var found *networkResponse
	for i := range networks {
		if networks[i].ID == state.ID.ValueString() {
			found = &networks[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	inspected, status, err := r.client.GetNetworkInspect(ctx, strings.TrimSpace(state.Env.ValueString()), state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error reading Dockhand network inspect", err.Error())
		return
	}

	state.ID = types.StringValue(found.ID)
	state.Name = types.StringValue(found.Name)
	state.Driver = types.StringValue(found.Driver)
	state.Internal = types.BoolValue(found.Internal)
	state.Attachable = types.BoolValue(found.Attachable)
	state.Options = types.MapNull(types.StringType)
	state.Labels = types.MapNull(types.StringType)
	if inspected != nil {
		applyNetworkInspectToState(ctx, &state, inspected, resp)
	} else {
		if found.Scope != nil {
			state.Scope = types.StringValue(*found.Scope)
		} else {
			state.Scope = types.StringNull()
		}
		if found.CreatedAt != nil {
			state.CreatedAt = types.StringValue(*found.CreatedAt)
		} else {
			state.CreatedAt = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable attributes require replacement.
	var state networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state networkModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteNetwork(ctx, strings.TrimSpace(state.Env.ValueString()), state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand network", err.Error())
		return
	}
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func applyNetworkInspectToState(ctx context.Context, state *networkModel, inspected *networkInspectResponse, resp *resource.ReadResponse) {
	state.Name = types.StringValue(inspected.Name)
	state.Driver = types.StringValue(inspected.Driver)
	state.Internal = types.BoolValue(inspected.Internal)
	state.Attachable = types.BoolValue(inspected.Attachable)
	if inspected.Scope != nil {
		state.Scope = types.StringValue(*inspected.Scope)
	} else {
		state.Scope = types.StringNull()
	}
	if inspected.CreatedAt != nil {
		state.CreatedAt = types.StringValue(*inspected.CreatedAt)
	} else {
		state.CreatedAt = types.StringNull()
	}
	if inspected.Options != nil {
		mv, diags := types.MapValueFrom(ctx, types.StringType, inspected.Options)
		if resp != nil {
			resp.Diagnostics.Append(diags...)
		}
		state.Options = mv
	} else {
		state.Options = types.MapNull(types.StringType)
	}
	if inspected.Labels != nil {
		mv, diags := types.MapValueFrom(ctx, types.StringType, inspected.Labels)
		if resp != nil {
			resp.Diagnostics.Append(diags...)
		}
		state.Labels = mv
	} else {
		state.Labels = types.MapNull(types.StringType)
	}
}

type boolRequiresReplace struct{}

func (boolRequiresReplace) Description(context.Context) string {
	return "Changing this value requires replacing the resource."
}

func (boolRequiresReplace) MarkdownDescription(context.Context) string {
	return "Changing this value requires replacing the resource."
}

func (boolRequiresReplace) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsUnknown() || req.StateValue.IsUnknown() {
		return
	}
	if req.StateValue.Equal(req.PlanValue) {
		return
	}
	resp.RequiresReplace = true
}

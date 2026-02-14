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
	_ resource.Resource                = (*networkConnectionActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*networkConnectionActionResource)(nil)
	_ resource.ResourceWithImportState = (*networkConnectionActionResource)(nil)
)

func NewNetworkConnectionActionResource() resource.Resource {
	return &networkConnectionActionResource{}
}

type networkConnectionActionResource struct {
	client *Client
}

type networkConnectionActionModel struct {
	ID          types.String `tfsdk:"id"`
	Env         types.String `tfsdk:"env"`
	NetworkID   types.String `tfsdk:"network_id"`
	ContainerID types.String `tfsdk:"container_id"`
	Action      types.String `tfsdk:"action"`
	Trigger     types.String `tfsdk:"trigger"`
}

func (r *networkConnectionActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_connection_action"
}

func (r *networkConnectionActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot network connect/disconnect action for a container. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"env": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"container_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Supported values: `connect`, `disconnect`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"trigger": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *networkConnectionActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *networkConnectionActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan networkConnectionActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	action := strings.ToLower(strings.TrimSpace(plan.Action.ValueString()))
	env := plan.Env.ValueString()
	networkID := strings.TrimSpace(plan.NetworkID.ValueString())
	containerID := strings.TrimSpace(plan.ContainerID.ValueString())

	if networkID == "" {
		resp.Diagnostics.AddError("Invalid network ID", "`network_id` cannot be empty.")
		return
	}
	if containerID == "" {
		resp.Diagnostics.AddError("Invalid container ID", "`container_id` cannot be empty.")
		return
	}

	var (
		status int
		err    error
	)
	switch action {
	case "connect":
		status, err = r.client.ConnectNetwork(ctx, env, networkID, containerID)
	case "disconnect":
		status, err = r.client.DisconnectNetwork(ctx, env, networkID, containerID)
	default:
		resp.Diagnostics.AddError("Invalid action", "Supported actions: connect, disconnect.")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error running Dockhand network connection action", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error running Dockhand network connection action", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s:%s:%s", env, networkID, containerID, action, plan.Trigger.ValueString()))
	plan.Action = types.StringValue(action)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkConnectionActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkConnectionActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *networkConnectionActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan networkConnectionActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *networkConnectionActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *networkConnectionActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<env>:<network_id>:<container_id>:<action>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

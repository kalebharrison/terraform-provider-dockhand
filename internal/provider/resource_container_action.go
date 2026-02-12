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
	_ resource.Resource                = (*containerActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*containerActionResource)(nil)
	_ resource.ResourceWithImportState = (*containerActionResource)(nil)
)

func NewContainerActionResource() resource.Resource {
	return &containerActionResource{}
}

type containerActionResource struct {
	client *Client
}

type containerActionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Env         types.String `tfsdk:"env"`
	ContainerID types.String `tfsdk:"container_id"`
	Action      types.String `tfsdk:"action"`
	Trigger     types.String `tfsdk:"trigger"`
}

func (r *containerActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_action"
}

func (r *containerActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot container action (`start`, `stop`, or `restart`). Change `trigger` to run it again.",
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
			"container_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"action": schema.StringAttribute{
				Required: true,
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

func (r *containerActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *containerActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan containerActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	action := strings.ToLower(strings.TrimSpace(plan.Action.ValueString()))
	env := plan.Env.ValueString()
	id := plan.ContainerID.ValueString()

	var (
		status int
		err    error
	)
	switch action {
	case "start":
		status, err = r.client.StartContainer(ctx, env, id)
	case "stop":
		status, err = r.client.StopContainer(ctx, env, id)
	case "restart":
		status, err = r.client.RestartContainer(ctx, env, id)
	default:
		resp.Diagnostics.AddError("Invalid action", "Supported actions: start, stop, restart.")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error running Dockhand container action", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error running Dockhand container action", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	trigger := plan.Trigger.ValueString()
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s:%s", env, id, action, trigger))
	plan.Action = types.StringValue(action)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// One-shot action resource; state existence is enough.
	var state containerActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *containerActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan containerActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No-op: action already executed at create time.
}

func (r *containerActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<env>:<container_id>:<action>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

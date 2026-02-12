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
	_ resource.Resource                = (*stackActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*stackActionResource)(nil)
	_ resource.ResourceWithImportState = (*stackActionResource)(nil)
)

func NewStackActionResource() resource.Resource {
	return &stackActionResource{}
}

type stackActionResource struct {
	client *Client
}

type stackActionResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Env       types.String `tfsdk:"env"`
	StackName types.String `tfsdk:"stack_name"`
	Action    types.String `tfsdk:"action"`
	Trigger   types.String `tfsdk:"trigger"`
}

func (r *stackActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_action"
}

func (r *stackActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot stack action (`start`, `stop`, or `restart`). Change `trigger` to run it again.",
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
			"stack_name": schema.StringAttribute{
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

func (r *stackActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *stackActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan stackActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	action := strings.ToLower(strings.TrimSpace(plan.Action.ValueString()))
	env := plan.Env.ValueString()
	name := plan.StackName.ValueString()

	var err error
	switch action {
	case "start":
		_, err = r.client.StartStackWithStatus(ctx, env, name)
	case "stop":
		_, err = r.client.StopStackWithStatus(ctx, env, name)
	case "restart":
		if _, err = r.client.StopStackWithStatus(ctx, env, name); err == nil {
			_, err = r.client.StartStackWithStatus(ctx, env, name)
		}
	default:
		resp.Diagnostics.AddError("Invalid action", "Supported actions: start, stop, restart.")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error running Dockhand stack action", err.Error())
		return
	}

	trigger := plan.Trigger.ValueString()
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s:%s", env, name, action, trigger))
	plan.Action = types.StringValue(action)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// One-shot action resource; state existence is enough.
	var state stackActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *stackActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stackActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No-op: action already executed at create time.
}

func (r *stackActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<env>:<stack_name>:<action>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

package provider

import (
	"context"
	"encoding/json"
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
	_ resource.Resource                = (*containerUpdateActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*containerUpdateActionResource)(nil)
	_ resource.ResourceWithImportState = (*containerUpdateActionResource)(nil)
)

func NewContainerUpdateActionResource() resource.Resource {
	return &containerUpdateActionResource{}
}

type containerUpdateActionResource struct {
	client *Client
}

type containerUpdateActionModel struct {
	ID          types.String `tfsdk:"id"`
	Env         types.String `tfsdk:"env"`
	ContainerID types.String `tfsdk:"container_id"`
	PayloadJSON types.String `tfsdk:"payload_json"`
	Trigger     types.String `tfsdk:"trigger"`
	ResultJSON  types.String `tfsdk:"result_json"`
}

func (r *containerUpdateActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_update_action"
}

func (r *containerUpdateActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot container update via `/api/containers/{id}/update` with a raw JSON payload object. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"env": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"container_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"payload_json": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"trigger": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"result_json": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *containerUpdateActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *containerUpdateActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan containerUpdateActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	containerID := strings.TrimSpace(plan.ContainerID.ValueString())
	if containerID == "" {
		resp.Diagnostics.AddError("Invalid container ID", "`container_id` cannot be empty.")
		return
	}

	payloadRaw := strings.TrimSpace(plan.PayloadJSON.ValueString())
	if payloadRaw == "" {
		payloadRaw = "{}"
	}
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(payloadRaw), &payload); err != nil {
		resp.Diagnostics.AddError("Invalid payload JSON", fmt.Sprintf("`payload_json` must be a valid JSON object: %s", err))
		return
	}

	result, status, err := r.client.UpdateContainer(ctx, plan.Env.ValueString(), containerID, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand container", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error updating Dockhand container", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", plan.Env.ValueString(), containerID, plan.Trigger.ValueString()))
	plan.PayloadJSON = types.StringValue(payloadRaw)
	plan.ResultJSON = types.StringValue(mustJSON(result))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerUpdateActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state containerUpdateActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *containerUpdateActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan containerUpdateActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerUpdateActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *containerUpdateActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<env>:<container_id>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

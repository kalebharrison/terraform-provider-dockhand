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
	_ resource.Resource                = (*scheduleRunActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*scheduleRunActionResource)(nil)
	_ resource.ResourceWithImportState = (*scheduleRunActionResource)(nil)
)

func NewScheduleRunActionResource() resource.Resource {
	return &scheduleRunActionResource{}
}

type scheduleRunActionResource struct {
	client *Client
}

type scheduleRunActionModel struct {
	ID         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	ScheduleID types.String `tfsdk:"schedule_id"`
	Trigger    types.String `tfsdk:"trigger"`
}

func (r *scheduleRunActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule_run_action"
}

func (r *scheduleRunActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot schedule execution via `/api/schedules/{type}/{id}/run`. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"type": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"schedule_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"trigger": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *scheduleRunActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *scheduleRunActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan scheduleRunActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleType := strings.TrimSpace(plan.Type.ValueString())
	scheduleID := strings.TrimSpace(plan.ScheduleID.ValueString())
	if scheduleType == "" || scheduleID == "" {
		resp.Diagnostics.AddError("Invalid schedule reference", "`type` and `schedule_id` must be non-empty.")
		return
	}

	status, err := r.client.RunSchedule(ctx, scheduleType, scheduleID)
	if err != nil {
		resp.Diagnostics.AddError("Error running Dockhand schedule", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error running Dockhand schedule", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", scheduleType, scheduleID, plan.Trigger.ValueString()))
	plan.Type = types.StringValue(scheduleType)
	plan.ScheduleID = types.StringValue(scheduleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scheduleRunActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scheduleRunActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *scheduleRunActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan scheduleRunActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scheduleRunActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No-op one-shot action.
}

func (r *scheduleRunActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<type>:<schedule_id>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

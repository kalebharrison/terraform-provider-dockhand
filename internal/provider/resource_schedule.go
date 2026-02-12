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
	_ resource.Resource                = (*scheduleResource)(nil)
	_ resource.ResourceWithConfigure   = (*scheduleResource)(nil)
	_ resource.ResourceWithImportState = (*scheduleResource)(nil)
)

func NewScheduleResource() resource.Resource {
	return &scheduleResource{}
}

type scheduleResource struct {
	client *Client
}

type scheduleModel struct {
	ID         types.String `tfsdk:"id"`
	ScheduleID types.String `tfsdk:"schedule_id"`
	Type       types.String `tfsdk:"type"`
	Name       types.String `tfsdk:"name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	IsSystem   types.Bool   `tfsdk:"is_system"`
	NextRun    types.String `tfsdk:"next_run"`
}

func (r *scheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (r *scheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages pause/resume state of an existing Dockhand schedule using `/api/schedules/*/toggle`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Synthetic Terraform ID: `<type>:<schedule_id>`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule_id": schema.StringAttribute{
				MarkdownDescription: "Dockhand schedule ID to manage.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Schedule type (for example `system_cleanup`, `container_update`, `git_stack_sync`, `env_update_check`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Desired enabled state (true = active, false = paused).",
				Required:            true,
			},
			"is_system": schema.BoolAttribute{
				MarkdownDescription: "Whether this is a system schedule. If true, resource uses `/api/schedules/system/{id}/toggle`.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Schedule name reported by Dockhand.",
				Computed:            true,
			},
			"next_run": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *scheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *scheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan scheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sched, err := r.resolveSchedule(ctx, plan.Type.ValueString(), plan.ScheduleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand schedule", err.Error())
		return
	}
	if sched == nil {
		resp.Diagnostics.AddError("Schedule not found", "No matching schedule was found for the provided `type` and `schedule_id`.")
		return
	}

	desired := plan.Enabled.ValueBool()
	if sched.Enabled != desired {
		if _, err := r.client.ToggleSchedule(ctx, sched.Type, strconv.FormatInt(sched.ID, 10), sched.IsSystem); err != nil {
			resp.Diagnostics.AddError("Error toggling Dockhand schedule", err.Error())
			return
		}
		sched, err = r.resolveSchedule(ctx, plan.Type.ValueString(), plan.ScheduleID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading Dockhand schedule", err.Error())
			return
		}
		if sched == nil {
			resp.Diagnostics.AddError("Schedule not found after toggle", "Schedule disappeared after applying toggle.")
			return
		}
	}

	state := modelFromScheduleResponse(plan, sched)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state scheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sched, err := r.resolveSchedule(ctx, state.Type.ValueString(), state.ScheduleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand schedule", err.Error())
		return
	}
	if sched == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := modelFromScheduleResponse(state, sched)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *scheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan scheduleModel
	var state scheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sched, err := r.resolveSchedule(ctx, plan.Type.ValueString(), plan.ScheduleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand schedule", err.Error())
		return
	}
	if sched == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	desired := plan.Enabled.ValueBool()
	if sched.Enabled != desired {
		if _, err := r.client.ToggleSchedule(ctx, sched.Type, strconv.FormatInt(sched.ID, 10), sched.IsSystem); err != nil {
			resp.Diagnostics.AddError("Error toggling Dockhand schedule", err.Error())
			return
		}
		sched, err = r.resolveSchedule(ctx, plan.Type.ValueString(), plan.ScheduleID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading Dockhand schedule", err.Error())
			return
		}
		if sched == nil {
			resp.State.RemoveResource(ctx)
			return
		}
	}

	newState := modelFromScheduleResponse(plan, sched)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *scheduleResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Intentionally no-op: destroying this Terraform resource does not delete the underlying schedule.
}

func (r *scheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Use `<type>:<schedule_id>` for import.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schedule_id"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *scheduleResource) resolveSchedule(ctx context.Context, scheduleType string, scheduleID string) (*scheduleResponse, error) {
	schedules, _, err := r.client.GetSchedules(ctx)
	if err != nil {
		return nil, err
	}
	id := strings.TrimSpace(scheduleID)
	typ := strings.TrimSpace(scheduleType)
	for i := range schedules.Schedules {
		s := schedules.Schedules[i]
		if s.Type == typ && strconv.FormatInt(s.ID, 10) == id {
			return &s, nil
		}
	}
	return nil, nil
}

func modelFromScheduleResponse(prior scheduleModel, sched *scheduleResponse) scheduleModel {
	out := scheduleModel{
		ID:         types.StringValue(fmt.Sprintf("%s:%d", sched.Type, sched.ID)),
		ScheduleID: types.StringValue(strconv.FormatInt(sched.ID, 10)),
		Type:       types.StringValue(sched.Type),
		Name:       types.StringValue(sched.Name),
		Enabled:    types.BoolValue(sched.Enabled),
		IsSystem:   types.BoolValue(sched.IsSystem),
	}
	if sched.NextRun != nil {
		out.NextRun = types.StringValue(*sched.NextRun)
	} else {
		out.NextRun = types.StringNull()
	}

	// Keep prior desired state on refresh when API and desired diverge during planning windows.
	if !prior.Enabled.IsNull() && !prior.Enabled.IsUnknown() {
		out.Enabled = types.BoolValue(sched.Enabled)
	}
	return out
}

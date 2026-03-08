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
	_ resource.Resource                = (*scheduleSettingsResource)(nil)
	_ resource.ResourceWithConfigure   = (*scheduleSettingsResource)(nil)
	_ resource.ResourceWithImportState = (*scheduleSettingsResource)(nil)
)

func NewScheduleSettingsResource() resource.Resource {
	return &scheduleSettingsResource{}
}

type scheduleSettingsResource struct {
	client *Client
}

type scheduleSettingsResourceModel struct {
	ID             types.String `tfsdk:"id"`
	HideSystemJobs types.Bool   `tfsdk:"hide_system_jobs"`
}

func (r *scheduleSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule_settings"
}

func (r *scheduleSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages global schedule UI/settings via `GET/PUT /api/schedules/settings`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Singleton resource ID. Use `schedule-settings`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hide_system_jobs": schema.BoolAttribute{
				MarkdownDescription: "Hide system jobs in schedule views.",
				Required:            true,
			},
		},
	}
}

func (r *scheduleSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *scheduleSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan scheduleSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.HideSystemJobs.IsNull() || plan.HideSystemJobs.IsUnknown() {
		resp.Diagnostics.AddError("Invalid schedule settings", "`hide_system_jobs` must be set.")
		return
	}

	out, _, err := r.client.UpdateScheduleSettings(ctx, scheduleSettingsPayload{
		HideSystemJobs: plan.HideSystemJobs.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand schedule settings", err.Error())
		return
	}

	plan.ID = types.StringValue("schedule-settings")
	plan.HideSystemJobs = types.BoolValue(out.HideSystemJobs)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scheduleSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state scheduleSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, status, err := r.client.GetScheduleSettings(ctx)
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand schedule settings", err.Error())
		return
	}

	state.ID = types.StringValue("schedule-settings")
	state.HideSystemJobs = types.BoolValue(out.HideSystemJobs)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *scheduleSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan scheduleSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.HideSystemJobs.IsNull() || plan.HideSystemJobs.IsUnknown() {
		resp.Diagnostics.AddError("Invalid schedule settings", "`hide_system_jobs` must be set.")
		return
	}

	out, _, err := r.client.UpdateScheduleSettings(ctx, scheduleSettingsPayload{
		HideSystemJobs: plan.HideSystemJobs.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand schedule settings", err.Error())
		return
	}

	plan.ID = types.StringValue("schedule-settings")
	plan.HideSystemJobs = types.BoolValue(out.HideSystemJobs)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *scheduleSettingsResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No delete endpoint exists for schedule settings. Removing from state is sufficient.
}

func (r *scheduleSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		raw = "schedule-settings"
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

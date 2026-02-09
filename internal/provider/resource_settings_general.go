package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*generalSettingsResource)(nil)
	_ resource.ResourceWithConfigure   = (*generalSettingsResource)(nil)
	_ resource.ResourceWithImportState = (*generalSettingsResource)(nil)
)

func NewGeneralSettingsResource() resource.Resource {
	return &generalSettingsResource{}
}

type generalSettingsResource struct {
	client *Client
}

type generalSettingsModel struct {
	ID                        types.String `tfsdk:"id"`
	ConfirmDestructive        types.Bool   `tfsdk:"confirm_destructive"`
	TimeFormat                types.String `tfsdk:"time_format"`
	DateFormat                types.String `tfsdk:"date_format"`
	ShowStoppedContainers     types.Bool   `tfsdk:"show_stopped_containers"`
	HighlightUpdates          types.Bool   `tfsdk:"highlight_updates"`
	DefaultTimezone           types.String `tfsdk:"default_timezone"`
	DownloadFormat            types.String `tfsdk:"download_format"`
	DefaultGrypeArgs          types.String `tfsdk:"default_grype_args"`
	DefaultTrivyArgs          types.String `tfsdk:"default_trivy_args"`
	ScheduleRetentionDays     types.Int64  `tfsdk:"schedule_retention_days"`
	ScheduleCleanupEnabled    types.Bool   `tfsdk:"schedule_cleanup_enabled"`
	ScheduleCleanupCron       types.String `tfsdk:"schedule_cleanup_cron"`
	EventRetentionDays        types.Int64  `tfsdk:"event_retention_days"`
	EventCleanupEnabled       types.Bool   `tfsdk:"event_cleanup_enabled"`
	EventCleanupCron          types.String `tfsdk:"event_cleanup_cron"`
	EventCollectionMode       types.String `tfsdk:"event_collection_mode"`
	EventPollInterval         types.Int64  `tfsdk:"event_poll_interval"`
	MetricsCollectionInterval types.Int64  `tfsdk:"metrics_collection_interval"`
	LogBufferSizeKb           types.Int64  `tfsdk:"log_buffer_size_kb"`
	Font                      types.String `tfsdk:"font"`
	FontSize                  types.String `tfsdk:"font_size"`
	GridFontSize              types.String `tfsdk:"grid_font_size"`
	EditorFont                types.String `tfsdk:"editor_font"`
	TerminalFont              types.String `tfsdk:"terminal_font"`
	LightTheme                types.String `tfsdk:"light_theme"`
	DarkTheme                 types.String `tfsdk:"dark_theme"`
	PrimaryStackLocation      types.String `tfsdk:"primary_stack_location"`
	ExternalStackPaths        types.List   `tfsdk:"external_stack_paths"`
}

func (r *generalSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_settings_general"
}

func (r *generalSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Dockhand general UI/settings via `GET/POST /api/settings/general`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Singleton ID. Always `general`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"confirm_destructive": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"time_format": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"date_format": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"show_stopped_containers": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"highlight_updates": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"default_timezone": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"download_format": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"default_grype_args": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"default_trivy_args": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"schedule_retention_days": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"schedule_cleanup_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"schedule_cleanup_cron": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"event_retention_days": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"event_cleanup_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"event_cleanup_cron": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"event_collection_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"event_poll_interval": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"metrics_collection_interval": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"log_buffer_size_kb": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"font": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"font_size": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"grid_font_size": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"editor_font": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"terminal_font": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"light_theme": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"dark_theme": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"primary_stack_location": schema.StringAttribute{
				MarkdownDescription: "Optional primary stack location. Set to `null` to clear.",
				Optional:            true,
				Computed:            true,
			},
			"external_stack_paths": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *generalSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *generalSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan generalSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, diags := r.applyPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *generalSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan generalSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, diags := r.applyPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *generalSettingsResource) applyPlan(ctx context.Context, plan generalSettingsModel) (generalSettingsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if r.client == nil {
		diags.AddError("Unconfigured client", "The provider client was not configured.")
		return generalSettingsModel{}, diags
	}

	current, _, err := r.client.GetGeneralSettings(ctx)
	if err != nil {
		diags.AddError("Error reading Dockhand general settings", err.Error())
		return generalSettingsModel{}, diags
	}

	payload, mergeDiags := mergeGeneralSettings(ctx, plan, current)
	diags.Append(mergeDiags...)
	if diags.HasError() {
		return generalSettingsModel{}, diags
	}

	updated, _, err := r.client.UpdateGeneralSettings(ctx, payload)
	if err != nil {
		diags.AddError("Error updating Dockhand general settings", err.Error())
		return generalSettingsModel{}, diags
	}

	state := modelFromGeneralSettings(ctx, updated)
	state.ID = types.StringValue("general")
	return state, diags
}

func (r *generalSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	current, status, err := r.client.GetGeneralSettings(ctx)
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand general settings", err.Error())
		return
	}

	state := modelFromGeneralSettings(ctx, current)
	state.ID = types.StringValue("general")
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *generalSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Dockhand does not expose a delete/reset endpoint for settings. Delete is a no-op.
}

func (r *generalSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Accept any ID, but normalize to the singleton value.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "general")...)
}

func mergeGeneralSettings(ctx context.Context, plan generalSettingsModel, current *generalSettings) (generalSettings, diag.Diagnostics) {
	var diags diag.Diagnostics

	if current == nil {
		diags.AddError("Missing current settings", "Dockhand returned empty settings response.")
		return generalSettings{}, diags
	}

	payload := *current

	// Helpers: treat Unknown as "keep current"; treat Null as "keep current" for scalars, except where explicitly documented.
	if !plan.ConfirmDestructive.IsUnknown() && !plan.ConfirmDestructive.IsNull() {
		payload.ConfirmDestructive = plan.ConfirmDestructive.ValueBool()
	}
	if !plan.TimeFormat.IsUnknown() && !plan.TimeFormat.IsNull() {
		payload.TimeFormat = plan.TimeFormat.ValueString()
	}
	if !plan.DateFormat.IsUnknown() && !plan.DateFormat.IsNull() {
		payload.DateFormat = plan.DateFormat.ValueString()
	}
	if !plan.ShowStoppedContainers.IsUnknown() && !plan.ShowStoppedContainers.IsNull() {
		payload.ShowStoppedContainers = plan.ShowStoppedContainers.ValueBool()
	}
	if !plan.HighlightUpdates.IsUnknown() && !plan.HighlightUpdates.IsNull() {
		payload.HighlightUpdates = plan.HighlightUpdates.ValueBool()
	}
	if !plan.DefaultTimezone.IsUnknown() && !plan.DefaultTimezone.IsNull() {
		payload.DefaultTimezone = plan.DefaultTimezone.ValueString()
	}
	if !plan.DownloadFormat.IsUnknown() && !plan.DownloadFormat.IsNull() {
		payload.DownloadFormat = plan.DownloadFormat.ValueString()
	}
	if !plan.DefaultGrypeArgs.IsUnknown() && !plan.DefaultGrypeArgs.IsNull() {
		payload.DefaultGrypeArgs = plan.DefaultGrypeArgs.ValueString()
	}
	if !plan.DefaultTrivyArgs.IsUnknown() && !plan.DefaultTrivyArgs.IsNull() {
		payload.DefaultTrivyArgs = plan.DefaultTrivyArgs.ValueString()
	}
	if !plan.ScheduleRetentionDays.IsUnknown() && !plan.ScheduleRetentionDays.IsNull() {
		payload.ScheduleRetentionDays = plan.ScheduleRetentionDays.ValueInt64()
	}
	if !plan.ScheduleCleanupEnabled.IsUnknown() && !plan.ScheduleCleanupEnabled.IsNull() {
		payload.ScheduleCleanupEnabled = plan.ScheduleCleanupEnabled.ValueBool()
	}
	if !plan.ScheduleCleanupCron.IsUnknown() && !plan.ScheduleCleanupCron.IsNull() {
		payload.ScheduleCleanupCron = plan.ScheduleCleanupCron.ValueString()
	}
	if !plan.EventRetentionDays.IsUnknown() && !plan.EventRetentionDays.IsNull() {
		payload.EventRetentionDays = plan.EventRetentionDays.ValueInt64()
	}
	if !plan.EventCleanupEnabled.IsUnknown() && !plan.EventCleanupEnabled.IsNull() {
		payload.EventCleanupEnabled = plan.EventCleanupEnabled.ValueBool()
	}
	if !plan.EventCleanupCron.IsUnknown() && !plan.EventCleanupCron.IsNull() {
		payload.EventCleanupCron = plan.EventCleanupCron.ValueString()
	}
	if !plan.EventCollectionMode.IsUnknown() && !plan.EventCollectionMode.IsNull() {
		payload.EventCollectionMode = plan.EventCollectionMode.ValueString()
	}
	if !plan.EventPollInterval.IsUnknown() && !plan.EventPollInterval.IsNull() {
		payload.EventPollInterval = plan.EventPollInterval.ValueInt64()
	}
	if !plan.MetricsCollectionInterval.IsUnknown() && !plan.MetricsCollectionInterval.IsNull() {
		payload.MetricsCollectionInterval = plan.MetricsCollectionInterval.ValueInt64()
	}
	if !plan.LogBufferSizeKb.IsUnknown() && !plan.LogBufferSizeKb.IsNull() {
		payload.LogBufferSizeKb = plan.LogBufferSizeKb.ValueInt64()
	}
	if !plan.Font.IsUnknown() && !plan.Font.IsNull() {
		payload.Font = plan.Font.ValueString()
	}
	if !plan.FontSize.IsUnknown() && !plan.FontSize.IsNull() {
		payload.FontSize = plan.FontSize.ValueString()
	}
	if !plan.GridFontSize.IsUnknown() && !plan.GridFontSize.IsNull() {
		payload.GridFontSize = plan.GridFontSize.ValueString()
	}
	if !plan.EditorFont.IsUnknown() && !plan.EditorFont.IsNull() {
		payload.EditorFont = plan.EditorFont.ValueString()
	}
	if !plan.TerminalFont.IsUnknown() && !plan.TerminalFont.IsNull() {
		payload.TerminalFont = plan.TerminalFont.ValueString()
	}
	if !plan.LightTheme.IsUnknown() && !plan.LightTheme.IsNull() {
		payload.LightTheme = plan.LightTheme.ValueString()
	}
	if !plan.DarkTheme.IsUnknown() && !plan.DarkTheme.IsNull() {
		payload.DarkTheme = plan.DarkTheme.ValueString()
	}

	// Nullable: allow explicit null to clear.
	if !plan.PrimaryStackLocation.IsUnknown() {
		if plan.PrimaryStackLocation.IsNull() {
			payload.PrimaryStackLocation = nil
		} else {
			v := plan.PrimaryStackLocation.ValueString()
			payload.PrimaryStackLocation = &v
		}
	}

	if !plan.ExternalStackPaths.IsUnknown() {
		if plan.ExternalStackPaths.IsNull() {
			payload.ExternalStackPaths = []string{}
		} else {
			var paths []string
			diags.Append(plan.ExternalStackPaths.ElementsAs(ctx, &paths, false)...)
			if diags.HasError() {
				return generalSettings{}, diags
			}
			payload.ExternalStackPaths = paths
		}
	}

	return payload, diags
}

func modelFromGeneralSettings(ctx context.Context, s *generalSettings) generalSettingsModel {
	if s == nil {
		return generalSettingsModel{}
	}

	out := generalSettingsModel{
		ConfirmDestructive:        types.BoolValue(s.ConfirmDestructive),
		TimeFormat:                types.StringValue(s.TimeFormat),
		DateFormat:                types.StringValue(s.DateFormat),
		ShowStoppedContainers:     types.BoolValue(s.ShowStoppedContainers),
		HighlightUpdates:          types.BoolValue(s.HighlightUpdates),
		DefaultTimezone:           types.StringValue(s.DefaultTimezone),
		DownloadFormat:            types.StringValue(s.DownloadFormat),
		DefaultGrypeArgs:          types.StringValue(s.DefaultGrypeArgs),
		DefaultTrivyArgs:          types.StringValue(s.DefaultTrivyArgs),
		ScheduleRetentionDays:     types.Int64Value(s.ScheduleRetentionDays),
		ScheduleCleanupEnabled:    types.BoolValue(s.ScheduleCleanupEnabled),
		ScheduleCleanupCron:       types.StringValue(s.ScheduleCleanupCron),
		EventRetentionDays:        types.Int64Value(s.EventRetentionDays),
		EventCleanupEnabled:       types.BoolValue(s.EventCleanupEnabled),
		EventCleanupCron:          types.StringValue(s.EventCleanupCron),
		EventCollectionMode:       types.StringValue(s.EventCollectionMode),
		EventPollInterval:         types.Int64Value(s.EventPollInterval),
		MetricsCollectionInterval: types.Int64Value(s.MetricsCollectionInterval),
		LogBufferSizeKb:           types.Int64Value(s.LogBufferSizeKb),
		Font:                      types.StringValue(s.Font),
		FontSize:                  types.StringValue(s.FontSize),
		GridFontSize:              types.StringValue(s.GridFontSize),
		EditorFont:                types.StringValue(s.EditorFont),
		TerminalFont:              types.StringValue(s.TerminalFont),
		LightTheme:                types.StringValue(s.LightTheme),
		DarkTheme:                 types.StringValue(s.DarkTheme),
	}

	if s.PrimaryStackLocation != nil {
		out.PrimaryStackLocation = types.StringValue(*s.PrimaryStackLocation)
	} else {
		out.PrimaryStackLocation = types.StringNull()
	}

	paths, diags := types.ListValueFrom(ctx, types.StringType, s.ExternalStackPaths)
	if diags.HasError() {
		out.ExternalStackPaths = types.ListNull(types.StringType)
	} else {
		out.ExternalStackPaths = paths
	}

	return out
}

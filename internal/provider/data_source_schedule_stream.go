package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*scheduleStreamDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*scheduleStreamDataSource)(nil)
)

func NewScheduleStreamDataSource() datasource.DataSource {
	return &scheduleStreamDataSource{}
}

type scheduleStreamDataSource struct {
	client *Client
}

type scheduleStreamEventModel struct {
	Event types.String `tfsdk:"event"`
	Data  types.String `tfsdk:"data"`
}

type scheduleStreamDataSourceModel struct {
	ID             types.String               `tfsdk:"id"`
	MaxEvents      types.Int64                `tfsdk:"max_events"`
	TimeoutSeconds types.Int64                `tfsdk:"timeout_seconds"`
	Connected      types.Bool                 `tfsdk:"connected"`
	EventCount     types.Int64                `tfsdk:"event_count"`
	Events         []scheduleStreamEventModel `tfsdk:"events"`
	EventsJSON     types.String               `tfsdk:"events_json"`
	SchedulesJSON  types.String               `tfsdk:"schedules_json"`
}

func (d *scheduleStreamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule_stream"
}

func (d *scheduleStreamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a bounded snapshot from the schedules SSE stream (`GET /api/schedules/stream`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"max_events": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of events to capture from the stream.",
				Optional:            true,
			},
			"timeout_seconds": schema.Int64Attribute{
				MarkdownDescription: "Maximum stream read duration in seconds.",
				Optional:            true,
			},
			"connected": schema.BoolAttribute{
				Computed: true,
			},
			"event_count": schema.Int64Attribute{
				Computed: true,
			},
			"events": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"event": schema.StringAttribute{Computed: true},
						"data":  schema.StringAttribute{Computed: true},
					},
				},
			},
			"events_json": schema.StringAttribute{
				Computed: true,
			},
			"schedules_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *scheduleStreamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *scheduleStreamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data scheduleStreamDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	maxEvents := int64(1)
	if !data.MaxEvents.IsNull() && !data.MaxEvents.IsUnknown() && data.MaxEvents.ValueInt64() > 0 {
		maxEvents = data.MaxEvents.ValueInt64()
	}
	timeoutSeconds := int64(5)
	if !data.TimeoutSeconds.IsNull() && !data.TimeoutSeconds.IsUnknown() && data.TimeoutSeconds.ValueInt64() > 0 {
		timeoutSeconds = data.TimeoutSeconds.ValueInt64()
	}

	events, _, err := d.client.ReadScheduleStream(ctx, maxEvents, time.Duration(timeoutSeconds)*time.Second)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand schedule stream", err.Error())
		return
	}

	connected := false
	schedulesJSON := ""
	items := make([]scheduleStreamEventModel, 0, len(events))
	for _, item := range events {
		eventType := strings.TrimSpace(item.Event)
		eventData := strings.TrimSpace(item.Data)
		if strings.EqualFold(eventType, "connected") {
			connected = true
		}
		if strings.EqualFold(eventType, "schedules") && eventData != "" {
			schedulesJSON = eventData
		}
		items = append(items, scheduleStreamEventModel{
			Event: types.StringValue(eventType),
			Data:  types.StringValue(eventData),
		})
	}

	data.ID = types.StringValue(fmt.Sprintf("schedule-stream:%d:%d", maxEvents, timeoutSeconds))
	data.MaxEvents = types.Int64Value(maxEvents)
	data.TimeoutSeconds = types.Int64Value(timeoutSeconds)
	data.Connected = types.BoolValue(connected)
	data.EventCount = types.Int64Value(int64(len(items)))
	data.Events = items
	data.EventsJSON = types.StringValue(mustJSON(events))
	if schedulesJSON == "" {
		data.SchedulesJSON = types.StringNull()
	} else {
		data.SchedulesJSON = types.StringValue(schedulesJSON)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

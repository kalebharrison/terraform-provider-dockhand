package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*schedulesExecutionsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*schedulesExecutionsDataSource)(nil)
)

func NewSchedulesExecutionsDataSource() datasource.DataSource {
	return &schedulesExecutionsDataSource{}
}

type schedulesExecutionsDataSource struct {
	client *Client
}

type scheduleExecutionItemModel struct {
	ID            types.String `tfsdk:"id"`
	ScheduleType  types.String `tfsdk:"schedule_type"`
	ScheduleID    types.String `tfsdk:"schedule_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	EntityName    types.String `tfsdk:"entity_name"`
	TriggeredBy   types.String `tfsdk:"triggered_by"`
	TriggeredAt   types.String `tfsdk:"triggered_at"`
	StartedAt     types.String `tfsdk:"started_at"`
	CompletedAt   types.String `tfsdk:"completed_at"`
	DurationMs    types.Int64  `tfsdk:"duration_ms"`
	Status        types.String `tfsdk:"status"`
	ErrorMessage  types.String `tfsdk:"error_message"`
	DetailsJSON   types.String `tfsdk:"details_json"`
	CreatedAt     types.String `tfsdk:"created_at"`
	Logs          types.String `tfsdk:"logs"`
}

type schedulesExecutionsDataSourceModel struct {
	ID             types.String                 `tfsdk:"id"`
	Limit          types.Int64                  `tfsdk:"limit"`
	Offset         types.Int64                  `tfsdk:"offset"`
	Total          types.Int64                  `tfsdk:"total"`
	ReturnedLimit  types.Int64                  `tfsdk:"returned_limit"`
	ReturnedOffset types.Int64                  `tfsdk:"returned_offset"`
	Executions     []scheduleExecutionItemModel `tfsdk:"executions"`
}

func (d *schedulesExecutionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedules_executions"
}

func (d *schedulesExecutionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads schedule execution history from `/api/schedules/executions`.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true},
			"limit":           schema.Int64Attribute{Optional: true},
			"offset":          schema.Int64Attribute{Optional: true},
			"total":           schema.Int64Attribute{Computed: true},
			"returned_limit":  schema.Int64Attribute{Computed: true},
			"returned_offset": schema.Int64Attribute{Computed: true},
			"executions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true},
						"schedule_type":  schema.StringAttribute{Computed: true},
						"schedule_id":    schema.StringAttribute{Computed: true},
						"environment_id": schema.StringAttribute{Computed: true},
						"entity_name":    schema.StringAttribute{Computed: true},
						"triggered_by":   schema.StringAttribute{Computed: true},
						"triggered_at":   schema.StringAttribute{Computed: true},
						"started_at":     schema.StringAttribute{Computed: true},
						"completed_at":   schema.StringAttribute{Computed: true},
						"duration_ms":    schema.Int64Attribute{Computed: true},
						"status":         schema.StringAttribute{Computed: true},
						"error_message":  schema.StringAttribute{Computed: true},
						"details_json":   schema.StringAttribute{Computed: true},
						"created_at":     schema.StringAttribute{Computed: true},
						"logs":           schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *schedulesExecutionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *schedulesExecutionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var config schedulesExecutionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	limit := int64(0)
	offset := int64(0)
	if !config.Limit.IsNull() && !config.Limit.IsUnknown() {
		limit = config.Limit.ValueInt64()
	}
	if !config.Offset.IsNull() && !config.Offset.IsUnknown() {
		offset = config.Offset.ValueInt64()
	}
	if limit < 0 {
		resp.Diagnostics.AddError("Invalid limit", "`limit` must be zero or greater.")
		return
	}
	if offset < 0 {
		resp.Diagnostics.AddError("Invalid offset", "`offset` must be zero or greater.")
		return
	}

	apiOut, _, err := d.client.GetScheduleExecutions(ctx, limit, offset)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand schedule executions", err.Error())
		return
	}

	out := schedulesExecutionsDataSourceModel{
		ID:             types.StringValue(fmt.Sprintf("dockhand-schedule-executions:%d:%d", limit, offset)),
		Limit:          types.Int64Value(limit),
		Offset:         types.Int64Value(offset),
		Total:          types.Int64Value(apiOut.Total),
		ReturnedLimit:  types.Int64Value(apiOut.Limit),
		ReturnedOffset: types.Int64Value(apiOut.Offset),
		Executions:     make([]scheduleExecutionItemModel, 0, len(apiOut.Executions)),
	}

	for _, e := range apiOut.Executions {
		item := scheduleExecutionItemModel{
			ID:           types.StringValue(strconv.FormatInt(e.ID, 10)),
			ScheduleType: types.StringValue(e.ScheduleType),
			ScheduleID:   types.StringValue(strconv.FormatInt(e.ScheduleID, 10)),
		}
		if e.EnvironmentID != nil {
			item.EnvironmentID = types.StringValue(strconv.FormatInt(*e.EnvironmentID, 10))
		} else {
			item.EnvironmentID = types.StringNull()
		}
		if e.EntityName != nil {
			item.EntityName = types.StringValue(*e.EntityName)
		} else {
			item.EntityName = types.StringNull()
		}
		if e.TriggeredBy != nil {
			item.TriggeredBy = types.StringValue(*e.TriggeredBy)
		} else {
			item.TriggeredBy = types.StringNull()
		}
		if e.TriggeredAt != nil {
			item.TriggeredAt = types.StringValue(*e.TriggeredAt)
		} else {
			item.TriggeredAt = types.StringNull()
		}
		if e.StartedAt != nil {
			item.StartedAt = types.StringValue(*e.StartedAt)
		} else {
			item.StartedAt = types.StringNull()
		}
		if e.CompletedAt != nil {
			item.CompletedAt = types.StringValue(*e.CompletedAt)
		} else {
			item.CompletedAt = types.StringNull()
		}
		if e.Duration != nil {
			item.DurationMs = types.Int64Value(*e.Duration)
		} else {
			item.DurationMs = types.Int64Null()
		}
		if e.Status != nil {
			item.Status = types.StringValue(*e.Status)
		} else {
			item.Status = types.StringNull()
		}
		if e.ErrorMessage != nil {
			item.ErrorMessage = types.StringValue(*e.ErrorMessage)
		} else {
			item.ErrorMessage = types.StringNull()
		}
		if e.Details != nil {
			item.DetailsJSON = types.StringValue(mustJSON(e.Details))
		} else {
			item.DetailsJSON = types.StringNull()
		}
		if e.CreatedAt != nil {
			item.CreatedAt = types.StringValue(*e.CreatedAt)
		} else {
			item.CreatedAt = types.StringNull()
		}
		if e.Logs != nil {
			item.Logs = types.StringValue(*e.Logs)
		} else {
			item.Logs = types.StringNull()
		}
		out.Executions = append(out.Executions, item)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

func mustJSON(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*schedulesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*schedulesDataSource)(nil)
)

func NewSchedulesDataSource() datasource.DataSource {
	return &schedulesDataSource{}
}

type schedulesDataSource struct {
	client *Client
}

type scheduleItemModel struct {
	ID              types.String `tfsdk:"id"`
	Type            types.String `tfsdk:"type"`
	Name            types.String `tfsdk:"name"`
	EntityName      types.String `tfsdk:"entity_name"`
	Description     types.String `tfsdk:"description"`
	EnvironmentID   types.String `tfsdk:"environment_id"`
	EnvironmentName types.String `tfsdk:"environment_name"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	ScheduleType    types.String `tfsdk:"schedule_type"`
	CronExpression  types.String `tfsdk:"cron_expression"`
	NextRun         types.String `tfsdk:"next_run"`
	IsSystem        types.Bool   `tfsdk:"is_system"`
	LastStatus      types.String `tfsdk:"last_status"`
	LastTriggeredAt types.String `tfsdk:"last_triggered_at"`
	LastCompletedAt types.String `tfsdk:"last_completed_at"`
}

type schedulesDataSourceModel struct {
	ID        types.String        `tfsdk:"id"`
	Schedules []scheduleItemModel `tfsdk:"schedules"`
}

func (d *schedulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedules"
}

func (d *schedulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads schedules from `/api/schedules`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"schedules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"entity_name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"environment_id": schema.StringAttribute{
							Computed: true,
						},
						"environment_name": schema.StringAttribute{
							Computed: true,
						},
						"enabled": schema.BoolAttribute{
							Computed: true,
						},
						"schedule_type": schema.StringAttribute{
							Computed: true,
						},
						"cron_expression": schema.StringAttribute{
							Computed: true,
						},
						"next_run": schema.StringAttribute{
							Computed: true,
						},
						"is_system": schema.BoolAttribute{
							Computed: true,
						},
						"last_status": schema.StringAttribute{
							Computed: true,
						},
						"last_triggered_at": schema.StringAttribute{
							Computed: true,
						},
						"last_completed_at": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *schedulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *schedulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	apiOut, _, err := d.client.GetSchedules(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand schedules", err.Error())
		return
	}

	out := schedulesDataSourceModel{
		ID:        types.StringValue("dockhand-schedules"),
		Schedules: make([]scheduleItemModel, 0, len(apiOut.Schedules)),
	}

	for _, s := range apiOut.Schedules {
		item := scheduleItemModel{
			ID:       types.StringValue(strconv.FormatInt(s.ID, 10)),
			Type:     types.StringValue(s.Type),
			Name:     types.StringValue(s.Name),
			Enabled:  types.BoolValue(s.Enabled),
			IsSystem: types.BoolValue(s.IsSystem),
		}

		if s.EntityName != nil {
			item.EntityName = types.StringValue(*s.EntityName)
		} else {
			item.EntityName = types.StringNull()
		}
		if s.Description != nil {
			item.Description = types.StringValue(*s.Description)
		} else {
			item.Description = types.StringNull()
		}
		if s.EnvironmentID != nil {
			item.EnvironmentID = types.StringValue(strconv.FormatInt(*s.EnvironmentID, 10))
		} else {
			item.EnvironmentID = types.StringNull()
		}
		if s.EnvironmentName != nil {
			item.EnvironmentName = types.StringValue(*s.EnvironmentName)
		} else {
			item.EnvironmentName = types.StringNull()
		}
		if s.ScheduleType != nil {
			item.ScheduleType = types.StringValue(*s.ScheduleType)
		} else {
			item.ScheduleType = types.StringNull()
		}
		if s.CronExpression != nil {
			item.CronExpression = types.StringValue(*s.CronExpression)
		} else {
			item.CronExpression = types.StringNull()
		}
		if s.NextRun != nil {
			item.NextRun = types.StringValue(*s.NextRun)
		} else {
			item.NextRun = types.StringNull()
		}

		if s.LastExecution != nil {
			item.LastStatus = types.StringValue(s.LastExecution.Status)
			if s.LastExecution.TriggeredAt != nil {
				item.LastTriggeredAt = types.StringValue(*s.LastExecution.TriggeredAt)
			} else {
				item.LastTriggeredAt = types.StringNull()
			}
			if s.LastExecution.CompletedAt != nil {
				item.LastCompletedAt = types.StringValue(*s.LastExecution.CompletedAt)
			} else {
				item.LastCompletedAt = types.StringNull()
			}
		} else {
			item.LastStatus = types.StringNull()
			item.LastTriggeredAt = types.StringNull()
			item.LastCompletedAt = types.StringNull()
		}

		out.Schedules = append(out.Schedules, item)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

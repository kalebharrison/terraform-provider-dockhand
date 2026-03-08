package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*scheduleSettingsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*scheduleSettingsDataSource)(nil)
)

func NewScheduleSettingsDataSource() datasource.DataSource {
	return &scheduleSettingsDataSource{}
}

type scheduleSettingsDataSource struct {
	client *Client
}

type scheduleSettingsDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	HideSystemJobs types.Bool   `tfsdk:"hide_system_jobs"`
}

func (d *scheduleSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule_settings"
}

func (d *scheduleSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads global schedule settings via `GET /api/schedules/settings`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"hide_system_jobs": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *scheduleSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *scheduleSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data scheduleSettingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := d.client.GetScheduleSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand schedule settings", err.Error())
		return
	}

	data.ID = types.StringValue("schedule-settings")
	data.HideSystemJobs = types.BoolValue(out.HideSystemJobs)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

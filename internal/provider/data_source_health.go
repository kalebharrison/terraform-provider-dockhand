package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const timeLayout = time.RFC3339

var (
	_ datasource.DataSource              = (*healthDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*healthDataSource)(nil)
)

func NewHealthDataSource() datasource.DataSource {
	return &healthDataSource{}
}

type healthDataSource struct {
	client *Client
}

type healthDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Env       types.String `tfsdk:"env"`
	Status    types.String `tfsdk:"status"`
	Version   types.String `tfsdk:"version"`
	CheckedAt types.String `tfsdk:"checked_at"`
}

func (d *healthDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_health"
}

func (d *healthDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Checks Dockhand API availability using `/api/dashboard/stats`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Static ID for this data source.",
				Computed:            true,
			},
			"env": schema.StringAttribute{
				MarkdownDescription: "Optional Dockhand environment ID for the health check.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current Dockhand status.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Dockhand version, if returned by the API.",
				Computed:            true,
			},
			"checked_at": schema.StringAttribute{
				MarkdownDescription: "Time this health check was performed.",
				Computed:            true,
			},
		},
	}
}

func (d *healthDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *healthDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var config healthDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	health, err := d.client.Health(ctx, config.Env.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error checking Dockhand API health", err.Error())
		return
	}

	state := healthDataSourceModel{
		ID:        types.StringValue("dockhand-health"),
		Status:    types.StringValue(health.Status),
		CheckedAt: types.StringValue(time.Now().UTC().Format(timeLayout)),
	}
	if !config.Env.IsNull() && !config.Env.IsUnknown() && config.Env.ValueString() != "" {
		state.Env = types.StringValue(config.Env.ValueString())
	} else {
		state.Env = types.StringNull()
	}
	if health.Version != "" {
		state.Version = types.StringValue(health.Version)
	} else {
		state.Version = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

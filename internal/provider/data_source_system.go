package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*systemDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*systemDataSource)(nil)
)

func NewSystemDataSource() datasource.DataSource {
	return &systemDataSource{}
}

type systemDataSource struct {
	client *Client
}

type systemDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	SystemJSON   types.String `tfsdk:"system_json"`
	RuntimeJSON  types.String `tfsdk:"runtime_json"`
	DatabaseJSON types.String `tfsdk:"database_json"`
	StatsJSON    types.String `tfsdk:"stats_json"`
	DockerJSON   types.String `tfsdk:"docker_json"`
	HostJSON     types.String `tfsdk:"host_json"`
}

func (d *systemDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system"
}

func (d *systemDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads Dockhand system summary via `GET /api/system`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"system_json": schema.StringAttribute{
				Computed: true,
			},
			"runtime_json": schema.StringAttribute{
				Computed: true,
			},
			"database_json": schema.StringAttribute{
				Computed: true,
			},
			"stats_json": schema.StringAttribute{
				Computed: true,
			},
			"docker_json": schema.StringAttribute{
				Computed: true,
			},
			"host_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *systemDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *systemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data systemDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := d.client.GetSystemInfo(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand system info", err.Error())
		return
	}

	data.ID = types.StringValue("system")
	data.SystemJSON = types.StringValue(mustJSON(out))
	data.RuntimeJSON = jsonAttrOrNull(out, "runtime")
	data.DatabaseJSON = jsonAttrOrNull(out, "database")
	data.StatsJSON = jsonAttrOrNull(out, "stats")
	data.DockerJSON = jsonAttrOrNull(out, "docker")
	data.HostJSON = jsonAttrOrNull(out, "host")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func jsonAttrOrNull(input map[string]any, key string) types.String {
	if input == nil {
		return types.StringNull()
	}
	value, ok := input[key]
	if !ok || value == nil {
		return types.StringNull()
	}
	return types.StringValue(mustJSON(value))
}

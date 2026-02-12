package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*containerLogsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*containerLogsDataSource)(nil)
)

func NewContainerLogsDataSource() datasource.DataSource {
	return &containerLogsDataSource{}
}

type containerLogsDataSource struct {
	client *Client
}

type containerLogsDataSourceModel struct {
	Env         types.String `tfsdk:"env"`
	ContainerID types.String `tfsdk:"container_id"`
	Tail        types.Int64  `tfsdk:"tail"`
	Logs        types.String `tfsdk:"logs"`
}

func (d *containerLogsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_logs"
}

func (d *containerLogsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches logs for a Dockhand container.",
		Attributes: map[string]schema.Attribute{
			"env": schema.StringAttribute{
				Optional: true,
			},
			"container_id": schema.StringAttribute{
				Required: true,
			},
			"tail": schema.Int64Attribute{
				Optional: true,
			},
			"logs": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *containerLogsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *containerLogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data containerLogsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tail := data.Tail.ValueInt64()
	if data.Tail.IsNull() || data.Tail.IsUnknown() || tail <= 0 {
		tail = 100
		data.Tail = types.Int64Value(tail)
	}

	out, _, err := d.client.GetContainerLogs(ctx, data.Env.ValueString(), data.ContainerID.ValueString(), tail)
	if err != nil {
		resp.Diagnostics.AddError("Error reading container logs", err.Error())
		return
	}

	data.Logs = types.StringValue(out.Logs)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

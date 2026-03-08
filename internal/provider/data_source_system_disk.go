package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*systemDiskDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*systemDiskDataSource)(nil)
)

func NewSystemDiskDataSource() datasource.DataSource {
	return &systemDiskDataSource{}
}

type systemDiskDataSource struct {
	client *Client
}

type systemDiskDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Env           types.String `tfsdk:"env"`
	ResponseJSON  types.String `tfsdk:"response_json"`
	DiskUsageJSON types.String `tfsdk:"disk_usage_json"`
}

func (d *systemDiskDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_disk"
}

func (d *systemDiskDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads environment-scoped disk usage summary via `GET /api/system/disk?env=<id>`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"env": schema.StringAttribute{
				Optional: true,
			},
			"response_json": schema.StringAttribute{
				Computed: true,
			},
			"disk_usage_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *systemDiskDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *systemDiskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data systemDiskDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := strings.TrimSpace(data.Env.ValueString())
	out, _, err := d.client.GetSystemDisk(ctx, env)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand system disk usage", err.Error())
		return
	}

	if env == "" {
		env = strings.TrimSpace(d.client.resolveEnv(""))
	}
	if env == "" {
		data.ID = types.StringValue("system-disk")
		data.Env = types.StringNull()
	} else {
		data.ID = types.StringValue("system-disk:" + env)
		data.Env = types.StringValue(env)
	}
	data.ResponseJSON = types.StringValue(mustJSON(out))
	data.DiskUsageJSON = jsonAttrOrNull(out, "diskUsage")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

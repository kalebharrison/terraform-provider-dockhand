package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*containerStatsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*containerStatsDataSource)(nil)
)

func NewContainerStatsDataSource() datasource.DataSource {
	return &containerStatsDataSource{}
}

type containerStatsDataSource struct {
	client *Client
}

type containerStatsItemModel struct {
	ID          types.String  `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	CPUPercent  types.Float64 `tfsdk:"cpu_percent"`
	MemoryUsage types.Int64   `tfsdk:"memory_usage"`
	MemoryRaw   types.Int64   `tfsdk:"memory_raw"`
	MemoryCache types.Int64   `tfsdk:"memory_cache"`
	MemoryLimit types.Int64   `tfsdk:"memory_limit"`
	MemoryPct   types.Float64 `tfsdk:"memory_percent"`
	NetworkRX   types.Int64   `tfsdk:"network_rx"`
	NetworkTX   types.Int64   `tfsdk:"network_tx"`
	BlockRead   types.Int64   `tfsdk:"block_read"`
	BlockWrite  types.Int64   `tfsdk:"block_write"`
}

type containerStatsDataSourceModel struct {
	ID    types.String              `tfsdk:"id"`
	Env   types.String              `tfsdk:"env"`
	Stats []containerStatsItemModel `tfsdk:"stats"`
}

func (d *containerStatsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_stats"
}

func (d *containerStatsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads live container stats from `/api/containers/stats`.",
		Attributes: map[string]schema.Attribute{
			"id":  schema.StringAttribute{Computed: true},
			"env": schema.StringAttribute{Optional: true},
			"stats": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":             schema.StringAttribute{Computed: true},
					"name":           schema.StringAttribute{Computed: true},
					"cpu_percent":    schema.Float64Attribute{Computed: true},
					"memory_usage":   schema.Int64Attribute{Computed: true},
					"memory_raw":     schema.Int64Attribute{Computed: true},
					"memory_cache":   schema.Int64Attribute{Computed: true},
					"memory_limit":   schema.Int64Attribute{Computed: true},
					"memory_percent": schema.Float64Attribute{Computed: true},
					"network_rx":     schema.Int64Attribute{Computed: true},
					"network_tx":     schema.Int64Attribute{Computed: true},
					"block_read":     schema.Int64Attribute{Computed: true},
					"block_write":    schema.Int64Attribute{Computed: true},
				}},
			},
		},
	}
}

func (d *containerStatsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *containerStatsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var config containerStatsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := d.client.GetContainerStats(ctx, config.Env.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading container stats", err.Error())
		return
	}

	items := make([]containerStatsItemModel, 0, len(out))
	for _, s := range out {
		items = append(items, containerStatsItemModel{
			ID:          types.StringValue(s.ID),
			Name:        types.StringValue(s.Name),
			CPUPercent:  types.Float64Value(s.CPUPercent),
			MemoryUsage: types.Int64Value(s.MemoryUsage),
			MemoryRaw:   types.Int64Value(s.MemoryRaw),
			MemoryCache: types.Int64Value(s.MemoryCache),
			MemoryLimit: types.Int64Value(s.MemoryLimit),
			MemoryPct:   types.Float64Value(s.MemoryPct),
			NetworkRX:   types.Int64Value(s.NetworkRX),
			NetworkTX:   types.Int64Value(s.NetworkTX),
			BlockRead:   types.Int64Value(s.BlockRead),
			BlockWrite:  types.Int64Value(s.BlockWrite),
		})
	}

	state := containerStatsDataSourceModel{ID: types.StringValue("container-stats:" + config.Env.ValueString()), Env: config.Env, Stats: items}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*environmentDetectSocketDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*environmentDetectSocketDataSource)(nil)
)

func NewEnvironmentDetectSocketDataSource() datasource.DataSource {
	return &environmentDetectSocketDataSource{}
}

type environmentDetectSocketDataSource struct {
	client *Client
}

type environmentDetectSocketDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	HomeDir     types.String `tfsdk:"home_dir"`
	SocketPaths types.List   `tfsdk:"socket_paths"`
	SocketsJSON types.String `tfsdk:"sockets_json"`
}

func (d *environmentDetectSocketDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_detect_socket"
}

func (d *environmentDetectSocketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads Dockhand local socket discovery output via `GET /api/environments/detect-socket`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"home_dir": schema.StringAttribute{
				Computed: true,
			},
			"socket_paths": schema.ListAttribute{
				MarkdownDescription: "Best-effort extracted socket paths from the `sockets` payload.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"sockets_json": schema.StringAttribute{
				MarkdownDescription: "Raw `sockets` payload encoded as JSON.",
				Computed:            true,
			},
		},
	}
}

func (d *environmentDetectSocketDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *environmentDetectSocketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	out, _, err := d.client.DetectEnvironmentSockets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error detecting Dockhand environment sockets", err.Error())
		return
	}

	data := environmentDetectSocketDataSourceModel{
		ID:          types.StringValue("dockhand-environment-detect-socket"),
		HomeDir:     types.StringValue(strings.TrimSpace(out.HomeDir)),
		SocketsJSON: types.StringValue(mustJSON(out.Sockets)),
	}

	paths := extractSocketPaths(out.Sockets)
	socketPaths, diags := types.ListValueFrom(ctx, types.StringType, paths)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.SocketPaths = socketPaths

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func extractSocketPaths(sockets []any) []string {
	if len(sockets) == 0 {
		return []string{}
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(sockets))
	for _, entry := range sockets {
		value := ""
		switch v := entry.(type) {
		case string:
			value = strings.TrimSpace(v)
		case map[string]any:
			value = strings.TrimSpace(firstString(v, "path", "socketPath", "socket", "value"))
		}
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

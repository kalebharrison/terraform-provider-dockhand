package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*systemFileContentDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*systemFileContentDataSource)(nil)
)

func NewSystemFileContentDataSource() datasource.DataSource {
	return &systemFileContentDataSource{}
}

type systemFileContentDataSource struct {
	client *Client
}

type systemFileContentDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Path       types.String `tfsdk:"path"`
	Content    types.String `tfsdk:"content"`
	Size       types.Int64  `tfsdk:"size"`
	Mtime      types.String `tfsdk:"mtime"`
	ResultJSON types.String `tfsdk:"result_json"`
}

func (d *systemFileContentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_file_content"
}

func (d *systemFileContentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads system file content via `GET /api/system/files/content?path=<file>`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"path": schema.StringAttribute{
				Required: true,
			},
			"content": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"size": schema.Int64Attribute{
				Computed: true,
			},
			"mtime": schema.StringAttribute{
				Computed: true,
			},
			"result_json": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func (d *systemFileContentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *systemFileContentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data systemFileContentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := strings.TrimSpace(data.Path.ValueString())
	if path == "" {
		resp.Diagnostics.AddError("Invalid path", "`path` cannot be empty.")
		return
	}

	out, status, err := d.client.GetSystemFileContent(ctx, path)
	if err != nil {
		if status == http.StatusNotFound {
			resp.Diagnostics.AddError("System file not found", fmt.Sprintf("Dockhand returned 404 for path %q", path))
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand system file content", err.Error())
		return
	}

	data.ID = types.StringValue("system-file-content:" + path)
	if out.Path == nil || strings.TrimSpace(*out.Path) == "" {
		data.Path = types.StringValue(path)
	} else {
		data.Path = types.StringValue(strings.TrimSpace(*out.Path))
	}
	if out.Content == nil {
		data.Content = types.StringNull()
	} else {
		data.Content = types.StringValue(*out.Content)
	}
	if out.Size == nil {
		data.Size = types.Int64Null()
	} else {
		data.Size = types.Int64Value(*out.Size)
	}
	if out.Mtime == nil || strings.TrimSpace(*out.Mtime) == "" {
		data.Mtime = types.StringNull()
	} else {
		data.Mtime = types.StringValue(strings.TrimSpace(*out.Mtime))
	}
	data.ResultJSON = types.StringValue(mustJSON(out))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

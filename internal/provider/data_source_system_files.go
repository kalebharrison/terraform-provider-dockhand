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
	_ datasource.DataSource              = (*systemFilesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*systemFilesDataSource)(nil)
)

func NewSystemFilesDataSource() datasource.DataSource {
	return &systemFilesDataSource{}
}

type systemFilesDataSource struct {
	client *Client
}

type systemFilesDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Path        types.String `tfsdk:"path"`
	Parent      types.String `tfsdk:"parent"`
	EntryCount  types.Int64  `tfsdk:"entry_count"`
	EntriesJSON types.String `tfsdk:"entries_json"`
	ResultJSON  types.String `tfsdk:"result_json"`
}

func (d *systemFilesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_files"
}

func (d *systemFilesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists system file entries via `GET /api/system/files?path=<dir>`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"path": schema.StringAttribute{
				Optional: true,
			},
			"parent": schema.StringAttribute{
				Computed: true,
			},
			"entry_count": schema.Int64Attribute{
				Computed: true,
			},
			"entries_json": schema.StringAttribute{
				Computed: true,
			},
			"result_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *systemFilesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *systemFilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data systemFilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := strings.TrimSpace(data.Path.ValueString())
	if path == "" {
		path = "/"
	}

	out, status, err := d.client.ListSystemFiles(ctx, path)
	if err != nil {
		if status == http.StatusNotFound {
			resp.Diagnostics.AddError("System path not found", fmt.Sprintf("Dockhand returned 404 for path %q", path))
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand system files", err.Error())
		return
	}

	data.ID = types.StringValue("system-files:" + path)
	if out.Path == nil || strings.TrimSpace(*out.Path) == "" {
		data.Path = types.StringValue(path)
	} else {
		data.Path = types.StringValue(strings.TrimSpace(*out.Path))
	}
	if out.Parent == nil || strings.TrimSpace(*out.Parent) == "" {
		data.Parent = types.StringNull()
	} else {
		data.Parent = types.StringValue(strings.TrimSpace(*out.Parent))
	}
	data.EntryCount = types.Int64Value(int64(len(out.Entries)))
	data.EntriesJSON = types.StringValue(mustJSON(out.Entries))
	data.ResultJSON = types.StringValue(mustJSON(out))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

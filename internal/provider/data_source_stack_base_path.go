package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*stackBasePathDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*stackBasePathDataSource)(nil)
)

func NewStackBasePathDataSource() datasource.DataSource {
	return &stackBasePathDataSource{}
}

type stackBasePathDataSource struct {
	client *Client
}

type stackBasePathDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	BasePath types.String `tfsdk:"base_path"`
}

func (d *stackBasePathDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_base_path"
}

func (d *stackBasePathDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads global stack base path via `GET /api/stacks/base-path`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"base_path": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *stackBasePathDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *stackBasePathDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data stackBasePathDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := d.client.GetStackBasePath(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand stack base path", err.Error())
		return
	}

	data.ID = types.StringValue("stack-base-path")
	data.BasePath = types.StringValue(out.BasePath)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

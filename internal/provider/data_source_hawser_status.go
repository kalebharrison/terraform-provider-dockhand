package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*hawserStatusDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*hawserStatusDataSource)(nil)
)

func NewHawserStatusDataSource() datasource.DataSource {
	return &hawserStatusDataSource{}
}

type hawserStatusDataSource struct {
	client *Client
}

type hawserStatusDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Status            types.String `tfsdk:"status"`
	Message           types.String `tfsdk:"message"`
	Protocol          types.String `tfsdk:"protocol"`
	ActiveConnections types.Int64  `tfsdk:"active_connections"`
}

func (d *hawserStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hawser_status"
}

func (d *hawserStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads Hawser connect endpoint status.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"message": schema.StringAttribute{
				Computed: true,
			},
			"protocol": schema.StringAttribute{
				Computed: true,
			},
			"active_connections": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *hawserStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *hawserStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data hawserStatusDataSourceModel
	status, _, err := d.client.GetHawserStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading hawser status", err.Error())
		return
	}

	data.ID = types.StringValue("dockhand-hawser-status")
	data.Status = types.StringValue(status.Status)
	data.Message = types.StringValue(status.Message)
	data.Protocol = types.StringValue(status.Protocol)
	data.ActiveConnections = types.Int64Value(status.ActiveConnections)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

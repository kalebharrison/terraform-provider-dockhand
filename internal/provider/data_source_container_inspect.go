package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*containerInspectDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*containerInspectDataSource)(nil)
)

func NewContainerInspectDataSource() datasource.DataSource {
	return &containerInspectDataSource{}
}

type containerInspectDataSource struct {
	client *Client
}

type containerInspectDataSourceModel struct {
	Env         types.String `tfsdk:"env"`
	ContainerID types.String `tfsdk:"container_id"`
	InspectJSON types.String `tfsdk:"inspect_json"`
}

func (d *containerInspectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_inspect"
}

func (d *containerInspectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads full container inspect payload from Dockhand.",
		Attributes: map[string]schema.Attribute{
			"env": schema.StringAttribute{
				Optional: true,
			},
			"container_id": schema.StringAttribute{
				Required: true,
			},
			"inspect_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *containerInspectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *containerInspectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data containerInspectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inspect, _, err := d.client.GetContainerInspect(ctx, data.Env.ValueString(), data.ContainerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading container inspect", err.Error())
		return
	}
	raw, err := json.Marshal(inspect)
	if err != nil {
		resp.Diagnostics.AddError("Error encoding inspect payload", err.Error())
		return
	}

	data.InspectJSON = types.StringValue(string(raw))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

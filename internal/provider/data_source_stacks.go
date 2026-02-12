package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*stacksDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*stacksDataSource)(nil)
)

func NewStacksDataSource() datasource.DataSource {
	return &stacksDataSource{}
}

type stacksDataSource struct {
	client *Client
}

type stackItemModel struct {
	Name           types.String `tfsdk:"name"`
	Status         types.String `tfsdk:"status"`
	ContainerCount types.Int64  `tfsdk:"container_count"`
}

type stacksDataSourceModel struct {
	Env    types.String     `tfsdk:"env"`
	Stacks []stackItemModel `tfsdk:"stacks"`
}

func (d *stacksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stacks"
}

func (d *stacksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists stacks for a Dockhand environment.",
		Attributes: map[string]schema.Attribute{
			"env": schema.StringAttribute{
				Optional: true,
			},
			"stacks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"container_count": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *stacksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *stacksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data stacksDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListStacks(ctx, data.Env.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading stacks", err.Error())
		return
	}

	flattened := make([]stackItemModel, 0, len(items))
	for _, item := range items {
		flattened = append(flattened, stackItemModel{
			Name:           types.StringValue(item.Name),
			Status:         types.StringValue(item.Status),
			ContainerCount: types.Int64Value(int64(len(item.Containers))),
		})
	}
	data.Stacks = flattened
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

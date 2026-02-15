package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*networksDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*networksDataSource)(nil)
)

func NewNetworksDataSource() datasource.DataSource {
	return &networksDataSource{}
}

type networksDataSource struct {
	client *Client
}

type networksDataSourceNetworkModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Driver     types.String `tfsdk:"driver"`
	Internal   types.Bool   `tfsdk:"internal"`
	Attachable types.Bool   `tfsdk:"attachable"`
	Scope      types.String `tfsdk:"scope"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

type networksDataSourceModel struct {
	Env      types.String                     `tfsdk:"env"`
	Networks []networksDataSourceNetworkModel `tfsdk:"networks"`
	Names    types.List                       `tfsdk:"names"`
	IDs      types.List                       `tfsdk:"ids"`
}

func (d *networksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networks"
}

func (d *networksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists networks from Dockhand `/api/networks`.",
		Attributes: map[string]schema.Attribute{
			"env": schema.StringAttribute{
				Optional: true,
			},
			"networks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true},
						"name":       schema.StringAttribute{Computed: true},
						"driver":     schema.StringAttribute{Computed: true},
						"internal":   schema.BoolAttribute{Computed: true},
						"attachable": schema.BoolAttribute{Computed: true},
						"scope":      schema.StringAttribute{Computed: true},
						"created_at": schema.StringAttribute{Computed: true},
					},
				},
			},
			"names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *networksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *networksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data networksDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListNetworks(ctx, data.Env.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand networks", err.Error())
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	out := make([]networksDataSourceNetworkModel, 0, len(items))
	names := make([]string, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, networksDataSourceNetworkModel{
			ID:         types.StringValue(item.ID),
			Name:       types.StringValue(item.Name),
			Driver:     types.StringValue(item.Driver),
			Internal:   types.BoolValue(item.Internal),
			Attachable: types.BoolValue(item.Attachable),
			Scope:      stringValueOrNull(item.Scope),
			CreatedAt:  stringValueOrNull(item.CreatedAt),
		})
		names = append(names, item.Name)
		ids = append(ids, item.ID)
	}

	namesVal, diags := types.ListValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	idsVal, diags := types.ListValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Networks = out
	data.Names = namesVal
	data.IDs = idsVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

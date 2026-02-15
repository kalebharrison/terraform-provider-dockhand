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
	_ datasource.DataSource              = (*volumesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*volumesDataSource)(nil)
)

func NewVolumesDataSource() datasource.DataSource {
	return &volumesDataSource{}
}

type volumesDataSource struct {
	client *Client
}

type volumesDataSourceVolumeModel struct {
	Name       types.String `tfsdk:"name"`
	Driver     types.String `tfsdk:"driver"`
	Mountpoint types.String `tfsdk:"mountpoint"`
	Scope      types.String `tfsdk:"scope"`
	CreatedAt  types.String `tfsdk:"created_at"`
	Labels     types.Map    `tfsdk:"labels"`
}

type volumesDataSourceModel struct {
	Env     types.String                   `tfsdk:"env"`
	Volumes []volumesDataSourceVolumeModel `tfsdk:"volumes"`
	Names   types.List                     `tfsdk:"names"`
}

func (d *volumesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

func (d *volumesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists volumes from Dockhand `/api/volumes`.",
		Attributes: map[string]schema.Attribute{
			"env": schema.StringAttribute{
				Optional: true,
			},
			"volumes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":       schema.StringAttribute{Computed: true},
						"driver":     schema.StringAttribute{Computed: true},
						"mountpoint": schema.StringAttribute{Computed: true},
						"scope":      schema.StringAttribute{Computed: true},
						"created_at": schema.StringAttribute{Computed: true},
						"labels": schema.MapAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *volumesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *volumesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data volumesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListVolumes(ctx, data.Env.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand volumes", err.Error())
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	out := make([]volumesDataSourceVolumeModel, 0, len(items))
	names := make([]string, 0, len(items))
	for _, item := range items {
		labels, diags := types.MapValueFrom(ctx, types.StringType, item.Labels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		out = append(out, volumesDataSourceVolumeModel{
			Name:       types.StringValue(item.Name),
			Driver:     types.StringValue(item.Driver),
			Mountpoint: stringValueOrNull(item.Mountpoint),
			Scope:      stringValueOrNull(item.Scope),
			CreatedAt:  stringValueOrNull(item.CreatedAt),
			Labels:     labels,
		})
		names = append(names, item.Name)
	}

	namesVal, diags := types.ListValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Volumes = out
	data.Names = namesVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

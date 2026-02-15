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
	_ datasource.DataSource              = (*imagesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*imagesDataSource)(nil)
)

func NewImagesDataSource() datasource.DataSource {
	return &imagesDataSource{}
}

type imagesDataSource struct {
	client *Client
}

type imagesDataSourceImageModel struct {
	ID      types.String `tfsdk:"id"`
	Tags    types.List   `tfsdk:"tags"`
	Size    types.Int64  `tfsdk:"size"`
	Created types.Int64  `tfsdk:"created"`
}

type imagesDataSourceModel struct {
	Env    types.String                 `tfsdk:"env"`
	Images []imagesDataSourceImageModel `tfsdk:"images"`
	IDs    types.List                   `tfsdk:"ids"`
}

func (d *imagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *imagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists images from Dockhand `/api/images`.",
		Attributes: map[string]schema.Attribute{
			"env": schema.StringAttribute{
				Optional: true,
			},
			"images": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true},
						"tags":    schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"size":    schema.Int64Attribute{Computed: true},
						"created": schema.Int64Attribute{Computed: true},
					},
				},
			},
			"ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *imagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *imagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data imagesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListImages(ctx, data.Env.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand images", err.Error())
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	out := make([]imagesDataSourceImageModel, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		tagsVal, diags := types.ListValueFrom(ctx, types.StringType, item.Tags)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		out = append(out, imagesDataSourceImageModel{
			ID:      types.StringValue(item.ID),
			Tags:    tagsVal,
			Size:    types.Int64Value(item.Size),
			Created: types.Int64Value(item.Created),
		})
		ids = append(ids, item.ID)
	}

	idsVal, diags := types.ListValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Images = out
	data.IDs = idsVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

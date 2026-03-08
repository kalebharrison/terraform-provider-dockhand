package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*registryTagsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*registryTagsDataSource)(nil)
)

func NewRegistryTagsDataSource() datasource.DataSource {
	return &registryTagsDataSource{}
}

type registryTagsDataSource struct {
	client *Client
}

type registryTagDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	Size        types.Int64  `tfsdk:"size"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Digest      types.String `tfsdk:"digest"`
}

type registryTagsDataSourceModel struct {
	ID       types.String                 `tfsdk:"id"`
	Image    types.String                 `tfsdk:"image"`
	Registry types.String                 `tfsdk:"registry"`
	Page     types.Int64                  `tfsdk:"page"`
	PageSize types.Int64                  `tfsdk:"page_size"`
	Total    types.Int64                  `tfsdk:"total"`
	HasNext  types.Bool                   `tfsdk:"has_next"`
	HasPrev  types.Bool                   `tfsdk:"has_prev"`
	Tags     []registryTagDataSourceModel `tfsdk:"tags"`
	TagNames types.List                   `tfsdk:"tag_names"`
	TagsJSON types.String                 `tfsdk:"tags_json"`
}

func (d *registryTagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_tags"
}

func (d *registryTagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists available image tags via `GET /api/registry/tags`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "Image/repository name (for example `library/busybox`).",
				Required:            true,
			},
			"registry": schema.StringAttribute{
				MarkdownDescription: "Optional registry selector sent as `registry` query parameter.",
				Optional:            true,
			},
			"page": schema.Int64Attribute{
				MarkdownDescription: "Optional page number query parameter.",
				Optional:            true,
			},
			"page_size": schema.Int64Attribute{
				MarkdownDescription: "Optional page size query parameter.",
				Optional:            true,
			},
			"total": schema.Int64Attribute{
				Computed: true,
			},
			"has_next": schema.BoolAttribute{
				Computed: true,
			},
			"has_prev": schema.BoolAttribute{
				Computed: true,
			},
			"tags": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":         schema.StringAttribute{Computed: true},
						"size":         schema.Int64Attribute{Computed: true},
						"last_updated": schema.StringAttribute{Computed: true},
						"digest":       schema.StringAttribute{Computed: true},
					},
				},
			},
			"tag_names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"tags_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *registryTagsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *registryTagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data registryTagsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	image := strings.TrimSpace(data.Image.ValueString())
	if image == "" {
		resp.Diagnostics.AddError("Invalid `image`", "`image` must be non-empty.")
		return
	}

	registry := strings.TrimSpace(data.Registry.ValueString())
	page := int64(0)
	if !data.Page.IsNull() && !data.Page.IsUnknown() {
		page = data.Page.ValueInt64()
	}
	pageSize := int64(0)
	if !data.PageSize.IsNull() && !data.PageSize.IsUnknown() {
		pageSize = data.PageSize.ValueInt64()
	}

	out, _, err := d.client.ListRegistryTags(ctx, image, registry, page, pageSize)
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand registry tags", err.Error())
		return
	}

	tags := make([]registryTagDataSourceModel, 0, len(out.Tags))
	tagNames := make([]string, 0, len(out.Tags))
	for _, item := range out.Tags {
		entry := registryTagDataSourceModel{
			Name: types.StringValue(strings.TrimSpace(item.Name)),
			Size: types.Int64Value(item.Size),
		}
		if item.LastUpdated == nil || strings.TrimSpace(*item.LastUpdated) == "" {
			entry.LastUpdated = types.StringNull()
		} else {
			entry.LastUpdated = types.StringValue(strings.TrimSpace(*item.LastUpdated))
		}
		if item.Digest == nil || strings.TrimSpace(*item.Digest) == "" {
			entry.Digest = types.StringNull()
		} else {
			entry.Digest = types.StringValue(strings.TrimSpace(*item.Digest))
		}
		tags = append(tags, entry)
		tagNames = append(tagNames, strings.TrimSpace(item.Name))
	}

	tagNameList, diags := types.ListValueFrom(ctx, types.StringType, tagNames)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s:%d:%d", image, registry, page, pageSize))
	data.Image = types.StringValue(image)
	if registry == "" {
		data.Registry = types.StringNull()
	} else {
		data.Registry = types.StringValue(registry)
	}
	if page > 0 {
		data.Page = types.Int64Value(page)
	} else {
		data.Page = types.Int64Null()
	}
	if pageSize > 0 {
		data.PageSize = types.Int64Value(pageSize)
	} else {
		data.PageSize = types.Int64Null()
	}
	data.Total = types.Int64Value(out.Total)
	data.HasNext = types.BoolValue(out.HasNext)
	data.HasPrev = types.BoolValue(out.HasPrev)
	data.Tags = tags
	data.TagNames = tagNameList
	data.TagsJSON = types.StringValue(mustJSON(out.Tags))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

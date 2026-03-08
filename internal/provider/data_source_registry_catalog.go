package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*registryCatalogDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*registryCatalogDataSource)(nil)
)

func NewRegistryCatalogDataSource() datasource.DataSource {
	return &registryCatalogDataSource{}
}

type registryCatalogDataSource struct {
	client *Client
}

type registryCatalogDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Registry        types.String `tfsdk:"registry"`
	Page            types.Int64  `tfsdk:"page"`
	PageSize        types.Int64  `tfsdk:"page_size"`
	RepositoryCount types.Int64  `tfsdk:"repository_count"`
	Repositories    types.List   `tfsdk:"repositories"`
	CatalogJSON     types.String `tfsdk:"catalog_json"`
}

func (d *registryCatalogDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_catalog"
}

func (d *registryCatalogDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads registry catalog data via `GET /api/registry/catalog`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
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
			"repository_count": schema.Int64Attribute{
				Computed: true,
			},
			"repositories": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"catalog_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *registryCatalogDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *registryCatalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data registryCatalogDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
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

	raw, _, err := d.client.GetRegistryCatalogRaw(ctx, registry, page, pageSize)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand registry catalog", err.Error())
		return
	}

	repositories := extractCatalogRepositories(raw)
	repoList, diags := types.ListValueFrom(ctx, types.StringType, repositories)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%d:%d", registry, page, pageSize))
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
	data.RepositoryCount = types.Int64Value(int64(len(repositories)))
	data.Repositories = repoList
	data.CatalogJSON = types.StringValue(strings.TrimSpace(string(raw)))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func extractCatalogRepositories(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return []string{}
	}

	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return []string{}
	}

	switch typed := decoded.(type) {
	case []any:
		return catalogStringsFromAnySlice(typed)
	case map[string]any:
		for _, key := range []string{"repositories", "items", "images", "results"} {
			value, ok := typed[key]
			if !ok {
				continue
			}
			switch v := value.(type) {
			case []any:
				return catalogStringsFromAnySlice(v)
			case []string:
				return append([]string{}, v...)
			}
		}
	}

	return []string{}
}

func catalogStringsFromAnySlice(values []any) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		name := ""
		switch typed := value.(type) {
		case string:
			name = typed
		case map[string]any:
			name = firstString(typed, "name", "image", "repository", "path")
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

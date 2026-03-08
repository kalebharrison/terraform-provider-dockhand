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
	_ datasource.DataSource              = (*registrySearchDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*registrySearchDataSource)(nil)
)

func NewRegistrySearchDataSource() datasource.DataSource {
	return &registrySearchDataSource{}
}

type registrySearchDataSource struct {
	client *Client
}

type registrySearchDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Term        types.String `tfsdk:"term"`
	Registry    types.String `tfsdk:"registry"`
	ResultCount types.Int64  `tfsdk:"result_count"`
	ImageNames  types.List   `tfsdk:"image_names"`
	ResultsJSON types.String `tfsdk:"results_json"`
}

func (d *registrySearchDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_search"
}

func (d *registrySearchDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Searches registry repositories via `GET /api/registry/search`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"term": schema.StringAttribute{
				MarkdownDescription: "Search term sent as `term` query parameter.",
				Required:            true,
			},
			"registry": schema.StringAttribute{
				MarkdownDescription: "Optional registry ID/name selector sent as `registry` query parameter.",
				Optional:            true,
			},
			"result_count": schema.Int64Attribute{
				Computed: true,
			},
			"image_names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"results_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *registrySearchDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *registrySearchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data registrySearchDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	term := strings.TrimSpace(data.Term.ValueString())
	if term == "" {
		resp.Diagnostics.AddError("Invalid `term`", "`term` must be non-empty.")
		return
	}
	registry := strings.TrimSpace(data.Registry.ValueString())

	results, _, err := d.client.SearchRegistry(ctx, term, registry)
	if err != nil {
		resp.Diagnostics.AddError("Error searching Dockhand registry", err.Error())
		return
	}

	names := extractRegistryImageNames(results)
	nameList, diags := types.ListValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s:%s", term, registry))
	data.Term = types.StringValue(term)
	if registry == "" {
		data.Registry = types.StringNull()
	} else {
		data.Registry = types.StringValue(registry)
	}
	data.ResultCount = types.Int64Value(int64(len(results)))
	data.ImageNames = nameList
	data.ResultsJSON = types.StringValue(mustJSON(results))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func extractRegistryImageNames(items []map[string]any) []string {
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		name := firstString(item, "name", "image", "repository", "fullName", "full_name")
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

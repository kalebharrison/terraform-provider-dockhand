package provider

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*registriesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*registriesDataSource)(nil)
)

func NewRegistriesDataSource() datasource.DataSource {
	return &registriesDataSource{}
}

type registriesDataSource struct {
	client *Client
}

type registriesDataSourceRegistryModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	URL            types.String `tfsdk:"url"`
	Username       types.String `tfsdk:"username"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
	HasCredentials types.Bool   `tfsdk:"has_credentials"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

type registriesDataSourceModel struct {
	Registries []registriesDataSourceRegistryModel `tfsdk:"registries"`
	Names      types.List                          `tfsdk:"names"`
	IDs        types.List                          `tfsdk:"ids"`
}

func (d *registriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registries"
}

func (d *registriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists registries from Dockhand `/api/registries`.",
		Attributes: map[string]schema.Attribute{
			"registries": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true},
						"name":            schema.StringAttribute{Computed: true},
						"url":             schema.StringAttribute{Computed: true},
						"username":        schema.StringAttribute{Computed: true},
						"is_default":      schema.BoolAttribute{Computed: true},
						"has_credentials": schema.BoolAttribute{Computed: true},
						"created_at":      schema.StringAttribute{Computed: true},
						"updated_at":      schema.StringAttribute{Computed: true},
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

func (d *registriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *registriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data registriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListRegistries(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand registries", err.Error())
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	out := make([]registriesDataSourceRegistryModel, 0, len(items))
	names := make([]string, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		id := strconv.FormatInt(item.ID, 10)
		out = append(out, registriesDataSourceRegistryModel{
			ID:             types.StringValue(id),
			Name:           types.StringValue(item.Name),
			URL:            types.StringValue(item.URL),
			Username:       stringValueOrNull(item.Username),
			IsDefault:      types.BoolValue(item.IsDefault),
			HasCredentials: types.BoolValue(item.HasCredentials),
			CreatedAt:      stringValueOrNull(item.CreatedAt),
			UpdatedAt:      stringValueOrNull(item.UpdatedAt),
		})
		names = append(names, item.Name)
		ids = append(ids, id)
	}

	namesVal, diags := types.ListValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	idsVal, diags := types.ListValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Registries = out
	data.Names = namesVal
	data.IDs = idsVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

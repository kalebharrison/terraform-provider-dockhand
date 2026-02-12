package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*authProvidersDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*authProvidersDataSource)(nil)
)

func NewAuthProvidersDataSource() datasource.DataSource {
	return &authProvidersDataSource{}
}

type authProvidersDataSource struct {
	client *Client
}

type authProviderItemModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

type authProvidersDataSourceModel struct {
	ID              types.String            `tfsdk:"id"`
	DefaultProvider types.String            `tfsdk:"default_provider"`
	Providers       []authProviderItemModel `tfsdk:"providers"`
}

func (d *authProvidersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_providers"
}

func (d *authProvidersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads authentication providers from `/api/auth/providers`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"default_provider": schema.StringAttribute{
				Computed: true,
			},
			"providers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *authProvidersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *authProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	apiOut, _, err := d.client.GetAuthProviders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand auth providers", err.Error())
		return
	}

	out := authProvidersDataSourceModel{
		ID:              types.StringValue("auth-providers"),
		DefaultProvider: types.StringValue(apiOut.DefaultProvider),
		Providers:       make([]authProviderItemModel, 0, len(apiOut.Providers)),
	}
	for _, p := range apiOut.Providers {
		out.Providers = append(out.Providers, authProviderItemModel{
			ID:   types.StringValue(p.ID),
			Name: types.StringValue(p.Name),
			Type: types.StringValue(p.Type),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &out)...)
}

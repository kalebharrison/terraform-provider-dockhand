package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*configSetsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*configSetsDataSource)(nil)
)

func NewConfigSetsDataSource() datasource.DataSource {
	return &configSetsDataSource{}
}

type configSetsDataSource struct {
	client *Client
}

type configSetsDataSourceSetModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	NetworkMode   types.String `tfsdk:"network_mode"`
	RestartPolicy types.String `tfsdk:"restart_policy"`
	SpecJSON      types.String `tfsdk:"spec_json"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

type configSetsDataSourceModel struct {
	ConfigSets []configSetsDataSourceSetModel `tfsdk:"config_sets"`
	Names      types.List                     `tfsdk:"names"`
	IDs        types.List                     `tfsdk:"ids"`
}

func (d *configSetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_sets"
}

func (d *configSetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists config sets from Dockhand `/api/config-sets`.",
		Attributes: map[string]schema.Attribute{
			"config_sets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true},
						"name":           schema.StringAttribute{Computed: true},
						"description":    schema.StringAttribute{Computed: true},
						"network_mode":   schema.StringAttribute{Computed: true},
						"restart_policy": schema.StringAttribute{Computed: true},
						"spec_json":      schema.StringAttribute{Computed: true},
						"created_at":     schema.StringAttribute{Computed: true},
						"updated_at":     schema.StringAttribute{Computed: true},
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

func (d *configSetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *configSetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data configSetsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListConfigSets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand config sets", err.Error())
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	out := make([]configSetsDataSourceSetModel, 0, len(items))
	names := make([]string, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		id := strconv.FormatInt(item.ID, 10)
		specBytes, err := json.Marshal(map[string]any{
			"envVars": item.EnvVars,
			"labels":  item.Labels,
			"ports":   item.Ports,
			"volumes": item.Volumes,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error encoding config set spec", err.Error())
			return
		}
		out = append(out, configSetsDataSourceSetModel{
			ID:            types.StringValue(id),
			Name:          types.StringValue(item.Name),
			Description:   stringValueOrNull(item.Description),
			NetworkMode:   types.StringValue(item.NetworkMode),
			RestartPolicy: types.StringValue(item.RestartPolicy),
			SpecJSON:      types.StringValue(string(specBytes)),
			CreatedAt:     stringValueOrNull(item.CreatedAt),
			UpdatedAt:     stringValueOrNull(item.UpdatedAt),
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

	data.ConfigSets = out
	data.Names = namesVal
	data.IDs = idsVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

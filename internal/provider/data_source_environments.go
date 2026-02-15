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
	_ datasource.DataSource              = (*environmentsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*environmentsDataSource)(nil)
)

func NewEnvironmentsDataSource() datasource.DataSource {
	return &environmentsDataSource{}
}

type environmentsDataSource struct {
	client *Client
}

type environmentsDataSourceEnvironmentModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	ConnectionType        types.String `tfsdk:"connection_type"`
	Host                  types.String `tfsdk:"host"`
	Port                  types.Int64  `tfsdk:"port"`
	Protocol              types.String `tfsdk:"protocol"`
	SocketPath            types.String `tfsdk:"socket_path"`
	TLSSkipVerify         types.Bool   `tfsdk:"tls_skip_verify"`
	Icon                  types.String `tfsdk:"icon"`
	CollectActivity       types.Bool   `tfsdk:"collect_activity"`
	CollectMetrics        types.Bool   `tfsdk:"collect_metrics"`
	HighlightChanges      types.Bool   `tfsdk:"highlight_changes"`
	Timezone              types.String `tfsdk:"timezone"`
	UpdateCheckEnabled    types.Bool   `tfsdk:"update_check_enabled"`
	UpdateCheckAutoUpdate types.Bool   `tfsdk:"update_check_auto_update"`
	ImagePruneEnabled     types.Bool   `tfsdk:"image_prune_enabled"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

type environmentsDataSourceModel struct {
	Environments []environmentsDataSourceEnvironmentModel `tfsdk:"environments"`
	Names        types.List                               `tfsdk:"names"`
	IDs          types.List                               `tfsdk:"ids"`
}

func (d *environmentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

func (d *environmentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists environments from Dockhand `/api/environments`.",
		Attributes: map[string]schema.Attribute{
			"environments": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                       schema.StringAttribute{Computed: true},
						"name":                     schema.StringAttribute{Computed: true},
						"connection_type":          schema.StringAttribute{Computed: true},
						"host":                     schema.StringAttribute{Computed: true},
						"port":                     schema.Int64Attribute{Computed: true},
						"protocol":                 schema.StringAttribute{Computed: true},
						"socket_path":              schema.StringAttribute{Computed: true},
						"tls_skip_verify":          schema.BoolAttribute{Computed: true},
						"icon":                     schema.StringAttribute{Computed: true},
						"collect_activity":         schema.BoolAttribute{Computed: true},
						"collect_metrics":          schema.BoolAttribute{Computed: true},
						"highlight_changes":        schema.BoolAttribute{Computed: true},
						"timezone":                 schema.StringAttribute{Computed: true},
						"update_check_enabled":     schema.BoolAttribute{Computed: true},
						"update_check_auto_update": schema.BoolAttribute{Computed: true},
						"image_prune_enabled":      schema.BoolAttribute{Computed: true},
						"created_at":               schema.StringAttribute{Computed: true},
						"updated_at":               schema.StringAttribute{Computed: true},
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

func (d *environmentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *environmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data environmentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListEnvironments(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand environments", err.Error())
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	out := make([]environmentsDataSourceEnvironmentModel, 0, len(items))
	names := make([]string, 0, len(items))
	ids := make([]string, 0, len(items))

	for _, item := range items {
		id := strconv.FormatInt(item.ID, 10)
		out = append(out, environmentsDataSourceEnvironmentModel{
			ID:                    types.StringValue(id),
			Name:                  types.StringValue(item.Name),
			ConnectionType:        types.StringValue(item.ConnectionType),
			Host:                  stringValueOrNull(item.Host),
			Port:                  types.Int64Value(item.Port),
			Protocol:              types.StringValue(item.Protocol),
			SocketPath:            stringValueOrNull(item.SocketPath),
			TLSSkipVerify:         types.BoolValue(item.TLSSkipVerify),
			Icon:                  types.StringValue(item.Icon),
			CollectActivity:       types.BoolValue(item.CollectActivity),
			CollectMetrics:        types.BoolValue(item.CollectMetrics),
			HighlightChanges:      types.BoolValue(item.HighlightChanges),
			Timezone:              stringValueOrNull(item.Timezone),
			UpdateCheckEnabled:    types.BoolValue(item.UpdateCheckEnabled),
			UpdateCheckAutoUpdate: types.BoolValue(item.UpdateCheckAutoUpdate),
			ImagePruneEnabled:     types.BoolValue(item.ImagePruneEnabled),
			CreatedAt:             stringValueOrNull(item.CreatedAt),
			UpdatedAt:             stringValueOrNull(item.UpdatedAt),
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

	data.Environments = out
	data.Names = namesVal
	data.IDs = idsVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

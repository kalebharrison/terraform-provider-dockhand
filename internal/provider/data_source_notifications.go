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
	_ datasource.DataSource              = (*notificationsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*notificationsDataSource)(nil)
)

func NewNotificationsDataSource() datasource.DataSource {
	return &notificationsDataSource{}
}

type notificationsDataSource struct {
	client *Client
}

type notificationsDataSourceNotificationModel struct {
	ID         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	Name       types.String `tfsdk:"name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	EventTypes types.List   `tfsdk:"event_types"`
	ConfigJSON types.String `tfsdk:"config_json"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

type notificationsDataSourceModel struct {
	Notifications []notificationsDataSourceNotificationModel `tfsdk:"notifications"`
	Names         types.List                                 `tfsdk:"names"`
	IDs           types.List                                 `tfsdk:"ids"`
}

func (d *notificationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notifications"
}

func (d *notificationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists notification configs from Dockhand `/api/notifications`.",
		Attributes: map[string]schema.Attribute{
			"notifications": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":      schema.StringAttribute{Computed: true},
						"type":    schema.StringAttribute{Computed: true},
						"name":    schema.StringAttribute{Computed: true},
						"enabled": schema.BoolAttribute{Computed: true},
						"event_types": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
						"config_json": schema.StringAttribute{Computed: true},
						"created_at":  schema.StringAttribute{Computed: true},
						"updated_at":  schema.StringAttribute{Computed: true},
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

func (d *notificationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *notificationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data notificationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListNotifications(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand notifications", err.Error())
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	out := make([]notificationsDataSourceNotificationModel, 0, len(items))
	names := make([]string, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		id := strconv.FormatInt(item.ID, 10)
		eventTypesVal, diags := types.ListValueFrom(ctx, types.StringType, item.EventTypes)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		configBytes, err := json.Marshal(item.Config)
		if err != nil {
			resp.Diagnostics.AddError("Error encoding notification config", err.Error())
			return
		}

		out = append(out, notificationsDataSourceNotificationModel{
			ID:         types.StringValue(id),
			Type:       types.StringValue(item.Type),
			Name:       types.StringValue(item.Name),
			Enabled:    types.BoolValue(item.Enabled),
			EventTypes: eventTypesVal,
			ConfigJSON: types.StringValue(string(configBytes)),
			CreatedAt:  stringValueOrNull(item.CreatedAt),
			UpdatedAt:  stringValueOrNull(item.UpdatedAt),
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

	data.Notifications = out
	data.Names = namesVal
	data.IDs = idsVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*activityDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*activityDataSource)(nil)
)

func NewActivityDataSource() datasource.DataSource {
	return &activityDataSource{}
}

type activityDataSource struct {
	client *Client
}

type activityEventModel struct {
	ID            types.String `tfsdk:"id"`
	Action        types.String `tfsdk:"action"`
	ContainerID   types.String `tfsdk:"container_id"`
	ContainerName types.String `tfsdk:"container_name"`
	Image         types.String `tfsdk:"image"`
	Timestamp     types.String `tfsdk:"timestamp"`
	Status        types.String `tfsdk:"status"`
}

type activityDataSourceModel struct {
	ID     types.String         `tfsdk:"id"`
	Limit  types.Int64          `tfsdk:"limit"`
	Events []activityEventModel `tfsdk:"events"`
}

func (d *activityDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_activity"
}

func (d *activityDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads recent activity events from Dockhand.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"limit": schema.Int64Attribute{
				Optional: true,
			},
			"events": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{Computed: true},
						"action": schema.StringAttribute{
							Computed: true,
						},
						"container_id": schema.StringAttribute{
							Computed: true,
						},
						"container_name": schema.StringAttribute{
							Computed: true,
						},
						"image": schema.StringAttribute{
							Computed: true,
						},
						"timestamp": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *activityDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *activityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data activityDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	limit := int64(50)
	if !data.Limit.IsNull() && !data.Limit.IsUnknown() && data.Limit.ValueInt64() > 0 {
		limit = data.Limit.ValueInt64()
	}
	data.Limit = types.Int64Value(limit)

	events, _, err := d.client.ListActivity(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading activity", err.Error())
		return
	}

	if int64(len(events)) > limit {
		events = events[:limit]
	}

	out := make([]activityEventModel, 0, len(events))
	for _, ev := range events {
		item := activityEventModel{
			ID:        types.StringValue(fmt.Sprintf("%d", ev.ID)),
			Action:    types.StringValue(ev.Action),
			Status:    types.StringNull(),
			Image:     types.StringNull(),
			Timestamp: types.StringNull(),
		}
		if ev.ContainerID != nil {
			item.ContainerID = types.StringValue(*ev.ContainerID)
		} else {
			item.ContainerID = types.StringNull()
		}
		if ev.ContainerName != nil {
			item.ContainerName = types.StringValue(*ev.ContainerName)
		} else {
			item.ContainerName = types.StringNull()
		}
		if ev.Image != nil {
			item.Image = types.StringValue(*ev.Image)
		}
		if ev.Timestamp != nil {
			item.Timestamp = types.StringValue(*ev.Timestamp)
		}
		if ev.Status != nil {
			item.Status = types.StringValue(*ev.Status)
		}
		out = append(out, item)
	}

	data.ID = types.StringValue("dockhand-activity")
	data.Events = out
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

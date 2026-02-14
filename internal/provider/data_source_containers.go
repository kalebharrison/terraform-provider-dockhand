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
	_ datasource.DataSource              = (*containersDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*containersDataSource)(nil)
)

func NewContainersDataSource() datasource.DataSource {
	return &containersDataSource{}
}

type containersDataSource struct {
	client *Client
}

type containersDataSourceContainerModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Image        types.String `tfsdk:"image"`
	State        types.String `tfsdk:"state"`
	Status       types.String `tfsdk:"status"`
	Health       types.String `tfsdk:"health"`
	RestartCount types.Int64  `tfsdk:"restart_count"`
	Command      types.String `tfsdk:"command"`
	Labels       types.Map    `tfsdk:"labels"`
}

type containersDataSourceModel struct {
	Env        types.String                         `tfsdk:"env"`
	Containers []containersDataSourceContainerModel `tfsdk:"containers"`
	Names      types.List                           `tfsdk:"names"`
	IDs        types.List                           `tfsdk:"ids"`
}

func (d *containersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_containers"
}

func (d *containersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists containers from Dockhand `/api/containers`.",
		Attributes: map[string]schema.Attribute{
			"env": schema.StringAttribute{
				MarkdownDescription: "Optional environment ID query parameter.",
				Optional:            true,
			},
			"containers": schema.ListNestedAttribute{
				MarkdownDescription: "Container objects returned by Dockhand.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true},
						"name":          schema.StringAttribute{Computed: true},
						"image":         schema.StringAttribute{Computed: true},
						"state":         schema.StringAttribute{Computed: true},
						"status":        schema.StringAttribute{Computed: true},
						"health":        schema.StringAttribute{Computed: true},
						"restart_count": schema.Int64Attribute{Computed: true},
						"command":       schema.StringAttribute{Computed: true},
						"labels": schema.MapAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"names": schema.ListAttribute{
				MarkdownDescription: "Sorted list of container names.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"ids": schema.ListAttribute{
				MarkdownDescription: "Sorted list of container IDs.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *containersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *containersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data containersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	containers, _, err := d.client.ListContainers(ctx, data.Env.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand containers", err.Error())
		return
	}

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Name < containers[j].Name
	})

	items := make([]containersDataSourceContainerModel, 0, len(containers))
	names := make([]string, 0, len(containers))
	ids := make([]string, 0, len(containers))

	for _, c := range containers {
		labels, diags := types.MapValueFrom(ctx, types.StringType, c.Labels)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		command := types.StringNull()
		if c.Command != nil {
			command = types.StringValue(*c.Command)
		}

		items = append(items, containersDataSourceContainerModel{
			ID:           types.StringValue(c.ID),
			Name:         types.StringValue(c.Name),
			Image:        types.StringValue(c.Image),
			State:        types.StringValue(c.State),
			Status:       types.StringValue(c.Status),
			Health:       types.StringValue(c.Health),
			RestartCount: types.Int64Value(c.RestartCount),
			Command:      command,
			Labels:       labels,
		})
		names = append(names, c.Name)
		ids = append(ids, c.ID)
	}

	namesVal, diags := types.ListValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	idsVal, diags := types.ListValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Containers = items
	data.Names = namesVal
	data.IDs = idsVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

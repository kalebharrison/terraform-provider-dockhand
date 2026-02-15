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
	_ datasource.DataSource              = (*containerProcessesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*containerProcessesDataSource)(nil)
)

func NewContainerProcessesDataSource() datasource.DataSource {
	return &containerProcessesDataSource{}
}

type containerProcessesDataSource struct {
	client *Client
}

type containerProcessRowModel struct {
	Values []types.String `tfsdk:"values"`
}

type containerProcessesDataSourceModel struct {
	ID          types.String               `tfsdk:"id"`
	Env         types.String               `tfsdk:"env"`
	ContainerID types.String               `tfsdk:"container_id"`
	Titles      []types.String             `tfsdk:"titles"`
	Processes   []containerProcessRowModel `tfsdk:"processes"`
	Error       types.String               `tfsdk:"error"`
}

func (d *containerProcessesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_processes"
}

func (d *containerProcessesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads process table information for a running container from `/api/containers/{id}/top`.",
		Attributes: map[string]schema.Attribute{
			"id":           schema.StringAttribute{Computed: true},
			"env":          schema.StringAttribute{Optional: true},
			"container_id": schema.StringAttribute{Required: true},
			"titles": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"processes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"values": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"error": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *containerProcessesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *containerProcessesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var config containerProcessesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	containerID := strings.TrimSpace(config.ContainerID.ValueString())
	if containerID == "" {
		resp.Diagnostics.AddError("Invalid container ID", "`container_id` cannot be empty.")
		return
	}

	out, status, err := d.client.GetContainerTop(ctx, config.Env.ValueString(), containerID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand container processes", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error reading Dockhand container processes", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	data := containerProcessesDataSourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s:%s", config.Env.ValueString(), containerID)),
		Env:         config.Env,
		ContainerID: types.StringValue(containerID),
		Titles:      make([]types.String, 0, len(out.Titles)),
		Processes:   make([]containerProcessRowModel, 0, len(out.Processes)),
		Error:       types.StringNull(),
	}
	for _, title := range out.Titles {
		data.Titles = append(data.Titles, types.StringValue(title))
	}
	for _, row := range out.Processes {
		vals := make([]types.String, 0, len(row))
		for _, col := range row {
			vals = append(vals, types.StringValue(col))
		}
		data.Processes = append(data.Processes, containerProcessRowModel{Values: vals})
	}
	if out.Error != nil && strings.TrimSpace(*out.Error) != "" {
		data.Error = types.StringValue(strings.TrimSpace(*out.Error))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

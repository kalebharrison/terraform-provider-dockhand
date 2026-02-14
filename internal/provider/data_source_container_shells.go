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
	_ datasource.DataSource              = (*containerShellsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*containerShellsDataSource)(nil)
)

func NewContainerShellsDataSource() datasource.DataSource {
	return &containerShellsDataSource{}
}

type containerShellsDataSource struct {
	client *Client
}

type containerShellOptionModel struct {
	Path      types.String `tfsdk:"path"`
	Label     types.String `tfsdk:"label"`
	Available types.Bool   `tfsdk:"available"`
}

type containerShellsDataSourceModel struct {
	Env          types.String                `tfsdk:"env"`
	ContainerID  types.String                `tfsdk:"container_id"`
	Shells       types.List                  `tfsdk:"shells"`
	DefaultShell types.String                `tfsdk:"default_shell"`
	AllShells    []containerShellOptionModel `tfsdk:"all_shells"`
}

func (d *containerShellsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_shells"
}

func (d *containerShellsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Detects available shells for a container using Dockhand `/api/containers/{id}/shells`.",
		Attributes: map[string]schema.Attribute{
			"env": schema.StringAttribute{
				MarkdownDescription: "Optional environment ID. Sent as `envId` query parameter for this endpoint.",
				Optional:            true,
			},
			"container_id": schema.StringAttribute{
				MarkdownDescription: "Container ID.",
				Required:            true,
			},
			"shells": schema.ListAttribute{
				MarkdownDescription: "Available shell paths.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"default_shell": schema.StringAttribute{
				MarkdownDescription: "Default shell selected by Dockhand, when provided.",
				Computed:            true,
			},
			"all_shells": schema.ListNestedAttribute{
				MarkdownDescription: "Detailed shell options including availability flags.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path":      schema.StringAttribute{Computed: true},
						"label":     schema.StringAttribute{Computed: true},
						"available": schema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *containerShellsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *containerShellsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data containerShellsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := d.client.GetContainerShells(ctx, data.Env.ValueString(), data.ContainerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading container shells", err.Error())
		return
	}

	shells := append([]string(nil), out.Shells...)
	sort.Strings(shells)
	shellsVal, diags := types.ListValueFrom(ctx, types.StringType, shells)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	defaultShell := types.StringNull()
	if out.DefaultShell != nil {
		defaultShell = types.StringValue(*out.DefaultShell)
	}

	all := make([]containerShellOptionModel, 0, len(out.AllShells))
	for _, s := range out.AllShells {
		all = append(all, containerShellOptionModel{
			Path:      types.StringValue(s.Path),
			Label:     types.StringValue(s.Label),
			Available: types.BoolValue(s.Available),
		})
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Path.ValueString() < all[j].Path.ValueString()
	})

	data.Shells = shellsVal
	data.DefaultShell = defaultShell
	data.AllShells = all
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

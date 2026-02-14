package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*containerPendingUpdatesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*containerPendingUpdatesDataSource)(nil)
)

func NewContainerPendingUpdatesDataSource() datasource.DataSource {
	return &containerPendingUpdatesDataSource{}
}

type containerPendingUpdatesDataSource struct {
	client *Client
}

type containerPendingUpdatesDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Env            types.String `tfsdk:"env"`
	EnvironmentID  types.String `tfsdk:"environment_id"`
	PendingUpdates types.String `tfsdk:"pending_updates_json"`
}

func (d *containerPendingUpdatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_pending_updates"
}

func (d *containerPendingUpdatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads pending container updates from `/api/containers/pending-updates`.",
		Attributes: map[string]schema.Attribute{
			"id":                   schema.StringAttribute{Computed: true},
			"env":                  schema.StringAttribute{Optional: true},
			"environment_id":       schema.StringAttribute{Computed: true},
			"pending_updates_json": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *containerPendingUpdatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *containerPendingUpdatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var config containerPendingUpdatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := d.client.GetContainerPendingUpdates(ctx, config.Env.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading pending container updates", err.Error())
		return
	}

	state := containerPendingUpdatesDataSourceModel{
		ID:             types.StringValue("container-pending-updates:" + config.Env.ValueString()),
		Env:            config.Env,
		EnvironmentID:  types.StringValue(strconv.FormatInt(out.EnvironmentID, 10)),
		PendingUpdates: types.StringValue(mustJSON(out.PendingUpdates)),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

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
	_ datasource.DataSource              = (*stackDefaultPathDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*stackDefaultPathDataSource)(nil)
)

func NewStackDefaultPathDataSource() datasource.DataSource {
	return &stackDefaultPathDataSource{}
}

type stackDefaultPathDataSource struct {
	client *Client
}

type stackDefaultPathDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	StackName   types.String `tfsdk:"stack_name"`
	StackDir    types.String `tfsdk:"stack_dir"`
	ComposePath types.String `tfsdk:"compose_path"`
	EnvPath     types.String `tfsdk:"env_path"`
	Source      types.String `tfsdk:"source"`
	ResultJSON  types.String `tfsdk:"result_json"`
}

func (d *stackDefaultPathDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_default_path"
}

func (d *stackDefaultPathDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads default stack directory paths via `GET /api/stacks/default-path?name=<stack>`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"stack_name": schema.StringAttribute{
				MarkdownDescription: "Stack name used for default path calculation.",
				Required:            true,
			},
			"stack_dir": schema.StringAttribute{
				Computed: true,
			},
			"compose_path": schema.StringAttribute{
				Computed: true,
			},
			"env_path": schema.StringAttribute{
				Computed: true,
			},
			"source": schema.StringAttribute{
				Computed: true,
			},
			"result_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *stackDefaultPathDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *stackDefaultPathDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data stackDefaultPathDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackName := strings.TrimSpace(data.StackName.ValueString())
	if stackName == "" {
		resp.Diagnostics.AddError("Invalid `stack_name`", "`stack_name` must be non-empty.")
		return
	}

	out, _, err := d.client.GetStackDefaultPath(ctx, stackName)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand stack default path", err.Error())
		return
	}

	data.ID = types.StringValue("stack-default-path:" + stackName)
	data.StackName = types.StringValue(stackName)
	data.StackDir = types.StringValue(out.StackDir)
	data.ComposePath = types.StringValue(out.ComposePath)
	data.EnvPath = types.StringValue(out.EnvPath)
	if strings.TrimSpace(out.Source) == "" {
		data.Source = types.StringNull()
	} else {
		data.Source = types.StringValue(strings.TrimSpace(out.Source))
	}
	data.ResultJSON = types.StringValue(mustJSON(out))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

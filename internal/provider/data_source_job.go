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
	_ datasource.DataSource              = (*jobDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*jobDataSource)(nil)
)

func NewJobDataSource() datasource.DataSource {
	return &jobDataSource{}
}

type jobDataSource struct {
	client *Client
}

type jobDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	JobID      types.String `tfsdk:"job_id"`
	Status     types.String `tfsdk:"status"`
	ResultJSON types.String `tfsdk:"result_json"`
	LinesJSON  types.String `tfsdk:"lines_json"`
}

func (d *jobDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

func (d *jobDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Dockhand async job status/result via `/api/jobs/{jobId}`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"job_id": schema.StringAttribute{
				Required: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"result_json": schema.StringAttribute{
				Computed: true,
			},
			"lines_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *jobDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *jobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data jobDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID := strings.TrimSpace(data.JobID.ValueString())
	if jobID == "" {
		resp.Diagnostics.AddError("Invalid `job_id`", "`job_id` must be non-empty.")
		return
	}

	out, _, err := d.client.GetJob(ctx, jobID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand job", err.Error())
		return
	}

	data.ID = types.StringValue(jobID)
	applyJobState(&data.Status, &data.ResultJSON, &data.LinesJSON, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

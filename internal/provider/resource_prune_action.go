package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*pruneActionResource)(nil)
	_ resource.ResourceWithConfigure = (*pruneActionResource)(nil)
)

func NewPruneActionResource() resource.Resource {
	return &pruneActionResource{}
}

type pruneActionResource struct {
	client *Client
}

type pruneActionResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Env               types.String `tfsdk:"env"`
	Mode              types.String `tfsdk:"mode"`
	WaitForCompletion types.Bool   `tfsdk:"wait_for_completion"`
	TimeoutSeconds    types.Int64  `tfsdk:"timeout_seconds"`
	PollIntervalMS    types.Int64  `tfsdk:"poll_interval_ms"`
	Trigger           types.String `tfsdk:"trigger"`

	Success    types.Bool   `tfsdk:"success"`
	StatusCode types.Int64  `tfsdk:"status_code"`
	JobID      types.String `tfsdk:"job_id"`
	JobStatus  types.String `tfsdk:"job_status"`
	ResultJSON types.String `tfsdk:"result_json"`
	LinesJSON  types.String `tfsdk:"lines_json"`
	Error      types.String `tfsdk:"error"`
}

func (r *pruneActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prune_action"
}

func (r *pruneActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs one-shot prune cleanup actions via `/api/prune/*`. Change `trigger` to execute again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"env": schema.StringAttribute{
				MarkdownDescription: "Optional environment ID query parameter. If omitted, provider default env is used.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "Prune mode: `all`, `containers`, `images`, `networks`, or `volumes`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"wait_for_completion": schema.BoolAttribute{
				MarkdownDescription: "Wait for terminal async job status when Dockhand returns `jobId`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"timeout_seconds": schema.Int64Attribute{
				MarkdownDescription: "Timeout in seconds while polling async prune jobs.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(120),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"poll_interval_ms": schema.Int64Attribute{
				MarkdownDescription: "Polling interval in milliseconds for async prune jobs.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1000),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"trigger": schema.StringAttribute{
				MarkdownDescription: "Arbitrary value that forces a new prune run when changed.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"success": schema.BoolAttribute{
				Computed: true,
			},
			"status_code": schema.Int64Attribute{
				Computed: true,
			},
			"job_id": schema.StringAttribute{
				Computed: true,
			},
			"job_status": schema.StringAttribute{
				Computed: true,
			},
			"result_json": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"lines_json": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"error": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *pruneActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *pruneActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan pruneActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mode := strings.ToLower(strings.TrimSpace(plan.Mode.ValueString()))
	switch mode {
	case "all", "containers", "images", "networks", "volumes":
	default:
		resp.Diagnostics.AddError("Invalid prune mode", "`mode` must be one of: all, containers, images, networks, volumes.")
		return
	}

	out, status, err := r.client.Prune(ctx, plan.Env.ValueString(), mode)
	if err != nil {
		resp.Diagnostics.AddError("Error running Dockhand prune action", err.Error())
		return
	}

	plan.Env = r.client.persistEnvAttr(plan.Env)
	plan.Mode = types.StringValue(mode)
	plan.StatusCode = types.Int64Value(int64(status))
	plan.Error = types.StringNull()
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", strings.TrimSpace(plan.Env.ValueString()), mode, strings.TrimSpace(plan.Trigger.ValueString())))

	jobID := strings.TrimSpace(firstString(out, "jobId", "jobID", "job_id"))
	if jobID != "" {
		plan.JobID = types.StringValue(jobID)
		job := &jobResponse{
			ID:     jobID,
			Status: "submitted",
			Result: out,
			Lines:  []jobLineResponse{},
		}

		waitForCompletion := true
		if !plan.WaitForCompletion.IsNull() && !plan.WaitForCompletion.IsUnknown() {
			waitForCompletion = plan.WaitForCompletion.ValueBool()
		}
		if waitForCompletion {
			timeoutSeconds := int64(120)
			if !plan.TimeoutSeconds.IsNull() && !plan.TimeoutSeconds.IsUnknown() && plan.TimeoutSeconds.ValueInt64() > 0 {
				timeoutSeconds = plan.TimeoutSeconds.ValueInt64()
			}
			pollIntervalMS := int64(1000)
			if !plan.PollIntervalMS.IsNull() && !plan.PollIntervalMS.IsUnknown() && plan.PollIntervalMS.ValueInt64() > 0 {
				pollIntervalMS = plan.PollIntervalMS.ValueInt64()
			}

			job, _, err = r.client.WaitForJob(
				ctx,
				jobID,
				time.Duration(timeoutSeconds)*time.Second,
				time.Duration(pollIntervalMS)*time.Millisecond,
			)
			if err != nil {
				resp.Diagnostics.AddError("Error waiting for prune job", err.Error())
				return
			}
		}

		applyJobState(&plan.JobStatus, &plan.ResultJSON, &plan.LinesJSON, job)
		plan.Success = types.BoolValue(!isFailureStatus(strings.TrimSpace(plan.JobStatus.ValueString())))
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	plan.JobID = types.StringNull()
	success := true
	if raw, ok := out["success"]; ok {
		if b, ok := raw.(bool); ok {
			success = b
		}
	}
	plan.Success = types.BoolValue(success)
	if success {
		plan.JobStatus = types.StringValue("done")
	} else {
		plan.JobStatus = types.StringValue("failed")
	}
	plan.ResultJSON = types.StringValue(mustJSON(out))
	plan.LinesJSON = types.StringValue("[]")

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pruneActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state pruneActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.JobID.IsNull() || state.JobID.IsUnknown() {
		return
	}
	jobID := strings.TrimSpace(state.JobID.ValueString())
	if jobID == "" {
		return
	}

	job, status, err := r.client.GetJob(ctx, jobID)
	if err != nil {
		if status == 404 {
			// Job records are ephemeral; retain prior state snapshot.
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand prune job", err.Error())
		return
	}

	applyJobState(&state.JobStatus, &state.ResultJSON, &state.LinesJSON, job)
	state.Success = types.BoolValue(!isFailureStatus(strings.TrimSpace(state.JobStatus.ValueString())))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *pruneActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pruneActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *pruneActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No-op one-shot action.
}

func isFailureStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "failed", "error", "cancelled", "canceled":
		return true
	default:
		return false
	}
}

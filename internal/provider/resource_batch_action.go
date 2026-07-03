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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*batchActionResource)(nil)
	_ resource.ResourceWithConfigure = (*batchActionResource)(nil)
)

func NewBatchActionResource() resource.Resource {
	return &batchActionResource{}
}

type batchActionResource struct {
	client *Client
}

type batchActionResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Env               types.String `tfsdk:"env"`
	EntityType        types.String `tfsdk:"entity_type"`
	Operation         types.String `tfsdk:"operation"`
	ItemIDs           types.List   `tfsdk:"item_ids"`
	WaitForCompletion types.Bool   `tfsdk:"wait_for_completion"`
	TimeoutSeconds    types.Int64  `tfsdk:"timeout_seconds"`
	PollIntervalMS    types.Int64  `tfsdk:"poll_interval_ms"`
	Trigger           types.String `tfsdk:"trigger"`
	JobID             types.String `tfsdk:"job_id"`
	JobStatus         types.String `tfsdk:"job_status"`
	Success           types.Bool   `tfsdk:"success"`
	ResultJSON        types.String `tfsdk:"result_json"`
	LinesJSON         types.String `tfsdk:"lines_json"`
}

func (r *batchActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_batch_action"
}

func (r *batchActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs an async Dockhand batch operation via `/api/batch` and optionally waits for terminal job status via `/api/jobs/{jobId}`. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"env": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"entity_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operation": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"item_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"wait_for_completion": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"timeout_seconds": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"poll_interval_ms": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"trigger": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"job_id":     schema.StringAttribute{Computed: true},
			"job_status": schema.StringAttribute{Computed: true},
			"success": schema.BoolAttribute{
				MarkdownDescription: "Whether the batch job completed successfully when `wait_for_completion` is true.",
				Computed:            true,
			},
			"result_json": schema.StringAttribute{Computed: true, Sensitive: true},
			"lines_json":  schema.StringAttribute{Computed: true, Sensitive: true},
		},
	}
}

func (r *batchActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *batchActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan batchActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var itemIDs []string
	resp.Diagnostics.Append(plan.ItemIDs.ElementsAs(ctx, &itemIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(itemIDs) == 0 {
		resp.Diagnostics.AddError("Invalid batch action input", "`item_ids` must contain at least one item id.")
		return
	}

	entityType := strings.TrimSpace(plan.EntityType.ValueString())
	operation := strings.TrimSpace(plan.Operation.ValueString())
	if entityType == "" || operation == "" {
		resp.Diagnostics.AddError("Invalid batch action input", "`entity_type` and `operation` must be non-empty.")
		return
	}

	submitted, _, err := r.client.SubmitBatch(ctx, plan.Env.ValueString(), entityType, operation, itemIDs)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand batch job", err.Error())
		return
	}

	plan.Env = r.client.persistEnvAttr(plan.Env)
	plan.EntityType = types.StringValue(entityType)
	plan.Operation = types.StringValue(operation)
	if strings.TrimSpace(submitted.JobID) != "" {
		plan.JobID = types.StringValue(submitted.JobID)
		plan.ID = types.StringValue(submitted.JobID)
	} else {
		plan.JobID = types.StringNull()
		plan.ID = types.StringValue(fmt.Sprintf("inline:%s:%s:%s", entityType, operation, strings.TrimSpace(plan.Trigger.ValueString())))
	}

	waitForCompletion := true
	if !plan.WaitForCompletion.IsNull() && !plan.WaitForCompletion.IsUnknown() {
		waitForCompletion = plan.WaitForCompletion.ValueBool()
	}

	job := &jobResponse{
		ID:     strings.TrimSpace(submitted.JobID),
		Status: "submitted",
		Result: map[string]any{},
		Lines:  []jobLineResponse{},
	}
	if strings.TrimSpace(submitted.JobID) == "" {
		job.Status = normalizeJobStatus(job.Status, submitted.Result)
		if s := strings.TrimSpace(submitted.Status); s != "" {
			job.Status = normalizeJobStatus(s, submitted.Result)
		}
		if submitted.Result != nil {
			job.Result = submitted.Result
		}
		if jobPayloadIndicatesFailure(job.Result) {
			job.Status = "failed"
		}
	}

	if waitForCompletion && strings.TrimSpace(submitted.JobID) != "" {
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
			submitted.JobID,
			time.Duration(timeoutSeconds)*time.Second,
			time.Duration(pollIntervalMS)*time.Millisecond,
		)
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for Dockhand batch job", err.Error())
			return
		}
	}

	applyJobState(&plan.JobStatus, &plan.ResultJSON, &plan.LinesJSON, job)
	if waitForCompletion {
		status := strings.TrimSpace(plan.JobStatus.ValueString())
		if jobPayloadIndicatesFailure(job.Result) {
			status = "failed"
			plan.JobStatus = types.StringValue(status)
		}
		plan.Success = types.BoolValue(!isFailureStatus(status))
		if isFailureStatus(status) {
			resp.Diagnostics.AddError(
				"Dockhand batch job failed",
				fmt.Sprintf("job %s finished with status %q", strings.TrimSpace(plan.JobID.ValueString()), status),
			)
			return
		}
	} else {
		plan.Success = types.BoolNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *batchActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state batchActionResourceModel
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
			// Jobs are ephemeral; keep prior one-shot action state if job record is gone.
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand batch job", err.Error())
		return
	}

	if strings.TrimSpace(job.ID) != "" {
		state.ID = types.StringValue(job.ID)
	}
	applyJobState(&state.JobStatus, &state.ResultJSON, &state.LinesJSON, job)
	waited := true
	if !state.WaitForCompletion.IsNull() && !state.WaitForCompletion.IsUnknown() {
		waited = state.WaitForCompletion.ValueBool()
	}
	if waited {
		state.Success = types.BoolValue(!isFailureStatus(strings.TrimSpace(state.JobStatus.ValueString())))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *batchActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan batchActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *batchActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No-op one-shot action.
}

func applyJobState(statusValue *types.String, resultJSONValue *types.String, linesJSONValue *types.String, job *jobResponse) {
	if statusValue != nil {
		*statusValue = types.StringValue(strings.TrimSpace(job.Status))
	}
	if resultJSONValue != nil {
		result := job.Result
		if result == nil {
			result = map[string]any{}
		}
		*resultJSONValue = types.StringValue(mustJSON(result))
	}
	if linesJSONValue != nil {
		lines := job.Lines
		if lines == nil {
			lines = []jobLineResponse{}
		}
		*linesJSONValue = types.StringValue(mustJSON(lines))
	}
}

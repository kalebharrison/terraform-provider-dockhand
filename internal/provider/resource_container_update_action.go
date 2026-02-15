package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*containerUpdateActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*containerUpdateActionResource)(nil)
	_ resource.ResourceWithImportState = (*containerUpdateActionResource)(nil)
)

func NewContainerUpdateActionResource() resource.Resource {
	return &containerUpdateActionResource{}
}

type containerUpdateActionResource struct {
	client *Client
}

type containerUpdateActionModel struct {
	ID                             types.String `tfsdk:"id"`
	Env                            types.String `tfsdk:"env"`
	ContainerID                    types.String `tfsdk:"container_id"`
	PayloadJSON                    types.String `tfsdk:"payload_json"`
	RestartPolicyName              types.String `tfsdk:"restart_policy_name"`
	RestartPolicyMaximumRetryCount types.Int64  `tfsdk:"restart_policy_maximum_retry_count"`
	CPUShares                      types.Int64  `tfsdk:"cpu_shares"`
	PidsLimit                      types.Int64  `tfsdk:"pids_limit"`
	MemoryBytes                    types.Int64  `tfsdk:"memory_bytes"`
	NanoCPUs                       types.Int64  `tfsdk:"nano_cpus"`
	Trigger                        types.String `tfsdk:"trigger"`
	ResultJSON                     types.String `tfsdk:"result_json"`
}

func (r *containerUpdateActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_update_action"
}

func (r *containerUpdateActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot container update via `/api/containers/{id}/update` with a raw JSON payload object. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"env": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"container_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"payload_json": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"restart_policy_name": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"restart_policy_maximum_retry_count": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"cpu_shares": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"pids_limit": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"memory_bytes": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"nano_cpus": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"trigger": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"result_json": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *containerUpdateActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *containerUpdateActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan containerUpdateActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	containerID := strings.TrimSpace(plan.ContainerID.ValueString())
	if containerID == "" {
		resp.Diagnostics.AddError("Invalid container ID", "`container_id` cannot be empty.")
		return
	}

	payloadRaw := strings.TrimSpace(plan.PayloadJSON.ValueString())
	payload, effectiveJSON, err := buildContainerUpdatePayload(plan, payloadRaw)
	if err != nil {
		resp.Diagnostics.AddError("Invalid update payload", err.Error())
		return
	}

	result, status, err := r.client.UpdateContainer(ctx, plan.Env.ValueString(), containerID, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand container", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error updating Dockhand container", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", plan.Env.ValueString(), containerID, plan.Trigger.ValueString()))
	plan.PayloadJSON = types.StringValue(effectiveJSON)
	plan.ResultJSON = types.StringValue(mustJSON(result))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerUpdateActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state containerUpdateActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *containerUpdateActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan containerUpdateActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerUpdateActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *containerUpdateActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<env>:<container_id>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

func buildContainerUpdatePayload(plan containerUpdateActionModel, payloadRaw string) (map[string]any, string, error) {
	merged := map[string]any{}

	if !plan.CPUShares.IsNull() && !plan.CPUShares.IsUnknown() {
		merged["CpuShares"] = plan.CPUShares.ValueInt64()
	}
	if !plan.PidsLimit.IsNull() && !plan.PidsLimit.IsUnknown() {
		merged["PidsLimit"] = plan.PidsLimit.ValueInt64()
	}
	if !plan.MemoryBytes.IsNull() && !plan.MemoryBytes.IsUnknown() {
		merged["Memory"] = plan.MemoryBytes.ValueInt64()
	}
	if !plan.NanoCPUs.IsNull() && !plan.NanoCPUs.IsUnknown() {
		merged["NanoCpus"] = plan.NanoCPUs.ValueInt64()
	}
	if !plan.RestartPolicyName.IsNull() && !plan.RestartPolicyName.IsUnknown() {
		restart := map[string]any{
			"Name": plan.RestartPolicyName.ValueString(),
		}
		if !plan.RestartPolicyMaximumRetryCount.IsNull() && !plan.RestartPolicyMaximumRetryCount.IsUnknown() {
			restart["MaximumRetryCount"] = plan.RestartPolicyMaximumRetryCount.ValueInt64()
		}
		merged["RestartPolicy"] = restart
	}

	raw := strings.TrimSpace(payloadRaw)
	if raw == "" {
		raw = "{}"
	}

	userPayload := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &userPayload); err != nil {
		return nil, "", fmt.Errorf("`payload_json` must be a valid JSON object: %w", err)
	}
	for k, v := range userPayload {
		merged[k] = v
	}
	if len(merged) == 0 {
		merged = map[string]any{}
	}
	return merged, mustJSON(merged), nil
}

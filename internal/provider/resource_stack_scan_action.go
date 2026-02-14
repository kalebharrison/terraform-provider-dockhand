package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*stackScanActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*stackScanActionResource)(nil)
	_ resource.ResourceWithImportState = (*stackScanActionResource)(nil)
)

func NewStackScanActionResource() resource.Resource {
	return &stackScanActionResource{}
}

type stackScanActionResource struct {
	client *Client
}

type stackScanActionModel struct {
	ID              types.String `tfsdk:"id"`
	Trigger         types.String `tfsdk:"trigger"`
	DiscoveredCount types.Int64  `tfsdk:"discovered_count"`
	AdoptedCount    types.Int64  `tfsdk:"adopted_count"`
	SkippedCount    types.Int64  `tfsdk:"skipped_count"`
	ErrorCount      types.Int64  `tfsdk:"error_count"`
	ResultJSON      types.String `tfsdk:"result_json"`
}

func (r *stackScanActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_scan_action"
}

func (r *stackScanActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot stack discovery scan via `/api/stacks/scan`. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"trigger": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"discovered_count": schema.Int64Attribute{Computed: true},
			"adopted_count":    schema.Int64Attribute{Computed: true},
			"skipped_count":    schema.Int64Attribute{Computed: true},
			"error_count":      schema.Int64Attribute{Computed: true},
			"result_json":      schema.StringAttribute{Computed: true},
		},
	}
}

func (r *stackScanActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *stackScanActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan stackScanActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, status, err := r.client.ScanStacks(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error scanning stacks", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error scanning stacks", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("scan:%s", plan.Trigger.ValueString()))
	plan.DiscoveredCount = types.Int64Value(int64(len(result.Discovered)))
	plan.AdoptedCount = types.Int64Value(int64(len(result.Adopted)))
	plan.SkippedCount = types.Int64Value(int64(len(result.Skipped)))
	plan.ErrorCount = types.Int64Value(int64(len(result.Errors)))
	plan.ResultJSON = types.StringValue(mustJSON(result))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackScanActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stackScanActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *stackScanActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stackScanActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackScanActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *stackScanActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `scan:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

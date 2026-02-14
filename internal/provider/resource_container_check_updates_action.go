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
	_ resource.Resource                = (*containerCheckUpdatesActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*containerCheckUpdatesActionResource)(nil)
	_ resource.ResourceWithImportState = (*containerCheckUpdatesActionResource)(nil)
)

func NewContainerCheckUpdatesActionResource() resource.Resource {
	return &containerCheckUpdatesActionResource{}
}

type containerCheckUpdatesActionResource struct {
	client *Client
}

type containerCheckUpdatesActionModel struct {
	ID           types.String `tfsdk:"id"`
	Env          types.String `tfsdk:"env"`
	Trigger      types.String `tfsdk:"trigger"`
	Total        types.Int64  `tfsdk:"total"`
	UpdatesFound types.Int64  `tfsdk:"updates_found"`
	ResultsJSON  types.String `tfsdk:"results_json"`
}

func (r *containerCheckUpdatesActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_check_updates_action"
}

func (r *containerCheckUpdatesActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot container update check via `/api/containers/check-updates`. Change `trigger` to run again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"env": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"trigger": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"total":         schema.Int64Attribute{Computed: true},
			"updates_found": schema.Int64Attribute{Computed: true},
			"results_json":  schema.StringAttribute{Computed: true},
		},
	}
}

func (r *containerCheckUpdatesActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *containerCheckUpdatesActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan containerCheckUpdatesActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, status, err := r.client.CheckContainerUpdates(ctx, plan.Env.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error checking container updates", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error checking container updates", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.Env.ValueString(), plan.Trigger.ValueString()))
	plan.Total = types.Int64Value(out.Total)
	plan.UpdatesFound = types.Int64Value(out.UpdatesFound)
	plan.ResultsJSON = types.StringValue(mustJSON(out.Results))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerCheckUpdatesActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state containerCheckUpdatesActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *containerCheckUpdatesActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan containerCheckUpdatesActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerCheckUpdatesActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *containerCheckUpdatesActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<env>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

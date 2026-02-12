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
	_ resource.Resource                = (*gitStackWebhookActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*gitStackWebhookActionResource)(nil)
	_ resource.ResourceWithImportState = (*gitStackWebhookActionResource)(nil)
)

func NewGitStackWebhookActionResource() resource.Resource {
	return &gitStackWebhookActionResource{}
}

type gitStackWebhookActionResource struct {
	client *Client
}

type gitStackWebhookActionModel struct {
	ID      types.String `tfsdk:"id"`
	StackID types.String `tfsdk:"stack_id"`
	Trigger types.String `tfsdk:"trigger"`
}

func (r *gitStackWebhookActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_stack_webhook_action"
}

func (r *gitStackWebhookActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Triggers a one-shot git stack webhook call via `/api/git/stacks/{id}/webhook`. Change `trigger` to run again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"stack_id": schema.StringAttribute{
				MarkdownDescription: "Git stack numeric ID.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"trigger": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *gitStackWebhookActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gitStackWebhookActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan gitStackWebhookActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackID := strings.TrimSpace(plan.StackID.ValueString())
	if stackID == "" {
		resp.Diagnostics.AddError("Invalid stack ID", "`stack_id` cannot be empty.")
		return
	}

	status, err := r.client.TriggerGitStackWebhook(ctx, stackID)
	if err != nil {
		resp.Diagnostics.AddError("Error triggering git stack webhook", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error triggering git stack webhook", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", stackID, plan.Trigger.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gitStackWebhookActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state gitStackWebhookActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *gitStackWebhookActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan gitStackWebhookActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gitStackWebhookActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *gitStackWebhookActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<stack_id>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

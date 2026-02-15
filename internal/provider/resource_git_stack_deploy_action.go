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
	_ resource.Resource                = (*gitStackDeployActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*gitStackDeployActionResource)(nil)
	_ resource.ResourceWithImportState = (*gitStackDeployActionResource)(nil)
)

func NewGitStackDeployActionResource() resource.Resource {
	return &gitStackDeployActionResource{}
}

type gitStackDeployActionResource struct {
	client *Client
}

type gitStackDeployActionModel struct {
	ID      types.String `tfsdk:"id"`
	StackID types.String `tfsdk:"stack_id"`
	Trigger types.String `tfsdk:"trigger"`
	Result  types.String `tfsdk:"result"`
	Output  types.String `tfsdk:"output"`
}

func (r *gitStackDeployActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_stack_deploy_action"
}

func (r *gitStackDeployActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot Git stack deploy request via `/api/git/stacks/{id}/deploy-stream`. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"stack_id": schema.StringAttribute{
				Required: true,
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
			"result": schema.StringAttribute{Computed: true},
			"output": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *gitStackDeployActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gitStackDeployActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan gitStackDeployActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackID := strings.TrimSpace(plan.StackID.ValueString())
	if stackID == "" {
		resp.Diagnostics.AddError("Invalid stack ID", "`stack_id` cannot be empty.")
		return
	}

	status, output, err := r.client.DeployGitStack(ctx, stackID)
	if err != nil {
		resp.Diagnostics.AddError("Error running Dockhand git stack deploy", err.Error())
		return
	}
	if status < 200 || status > 299 {
		msg := fmt.Sprintf("Dockhand returned status %d", status)
		if strings.TrimSpace(output) != "" {
			msg = fmt.Sprintf("%s: %s", msg, strings.TrimSpace(output))
		}
		resp.Diagnostics.AddError("Error running Dockhand git stack deploy", msg)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", stackID, plan.Trigger.ValueString()))
	plan.Result = types.StringValue("deploy_requested")
	if strings.TrimSpace(output) == "" {
		plan.Output = types.StringNull()
	} else {
		plan.Output = types.StringValue(strings.TrimSpace(output))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gitStackDeployActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state gitStackDeployActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *gitStackDeployActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan gitStackDeployActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gitStackDeployActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *gitStackDeployActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<stack_id>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

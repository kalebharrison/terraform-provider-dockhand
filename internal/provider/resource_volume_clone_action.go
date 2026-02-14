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
	_ resource.Resource                = (*volumeCloneActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*volumeCloneActionResource)(nil)
	_ resource.ResourceWithImportState = (*volumeCloneActionResource)(nil)
)

func NewVolumeCloneActionResource() resource.Resource {
	return &volumeCloneActionResource{}
}

type volumeCloneActionResource struct {
	client *Client
}

type volumeCloneActionModel struct {
	ID         types.String `tfsdk:"id"`
	Env        types.String `tfsdk:"env"`
	SourceName types.String `tfsdk:"source_name"`
	TargetName types.String `tfsdk:"target_name"`
	Trigger    types.String `tfsdk:"trigger"`
}

func (r *volumeCloneActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_clone_action"
}

func (r *volumeCloneActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot volume clone action. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"env": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_name": schema.StringAttribute{
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
		},
	}
}

func (r *volumeCloneActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *volumeCloneActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan volumeCloneActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceName := strings.TrimSpace(plan.SourceName.ValueString())
	targetName := strings.TrimSpace(plan.TargetName.ValueString())
	if sourceName == "" {
		resp.Diagnostics.AddError("Invalid source volume name", "`source_name` cannot be empty.")
		return
	}
	if targetName == "" {
		resp.Diagnostics.AddError("Invalid target volume name", "`target_name` cannot be empty.")
		return
	}

	status, err := r.client.CloneVolume(ctx, plan.Env.ValueString(), sourceName, targetName)
	if err != nil {
		resp.Diagnostics.AddError("Error cloning Dockhand volume", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error cloning Dockhand volume", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s:%s", plan.Env.ValueString(), sourceName, targetName, plan.Trigger.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeCloneActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeCloneActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *volumeCloneActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan volumeCloneActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeCloneActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state volumeCloneActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteVolume(ctx, strings.TrimSpace(state.Env.ValueString()), strings.TrimSpace(state.TargetName.ValueString()))
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting cloned Dockhand volume", err.Error())
		return
	}
}

func (r *volumeCloneActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<env>:<source_name>:<target_name>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

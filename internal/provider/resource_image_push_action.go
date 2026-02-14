package provider

import (
	"context"
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
	_ resource.Resource                = (*imagePushActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*imagePushActionResource)(nil)
	_ resource.ResourceWithImportState = (*imagePushActionResource)(nil)
)

func NewImagePushActionResource() resource.Resource {
	return &imagePushActionResource{}
}

type imagePushActionResource struct {
	client *Client
}

type imagePushActionModel struct {
	ID         types.String `tfsdk:"id"`
	Env        types.String `tfsdk:"env"`
	ImageID    types.String `tfsdk:"image_id"`
	RegistryID types.Int64  `tfsdk:"registry_id"`
	Trigger    types.String `tfsdk:"trigger"`
	Result     types.String `tfsdk:"result"`
}

func (r *imagePushActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_push_action"
}

func (r *imagePushActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot image push to a Dockhand registry. Change `trigger` to run it again.",
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
			"image_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"registry_id": schema.Int64Attribute{
				Required: true,
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
			"result": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *imagePushActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imagePushActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan imagePushActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	imageID := strings.TrimSpace(plan.ImageID.ValueString())
	if imageID == "" {
		resp.Diagnostics.AddError("Invalid image ID", "`image_id` cannot be empty.")
		return
	}
	if plan.RegistryID.IsNull() || plan.RegistryID.IsUnknown() || plan.RegistryID.ValueInt64() <= 0 {
		resp.Diagnostics.AddError("Invalid registry ID", "`registry_id` must be a positive integer.")
		return
	}

	status, err := r.client.PushImage(ctx, plan.Env.ValueString(), imageID, plan.RegistryID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Error pushing Dockhand image", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error pushing Dockhand image", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%d:%s", plan.Env.ValueString(), imageID, plan.RegistryID.ValueInt64(), plan.Trigger.ValueString()))
	plan.Result = types.StringValue("push_requested")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *imagePushActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imagePushActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *imagePushActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan imagePushActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *imagePushActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *imagePushActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<env>:<image_id>:<registry_id>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

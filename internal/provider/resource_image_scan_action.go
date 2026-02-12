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
	_ resource.Resource                = (*imageScanActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*imageScanActionResource)(nil)
	_ resource.ResourceWithImportState = (*imageScanActionResource)(nil)
)

func NewImageScanActionResource() resource.Resource {
	return &imageScanActionResource{}
}

type imageScanActionResource struct {
	client *Client
}

type imageScanActionResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Env       types.String `tfsdk:"env"`
	ImageName types.String `tfsdk:"image_name"`
	Trigger   types.String `tfsdk:"trigger"`
	Result    types.String `tfsdk:"result"`
}

func (r *imageScanActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image_scan_action"
}

func (r *imageScanActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot image vulnerability scan via `/api/images/scan`. Change `trigger` to re-run.",
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
			"image_name": schema.StringAttribute{
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
			"result": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *imageScanActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imageScanActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan imageScanActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	imageName := strings.TrimSpace(plan.ImageName.ValueString())
	if imageName == "" {
		resp.Diagnostics.AddError("Invalid image name", "`image_name` cannot be empty.")
		return
	}

	result, status, err := r.client.ScanImage(ctx, plan.Env.ValueString(), imageName)
	if err != nil {
		resp.Diagnostics.AddError("Error scanning image", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error scanning image", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	trigger := plan.Trigger.ValueString()
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", plan.Env.ValueString(), imageName, trigger))
	plan.Result = types.StringValue(result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *imageScanActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageScanActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *imageScanActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan imageScanActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *imageScanActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *imageScanActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<env>:<image_name>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

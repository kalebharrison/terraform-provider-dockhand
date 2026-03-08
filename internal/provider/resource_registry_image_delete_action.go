package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*registryImageDeleteActionResource)(nil)
	_ resource.ResourceWithConfigure = (*registryImageDeleteActionResource)(nil)
)

func NewRegistryImageDeleteActionResource() resource.Resource {
	return &registryImageDeleteActionResource{}
}

type registryImageDeleteActionResource struct {
	client *Client
}

type registryImageDeleteActionModel struct {
	ID          types.String `tfsdk:"id"`
	Registry    types.String `tfsdk:"registry"`
	Image       types.String `tfsdk:"image"`
	Tag         types.String `tfsdk:"tag"`
	FailOnError types.Bool   `tfsdk:"fail_on_error"`
	Trigger     types.String `tfsdk:"trigger"`

	Success    types.Bool   `tfsdk:"success"`
	StatusCode types.Int64  `tfsdk:"status_code"`
	Error      types.String `tfsdk:"error"`
	ResultJSON types.String `tfsdk:"result_json"`
}

func (r *registryImageDeleteActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_image_delete_action"
}

func (r *registryImageDeleteActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs one-shot remote registry image-tag delete via `DELETE /api/registry/image`. Change `trigger` to run again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"registry": schema.StringAttribute{
				MarkdownDescription: "Registry selector (ID or key expected by Dockhand `registry` query parameter).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image": schema.StringAttribute{
				MarkdownDescription: "Image repository name (for example `library/busybox`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tag": schema.StringAttribute{
				MarkdownDescription: "Image tag to delete.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fail_on_error": schema.BoolAttribute{
				MarkdownDescription: "If true (default), apply fails when Dockhand returns non-2xx.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"trigger": schema.StringAttribute{
				MarkdownDescription: "Arbitrary value that forces a new action run when changed.",
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"error": schema.StringAttribute{
				Computed: true,
			},
			"result_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *registryImageDeleteActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *registryImageDeleteActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan registryImageDeleteActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry := strings.TrimSpace(plan.Registry.ValueString())
	image := strings.TrimSpace(plan.Image.ValueString())
	tag := strings.TrimSpace(plan.Tag.ValueString())
	if registry == "" || image == "" || tag == "" {
		resp.Diagnostics.AddError("Invalid registry image delete input", "`registry`, `image`, and `tag` must be non-empty.")
		return
	}

	result, status, err := r.client.DeleteRegistryImage(ctx, registry, image, tag)
	success := err == nil

	plan.Registry = types.StringValue(registry)
	plan.Image = types.StringValue(image)
	plan.Tag = types.StringValue(tag)
	plan.Success = types.BoolValue(success)
	plan.StatusCode = types.Int64Value(int64(status))
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s:%s", registry, image, tag, strings.TrimSpace(plan.Trigger.ValueString())))

	if success {
		plan.Error = types.StringNull()
		if result == nil {
			result = map[string]any{}
		}
		plan.ResultJSON = types.StringValue(mustJSON(result))
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	errorMessage := strings.TrimSpace(err.Error())
	plan.Error = types.StringValue(errorMessage)
	plan.ResultJSON = types.StringValue(mustJSON(map[string]any{
		"error":  errorMessage,
		"status": status,
	}))

	failOnError := true
	if !plan.FailOnError.IsNull() && !plan.FailOnError.IsUnknown() {
		failOnError = plan.FailOnError.ValueBool()
	}
	if failOnError {
		resp.Diagnostics.AddError("Error deleting remote registry image tag", errorMessage)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *registryImageDeleteActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state registryImageDeleteActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *registryImageDeleteActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan registryImageDeleteActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *registryImageDeleteActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No-op one-shot action.
}

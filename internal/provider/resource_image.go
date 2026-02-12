package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*imageResource)(nil)
	_ resource.ResourceWithConfigure   = (*imageResource)(nil)
	_ resource.ResourceWithImportState = (*imageResource)(nil)
)

func NewImageResource() resource.Resource {
	return &imageResource{}
}

type imageResource struct {
	client *Client
}

type imageModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Env           types.String `tfsdk:"env"`
	ScanAfterPull types.Bool   `tfsdk:"scan_after_pull"`
	Tags          types.List   `tfsdk:"tags"`
	Size          types.Int64  `tfsdk:"size"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func (r *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (r *imageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand image via `/api/images` pull/delete endpoints.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Image ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Image reference to pull (for example `nginx:latest`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env": schema.StringAttribute{
				MarkdownDescription: "Optional environment ID. If omitted, provider `default_env` is used.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scan_after_pull": schema.BoolAttribute{
				MarkdownDescription: "Whether to trigger vulnerability scanning during pull.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolRequiresReplace{},
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Image tags reported by Dockhand.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"size": schema.Int64Attribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *imageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan imageModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := strings.TrimSpace(plan.Name.ValueString())
	if name == "" {
		resp.Diagnostics.AddError("Missing image name", "`name` must be set.")
		return
	}

	env := strings.TrimSpace(plan.Env.ValueString())
	scanAfterPull := false
	if !plan.ScanAfterPull.IsNull() && !plan.ScanAfterPull.IsUnknown() {
		scanAfterPull = plan.ScanAfterPull.ValueBool()
	}
	if _, err := r.client.PullImage(ctx, env, name, scanAfterPull); err != nil {
		resp.Diagnostics.AddError("Error pulling image", err.Error())
		return
	}

	var (
		found *imageResponse
		err   error
	)
	for range 5 {
		found, err = findImageByName(ctx, r.client, env, name)
		if err == nil && found != nil {
			break
		}
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Error pulling image", ctx.Err().Error())
			return
		case <-time.After(1200 * time.Millisecond):
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("Error reading pulled image", err.Error())
		return
	}
	if found == nil {
		resp.Diagnostics.AddError("Pulled image not found", "Dockhand pull completed but the image was not returned by `/api/images`.")
		return
	}

	state, diags := modelFromImageResponse(ctx, plan.Env, name, found)
	state.ScanAfterPull = plan.ScanAfterPull
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state imageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := strings.TrimSpace(state.Env.ValueString())
	images, _, err := r.client.ListImages(ctx, env)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand image", err.Error())
		return
	}

	var found *imageResponse
	for i := range images {
		if images[i].ID == state.ID.ValueString() {
			found = &images[i]
			break
		}
	}
	if found == nil {
		found, err = findImageMatch(images, state.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error reading Dockhand image", err.Error())
			return
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState, diags := modelFromImageResponse(ctx, state.Env, state.Name.ValueString(), found)
	newState.ScanAfterPull = state.ScanAfterPull
	resp.Diagnostics.Append(diags...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All configurable attributes require replacement.
	var state imageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state imageModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := strings.TrimSpace(state.Env.ValueString())
	id := strings.TrimSpace(state.ID.ValueString())
	if id == "" {
		match, err := findImageByName(ctx, r.client, env, state.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error locating image for delete", err.Error())
			return
		}
		if match == nil {
			return
		}
		id = match.ID
	}

	status, err := r.client.DeleteImage(ctx, env, id)
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand image", err.Error())
		return
	}
}

func (r *imageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func modelFromImageResponse(ctx context.Context, env types.String, name string, img *imageResponse) (imageModel, diag.Diagnostics) {
	out := imageModel{
		ID:   types.StringValue(img.ID),
		Env:  env,
		Name: types.StringValue(name),
		Size: types.Int64Value(img.Size),
	}

	tags := append([]string(nil), img.Tags...)
	slices.Sort(tags)
	tagsVal, diags := types.ListValueFrom(ctx, types.StringType, tags)
	out.Tags = tagsVal
	if img.Created > 0 {
		out.CreatedAt = types.StringValue(time.Unix(img.Created, 0).UTC().Format(time.RFC3339))
	} else {
		out.CreatedAt = types.StringNull()
	}

	return out, diags
}

func findImageByName(ctx context.Context, client *Client, env string, name string) (*imageResponse, error) {
	images, _, err := client.ListImages(ctx, env)
	if err != nil {
		return nil, err
	}
	return findImageMatch(images, name)
}

func findImageMatch(images []imageResponse, name string) (*imageResponse, error) {
	want := strings.TrimSpace(name)
	if want == "" {
		return nil, fmt.Errorf("image name is required")
	}

	base := want
	if i := strings.Index(want, ":"); i >= 0 {
		base = want[:i]
	}
	wantWithLatest := want
	if !strings.Contains(want, ":") {
		wantWithLatest = want + ":latest"
	}

	for i := range images {
		for _, tag := range images[i].Tags {
			if tag == want || tag == wantWithLatest || tag == "library/"+want || tag == "library/"+wantWithLatest || strings.HasPrefix(tag, base+":") {
				return &images[i], nil
			}
		}
	}
	return nil, nil
}

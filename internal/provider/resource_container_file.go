package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*containerFileResource)(nil)
	_ resource.ResourceWithConfigure = (*containerFileResource)(nil)
)

func NewContainerFileResource() resource.Resource {
	return &containerFileResource{}
}

type containerFileResource struct {
	client *Client
}

type containerFileResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Env         types.String `tfsdk:"env"`
	ContainerID types.String `tfsdk:"container_id"`
	Path        types.String `tfsdk:"path"`
	Content     types.String `tfsdk:"content"`
}

func (r *containerFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_container_file"
}

func (r *containerFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a text file inside a running container via Dockhand file APIs.",
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
			"container_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *containerFileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *containerFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan containerFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := plan.Env.ValueString()
	containerID := strings.TrimSpace(plan.ContainerID.ValueString())
	filePath := strings.TrimSpace(plan.Path.ValueString())
	content := plan.Content.ValueString()

	if containerID == "" {
		resp.Diagnostics.AddError("Invalid container ID", "`container_id` cannot be empty.")
		return
	}
	if filePath == "" {
		resp.Diagnostics.AddError("Invalid path", "`path` cannot be empty.")
		return
	}

	status, err := r.client.CreateContainerFile(ctx, env, containerID, filePath, "file")
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand container file", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error creating Dockhand container file", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	status, err = r.client.UpdateContainerFileContent(ctx, env, containerID, filePath, content)
	if err != nil {
		resp.Diagnostics.AddError("Error writing Dockhand container file", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error writing Dockhand container file", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", env, containerID, filePath))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state containerFileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	content, status, err := r.client.GetContainerFileContent(ctx, state.Env.ValueString(), state.ContainerID.ValueString(), state.Path.ValueString())
	if err != nil {
		if status == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand container file", err.Error())
		return
	}
	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error reading Dockhand container file", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	state.Content = types.StringValue(content)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *containerFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan containerFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.UpdateContainerFileContent(ctx, plan.Env.ValueString(), plan.ContainerID.ValueString(), plan.Path.ValueString(), plan.Content.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand container file", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error updating Dockhand container file", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *containerFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state containerFileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteContainerFile(ctx, state.Env.ValueString(), state.ContainerID.ValueString(), state.Path.ValueString())
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Error deleting Dockhand container file", err.Error())
		return
	}
	if status != 0 && status != http.StatusNotFound && (status < 200 || status > 299) {
		resp.Diagnostics.AddError("Error deleting Dockhand container file", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}
}

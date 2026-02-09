package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*stackResource)(nil)
	_ resource.ResourceWithConfigure   = (*stackResource)(nil)
	_ resource.ResourceWithImportState = (*stackResource)(nil)
)

func NewStackResource() resource.Resource {
	return &stackResource{}
}

type stackResource struct {
	client *Client
}

type stackResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Env     types.String `tfsdk:"env"`
	Compose types.String `tfsdk:"compose"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (r *stackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

func (r *stackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand stack using the documented /api/stacks endpoints.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Stack identifier in the format `<env>:<name>` or `<name>` when env is omitted.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Stack name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"env": schema.StringAttribute{
				MarkdownDescription: "Dockhand environment ID sent as the `env` query parameter.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compose": schema.StringAttribute{
				MarkdownDescription: "Stack Docker Compose manifest content.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the stack should be started after create and kept running.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}

func (r *stackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *stackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan stackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := plan.Env.ValueString()
	name := plan.Name.ValueString()

	if err := r.client.CreateStack(ctx, env, stackPayload{
		Name:    name,
		Compose: plan.Compose.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand stack", err.Error())
		return
	}

	// Dockhand create currently starts stacks automatically. If desired state is disabled,
	// explicitly stop it after creation.
	if !plan.Enabled.ValueBool() {
		if err := r.client.StopStack(ctx, env, name); err != nil {
			resp.Diagnostics.AddError("Error stopping Dockhand stack after create", err.Error())
			return
		}
	}

	plan.ID = types.StringValue(formatStackID(env, name))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state stackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stack, found, err := r.client.GetStackByName(ctx, state.Env.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand stack", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Keep configured compose if API doesn't return it in list responses.
	if stack.Compose != "" {
		state.Compose = types.StringValue(stack.Compose)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *stackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan stackResourceModel
	var state stackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Enabled.ValueBool() == state.Enabled.ValueBool() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	env := plan.Env.ValueString()
	name := plan.Name.ValueString()

	var err error
	if plan.Enabled.ValueBool() {
		err = r.client.StartStack(ctx, env, name)
	} else {
		err = r.client.StopStack(ctx, env, name)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand stack runtime state", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state stackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteStack(ctx, state.Env.ValueString(), state.Name.ValueString())
	if err != nil && status != 404 {
		// Dockhand may return non-2xx with a successful backend remove.
		_, found, readErr := r.client.GetStackByName(ctx, state.Env.ValueString(), state.Name.ValueString())
		if readErr != nil || found {
			resp.Diagnostics.AddError("Error deleting Dockhand stack", err.Error())
			return
		}
	}
}

func (r *stackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<name>` or `<env>:<name>`.")
		return
	}

	env := ""
	name := raw
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) == 2 {
		env = strings.TrimSpace(parts[0])
		name = strings.TrimSpace(parts[1])
	}
	if name == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Stack name cannot be empty.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), formatStackID(env, name))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	if env != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("env"), env)...)
	}
}

func formatStackID(env string, name string) string {
	if env == "" {
		return name
	}
	return env + ":" + name
}

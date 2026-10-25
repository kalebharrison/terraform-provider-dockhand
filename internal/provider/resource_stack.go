package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Env                types.String `tfsdk:"env"`
	Compose            types.String `tfsdk:"compose"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	SecretProviderID   types.Int64  `tfsdk:"secret_provider_id"`
	SecretProviderPath types.String `tfsdk:"secret_provider_path"`
	Status             types.String `tfsdk:"status"`
	ContainerIDs       types.List   `tfsdk:"container_ids"`
	ContainerCount     types.Int64  `tfsdk:"container_count"`
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
				Sensitive:           true,
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
			"secret_provider_id": schema.Int64Attribute{
				MarkdownDescription: "Dockhand secret provider id (`secretProviderId`). Omit it to leave an existing binding unchanged. Remove a previously set value to clear the binding. Binding a provider requires the Dockhand `secrets:view` permission.",
				Optional:            true,
			},
			"secret_provider_path": schema.StringAttribute{
				MarkdownDescription: "Bulk secret selector written as the `DOCKHAND_SECRET_SELECTOR` env var (the path field next to the provider dropdown). A legacy `OP_ENVIRONMENT_ID` row is replaced. Omit it to leave stack env vars unchanged.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current stack runtime status as reported by Dockhand.",
				Computed:            true,
			},
			"container_ids": schema.ListAttribute{
				MarkdownDescription: "Container IDs currently associated with the stack.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"container_count": schema.Int64Attribute{
				MarkdownDescription: "Current number of containers associated with the stack.",
				Computed:            true,
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

	env, err := r.client.requireResolvedEnv(plan.Env.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Missing environment", err.Error())
		return
	}
	plan.Env = r.client.persistEnvAttr(plan.Env)
	name := plan.Name.ValueString()

	payload := stackPayload{
		Name:             name,
		Compose:          plan.Compose.ValueString(),
		SecretProviderID: secretProviderIDFromAttrs(plan.SecretProviderID, types.Int64Null(), false),
	}
	if !plan.SecretProviderPath.IsNull() && !plan.SecretProviderPath.IsUnknown() {
		payload.EnvVars = []stackEnvVariable{{
			Key:   dockhandSecretSelectorVar,
			Value: plan.SecretProviderPath.ValueString(),
		}}
	}
	if err := r.client.CreateStack(ctx, env, payload); err != nil {
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
	if err := r.refreshStackObservedFields(ctx, env, name, &plan, nil); err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand stack after create", err.Error())
		return
	}
	if err := r.refreshManagedSecretProvider(ctx, env, name, &plan); err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand stack secret provider", err.Error())
		return
	}
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

	resolvedEnv := r.client.persistEnvAttr(state.Env)
	if !resolvedEnv.IsNull() && !resolvedEnv.IsUnknown() {
		state.Env = resolvedEnv
	}

	// Keep configured compose if API doesn't return it in list responses.
	if stack.Compose != "" {
		state.Compose = types.StringValue(stack.Compose)
	}
	state.Status = types.StringValue(stack.Status)
	state.ContainerIDs = stringSliceToListValue(stack.Containers)
	state.ContainerCount = types.Int64Value(int64(len(stack.Containers)))
	if enabled, ok := runtimeEnabledFromStatus(stack.Status); ok {
		state.Enabled = types.BoolValue(enabled)
	}
	if err := r.refreshManagedSecretProvider(ctx, state.Env.ValueString(), state.Name.ValueString(), &state); err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand stack secret provider", err.Error())
		return
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

	env := plan.Env.ValueString()
	name := plan.Name.ValueString()

	if err := r.syncStackSecretProvider(ctx, env, name, plan, state); err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand stack secret provider", err.Error())
		return
	}

	if plan.Enabled.ValueBool() != state.Enabled.ValueBool() {
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
	}

	if err := r.refreshStackObservedFields(ctx, env, name, &plan, &state); err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand stack after update", err.Error())
		return
	}
	if err := r.refreshManagedSecretProvider(ctx, env, name, &plan); err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand stack secret provider", err.Error())
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

func (r *stackResource) syncStackSecretProvider(ctx context.Context, env string, name string, plan stackResourceModel, state stackResourceModel) error {
	if !plan.SecretProviderID.Equal(state.SecretProviderID) {
		if err := r.client.UpdateStackSecretProvider(ctx, env, name, plan.Compose.ValueString(), secretProviderIDFromAttrs(plan.SecretProviderID, state.SecretProviderID, true)); err != nil {
			return err
		}
	}
	if plan.SecretProviderPath.Equal(state.SecretProviderPath) {
		return nil
	}
	clear := plan.SecretProviderPath.IsNull() || plan.SecretProviderPath.IsUnknown()
	path := ""
	if !clear {
		path = plan.SecretProviderPath.ValueString()
	}
	return r.client.SetStackSecretSelector(ctx, env, name, path, clear)
}

func (r *stackResource) refreshManagedSecretProvider(ctx context.Context, env string, name string, target *stackResourceModel) error {
	if target == nil {
		return nil
	}
	if !target.SecretProviderID.IsNull() && !target.SecretProviderID.IsUnknown() {
		id, err := r.client.GetStackSecretProviderID(ctx, env, name)
		if err != nil {
			return err
		}
		if id == nil {
			target.SecretProviderID = types.Int64Null()
		} else {
			target.SecretProviderID = types.Int64Value(*id)
		}
	}
	if target.SecretProviderPath.IsNull() || target.SecretProviderPath.IsUnknown() {
		return nil
	}
	vars, _, err := r.client.GetStackEnvVars(ctx, env, name)
	if err != nil {
		return err
	}
	current := make([]selectorVar, 0, len(vars))
	for _, item := range vars {
		current = append(current, selectorVar(item))
	}
	if value, ok := selectorValue(current); ok {
		target.SecretProviderPath = types.StringValue(value)
		return nil
	}
	target.SecretProviderPath = types.StringNull()
	return nil
}

func (r *stackResource) refreshStackObservedFields(ctx context.Context, env string, name string, target *stackResourceModel, fallback *stackResourceModel) error {
	stack, found, err := r.getStackByNameEventually(ctx, env, name)
	if err != nil {
		return err
	}
	if found {
		target.Status = types.StringValue(stack.Status)
		target.ContainerIDs = stringSliceToListValue(stack.Containers)
		target.ContainerCount = types.Int64Value(int64(len(stack.Containers)))
		if enabled, ok := runtimeEnabledFromStatus(stack.Status); ok {
			target.Enabled = types.BoolValue(enabled)
		}
		return nil
	}

	if fallback != nil {
		if !fallback.Status.IsUnknown() {
			target.Status = fallback.Status
		}
		if !fallback.ContainerIDs.IsUnknown() {
			target.ContainerIDs = fallback.ContainerIDs
		}
		if !fallback.ContainerCount.IsUnknown() {
			target.ContainerCount = fallback.ContainerCount
		}
	}

	if target.Status.IsUnknown() {
		target.Status = types.StringNull()
	}
	if target.ContainerIDs.IsUnknown() || target.ContainerIDs.IsNull() {
		target.ContainerIDs = stringSliceToListValue(nil)
	}
	if target.ContainerCount.IsUnknown() || target.ContainerCount.IsNull() {
		target.ContainerCount = types.Int64Value(0)
	}

	return nil
}

func (r *stackResource) getStackByNameEventually(ctx context.Context, env string, name string) (*stackResponse, bool, error) {
	var lastErr error
	for attempt := 0; attempt < 6; attempt++ {
		stack, found, err := r.client.GetStackByName(ctx, env, name)
		if err == nil {
			if found {
				return stack, true, nil
			}
			if attempt == 5 {
				return nil, false, nil
			}
		} else {
			lastErr = err
			if attempt == 5 {
				return nil, false, err
			}
		}

		select {
		case <-ctx.Done():
			if lastErr != nil {
				return nil, false, lastErr
			}
			return nil, false, ctx.Err()
		case <-time.After(350 * time.Millisecond):
		}
	}

	return nil, false, lastErr
}

func stringSliceToListValue(values []string) types.List {
	if len(values) == 0 {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}

	out := make([]attr.Value, 0, len(values))
	for _, item := range values {
		out = append(out, types.StringValue(item))
	}
	return types.ListValueMust(types.StringType, out)
}

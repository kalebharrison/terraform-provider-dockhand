package provider

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*stackEnvResource)(nil)
	_ resource.ResourceWithConfigure = (*stackEnvResource)(nil)
)

func NewStackEnvResource() resource.Resource {
	return &stackEnvResource{}
}

type stackEnvResource struct {
	client *Client
}

type stackEnvVariableModel struct {
	Key      types.String `tfsdk:"key"`
	Value    types.String `tfsdk:"value"`
	IsSecret types.Bool   `tfsdk:"is_secret"`
}

type stackEnvResourceModel struct {
	ID              types.String            `tfsdk:"id"`
	Env             types.String            `tfsdk:"env"`
	StackName       types.String            `tfsdk:"stack_name"`
	RawContent      types.String            `tfsdk:"raw_content"`
	SecretVariables []stackEnvVariableModel `tfsdk:"secret_variables"`
}

func (r *stackEnvResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_env"
}

func (r *stackEnvResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages stack env content and secret variables via `/api/stacks/{name}/env` and `/api/stacks/{name}/env/raw`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"env": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"stack_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"raw_content": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"secret_variables": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":       schema.StringAttribute{Required: true},
						"value":     schema.StringAttribute{Required: true},
						"is_secret": schema.BoolAttribute{Optional: true, Computed: true},
					},
				},
			},
		},
	}
}

func (r *stackEnvResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func normalizeStackEnvVars(items []stackEnvVariableModel) []stackEnvVariableModel {
	out := make([]stackEnvVariableModel, 0, len(items))
	for _, it := range items {
		key := strings.TrimSpace(it.Key.ValueString())
		if key == "" {
			continue
		}
		val := ""
		if !it.Value.IsNull() && !it.Value.IsUnknown() {
			val = it.Value.ValueString()
		}
		isSecret := true
		if !it.IsSecret.IsNull() && !it.IsSecret.IsUnknown() {
			isSecret = it.IsSecret.ValueBool()
		}
		if !isSecret {
			continue
		}
		out = append(out, stackEnvVariableModel{
			Key:      types.StringValue(key),
			Value:    types.StringValue(val),
			IsSecret: types.BoolValue(true),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key.ValueString() < out[j].Key.ValueString()
	})
	return out
}

func flattenStackEnvVars(items []stackEnvVariableModel) []stackEnvVariable {
	out := make([]stackEnvVariable, 0, len(items))
	for _, it := range normalizeStackEnvVars(items) {
		out = append(out, stackEnvVariable{
			Key:      it.Key.ValueString(),
			Value:    it.Value.ValueString(),
			IsSecret: true,
		})
	}
	return out
}

func expandStackEnvVars(items []stackEnvVariable) []stackEnvVariableModel {
	out := make([]stackEnvVariableModel, 0, len(items))
	for _, it := range items {
		if strings.TrimSpace(it.Key) == "" || !it.IsSecret {
			continue
		}
		out = append(out, stackEnvVariableModel{
			Key:      types.StringValue(it.Key),
			Value:    types.StringValue(it.Value),
			IsSecret: types.BoolValue(true),
		})
	}
	return normalizeStackEnvVars(out)
}

func mergeMaskedSecretValues(previous []stackEnvVariableModel, current []stackEnvVariableModel) []stackEnvVariableModel {
	prevByKey := make(map[string]string, len(previous))
	for _, it := range previous {
		key := strings.TrimSpace(it.Key.ValueString())
		if key == "" || it.Value.IsNull() || it.Value.IsUnknown() {
			continue
		}
		prevByKey[key] = it.Value.ValueString()
	}

	for i := range current {
		key := strings.TrimSpace(current[i].Key.ValueString())
		if key == "" || current[i].Value.IsNull() || current[i].Value.IsUnknown() {
			continue
		}
		remote := strings.TrimSpace(current[i].Value.ValueString())
		if remote == "***" {
			if prev, ok := prevByKey[key]; ok {
				current[i].Value = types.StringValue(prev)
			}
		}
	}

	return normalizeStackEnvVars(current)
}

func (r *stackEnvResource) upsert(ctx context.Context, env string, stackName string, rawContent string, vars []stackEnvVariableModel) (int, error) {
	status, err := r.client.UpdateStackEnvRaw(ctx, env, stackName, rawContent)
	if err != nil {
		return status, err
	}
	if status < 200 || status > 299 {
		return status, fmt.Errorf("raw env update returned status %d", status)
	}
	status, err = r.client.UpdateStackEnvVars(ctx, env, stackName, flattenStackEnvVars(vars))
	if err != nil {
		return status, err
	}
	if status < 200 || status > 299 {
		return status, fmt.Errorf("env vars update returned status %d", status)
	}
	return status, nil
}

func (r *stackEnvResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan stackEnvResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := plan.Env.ValueString()
	stackName := strings.TrimSpace(plan.StackName.ValueString())
	if stackName == "" {
		resp.Diagnostics.AddError("Invalid stack name", "`stack_name` cannot be empty.")
		return
	}

	raw := ""
	if !plan.RawContent.IsNull() && !plan.RawContent.IsUnknown() {
		raw = plan.RawContent.ValueString()
	}

	if _, err := r.upsert(ctx, env, stackName, raw, plan.SecretVariables); err != nil {
		resp.Diagnostics.AddError("Error updating stack env", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", env, stackName))
	plan.RawContent = types.StringValue(raw)
	plan.SecretVariables = normalizeStackEnvVars(plan.SecretVariables)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackEnvResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state stackEnvResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := state.Env.ValueString()
	stackName := state.StackName.ValueString()

	raw, status, err := r.client.GetStackEnvRaw(ctx, env, stackName)
	if err != nil {
		if status == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading stack env raw content", err.Error())
		return
	}
	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	vars, status, err := r.client.GetStackEnvVars(ctx, env, stackName)
	if err != nil {
		if status == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading stack env variables", err.Error())
		return
	}

	state.RawContent = types.StringValue(raw)
	state.SecretVariables = mergeMaskedSecretValues(state.SecretVariables, expandStackEnvVars(vars))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *stackEnvResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan stackEnvResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	raw := ""
	if !plan.RawContent.IsNull() && !plan.RawContent.IsUnknown() {
		raw = plan.RawContent.ValueString()
	}

	if _, err := r.upsert(ctx, plan.Env.ValueString(), plan.StackName.ValueString(), raw, plan.SecretVariables); err != nil {
		resp.Diagnostics.AddError("Error updating stack env", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", plan.Env.ValueString(), plan.StackName.ValueString()))
	plan.RawContent = types.StringValue(raw)
	plan.SecretVariables = normalizeStackEnvVars(plan.SecretVariables)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackEnvResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state stackEnvResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.UpdateStackEnvRaw(ctx, state.Env.ValueString(), state.StackName.ValueString(), "")
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Error clearing stack env raw content", err.Error())
		return
	}

	status, err = r.client.UpdateStackEnvVars(ctx, state.Env.ValueString(), state.StackName.ValueString(), []stackEnvVariable{})
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Error clearing stack env variables", err.Error())
		return
	}
}

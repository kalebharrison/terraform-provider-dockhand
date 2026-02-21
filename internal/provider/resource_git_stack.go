package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*gitStackResource)(nil)
	_ resource.ResourceWithConfigure   = (*gitStackResource)(nil)
	_ resource.ResourceWithImportState = (*gitStackResource)(nil)
)

func NewGitStackResource() resource.Resource {
	return &gitStackResource{}
}

type gitStackResource struct {
	client *Client
}

type gitStackModel struct {
	ID                        types.String `tfsdk:"id"`
	Env                       types.String `tfsdk:"env"`
	StackName                 types.String `tfsdk:"stack_name"`
	RepositoryID              types.String `tfsdk:"repository_id"`
	RepoName                  types.String `tfsdk:"repo_name"`
	URL                       types.String `tfsdk:"url"`
	Branch                    types.String `tfsdk:"branch"`
	CredentialID              types.String `tfsdk:"credential_id"`
	ComposePath               types.String `tfsdk:"compose_path"`
	EnvFilePath               types.String `tfsdk:"env_file_path"`
	AutoUpdateEnabled         types.Bool   `tfsdk:"auto_update_enabled"`
	AutoUpdateCron            types.String `tfsdk:"auto_update_cron"`
	WebhookEnabled            types.Bool   `tfsdk:"webhook_enabled"`
	WebhookSecretAutoGenerate types.Bool   `tfsdk:"webhook_secret_auto_generate"`
	WebhookSecret             types.String `tfsdk:"webhook_secret"`
	DeployNow                 types.Bool   `tfsdk:"deploy_now"`
	EnvVarsJSON               types.String `tfsdk:"env_vars_json"`
	LastSync                  types.String `tfsdk:"last_sync"`
	LastCommit                types.String `tfsdk:"last_commit"`
	SyncStatus                types.String `tfsdk:"sync_status"`
	SyncError                 types.String `tfsdk:"sync_error"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
	RepositoryName            types.String `tfsdk:"repository_name"`
	RepositoryURL             types.String `tfsdk:"repository_url"`
	RepositoryBranch          types.String `tfsdk:"repository_branch"`
}

func (r *gitStackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_stack"
}

func (r *gitStackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Dockhand Git-backed stacks via `/api/git/stacks` in a target environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"env": schema.StringAttribute{
				MarkdownDescription: "Dockhand environment ID used as `env` query parameter.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stack_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"repository_id": schema.StringAttribute{
				MarkdownDescription: "Existing Dockhand Git repository ID. If unset, `url` (and optional `repo_name`) are used.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repo_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"url": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"branch": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("main"),
			},
			"credential_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"compose_path": schema.StringAttribute{
				Required: true,
			},
			"env_file_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"auto_update_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"auto_update_cron": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("0 3 * * *"),
			},
			"webhook_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"webhook_secret_auto_generate": schema.BoolAttribute{
				MarkdownDescription: "When `true` and `webhook_enabled=true`, allow Dockhand to auto-generate a webhook secret if `webhook_secret` is not provided.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"webhook_secret": schema.StringAttribute{
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"deploy_now": schema.BoolAttribute{
				MarkdownDescription: "Whether to request immediate deployment when creating/updating this git stack.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"env_vars_json": schema.StringAttribute{
				MarkdownDescription: "JSON array of env vars: `[{\"key\":\"A\",\"value\":\"B\",\"isSecret\":false}]`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("[]"),
			},
			"last_sync":         schema.StringAttribute{Computed: true},
			"last_commit":       schema.StringAttribute{Computed: true},
			"sync_status":       schema.StringAttribute{Computed: true},
			"sync_error":        schema.StringAttribute{Computed: true},
			"created_at":        schema.StringAttribute{Computed: true},
			"updated_at":        schema.StringAttribute{Computed: true},
			"repository_name":   schema.StringAttribute{Computed: true},
			"repository_url":    schema.StringAttribute{Computed: true},
			"repository_branch": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *gitStackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gitStackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan gitStackModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := strings.TrimSpace(r.client.resolveEnv(plan.Env.ValueString()))
	if env == "" {
		resp.Diagnostics.AddError("Missing environment", "Set `env` on `dockhand_git_stack` or provider `default_env`.")
		return
	}

	payload, err := buildGitStackPayload(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid git stack configuration", err.Error())
		return
	}
	if err := setGitStackPayloadEnvironment(&payload, env); err != nil {
		resp.Diagnostics.AddError("Invalid environment", err.Error())
		return
	}

	created, _, err := r.client.CreateGitStack(ctx, env, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand git stack", err.Error())
		return
	}

	state := mergeGitStackState(plan, modelFromGitStackResponse(created))
	state.Env = types.StringValue(env)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gitStackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state gitStackModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := strings.TrimSpace(r.client.resolveEnv(state.Env.ValueString()))
	item, _, err := r.client.GetGitStackByID(ctx, env, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Dockhand git stack", err.Error())
		return
	}
	if item == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := mergeGitStackState(state, modelFromGitStackResponse(item))
	newState.Env = types.StringValue(env)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *gitStackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan gitStackModel
	var state gitStackModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env := strings.TrimSpace(r.client.resolveEnv(firstKnownString(plan.Env, state.Env)))
	if env == "" {
		resp.Diagnostics.AddError("Missing environment", "Set `env` on `dockhand_git_stack` or provider `default_env`.")
		return
	}

	payload, err := buildGitStackPayload(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid git stack configuration", err.Error())
		return
	}
	if err := setGitStackPayloadEnvironment(&payload, env); err != nil {
		resp.Diagnostics.AddError("Invalid environment", err.Error())
		return
	}

	updated, _, err := r.client.UpdateGitStack(ctx, env, state.ID.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand git stack", err.Error())
		return
	}

	newState := mergeGitStackState(plan, modelFromGitStackResponse(updated))
	newState.Env = types.StringValue(env)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *gitStackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state gitStackModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteGitStack(ctx, state.Env.ValueString(), state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand git stack", err.Error())
		return
	}
}

func (r *gitStackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildGitStackPayload(plan gitStackModel) (gitStackPayload, error) {
	stackName := strings.TrimSpace(plan.StackName.ValueString())
	if stackName == "" {
		return gitStackPayload{}, fmt.Errorf("stack_name is required")
	}

	composePath := strings.TrimSpace(plan.ComposePath.ValueString())
	if composePath == "" {
		return gitStackPayload{}, fmt.Errorf("compose_path is required")
	}

	autoUpdateEnabled := false
	if !plan.AutoUpdateEnabled.IsNull() && !plan.AutoUpdateEnabled.IsUnknown() {
		autoUpdateEnabled = plan.AutoUpdateEnabled.ValueBool()
	}

	autoUpdateCron := "0 3 * * *"
	if !plan.AutoUpdateCron.IsNull() && !plan.AutoUpdateCron.IsUnknown() {
		if v := strings.TrimSpace(plan.AutoUpdateCron.ValueString()); v != "" {
			autoUpdateCron = v
		}
	}

	webhookEnabled := false
	if !plan.WebhookEnabled.IsNull() && !plan.WebhookEnabled.IsUnknown() {
		webhookEnabled = plan.WebhookEnabled.ValueBool()
	}

	webhookSecretAutoGenerate := false
	if !plan.WebhookSecretAutoGenerate.IsNull() && !plan.WebhookSecretAutoGenerate.IsUnknown() {
		webhookSecretAutoGenerate = plan.WebhookSecretAutoGenerate.ValueBool()
	}

	deployNow := false
	if !plan.DeployNow.IsNull() && !plan.DeployNow.IsUnknown() {
		deployNow = plan.DeployNow.ValueBool()
	}

	payload := gitStackPayload{
		StackName:         stackName,
		ComposePath:       composePath,
		AutoUpdateEnabled: autoUpdateEnabled,
		AutoUpdateCron:    autoUpdateCron,
		WebhookEnabled:    webhookEnabled,
		DeployNow:         deployNow,
	}

	if !plan.EnvFilePath.IsNull() && !plan.EnvFilePath.IsUnknown() {
		v := strings.TrimSpace(plan.EnvFilePath.ValueString())
		if v != "" {
			payload.EnvFilePath = &v
		}
	}

	if webhookEnabled {
		if !plan.WebhookSecret.IsNull() && !plan.WebhookSecret.IsUnknown() {
			secret := strings.TrimSpace(plan.WebhookSecret.ValueString())
			payload.WebhookSecret = &secret
		} else if !webhookSecretAutoGenerate {
			// Explicitly send empty secret when webhook is enabled and auto-generation is disabled.
			empty := ""
			payload.WebhookSecret = &empty
		}
	}

	envVars, err := parseGitStackEnvVarsJSON(plan.EnvVarsJSON)
	if err != nil {
		return gitStackPayload{}, err
	}
	payload.EnvVars = envVars

	if !plan.RepositoryID.IsNull() && !plan.RepositoryID.IsUnknown() && strings.TrimSpace(plan.RepositoryID.ValueString()) != "" {
		v, err := strconv.ParseInt(strings.TrimSpace(plan.RepositoryID.ValueString()), 10, 64)
		if err != nil {
			return gitStackPayload{}, fmt.Errorf("repository_id must be numeric: %w", err)
		}
		payload.RepositoryID = &v
		return payload, nil
	}

	url := ""
	if !plan.URL.IsNull() && !plan.URL.IsUnknown() {
		url = strings.TrimSpace(plan.URL.ValueString())
	}
	if url == "" {
		return gitStackPayload{}, fmt.Errorf("set `repository_id` or `url`")
	}

	repoName := strings.TrimSpace(plan.RepoName.ValueString())
	if repoName != "" {
		payload.RepoName = &repoName
	}
	payload.URL = &url

	branch := strings.TrimSpace(plan.Branch.ValueString())
	if branch == "" {
		branch = "main"
	}
	payload.Branch = &branch

	if !plan.CredentialID.IsNull() && !plan.CredentialID.IsUnknown() && strings.TrimSpace(plan.CredentialID.ValueString()) != "" {
		v, err := strconv.ParseInt(strings.TrimSpace(plan.CredentialID.ValueString()), 10, 64)
		if err != nil {
			return gitStackPayload{}, fmt.Errorf("credential_id must be numeric: %w", err)
		}
		payload.CredentialID = &v
	}

	return payload, nil
}

func parseGitStackEnvVarsJSON(raw types.String) ([]gitStackEnvVarPayload, error) {
	if raw.IsNull() || raw.IsUnknown() {
		return nil, nil
	}
	text := strings.TrimSpace(raw.ValueString())
	if text == "" {
		return nil, nil
	}
	var vars []gitStackEnvVarPayload
	if err := json.Unmarshal([]byte(text), &vars); err != nil {
		return nil, fmt.Errorf("env_vars_json must be a JSON array of {key,value,isSecret}: %w", err)
	}
	clean := make([]gitStackEnvVarPayload, 0, len(vars))
	for _, item := range vars {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			continue
		}
		clean = append(clean, gitStackEnvVarPayload{
			Key:      key,
			Value:    item.Value,
			IsSecret: item.IsSecret,
		})
	}
	return clean, nil
}

func modelFromGitStackResponse(in *gitStackResponse) gitStackModel {
	out := gitStackModel{
		ID:                        types.StringValue(fmt.Sprintf("%d", in.ID)),
		StackName:                 types.StringValue(in.StackName),
		AutoUpdateEnabled:         types.BoolValue(in.AutoUpdate),
		WebhookEnabled:            types.BoolValue(in.WebhookEnabled),
		WebhookSecretAutoGenerate: types.BoolValue(false),
		EnvVarsJSON:               types.StringValue("[]"),
	}

	if in.EnvironmentID != nil {
		out.Env = types.StringValue(fmt.Sprintf("%d", *in.EnvironmentID))
	} else {
		out.Env = types.StringNull()
	}
	if in.RepositoryID != nil {
		out.RepositoryID = types.StringValue(fmt.Sprintf("%d", *in.RepositoryID))
	} else {
		out.RepositoryID = types.StringNull()
	}
	if in.ComposePath != nil {
		out.ComposePath = types.StringValue(*in.ComposePath)
	} else {
		out.ComposePath = types.StringNull()
	}
	if in.EnvFilePath != nil {
		out.EnvFilePath = types.StringValue(*in.EnvFilePath)
	} else {
		out.EnvFilePath = types.StringNull()
	}
	if in.AutoUpdateCron != nil {
		out.AutoUpdateCron = types.StringValue(*in.AutoUpdateCron)
	} else {
		out.AutoUpdateCron = types.StringNull()
	}
	if in.WebhookSecret != nil {
		out.WebhookSecret = types.StringValue(*in.WebhookSecret)
	} else {
		out.WebhookSecret = types.StringNull()
	}
	if in.LastSync != nil {
		out.LastSync = types.StringValue(*in.LastSync)
	} else {
		out.LastSync = types.StringNull()
	}
	if in.LastCommit != nil {
		out.LastCommit = types.StringValue(*in.LastCommit)
	} else {
		out.LastCommit = types.StringNull()
	}
	if in.SyncStatus != nil {
		out.SyncStatus = types.StringValue(*in.SyncStatus)
	} else {
		out.SyncStatus = types.StringNull()
	}
	if in.SyncError != nil {
		out.SyncError = types.StringValue(*in.SyncError)
	} else {
		out.SyncError = types.StringNull()
	}
	if in.CreatedAt != nil {
		out.CreatedAt = types.StringValue(*in.CreatedAt)
	} else {
		out.CreatedAt = types.StringNull()
	}
	if in.UpdatedAt != nil {
		out.UpdatedAt = types.StringValue(*in.UpdatedAt)
	} else {
		out.UpdatedAt = types.StringNull()
	}
	if in.Repository != nil {
		out.RepositoryName = types.StringValue(in.Repository.Name)
		out.RepositoryURL = types.StringValue(in.Repository.URL)
		if in.Repository.Branch != nil {
			out.RepositoryBranch = types.StringValue(*in.Repository.Branch)
		} else {
			out.RepositoryBranch = types.StringNull()
		}
		if in.Repository.CredentialID != nil {
			out.CredentialID = types.StringValue(fmt.Sprintf("%d", *in.Repository.CredentialID))
		} else {
			out.CredentialID = types.StringNull()
		}
	} else {
		out.RepositoryName = types.StringNull()
		out.RepositoryURL = types.StringNull()
		out.RepositoryBranch = types.StringNull()
		out.CredentialID = types.StringNull()
	}

	out.RepoName = out.RepositoryName
	out.URL = out.RepositoryURL
	out.Branch = out.RepositoryBranch
	out.DeployNow = types.BoolValue(false)

	return out
}

func mergeGitStackState(preferred gitStackModel, remote gitStackModel) gitStackModel {
	out := remote

	if (out.Env.IsNull() || out.Env.IsUnknown()) && !preferred.Env.IsNull() && !preferred.Env.IsUnknown() {
		out.Env = preferred.Env
	}
	if (out.RepositoryID.IsNull() || out.RepositoryID.IsUnknown()) && !preferred.RepositoryID.IsNull() && !preferred.RepositoryID.IsUnknown() {
		out.RepositoryID = preferred.RepositoryID
	}
	if (out.ComposePath.IsNull() || out.ComposePath.IsUnknown()) && !preferred.ComposePath.IsNull() && !preferred.ComposePath.IsUnknown() {
		out.ComposePath = preferred.ComposePath
	}
	if (out.EnvFilePath.IsNull() || out.EnvFilePath.IsUnknown()) && !preferred.EnvFilePath.IsNull() && !preferred.EnvFilePath.IsUnknown() {
		out.EnvFilePath = preferred.EnvFilePath
	}
	if (out.AutoUpdateCron.IsNull() || out.AutoUpdateCron.IsUnknown()) && !preferred.AutoUpdateCron.IsNull() && !preferred.AutoUpdateCron.IsUnknown() {
		out.AutoUpdateCron = preferred.AutoUpdateCron
	}
	if preferred.WebhookSecret.IsNull() || preferred.WebhookSecret.IsUnknown() {
		// Keep this unset unless explicitly configured in HCL.
		// Dockhand may generate a secret server-side when webhook is enabled.
		out.WebhookSecret = types.StringNull()
	} else {
		// Always prefer configured value (including empty string) to avoid
		// server-generated secret values leaking into Terraform state.
		out.WebhookSecret = preferred.WebhookSecret
	}
	if !preferred.WebhookSecretAutoGenerate.IsNull() && !preferred.WebhookSecretAutoGenerate.IsUnknown() {
		out.WebhookSecretAutoGenerate = preferred.WebhookSecretAutoGenerate
	}
	if (out.CredentialID.IsNull() || out.CredentialID.IsUnknown()) && !preferred.CredentialID.IsNull() && !preferred.CredentialID.IsUnknown() {
		out.CredentialID = preferred.CredentialID
	}
	if (out.RepoName.IsNull() || out.RepoName.IsUnknown()) && !preferred.RepoName.IsNull() && !preferred.RepoName.IsUnknown() {
		out.RepoName = preferred.RepoName
	}
	if (out.URL.IsNull() || out.URL.IsUnknown()) && !preferred.URL.IsNull() && !preferred.URL.IsUnknown() {
		out.URL = preferred.URL
	}
	if (out.Branch.IsNull() || out.Branch.IsUnknown()) && !preferred.Branch.IsNull() && !preferred.Branch.IsUnknown() {
		out.Branch = preferred.Branch
	}

	if !preferred.DeployNow.IsNull() && !preferred.DeployNow.IsUnknown() {
		out.DeployNow = preferred.DeployNow
	}
	if !preferred.EnvVarsJSON.IsNull() && !preferred.EnvVarsJSON.IsUnknown() {
		out.EnvVarsJSON = preferred.EnvVarsJSON
	}

	return out
}

func setGitStackPayloadEnvironment(payload *gitStackPayload, env string) error {
	if payload == nil {
		return nil
	}
	value := strings.TrimSpace(env)
	if value == "" {
		payload.EnvironmentID = nil
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("env must be a numeric string: %w", err)
	}
	payload.EnvironmentID = &parsed
	return nil
}

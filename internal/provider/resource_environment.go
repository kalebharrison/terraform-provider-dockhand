package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*environmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*environmentResource)(nil)
	_ resource.ResourceWithImportState = (*environmentResource)(nil)
)

func NewEnvironmentResource() resource.Resource {
	return &environmentResource{}
}

type environmentResource struct {
	client *Client
}

type environmentModel struct {
	ID types.String `tfsdk:"id"`

	Name           types.String `tfsdk:"name"`
	ConnectionType types.String `tfsdk:"connection_type"`
	Host           types.String `tfsdk:"host"`
	Port           types.Int64  `tfsdk:"port"`
	Protocol       types.String `tfsdk:"protocol"`
	SocketPath     types.String `tfsdk:"socket_path"`
	TLSSkipVerify  types.Bool   `tfsdk:"tls_skip_verify"`
	CACert         types.String `tfsdk:"ca_cert"`
	ClientCert     types.String `tfsdk:"client_cert"`
	ClientKey      types.String `tfsdk:"client_key"`
	Icon           types.String `tfsdk:"icon"`

	CollectActivity  types.Bool `tfsdk:"collect_activity"`
	CollectMetrics   types.Bool `tfsdk:"collect_metrics"`
	HighlightChanges types.Bool `tfsdk:"highlight_changes"`

	Timezone types.String `tfsdk:"timezone"`

	UpdateCheckEnabled    types.Bool   `tfsdk:"update_check_enabled"`
	UpdateCheckAutoUpdate types.Bool   `tfsdk:"update_check_auto_update"`
	UpdateCheckCron       types.String `tfsdk:"update_check_cron"`
	UpdateCheckVulnCrit   types.String `tfsdk:"update_check_vulnerability_criteria"`
	ImagePruneEnabled     types.Bool   `tfsdk:"image_prune_enabled"`
	ImagePruneCron        types.String `tfsdk:"image_prune_cron"`
	ImagePruneMode        types.String `tfsdk:"image_prune_mode"`
	VulnScanEnabled       types.Bool   `tfsdk:"vulnerability_scanning_enabled"`
	VulnScanner           types.String `tfsdk:"vulnerability_scanner"`
	EnsureGrypeInstalled  types.Bool   `tfsdk:"ensure_grype_installed"`
	EnsureTrivyInstalled  types.Bool   `tfsdk:"ensure_trivy_installed"`
	GrypeInstalled        types.Bool   `tfsdk:"grype_installed"`
	TrivyInstalled        types.Bool   `tfsdk:"trivy_installed"`
	GrypeVersion          types.String `tfsdk:"grype_version"`
	TrivyVersion          types.String `tfsdk:"trivy_version"`

	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *environmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *environmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand environment via `/api/environments`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Numeric Dockhand environment ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"connection_type": schema.StringAttribute{
				MarkdownDescription: "Environment connection type. Example: `socket`.",
				Optional:            true,
				Computed:            true,
			},
			"host": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"protocol": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"socket_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"tls_skip_verify": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"ca_cert": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded CA certificate for mTLS-enabled Docker API endpoints.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"client_cert": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded client certificate for mTLS-enabled Docker API endpoints.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"client_key": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded client private key for mTLS-enabled Docker API endpoints.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"icon": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"collect_activity": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"collect_metrics": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"highlight_changes": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"timezone": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"update_check_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"update_check_auto_update": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"update_check_cron": schema.StringAttribute{
				MarkdownDescription: "Cron schedule for environment update checks (Dockhand `/api/environments/{id}/update-check`).",
				Optional:            true,
				Computed:            true,
			},
			"update_check_vulnerability_criteria": schema.StringAttribute{
				MarkdownDescription: "Vulnerability gating criteria for auto-updates (`never`, `critical`, etc.).",
				Optional:            true,
				Computed:            true,
			},
			"image_prune_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"image_prune_cron": schema.StringAttribute{
				MarkdownDescription: "Cron schedule for automatic image pruning (Dockhand `/api/environments/{id}/image-prune`).",
				Optional:            true,
				Computed:            true,
			},
			"image_prune_mode": schema.StringAttribute{
				MarkdownDescription: "Image prune mode (`dangling` or `all`).",
				Optional:            true,
				Computed:            true,
			},
			"vulnerability_scanning_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable vulnerability scanning for this environment (Dockhand `/api/settings/scanner` with env scope).",
				Optional:            true,
				Computed:            true,
			},
			"vulnerability_scanner": schema.StringAttribute{
				MarkdownDescription: "Scanner selection for this environment: `grype`, `trivy`, or `both`.",
				Optional:            true,
				Computed:            true,
			},
			"ensure_grype_installed": schema.BoolAttribute{
				MarkdownDescription: "If true, ensures Grype scanner image is present by pulling `anchore/grype:latest` when missing.",
				Optional:            true,
				Computed:            true,
			},
			"ensure_trivy_installed": schema.BoolAttribute{
				MarkdownDescription: "If true, ensures Trivy scanner image is present by pulling `aquasec/trivy:latest` when missing.",
				Optional:            true,
				Computed:            true,
			},
			"grype_installed": schema.BoolAttribute{
				MarkdownDescription: "Whether Grype is currently available for this environment.",
				Computed:            true,
			},
			"trivy_installed": schema.BoolAttribute{
				MarkdownDescription: "Whether Trivy is currently available for this environment.",
				Computed:            true,
			},
			"grype_version": schema.StringAttribute{
				MarkdownDescription: "Reported Grype version when available.",
				Computed:            true,
			},
			"trivy_version": schema.StringAttribute{
				MarkdownDescription: "Reported Trivy version when available.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *environmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *environmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan environmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildEnvironmentPayload(plan, environmentModel{})
	if err != nil {
		resp.Diagnostics.AddError("Invalid environment configuration", err.Error())
		return
	}

	created, _, err := r.client.CreateEnvironment(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand environment", err.Error())
		return
	}

	state := modelFromEnvironmentResponse(plan, created)
	state = r.applyEnvironmentAux(ctx, state, plan, state.ID.ValueString(), &resp.Diagnostics)
	state = r.readEnvironmentAux(ctx, state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *environmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state environmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	env, status, err := r.client.GetEnvironment(ctx, state.ID.ValueString())
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand environment", err.Error())
		return
	}

	newState := modelFromEnvironmentResponse(state, env)
	newState = r.readEnvironmentAux(ctx, newState, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *environmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan environmentModel
	var state environmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Missing environment ID", "Cannot update Dockhand environment because ID is unknown.")
		return
	}

	payload, err := buildEnvironmentPayload(plan, state)
	if err != nil {
		resp.Diagnostics.AddError("Invalid environment configuration", err.Error())
		return
	}

	updated, _, err := r.client.UpdateEnvironment(ctx, id, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand environment", err.Error())
		return
	}

	newState := modelFromEnvironmentResponse(plan, updated)
	newState = r.applyEnvironmentAux(ctx, newState, plan, id, &resp.Diagnostics)
	newState = r.readEnvironmentAux(ctx, newState, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *environmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state environmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteEnvironment(ctx, state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand environment", err.Error())
		return
	}
}

func (r *environmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildEnvironmentPayload(plan environmentModel, prior environmentModel) (environmentPayload, error) {
	name := strings.TrimSpace(plan.Name.ValueString())
	if name == "" {
		return environmentPayload{}, fmt.Errorf("name is required")
	}

	connectionType := firstKnownString(plan.ConnectionType, prior.ConnectionType)
	if connectionType == "" {
		connectionType = "socket"
	}

	payload := environmentPayload{
		Name:           name,
		ConnectionType: connectionType,
	}

	if v := firstKnownString(plan.Host, prior.Host); v != "" {
		payload.Host = &v
	}
	if v := firstKnownInt64(plan.Port, prior.Port); v != 0 {
		payload.Port = &v
	}
	if v := firstKnownString(plan.Protocol, prior.Protocol); v != "" {
		payload.Protocol = &v
	}
	if v := firstKnownString(plan.SocketPath, prior.SocketPath); v != "" {
		payload.SocketPath = &v
	}
	if v, ok := firstKnownBoolPtr(plan.TLSSkipVerify, prior.TLSSkipVerify); ok {
		payload.TLSSkipVerify = &v
	}
	if v := firstKnownString(plan.CACert, prior.CACert); v != "" {
		payload.CACert = &v
	}
	if v := firstKnownString(plan.ClientCert, prior.ClientCert); v != "" {
		payload.ClientCert = &v
	}
	if v := firstKnownString(plan.ClientKey, prior.ClientKey); v != "" {
		payload.ClientKey = &v
	}
	if v := firstKnownString(plan.Icon, prior.Icon); v != "" {
		payload.Icon = &v
	}
	if v, ok := firstKnownBoolPtr(plan.CollectActivity, prior.CollectActivity); ok {
		payload.CollectActivity = &v
	}
	if v, ok := firstKnownBoolPtr(plan.CollectMetrics, prior.CollectMetrics); ok {
		payload.CollectMetrics = &v
	}
	if v, ok := firstKnownBoolPtr(plan.HighlightChanges, prior.HighlightChanges); ok {
		payload.HighlightChanges = &v
	}
	if v := firstKnownString(plan.Timezone, prior.Timezone); v != "" {
		payload.Timezone = &v
	}
	if v, ok := firstKnownBoolPtr(plan.UpdateCheckEnabled, prior.UpdateCheckEnabled); ok {
		payload.UpdateCheckEnabled = &v
	}
	if v, ok := firstKnownBoolPtr(plan.UpdateCheckAutoUpdate, prior.UpdateCheckAutoUpdate); ok {
		payload.UpdateCheckAutoUpdate = &v
	}
	if v, ok := firstKnownBoolPtr(plan.ImagePruneEnabled, prior.ImagePruneEnabled); ok {
		payload.ImagePruneEnabled = &v
	}

	if connectionType == "socket" && payload.SocketPath == nil {
		return environmentPayload{}, fmt.Errorf("socket_path is required when connection_type is \"socket\"")
	}

	return payload, nil
}

func modelFromEnvironmentResponse(prior environmentModel, in *environmentResponse) environmentModel {
	out := environmentModel{
		ID:               types.StringValue(strconv.FormatInt(in.ID, 10)),
		Name:             types.StringValue(in.Name),
		ConnectionType:   types.StringValue(in.ConnectionType),
		Port:             types.Int64Value(in.Port),
		Protocol:         types.StringValue(in.Protocol),
		TLSSkipVerify:    types.BoolValue(in.TLSSkipVerify),
		Icon:             types.StringValue(in.Icon),
		CollectActivity:  types.BoolValue(in.CollectActivity),
		CollectMetrics:   types.BoolValue(in.CollectMetrics),
		HighlightChanges: types.BoolValue(in.HighlightChanges),
	}

	if in.UpdateCheckEnabled != nil {
		out.UpdateCheckEnabled = types.BoolValue(*in.UpdateCheckEnabled)
	} else if !prior.UpdateCheckEnabled.IsNull() && !prior.UpdateCheckEnabled.IsUnknown() {
		out.UpdateCheckEnabled = prior.UpdateCheckEnabled
	} else {
		out.UpdateCheckEnabled = types.BoolNull()
	}
	if in.UpdateCheckAutoUpdate != nil {
		out.UpdateCheckAutoUpdate = types.BoolValue(*in.UpdateCheckAutoUpdate)
	} else if !prior.UpdateCheckAutoUpdate.IsNull() && !prior.UpdateCheckAutoUpdate.IsUnknown() {
		out.UpdateCheckAutoUpdate = prior.UpdateCheckAutoUpdate
	} else {
		out.UpdateCheckAutoUpdate = types.BoolNull()
	}
	if in.ImagePruneEnabled != nil {
		out.ImagePruneEnabled = types.BoolValue(*in.ImagePruneEnabled)
	} else if !prior.ImagePruneEnabled.IsNull() && !prior.ImagePruneEnabled.IsUnknown() {
		out.ImagePruneEnabled = prior.ImagePruneEnabled
	} else {
		out.ImagePruneEnabled = types.BoolNull()
	}

	if in.Host != nil {
		out.Host = types.StringValue(*in.Host)
	} else {
		out.Host = types.StringNull()
	}
	if in.SocketPath != nil {
		out.SocketPath = types.StringValue(*in.SocketPath)
	} else {
		out.SocketPath = types.StringNull()
	}
	if !prior.CACert.IsNull() && !prior.CACert.IsUnknown() {
		out.CACert = prior.CACert
	} else if in.CACert != nil && *in.CACert != "" {
		out.CACert = types.StringValue(*in.CACert)
	} else {
		out.CACert = types.StringNull()
	}
	if !prior.ClientCert.IsNull() && !prior.ClientCert.IsUnknown() {
		out.ClientCert = prior.ClientCert
	} else if in.ClientCert != nil && *in.ClientCert != "" {
		out.ClientCert = types.StringValue(*in.ClientCert)
	} else {
		out.ClientCert = types.StringNull()
	}
	if !prior.ClientKey.IsNull() && !prior.ClientKey.IsUnknown() {
		out.ClientKey = prior.ClientKey
	} else if in.ClientKey != nil && *in.ClientKey != "" {
		out.ClientKey = types.StringValue(*in.ClientKey)
	} else {
		out.ClientKey = types.StringNull()
	}
	if in.Timezone != nil {
		out.Timezone = types.StringValue(*in.Timezone)
	} else {
		if !prior.Timezone.IsNull() && !prior.Timezone.IsUnknown() {
			out.Timezone = prior.Timezone
		} else {
			out.Timezone = types.StringNull()
		}
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

	if !prior.UpdateCheckCron.IsNull() && !prior.UpdateCheckCron.IsUnknown() {
		out.UpdateCheckCron = prior.UpdateCheckCron
	} else {
		out.UpdateCheckCron = types.StringNull()
	}
	if !prior.UpdateCheckVulnCrit.IsNull() && !prior.UpdateCheckVulnCrit.IsUnknown() {
		out.UpdateCheckVulnCrit = prior.UpdateCheckVulnCrit
	} else {
		out.UpdateCheckVulnCrit = types.StringNull()
	}
	if !prior.ImagePruneCron.IsNull() && !prior.ImagePruneCron.IsUnknown() {
		out.ImagePruneCron = prior.ImagePruneCron
	} else {
		out.ImagePruneCron = types.StringNull()
	}
	if !prior.ImagePruneMode.IsNull() && !prior.ImagePruneMode.IsUnknown() {
		out.ImagePruneMode = prior.ImagePruneMode
	} else {
		out.ImagePruneMode = types.StringNull()
	}
	if !prior.VulnScanEnabled.IsNull() && !prior.VulnScanEnabled.IsUnknown() {
		out.VulnScanEnabled = prior.VulnScanEnabled
	} else {
		out.VulnScanEnabled = types.BoolNull()
	}
	if !prior.VulnScanner.IsNull() && !prior.VulnScanner.IsUnknown() {
		out.VulnScanner = prior.VulnScanner
	} else {
		out.VulnScanner = types.StringNull()
	}
	if !prior.EnsureGrypeInstalled.IsNull() && !prior.EnsureGrypeInstalled.IsUnknown() {
		out.EnsureGrypeInstalled = prior.EnsureGrypeInstalled
	} else {
		out.EnsureGrypeInstalled = types.BoolNull()
	}
	if !prior.EnsureTrivyInstalled.IsNull() && !prior.EnsureTrivyInstalled.IsUnknown() {
		out.EnsureTrivyInstalled = prior.EnsureTrivyInstalled
	} else {
		out.EnsureTrivyInstalled = types.BoolNull()
	}
	out.GrypeInstalled = types.BoolNull()
	out.TrivyInstalled = types.BoolNull()
	out.GrypeVersion = types.StringNull()
	out.TrivyVersion = types.StringNull()

	return out
}

func (r *environmentResource) applyEnvironmentAux(ctx context.Context, current environmentModel, plan environmentModel, id string, diags *diag.Diagnostics) environmentModel {
	if r.client == nil || id == "" {
		return current
	}

	// Timezone
	tz := strings.TrimSpace(firstKnownString(plan.Timezone, current.Timezone))
	if tz == "" {
		tz = "UTC"
	}
	if _, err := r.client.SetEnvironmentTimezone(ctx, id, tz); err != nil {
		diags.AddWarning("Failed to set environment timezone", err.Error())
	} else {
		current.Timezone = types.StringValue(tz)
	}

	// Update-check settings
	updateEnabled, _ := firstKnownBoolPtr(plan.UpdateCheckEnabled, current.UpdateCheckEnabled)
	updateAuto, _ := firstKnownBoolPtr(plan.UpdateCheckAutoUpdate, current.UpdateCheckAutoUpdate)
	updateCron := strings.TrimSpace(firstKnownString(plan.UpdateCheckCron, current.UpdateCheckCron))
	if updateCron == "" {
		updateCron = "0 4 * * *"
	}
	updateVuln := strings.TrimSpace(firstKnownString(plan.UpdateCheckVulnCrit, current.UpdateCheckVulnCrit))
	if updateVuln == "" {
		updateVuln = "never"
	}
	upPayload := environmentUpdateCheckPayload{
		Enabled:               updateEnabled,
		Cron:                  updateCron,
		AutoUpdate:            updateAuto,
		VulnerabilityCriteria: updateVuln,
	}
	if _, err := r.client.SetEnvironmentUpdateCheck(ctx, id, upPayload); err != nil {
		diags.AddWarning("Failed to set environment update-check settings", err.Error())
	}

	// Image prune settings
	pruneEnabled, pruneEnabledKnown := firstKnownBoolPtr(plan.ImagePruneEnabled, current.ImagePruneEnabled)
	if !pruneEnabledKnown {
		pruneEnabled = false
	}
	pruneCron := strings.TrimSpace(firstKnownString(plan.ImagePruneCron, current.ImagePruneCron))
	if pruneCron == "" {
		pruneCron = "0 3 * * 0"
	}
	pruneMode := strings.TrimSpace(firstKnownString(plan.ImagePruneMode, current.ImagePruneMode))
	if pruneMode == "" {
		pruneMode = "dangling"
	}
	prPayload := environmentImagePrunePayload{
		Enabled:        pruneEnabled,
		CronExpression: pruneCron,
		PruneMode:      pruneMode,
	}
	if _, err := r.client.SetEnvironmentImagePrune(ctx, id, prPayload); err != nil {
		diags.AddWarning("Failed to set environment image-prune settings", err.Error())
	}

	// Vulnerability scanner settings (environment-scoped).
	scanEnabled, scanEnabledKnown := firstKnownBoolPtr(plan.VulnScanEnabled, current.VulnScanEnabled)
	if !scanEnabledKnown {
		scanEnabled = false
	}
	scanner := strings.ToLower(strings.TrimSpace(firstKnownString(plan.VulnScanner, current.VulnScanner)))
	if scanner == "" {
		scanner = "both"
	}
	switch scanner {
	case "grype", "trivy", "both", "none":
	default:
		scanner = "both"
	}
	if !scanEnabled {
		scanner = "none"
	}
	if _, err := r.client.SetScannerSettings(ctx, id, scanner); err != nil {
		diags.AddWarning("Failed to set environment scanner settings", err.Error())
	}

	ensureGrype, _ := firstKnownBoolPtr(plan.EnsureGrypeInstalled, current.EnsureGrypeInstalled)
	ensureTrivy, _ := firstKnownBoolPtr(plan.EnsureTrivyInstalled, current.EnsureTrivyInstalled)
	if ensureGrype || ensureTrivy {
		scanState, _, err := r.client.GetScannerSettings(ctx, id, false)
		if err != nil {
			diags.AddWarning("Failed to read scanner availability before install checks", err.Error())
			return current
		}
		if ensureGrype && !scannerAvailability(scanState, "grype") {
			if _, err := r.client.PullImage(ctx, id, "anchore/grype:latest", false); err != nil {
				diags.AddWarning("Failed to install Grype scanner image", err.Error())
			}
		}
		if ensureTrivy && !scannerAvailability(scanState, "trivy") {
			if _, err := r.client.PullImage(ctx, id, "aquasec/trivy:latest", false); err != nil {
				diags.AddWarning("Failed to install Trivy scanner image", err.Error())
			}
		}
	}

	return current
}

func (r *environmentResource) readEnvironmentAux(ctx context.Context, current environmentModel, diags *diag.Diagnostics) environmentModel {
	if r.client == nil || current.ID.IsNull() || current.ID.IsUnknown() {
		return current
	}
	id := current.ID.ValueString()
	if id == "" {
		return current
	}

	tzResp, _, err := r.client.GetEnvironmentTimezone(ctx, id)
	if err != nil {
		diags.AddWarning("Failed to read environment timezone", err.Error())
	} else if tzResp != nil && strings.TrimSpace(tzResp.Timezone) != "" {
		current.Timezone = types.StringValue(strings.TrimSpace(tzResp.Timezone))
	}

	upResp, _, err := r.client.GetEnvironmentUpdateCheck(ctx, id)
	if err != nil {
		diags.AddWarning("Failed to read environment update-check settings", err.Error())
	} else if upResp != nil && upResp.Settings != nil {
		current.UpdateCheckEnabled = types.BoolValue(upResp.Settings.Enabled)
		current.UpdateCheckAutoUpdate = types.BoolValue(upResp.Settings.AutoUpdate)
		current.UpdateCheckCron = types.StringValue(upResp.Settings.Cron)
		current.UpdateCheckVulnCrit = types.StringValue(upResp.Settings.VulnerabilityCriteria)
	}

	prResp, _, err := r.client.GetEnvironmentImagePrune(ctx, id)
	if err != nil {
		diags.AddWarning("Failed to read environment image-prune settings", err.Error())
	} else if prResp != nil && prResp.Settings != nil {
		current.ImagePruneEnabled = types.BoolValue(prResp.Settings.Enabled)
		current.ImagePruneCron = types.StringValue(prResp.Settings.CronExpression)
		current.ImagePruneMode = types.StringValue(prResp.Settings.PruneMode)
	}

	scannerSettingsResp, _, err := r.client.GetScannerSettings(ctx, id, true)
	if err != nil {
		diags.AddWarning("Failed to read environment scanner settings", err.Error())
	} else if scannerSettingsResp != nil && scannerSettingsResp.Settings != nil {
		scanner := strings.ToLower(strings.TrimSpace(scannerSettingsResp.Settings.Scanner))
		if scanner == "" {
			scanner = "none"
		}
		current.VulnScanEnabled = types.BoolValue(scanner != "none")
		if scanner == "none" {
			current.VulnScanner = types.StringValue("both")
		} else {
			current.VulnScanner = types.StringValue(scanner)
		}
	}

	scannerFullResp, _, err := r.client.GetScannerSettings(ctx, id, false)
	if err != nil {
		diags.AddWarning("Failed to read scanner availability", err.Error())
	} else if scannerFullResp != nil {
		current.GrypeInstalled = types.BoolValue(scannerAvailability(scannerFullResp, "grype"))
		current.TrivyInstalled = types.BoolValue(scannerAvailability(scannerFullResp, "trivy"))
		grypeVersion := scannerVersion(scannerFullResp, "grype")
		if grypeVersion == "" {
			current.GrypeVersion = types.StringNull()
		} else {
			current.GrypeVersion = types.StringValue(grypeVersion)
		}
		trivyVersion := scannerVersion(scannerFullResp, "trivy")
		if trivyVersion == "" {
			current.TrivyVersion = types.StringNull()
		} else {
			current.TrivyVersion = types.StringValue(trivyVersion)
		}
	}

	return current
}

func scannerAvailability(resp *scannerSettingsResponse, key string) bool {
	if resp == nil || resp.Availability == nil {
		return false
	}
	v, ok := resp.Availability[key]
	return ok && v
}

func scannerVersion(resp *scannerSettingsResponse, key string) string {
	if resp == nil || resp.Versions == nil {
		return ""
	}
	raw, ok := resp.Versions[key]
	if !ok || raw == nil {
		return ""
	}
	if s, ok := raw.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprintf("%v", raw))
}

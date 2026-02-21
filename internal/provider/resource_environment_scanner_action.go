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
	_ resource.Resource                = (*environmentScannerActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*environmentScannerActionResource)(nil)
	_ resource.ResourceWithImportState = (*environmentScannerActionResource)(nil)
)

func NewEnvironmentScannerActionResource() resource.Resource {
	return &environmentScannerActionResource{}
}

type environmentScannerActionResource struct {
	client *Client
}

type environmentScannerActionModel struct {
	ID             types.String `tfsdk:"id"`
	Env            types.String `tfsdk:"env"`
	Action         types.String `tfsdk:"action"`
	Trigger        types.String `tfsdk:"trigger"`
	GrypeInstalled types.Bool   `tfsdk:"grype_installed"`
	TrivyInstalled types.Bool   `tfsdk:"trivy_installed"`
	GrypeVersion   types.String `tfsdk:"grype_version"`
	TrivyVersion   types.String `tfsdk:"trivy_version"`
	GrypeHasUpdate types.Bool   `tfsdk:"grype_has_update"`
	TrivyHasUpdate types.Bool   `tfsdk:"trivy_has_update"`
	ResultJSON     types.String `tfsdk:"result_json"`
}

func (r *environmentScannerActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_scanner_action"
}

func (r *environmentScannerActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs one-shot scanner operations for an environment: install/remove Grype or Trivy images, or check scanner updates. Change `trigger` to run again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"env": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"action": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"trigger": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"grype_installed": schema.BoolAttribute{Computed: true},
			"trivy_installed": schema.BoolAttribute{Computed: true},
			"grype_version":   schema.StringAttribute{Computed: true},
			"trivy_version":   schema.StringAttribute{Computed: true},
			"grype_has_update": schema.BoolAttribute{
				Computed: true,
			},
			"trivy_has_update": schema.BoolAttribute{
				Computed: true,
			},
			"result_json": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *environmentScannerActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *environmentScannerActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan environmentScannerActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	envID := strings.TrimSpace(r.client.resolveEnv(plan.Env.ValueString()))
	if envID == "" {
		resp.Diagnostics.AddError("Missing environment", "Set `env` on this action resource or provider `default_env`.")
		return
	}

	action := strings.TrimSpace(plan.Action.ValueString())
	result := map[string]any{
		"action": action,
		"env":    envID,
	}

	switch action {
	case "install_grype":
		status, err := r.client.PullImage(ctx, envID, "anchore/grype:latest", false)
		if err != nil {
			resp.Diagnostics.AddError("Error installing Grype scanner image", err.Error())
			return
		}
		result["status_code"] = status
		result["success"] = true
	case "install_trivy":
		status, err := r.client.PullImage(ctx, envID, "aquasec/trivy:latest", false)
		if err != nil {
			resp.Diagnostics.AddError("Error installing Trivy scanner image", err.Error())
			return
		}
		result["status_code"] = status
		result["success"] = true
	case "remove_grype":
		success, status, err := r.client.RemoveScannerImage(ctx, envID, "grype")
		if err != nil {
			resp.Diagnostics.AddError("Error removing Grype scanner image", err.Error())
			return
		}
		result["status_code"] = status
		result["success"] = success
	case "remove_trivy":
		success, status, err := r.client.RemoveScannerImage(ctx, envID, "trivy")
		if err != nil {
			resp.Diagnostics.AddError("Error removing Trivy scanner image", err.Error())
			return
		}
		result["status_code"] = status
		result["success"] = success
	case "check_updates":
		updates, status, err := r.client.CheckScannerUpdates(ctx, envID)
		if err != nil {
			resp.Diagnostics.AddError("Error checking scanner updates", err.Error())
			return
		}
		result["status_code"] = status
		if updates != nil {
			result["updates"] = updates.Updates
			applyScannerUpdateState(&plan, updates)
		}
	default:
		resp.Diagnostics.AddError("Invalid scanner action", fmt.Sprintf("Unsupported action %q", action))
		return
	}

	scanState, _, err := r.client.GetScannerSettings(ctx, envID, false)
	if err != nil {
		resp.Diagnostics.AddWarning("Failed to read scanner status after action", err.Error())
	} else if scanState != nil {
		plan.GrypeInstalled = types.BoolValue(scannerActionAvailability(scanState, "grype"))
		plan.TrivyInstalled = types.BoolValue(scannerActionAvailability(scanState, "trivy"))

		grypeVersion := scannerActionVersion(scanState, "grype")
		if grypeVersion == "" {
			plan.GrypeVersion = types.StringNull()
		} else {
			plan.GrypeVersion = types.StringValue(grypeVersion)
		}

		trivyVersion := scannerActionVersion(scanState, "trivy")
		if trivyVersion == "" {
			plan.TrivyVersion = types.StringNull()
		} else {
			plan.TrivyVersion = types.StringValue(trivyVersion)
		}
	}

	if action != "check_updates" {
		plan.GrypeHasUpdate = types.BoolNull()
		plan.TrivyHasUpdate = types.BoolNull()
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", envID, action, plan.Trigger.ValueString()))
	plan.Env = types.StringValue(envID)
	plan.Action = types.StringValue(action)
	plan.ResultJSON = types.StringValue(mustJSON(result))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentScannerActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state environmentScannerActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *environmentScannerActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan environmentScannerActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentScannerActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No-op one-shot action.
}

func (r *environmentScannerActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<env>:<action>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

func scannerActionAvailability(resp *scannerSettingsResponse, key string) bool {
	if resp == nil || resp.Availability == nil {
		return false
	}
	v, ok := resp.Availability[key]
	return ok && v
}

func scannerActionVersion(resp *scannerSettingsResponse, key string) string {
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

func applyScannerUpdateState(plan *environmentScannerActionModel, updates *scannerCheckUpdatesResponse) {
	if plan == nil {
		return
	}
	if updates == nil || updates.Updates == nil {
		plan.GrypeHasUpdate = types.BoolNull()
		plan.TrivyHasUpdate = types.BoolNull()
		return
	}

	if entry, ok := updates.Updates["grype"]; ok {
		plan.GrypeHasUpdate = types.BoolValue(entry.HasUpdate)
	} else {
		plan.GrypeHasUpdate = types.BoolNull()
	}
	if entry, ok := updates.Updates["trivy"]; ok {
		plan.TrivyHasUpdate = types.BoolValue(entry.HasUpdate)
	} else {
		plan.TrivyHasUpdate = types.BoolNull()
	}
}

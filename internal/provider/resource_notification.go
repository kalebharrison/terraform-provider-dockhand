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
	_ resource.Resource                = (*notificationResource)(nil)
	_ resource.ResourceWithConfigure   = (*notificationResource)(nil)
	_ resource.ResourceWithImportState = (*notificationResource)(nil)
)

func NewNotificationResource() resource.Resource {
	return &notificationResource{}
}

type notificationResource struct {
	client *Client
}

type notificationModel struct {
	ID types.String `tfsdk:"id"`

	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`

	Enabled    types.Bool `tfsdk:"enabled"`
	EventTypes types.List `tfsdk:"event_types"`

	AppriseURLs types.List `tfsdk:"apprise_urls"`

	SMTPHost          types.String `tfsdk:"smtp_host"`
	SMTPPort          types.Int64  `tfsdk:"smtp_port"`
	SMTPFromEmail     types.String `tfsdk:"smtp_from_email"`
	SMTPToEmails      types.List   `tfsdk:"smtp_to_emails"`
	SMTPUsername      types.String `tfsdk:"smtp_username"`
	SMTPPassword      types.String `tfsdk:"smtp_password"`
	SMTPUseTLS        types.Bool   `tfsdk:"smtp_use_tls"`
	SMTPStartTLS      types.Bool   `tfsdk:"smtp_starttls"`
	SMTPSkipTLSVerify types.Bool   `tfsdk:"smtp_skip_tls_verify"`

	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *notificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification"
}

func (r *notificationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Dockhand notification integration via `/api/notifications`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Numeric Dockhand notification ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": schema.StringAttribute{
				MarkdownDescription: "Notification display name.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Notification type. Known values observed: `apprise`, `smtp`.",
				Required:            true,
			},

			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether this notification is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"event_types": schema.ListAttribute{
				MarkdownDescription: "Event types that will trigger this notification. If omitted, Dockhand's defaults are used on create and then stored in state on read.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},

			"apprise_urls": schema.ListAttribute{
				MarkdownDescription: "Apprise URLs (required when `type = \"apprise\"`).",
				ElementType:         types.StringType,
				Optional:            true,
			},

			"smtp_host": schema.StringAttribute{
				MarkdownDescription: "SMTP host (required when `type = \"smtp\"`).",
				Optional:            true,
			},
			"smtp_port": schema.Int64Attribute{
				MarkdownDescription: "SMTP port (required when `type = \"smtp\"`).",
				Optional:            true,
			},
			"smtp_from_email": schema.StringAttribute{
				MarkdownDescription: "From email address (required when `type = \"smtp\"`).",
				Optional:            true,
			},
			"smtp_to_emails": schema.ListAttribute{
				MarkdownDescription: "Recipient email addresses (required when `type = \"smtp\"`).",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"smtp_username": schema.StringAttribute{
				MarkdownDescription: "SMTP username (optional when `type = \"smtp\"`).",
				Optional:            true,
			},
			"smtp_password": schema.StringAttribute{
				MarkdownDescription: "SMTP password (optional when `type = \"smtp\"`).",
				Optional:            true,
				Sensitive:           true,
			},
			"smtp_use_tls": schema.BoolAttribute{
				MarkdownDescription: "Whether to use implicit TLS (optional when `type = \"smtp\"`).",
				Optional:            true,
				Computed:            true,
			},
			"smtp_starttls": schema.BoolAttribute{
				MarkdownDescription: "Whether to use STARTTLS (optional when `type = \"smtp\"`).",
				Optional:            true,
				Computed:            true,
			},
			"smtp_skip_tls_verify": schema.BoolAttribute{
				MarkdownDescription: "Whether to skip TLS certificate verification (optional when `type = \"smtp\"`).",
				Optional:            true,
				Computed:            true,
			},

			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *notificationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *Client, got: %T", req.ProviderData))
		return
	}

	r.client = c
}

func (r *notificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan notificationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := buildNotificationPayload(ctx, plan, notificationModel{}, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, _, err := r.client.CreateNotification(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Dockhand notification", err.Error())
		return
	}

	state := modelFromNotificationResponse(plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state notificationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	n, status, err := r.client.GetNotification(ctx, state.ID.ValueString())
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand notification", err.Error())
		return
	}

	newState := modelFromNotificationResponse(state, n)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *notificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan notificationModel
	var state notificationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Missing notification ID", "Cannot update Dockhand notification because ID is unknown.")
		return
	}

	payload, diags := buildNotificationPayload(ctx, plan, state, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, _, err := r.client.UpdateNotification(ctx, id, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Dockhand notification", err.Error())
		return
	}

	newState := modelFromNotificationResponse(plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *notificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var state notificationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.DeleteNotification(ctx, state.ID.ValueString())
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand notification", err.Error())
		return
	}
}

func (r *notificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildNotificationPayload(ctx context.Context, plan notificationModel, prior notificationModel, requireConfig bool) (notificationPayload, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := strings.TrimSpace(plan.Name.ValueString())
	typ := strings.TrimSpace(plan.Type.ValueString())

	if name == "" {
		diags.AddError("Invalid notification configuration", "name is required")
		return notificationPayload{}, diags
	}
	if typ == "" {
		diags.AddError("Invalid notification configuration", "type is required")
		return notificationPayload{}, diags
	}

	payload := notificationPayload{
		Type:   typ,
		Name:   name,
		Config: map[string]any{},
	}

	// enabled
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		v := plan.Enabled.ValueBool()
		payload.Enabled = &v
	} else if !prior.Enabled.IsNull() && !prior.Enabled.IsUnknown() {
		v := prior.Enabled.ValueBool()
		payload.Enabled = &v
	}

	// event_types
	var eventTypes []string
	if !plan.EventTypes.IsNull() && !plan.EventTypes.IsUnknown() {
		diags.Append(plan.EventTypes.ElementsAs(ctx, &eventTypes, false)...)
	} else if !prior.EventTypes.IsNull() && !prior.EventTypes.IsUnknown() {
		diags.Append(prior.EventTypes.ElementsAs(ctx, &eventTypes, false)...)
	}
	if diags.HasError() {
		return notificationPayload{}, diags
	}
	if len(eventTypes) > 0 {
		payload.EventTypes = eventTypes
	}

	switch typ {
	case "apprise":
		var urls []string
		if !plan.AppriseURLs.IsNull() && !plan.AppriseURLs.IsUnknown() {
			diags.Append(plan.AppriseURLs.ElementsAs(ctx, &urls, false)...)
		} else if !prior.AppriseURLs.IsNull() && !prior.AppriseURLs.IsUnknown() {
			diags.Append(prior.AppriseURLs.ElementsAs(ctx, &urls, false)...)
		}
		if diags.HasError() {
			return notificationPayload{}, diags
		}
		if requireConfig && len(urls) == 0 {
			diags.AddError("Invalid apprise notification configuration", "apprise_urls is required and must contain at least 1 URL")
			return notificationPayload{}, diags
		}
		if len(urls) > 0 {
			payload.Config["urls"] = urls
		}

	case "smtp":
		host := firstKnownString(plan.SMTPHost, prior.SMTPHost)
		fromEmail := firstKnownString(plan.SMTPFromEmail, prior.SMTPFromEmail)

		port := firstKnownInt64(plan.SMTPPort, prior.SMTPPort)

		var toEmails []string
		if !plan.SMTPToEmails.IsNull() && !plan.SMTPToEmails.IsUnknown() {
			diags.Append(plan.SMTPToEmails.ElementsAs(ctx, &toEmails, false)...)
		} else if !prior.SMTPToEmails.IsNull() && !prior.SMTPToEmails.IsUnknown() {
			diags.Append(prior.SMTPToEmails.ElementsAs(ctx, &toEmails, false)...)
		}
		if diags.HasError() {
			return notificationPayload{}, diags
		}

		if requireConfig {
			var missing []string
			if host == "" {
				missing = append(missing, "smtp_host")
			}
			if port == 0 {
				missing = append(missing, "smtp_port")
			}
			if fromEmail == "" {
				missing = append(missing, "smtp_from_email")
			}
			if len(toEmails) == 0 {
				missing = append(missing, "smtp_to_emails")
			}
			if len(missing) > 0 {
				diags.AddError("Invalid smtp notification configuration", "Missing required attributes: "+strings.Join(missing, ", "))
				return notificationPayload{}, diags
			}
		}

		if host != "" {
			payload.Config["host"] = host
		}
		if port != 0 {
			payload.Config["port"] = port
		}
		if fromEmail != "" {
			payload.Config["from_email"] = fromEmail
		}
		if len(toEmails) > 0 {
			payload.Config["to_emails"] = toEmails
		}

		if v := firstKnownString(plan.SMTPUsername, prior.SMTPUsername); v != "" {
			payload.Config["username"] = v
		}
		if !plan.SMTPPassword.IsNull() && !plan.SMTPPassword.IsUnknown() {
			if v := plan.SMTPPassword.ValueString(); v != "" {
				payload.Config["password"] = v
			}
		} else if !prior.SMTPPassword.IsNull() && !prior.SMTPPassword.IsUnknown() {
			if v := prior.SMTPPassword.ValueString(); v != "" {
				payload.Config["password"] = v
			}
		}

		if v, ok := firstKnownBoolPtr(plan.SMTPUseTLS, prior.SMTPUseTLS); ok {
			payload.Config["use_tls"] = v
		}
		if v, ok := firstKnownBoolPtr(plan.SMTPStartTLS, prior.SMTPStartTLS); ok {
			payload.Config["starttls"] = v
		}
		if v, ok := firstKnownBoolPtr(plan.SMTPSkipTLSVerify, prior.SMTPSkipTLSVerify); ok {
			payload.Config["skip_tls_verify"] = v
		}

	default:
		diags.AddError("Invalid notification configuration", "type must be \"apprise\" or \"smtp\"")
		return notificationPayload{}, diags
	}

	// Guard against Dockhand rejecting empty config.
	if requireConfig && len(payload.Config) == 0 {
		diags.AddError("Invalid notification configuration", "config is required but was empty")
		return notificationPayload{}, diags
	}

	return payload, diags
}

func modelFromNotificationResponse(prior notificationModel, in *notificationResponse) notificationModel {
	out := notificationModel{
		ID:      types.StringValue(strconv.FormatInt(in.ID, 10)),
		Name:    types.StringValue(in.Name),
		Type:    types.StringValue(in.Type),
		Enabled: types.BoolValue(in.Enabled),
	}

	if len(in.EventTypes) > 0 {
		l, diags := types.ListValueFrom(context.Background(), types.StringType, in.EventTypes)
		if diags.HasError() {
			out.EventTypes = types.ListNull(types.StringType)
		} else {
			out.EventTypes = l
		}
	} else {
		out.EventTypes = types.ListNull(types.StringType)
	}

	switch in.Type {
	case "apprise":
		out.AppriseURLs = urlsFromConfig(in.Config)
		out.SMTPHost = types.StringNull()
		out.SMTPPort = types.Int64Null()
		out.SMTPFromEmail = types.StringNull()
		out.SMTPToEmails = types.ListNull(types.StringType)
		out.SMTPUsername = types.StringNull()
		out.SMTPPassword = types.StringNull()
		out.SMTPUseTLS = types.BoolNull()
		out.SMTPStartTLS = types.BoolNull()
		out.SMTPSkipTLSVerify = types.BoolNull()
	case "smtp":
		out.AppriseURLs = types.ListNull(types.StringType)
		applySMTPConfig(&out, prior, in.Config)
	default:
		// Unknown type: preserve prior config fields as-is to avoid flapping state.
		out.AppriseURLs = prior.AppriseURLs
		out.SMTPHost = prior.SMTPHost
		out.SMTPPort = prior.SMTPPort
		out.SMTPFromEmail = prior.SMTPFromEmail
		out.SMTPToEmails = prior.SMTPToEmails
		out.SMTPUsername = prior.SMTPUsername
		out.SMTPPassword = prior.SMTPPassword
		out.SMTPUseTLS = prior.SMTPUseTLS
		out.SMTPStartTLS = prior.SMTPStartTLS
		out.SMTPSkipTLSVerify = prior.SMTPSkipTLSVerify
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

	return out
}

func urlsFromConfig(cfg map[string]any) types.List {
	raw, ok := cfg["urls"]
	if !ok || raw == nil {
		return types.ListNull(types.StringType)
	}

	switch v := raw.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if ok && s != "" {
				out = append(out, s)
			}
		}
		l, _ := types.ListValueFrom(context.Background(), types.StringType, out)
		return l
	case []string:
		l, _ := types.ListValueFrom(context.Background(), types.StringType, v)
		return l
	default:
		return types.ListNull(types.StringType)
	}
}

func applySMTPConfig(out *notificationModel, prior notificationModel, cfg map[string]any) {
	if host, ok := cfg["host"].(string); ok {
		out.SMTPHost = types.StringValue(host)
	} else {
		out.SMTPHost = types.StringNull()
	}

	if port, ok := cfg["port"].(float64); ok {
		out.SMTPPort = types.Int64Value(int64(port))
	} else if port, ok := cfg["port"].(int64); ok {
		out.SMTPPort = types.Int64Value(port)
	} else {
		out.SMTPPort = types.Int64Null()
	}

	if v, ok := cfg["from_email"].(string); ok {
		out.SMTPFromEmail = types.StringValue(v)
	} else {
		out.SMTPFromEmail = types.StringNull()
	}

	if raw, ok := cfg["to_emails"]; ok && raw != nil {
		switch v := raw.(type) {
		case []any:
			vals := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok && s != "" {
					vals = append(vals, s)
				}
			}
			out.SMTPToEmails, _ = types.ListValueFrom(context.Background(), types.StringType, vals)
		case []string:
			out.SMTPToEmails, _ = types.ListValueFrom(context.Background(), types.StringType, v)
		default:
			out.SMTPToEmails = types.ListNull(types.StringType)
		}
	} else {
		out.SMTPToEmails = types.ListNull(types.StringType)
	}

	if v, ok := cfg["username"].(string); ok {
		out.SMTPUsername = types.StringValue(v)
	} else {
		out.SMTPUsername = types.StringNull()
	}

	// If Dockhand returns a password, prefer it; otherwise preserve prior (so plans don't
	// force replacement when Dockhand returns masked or empty secrets in the future).
	if v, ok := cfg["password"].(string); ok {
		out.SMTPPassword = types.StringValue(v)
	} else {
		out.SMTPPassword = prior.SMTPPassword
	}

	out.SMTPUseTLS = boolFromConfig(cfg, "use_tls", prior.SMTPUseTLS)
	out.SMTPStartTLS = boolFromConfig(cfg, "starttls", prior.SMTPStartTLS)
	out.SMTPSkipTLSVerify = boolFromConfig(cfg, "skip_tls_verify", prior.SMTPSkipTLSVerify)
}

func boolFromConfig(cfg map[string]any, key string, prior types.Bool) types.Bool {
	if v, ok := cfg[key].(bool); ok {
		return types.BoolValue(v)
	}
	if !prior.IsNull() && !prior.IsUnknown() {
		return prior
	}
	return types.BoolNull()
}

func firstKnownString(plan types.String, prior types.String) string {
	if !plan.IsNull() && !plan.IsUnknown() {
		return strings.TrimSpace(plan.ValueString())
	}
	if !prior.IsNull() && !prior.IsUnknown() {
		return strings.TrimSpace(prior.ValueString())
	}
	return ""
}

func firstKnownInt64(plan types.Int64, prior types.Int64) int64 {
	if !plan.IsNull() && !plan.IsUnknown() {
		return plan.ValueInt64()
	}
	if !prior.IsNull() && !prior.IsUnknown() {
		return prior.ValueInt64()
	}
	return 0
}

func firstKnownBoolPtr(plan types.Bool, prior types.Bool) (bool, bool) {
	if !plan.IsNull() && !plan.IsUnknown() {
		return plan.ValueBool(), true
	}
	if !prior.IsNull() && !prior.IsUnknown() {
		return prior.ValueBool(), true
	}
	return false, false
}

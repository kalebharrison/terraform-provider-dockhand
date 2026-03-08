package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*notificationTestActionResource)(nil)
	_ resource.ResourceWithConfigure = (*notificationTestActionResource)(nil)
)

func NewNotificationTestActionResource() resource.Resource {
	return &notificationTestActionResource{}
}

type notificationTestActionResource struct {
	client *Client
}

type notificationTestActionModel struct {
	ID             types.String `tfsdk:"id"`
	NotificationID types.String `tfsdk:"notification_id"`
	Type           types.String `tfsdk:"type"`
	ConfigJSON     types.String `tfsdk:"config_json"`
	FailOnError    types.Bool   `tfsdk:"fail_on_error"`
	Trigger        types.String `tfsdk:"trigger"`

	Success    types.Bool   `tfsdk:"success"`
	Error      types.String `tfsdk:"error"`
	Message    types.String `tfsdk:"message"`
	ResultJSON types.String `tfsdk:"result_json"`
}

func (r *notificationTestActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_test_action"
}

func (r *notificationTestActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot notification test via `POST /api/notifications/test`. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"notification_id": schema.StringAttribute{
				MarkdownDescription: "Existing Dockhand notification ID to use as test payload source.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Notification type (`apprise` or `smtp`) when `notification_id` is not used.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config_json": schema.StringAttribute{
				MarkdownDescription: "JSON object payload for `config` when `notification_id` is not used.",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fail_on_error": schema.BoolAttribute{
				MarkdownDescription: "If true (default), apply fails when Dockhand returns `success = false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"trigger": schema.StringAttribute{
				MarkdownDescription: "Arbitrary value that forces a new test run when changed.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"success": schema.BoolAttribute{
				Computed: true,
			},
			"error": schema.StringAttribute{
				Computed: true,
			},
			"message": schema.StringAttribute{
				Computed: true,
			},
			"result_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *notificationTestActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *notificationTestActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan notificationTestActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	testType, testConfig, sourceLabel, err := r.resolveTestPayload(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid notification test configuration", err.Error())
		return
	}

	out, _, err := r.client.TestNotification(ctx, testType, testConfig)
	if err != nil {
		resp.Diagnostics.AddError("Error testing Dockhand notification", err.Error())
		return
	}

	plan.Success = types.BoolValue(out.Success)
	if v := strings.TrimSpace(out.Error); v != "" {
		plan.Error = types.StringValue(v)
	} else {
		plan.Error = types.StringNull()
	}
	if v := strings.TrimSpace(out.Message); v != "" {
		plan.Message = types.StringValue(v)
	} else {
		plan.Message = types.StringNull()
	}

	plan.ResultJSON = types.StringValue(mustJSON(map[string]any{
		"type":    testType,
		"success": out.Success,
		"error":   strings.TrimSpace(out.Error),
		"message": strings.TrimSpace(out.Message),
	}))
	plan.Type = types.StringValue(testType)
	if len(testConfig) > 0 {
		plan.ConfigJSON = types.StringValue(mustJSON(testConfig))
	} else {
		plan.ConfigJSON = types.StringValue("{}")
	}
	plan.ID = types.StringValue(fmt.Sprintf("test:%s:%s", sourceLabel, strings.TrimSpace(plan.Trigger.ValueString())))

	failOnError := true
	if !plan.FailOnError.IsNull() && !plan.FailOnError.IsUnknown() {
		failOnError = plan.FailOnError.ValueBool()
	}
	if failOnError && !out.Success {
		msg := strings.TrimSpace(out.Error)
		if msg == "" {
			msg = "Dockhand notification test returned success=false"
		}
		resp.Diagnostics.AddError("Dockhand notification test failed", msg)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationTestActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationTestActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *notificationTestActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationTestActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *notificationTestActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No-op one-shot action.
}

func (r *notificationTestActionResource) resolveTestPayload(ctx context.Context, plan notificationTestActionModel) (string, map[string]any, string, error) {
	notificationID := strings.TrimSpace(plan.NotificationID.ValueString())
	inputType := strings.TrimSpace(plan.Type.ValueString())
	inputConfig := strings.TrimSpace(plan.ConfigJSON.ValueString())

	if notificationID != "" {
		n, _, err := r.client.GetNotification(ctx, notificationID)
		if err != nil {
			return "", nil, "", err
		}
		if strings.TrimSpace(n.Type) == "" {
			return "", nil, "", fmt.Errorf("notification %s has empty type", notificationID)
		}
		cfg := n.Config
		if cfg == nil {
			cfg = map[string]any{}
		}
		return strings.TrimSpace(n.Type), cfg, "id-" + notificationID, nil
	}

	if inputType == "" {
		return "", nil, "", fmt.Errorf("set `notification_id` or provide `type`")
	}
	if inputConfig == "" {
		return "", nil, "", fmt.Errorf("`config_json` is required when `notification_id` is not set")
	}

	var cfg map[string]any
	if err := json.Unmarshal([]byte(inputConfig), &cfg); err != nil {
		return "", nil, "", fmt.Errorf("config_json must be a valid JSON object: %w", err)
	}
	if cfg == nil {
		cfg = map[string]any{}
	}

	return inputType, cfg, "type-" + inputType, nil
}

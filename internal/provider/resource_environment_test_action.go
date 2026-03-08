package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*environmentTestActionResource)(nil)
	_ resource.ResourceWithConfigure = (*environmentTestActionResource)(nil)
)

func NewEnvironmentTestActionResource() resource.Resource {
	return &environmentTestActionResource{}
}

type environmentTestActionResource struct {
	client *Client
}

type environmentTestActionModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ConnectionType types.String `tfsdk:"connection_type"`
	AgentToken     types.String `tfsdk:"agent_token"`
	Host           types.String `tfsdk:"host"`
	Port           types.Int64  `tfsdk:"port"`
	Protocol       types.String `tfsdk:"protocol"`
	SocketPath     types.String `tfsdk:"socket_path"`
	TLSSkipVerify  types.Bool   `tfsdk:"tls_skip_verify"`
	CACert         types.String `tfsdk:"ca_cert"`
	ClientCert     types.String `tfsdk:"client_cert"`
	ClientKey      types.String `tfsdk:"client_key"`
	FailOnError    types.Bool   `tfsdk:"fail_on_error"`
	Trigger        types.String `tfsdk:"trigger"`

	Success       types.Bool   `tfsdk:"success"`
	Error         types.String `tfsdk:"error"`
	InfoJSON      types.String `tfsdk:"info_json"`
	HawserJSON    types.String `tfsdk:"hawser_json"`
	ServerVersion types.String `tfsdk:"server_version"`
	DaemonName    types.String `tfsdk:"daemon_name"`
	Containers    types.Int64  `tfsdk:"containers"`
	Images        types.Int64  `tfsdk:"images"`
}

func (r *environmentTestActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_test_action"
}

func (r *environmentTestActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs one-shot environment connectivity validation via `/api/environments/test`. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Optional environment name label sent to Dockhand.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"connection_type": schema.StringAttribute{
				MarkdownDescription: "Connection type (`direct`, `socket`, or `agent`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_token": schema.StringAttribute{
				MarkdownDescription: "Agent token used when `connection_type = \"agent\"`.",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "Docker API host (required for `connection_type = \"direct\"`).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Docker API port (required for `connection_type = \"direct\"`).",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Docker API protocol (`http` or `https`) for direct connections.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"socket_path": schema.StringAttribute{
				MarkdownDescription: "Socket path (required for `connection_type = \"socket\"`).",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tls_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate validation when testing HTTPS direct connections.",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"ca_cert": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded CA certificate (`tlsCa`).",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_cert": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded client certificate (`tlsCert`).",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_key": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded client private key (`tlsKey`).",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fail_on_error": schema.BoolAttribute{
				MarkdownDescription: "If true (default), Terraform apply fails when Dockhand returns `success = false`.",
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
			"info_json": schema.StringAttribute{
				Computed: true,
			},
			"hawser_json": schema.StringAttribute{
				Computed: true,
			},
			"server_version": schema.StringAttribute{
				Computed: true,
			},
			"daemon_name": schema.StringAttribute{
				Computed: true,
			},
			"containers": schema.Int64Attribute{
				Computed: true,
			},
			"images": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (r *environmentTestActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *environmentTestActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan environmentTestActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildEnvironmentTestPayload(plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid environment test configuration", err.Error())
		return
	}

	out, _, err := r.client.TestEnvironmentConnection(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error testing Dockhand environment connection", err.Error())
		return
	}

	plan.Success = types.BoolValue(out.Success)
	if msg := strings.TrimSpace(out.Error); msg != "" {
		plan.Error = types.StringValue(msg)
	} else {
		plan.Error = types.StringNull()
	}
	plan.InfoJSON = types.StringValue(mustJSON(out.Info))
	plan.HawserJSON = types.StringValue(mustJSON(out.Hawser))

	serverVersion := strings.TrimSpace(firstString(out.Info, "serverVersion", "version"))
	if serverVersion == "" {
		plan.ServerVersion = types.StringNull()
	} else {
		plan.ServerVersion = types.StringValue(serverVersion)
	}

	daemonName := strings.TrimSpace(firstString(out.Info, "name", "daemonName"))
	if daemonName == "" {
		plan.DaemonName = types.StringNull()
	} else {
		plan.DaemonName = types.StringValue(daemonName)
	}

	if count := firstInt64(out.Info, "containers", "containerCount"); count > 0 {
		plan.Containers = types.Int64Value(count)
	} else {
		plan.Containers = types.Int64Value(0)
	}
	if count := firstInt64(out.Info, "images", "imageCount"); count > 0 {
		plan.Images = types.Int64Value(count)
	} else {
		plan.Images = types.Int64Value(0)
	}

	connectionType := strings.TrimSpace(plan.ConnectionType.ValueString())
	target := strings.TrimSpace(plan.Host.ValueString())
	if target == "" {
		target = strings.TrimSpace(plan.SocketPath.ValueString())
	}
	if target == "" {
		target = "n/a"
	}
	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", connectionType, target, strings.TrimSpace(plan.Trigger.ValueString())))
	plan.ConnectionType = types.StringValue(connectionType)

	failOnError := true
	if !plan.FailOnError.IsNull() && !plan.FailOnError.IsUnknown() {
		failOnError = plan.FailOnError.ValueBool()
	}
	if failOnError && !out.Success {
		msg := strings.TrimSpace(out.Error)
		if msg == "" {
			msg = "Dockhand returned success=false"
		}
		resp.Diagnostics.AddError("Dockhand environment test failed", msg)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentTestActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state environmentTestActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *environmentTestActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan environmentTestActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *environmentTestActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No-op one-shot action.
}

func buildEnvironmentTestPayload(plan environmentTestActionModel) (environmentPayload, error) {
	connectionType := strings.TrimSpace(plan.ConnectionType.ValueString())
	if connectionType == "" {
		return environmentPayload{}, fmt.Errorf("connection_type is required")
	}

	payload := environmentPayload{
		ConnectionType: normalizeEnvironmentConnectionTypeForAPI(connectionType),
	}

	if v := strings.TrimSpace(plan.Name.ValueString()); v != "" {
		payload.Name = v
	}
	if v := strings.TrimSpace(plan.AgentToken.ValueString()); v != "" {
		payload.HawserToken = &v
	}
	if v := strings.TrimSpace(plan.Host.ValueString()); v != "" {
		payload.Host = &v
	}
	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		port := plan.Port.ValueInt64()
		payload.Port = &port
	}
	if v := strings.TrimSpace(plan.Protocol.ValueString()); v != "" {
		payload.Protocol = &v
	}
	if v := strings.TrimSpace(plan.SocketPath.ValueString()); v != "" {
		payload.SocketPath = &v
	}
	if !plan.TLSSkipVerify.IsNull() && !plan.TLSSkipVerify.IsUnknown() {
		v := plan.TLSSkipVerify.ValueBool()
		payload.TLSSkipVerify = &v
	}
	if v := strings.TrimSpace(plan.CACert.ValueString()); v != "" {
		payload.CACert = &v
	}
	if v := strings.TrimSpace(plan.ClientCert.ValueString()); v != "" {
		payload.ClientCert = &v
	}
	if v := strings.TrimSpace(plan.ClientKey.ValueString()); v != "" {
		payload.ClientKey = &v
	}

	switch strings.ToLower(connectionType) {
	case "direct":
		if payload.Host == nil || strings.TrimSpace(*payload.Host) == "" {
			return environmentPayload{}, fmt.Errorf("host is required when connection_type is \"direct\"")
		}
		if payload.Port == nil || *payload.Port <= 0 {
			return environmentPayload{}, fmt.Errorf("port must be > 0 when connection_type is \"direct\"")
		}
		if payload.Protocol == nil || strings.TrimSpace(*payload.Protocol) == "" {
			return environmentPayload{}, fmt.Errorf("protocol is required when connection_type is \"direct\"")
		}
	case "socket":
		if payload.SocketPath == nil || strings.TrimSpace(*payload.SocketPath) == "" {
			return environmentPayload{}, fmt.Errorf("socket_path is required when connection_type is \"socket\"")
		}
	default:
		if isAgentConnectionType(connectionType) && (payload.HawserToken == nil || strings.TrimSpace(*payload.HawserToken) == "") {
			return environmentPayload{}, fmt.Errorf("agent_token is required when connection_type is %q", connectionType)
		}
	}

	return payload, nil
}

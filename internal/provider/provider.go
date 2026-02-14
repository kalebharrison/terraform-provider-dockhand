package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*dockhandProvider)(nil)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dockhandProvider{
			version: version,
		}
	}
}

type dockhandProvider struct {
	version string
}

type dockhandProviderModel struct {
	Endpoint     types.String `tfsdk:"endpoint"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	MFAToken     types.String `tfsdk:"mfa_token"`
	AuthProvider types.String `tfsdk:"auth_provider"`
	DefaultEnv   types.String `tfsdk:"default_env"`
	Insecure     types.Bool   `tfsdk:"insecure"`
}

func (p *dockhandProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dockhand"
	resp.Version = p.version
}

func (p *dockhandProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Dockhand API base URL. Can also be set with `DOCKHAND_ENDPOINT`.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Dockhand username for login-based auth. Can also be set with `DOCKHAND_USERNAME`.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Dockhand password for login-based auth. Can also be set with `DOCKHAND_PASSWORD`.",
				Optional:            true,
				Sensitive:           true,
			},
			"mfa_token": schema.StringAttribute{
				MarkdownDescription: "Optional MFA token for login-based auth. Can also be set with `DOCKHAND_MFA_TOKEN`.",
				Optional:            true,
				Sensitive:           true,
			},
			"auth_provider": schema.StringAttribute{
				MarkdownDescription: "Auth provider id for login-based auth (e.g. `local`). Can also be set with `DOCKHAND_AUTH_PROVIDER`.",
				Optional:            true,
			},
			"default_env": schema.StringAttribute{
				MarkdownDescription: "Default Dockhand environment ID sent as `env` query parameter when omitted by resources. Can also be set with `DOCKHAND_DEFAULT_ENV`.",
				Optional:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Disable TLS verification for API requests. Useful only for local development.",
				Optional:            true,
			},
		},
	}
}

func (p *dockhandProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config dockhandProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("DOCKHAND_ENDPOINT")
	if !config.Endpoint.IsNull() && !config.Endpoint.IsUnknown() {
		endpoint = config.Endpoint.ValueString()
	}

	username := os.Getenv("DOCKHAND_USERNAME")
	if !config.Username.IsNull() && !config.Username.IsUnknown() {
		username = config.Username.ValueString()
	}

	password := os.Getenv("DOCKHAND_PASSWORD")
	if !config.Password.IsNull() && !config.Password.IsUnknown() {
		password = config.Password.ValueString()
	}

	mfaToken := os.Getenv("DOCKHAND_MFA_TOKEN")
	if !config.MFAToken.IsNull() && !config.MFAToken.IsUnknown() {
		mfaToken = config.MFAToken.ValueString()
	}

	authProvider := os.Getenv("DOCKHAND_AUTH_PROVIDER")
	if !config.AuthProvider.IsNull() && !config.AuthProvider.IsUnknown() {
		authProvider = config.AuthProvider.ValueString()
	}
	if authProvider == "" {
		authProvider = "local"
	}

	defaultEnv := os.Getenv("DOCKHAND_DEFAULT_ENV")
	if !config.DefaultEnv.IsNull() && !config.DefaultEnv.IsUnknown() {
		defaultEnv = config.DefaultEnv.ValueString()
	}

	insecure := false
	if !config.Insecure.IsNull() && !config.Insecure.IsUnknown() {
		insecure = config.Insecure.ValueBool()
	}

	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Dockhand endpoint",
			"Set `endpoint` in the provider block or export `DOCKHAND_ENDPOINT`.",
		)
		return
	}

	// Fail fast on missing auth rather than letting every resource/data source call fail later.
	if username == "" && password == "" {
		resp.Diagnostics.AddError(
			"Missing Dockhand authentication",
			"Set provider `username` and `password` (or export `DOCKHAND_USERNAME`/`DOCKHAND_PASSWORD`).",
		)
		return
	}
	if username != "" && password == "" {
		resp.Diagnostics.AddError(
			"Incomplete Dockhand authentication",
			"`username` was set but `password` was not. Set both `username` and `password`.",
		)
		return
	}
	if password != "" && username == "" {
		resp.Diagnostics.AddError(
			"Incomplete Dockhand authentication",
			"`password` was set but `username` was not. Set both `username` and `password`.",
		)
		return
	}

	sessionCookie, err := Login(ctx, endpoint, username, password, mfaToken, authProvider, insecure)
	if err != nil {
		resp.Diagnostics.AddError("Authentication failed", err.Error())
		return
	}

	client, err := NewClient(endpoint, sessionCookie, defaultEnv, insecure)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid provider configuration",
			fmt.Sprintf("Unable to create Dockhand client: %s", err),
		)
		return
	}

	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *dockhandProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewStackResource,
		NewUserResource,
		NewGeneralSettingsResource,
		NewAuthSettingsResource,
		NewLicenseResource,
		NewRegistryResource,
		NewGitCredentialResource,
		NewGitRepositoryResource,
		NewGitStackWebhookActionResource,
		NewConfigSetResource,
		NewNotificationResource,
		NewEnvironmentResource,
		NewNetworkResource,
		NewNetworkConnectionActionResource,
		NewVolumeResource,
		NewVolumeCloneActionResource,
		NewImageResource,
		NewImagePushActionResource,
		NewImageScanActionResource,
		NewContainerResource,
		NewContainerActionResource,
		NewContainerRenameActionResource,
		NewContainerUpdateActionResource,
		NewContainerCheckUpdatesActionResource,
		NewScheduleResource,
		NewStackActionResource,
		NewStackScanActionResource,
		NewStackAdoptActionResource,
	}
}

func (p *dockhandProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewHealthDataSource,
		NewActivityDataSource,
		NewHawserStatusDataSource,
		NewAuthProvidersDataSource,
		NewSchedulesDataSource,
		NewSchedulesExecutionsDataSource,
		NewContainersDataSource,
		NewContainerStatsDataSource,
		NewContainerPendingUpdatesDataSource,
		NewStackSourcesDataSource,
		NewContainerLogsDataSource,
		NewContainerInspectDataSource,
		NewStacksDataSource,
	}
}

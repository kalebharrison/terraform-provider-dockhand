package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL       *url.URL
	httpClient    *http.Client
	sessionCookie string
	defaultEnv    string
}

type stackPayload struct {
	Name    string `json:"name"`
	Compose string `json:"compose"`
}

type stackResponse struct {
	Name    string `json:"name"`
	Compose string `json:"compose"`
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

type userPayload struct {
	Username    string  `json:"username"`
	Password    *string `json:"password,omitempty"`
	Email       *string `json:"email,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
	IsAdmin     bool    `json:"isAdmin"`
	IsActive    bool    `json:"isActive"`
}

type userResponse struct {
	ID          int64   `json:"id"`
	Username    string  `json:"username"`
	Email       *string `json:"email"`
	DisplayName *string `json:"displayName"`
	MFAEnabled  bool    `json:"mfaEnabled"`
	IsAdmin     bool    `json:"isAdmin"`
	IsActive    bool    `json:"isActive"`
	LastLogin   *string `json:"lastLogin"`
	CreatedAt   *string `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt"`
}

type registryResponse struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	URL            string  `json:"url"`
	Username       *string `json:"username"`
	IsDefault      bool    `json:"isDefault"`
	CreatedAt      *string `json:"createdAt"`
	UpdatedAt      *string `json:"updatedAt"`
	HasCredentials bool    `json:"hasCredentials"`
}

type gitCredentialPayload struct {
	Name     string  `json:"name"`
	AuthType string  `json:"authType"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	SSHKey   *string `json:"sshKey,omitempty"`
}

type gitCredentialResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	AuthType    string  `json:"authType"`
	Username    *string `json:"username"`
	HasPassword bool    `json:"hasPassword"`
	HasSSHKey   bool    `json:"hasSshKey"`
	CreatedAt   *string `json:"createdAt"`
	UpdatedAt   *string `json:"updatedAt"`
}

type gitRepositoryPayload struct {
	Name               string  `json:"name"`
	URL                string  `json:"url"`
	Branch             *string `json:"branch,omitempty"`
	ComposePath        *string `json:"composePath,omitempty"`
	CredentialID       *int64  `json:"credentialId,omitempty"`
	EnvironmentID      *int64  `json:"environmentId,omitempty"`
	AutoUpdate         *bool   `json:"autoUpdate,omitempty"`
	AutoUpdateSchedule *string `json:"autoUpdateSchedule,omitempty"`
	AutoUpdateCron     *string `json:"autoUpdateCron,omitempty"`
	WebhookEnabled     *bool   `json:"webhookEnabled,omitempty"`
}

type gitRepositoryResponse struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	URL                string  `json:"url"`
	Branch             *string `json:"branch"`
	ComposePath        *string `json:"composePath"`
	CredentialID       *int64  `json:"credentialId"`
	EnvironmentID      *int64  `json:"environmentId"`
	AutoUpdate         bool    `json:"autoUpdate"`
	AutoUpdateSchedule *string `json:"autoUpdateSchedule"`
	AutoUpdateCron     *string `json:"autoUpdateCron"`
	WebhookEnabled     bool    `json:"webhookEnabled"`
	WebhookSecret      *string `json:"webhookSecret"`
	LastSync           *string `json:"lastSync"`
	LastCommit         *string `json:"lastCommit"`
	SyncStatus         *string `json:"syncStatus"`
	SyncError          *string `json:"syncError"`
	CreatedAt          *string `json:"createdAt"`
	UpdatedAt          *string `json:"updatedAt"`
}

type configSetKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type configSetPort struct {
	ContainerPort int64  `json:"containerPort"`
	HostPort      int64  `json:"hostPort"`
	Protocol      string `json:"protocol"`
}

type configSetVolume struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Type     string `json:"type"`
	ReadOnly bool   `json:"readOnly"`
}

type configSetPayload struct {
	Name          string            `json:"name"`
	Description   *string           `json:"description,omitempty"`
	EnvVars       []configSetKV     `json:"envVars,omitempty"`
	Labels        []configSetKV     `json:"labels,omitempty"`
	Ports         []configSetPort   `json:"ports,omitempty"`
	Volumes       []configSetVolume `json:"volumes,omitempty"`
	NetworkMode   *string           `json:"networkMode,omitempty"`
	RestartPolicy *string           `json:"restartPolicy,omitempty"`
}

type configSetResponse struct {
	ID            int64             `json:"id"`
	Name          string            `json:"name"`
	Description   *string           `json:"description"`
	EnvVars       []configSetKV     `json:"envVars"`
	Labels        []configSetKV     `json:"labels"`
	Ports         []configSetPort   `json:"ports"`
	Volumes       []configSetVolume `json:"volumes"`
	NetworkMode   string            `json:"networkMode"`
	RestartPolicy string            `json:"restartPolicy"`
	CreatedAt     *string           `json:"createdAt"`
	UpdatedAt     *string           `json:"updatedAt"`
}

type notificationPayload struct {
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Enabled    *bool          `json:"enabled,omitempty"`
	EventTypes []string       `json:"eventTypes,omitempty"`
	Config     map[string]any `json:"config"`
}

type notificationResponse struct {
	ID         int64          `json:"id"`
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Enabled    bool           `json:"enabled"`
	Config     map[string]any `json:"config"`
	EventTypes []string       `json:"eventTypes"`
	CreatedAt  *string        `json:"createdAt"`
	UpdatedAt  *string        `json:"updatedAt"`
}

type environmentPayload struct {
	Name                  string  `json:"name"`
	ConnectionType        string  `json:"connectionType"`
	Host                  *string `json:"host,omitempty"`
	Port                  *int64  `json:"port,omitempty"`
	Protocol              *string `json:"protocol,omitempty"`
	SocketPath            *string `json:"socketPath,omitempty"`
	TLSSkipVerify         *bool   `json:"tlsSkipVerify,omitempty"`
	Icon                  *string `json:"icon,omitempty"`
	CollectActivity       *bool   `json:"collectActivity,omitempty"`
	CollectMetrics        *bool   `json:"collectMetrics,omitempty"`
	HighlightChanges      *bool   `json:"highlightChanges,omitempty"`
	Timezone              *string `json:"timezone,omitempty"`
	UpdateCheckEnabled    *bool   `json:"updateCheckEnabled,omitempty"`
	UpdateCheckAutoUpdate *bool   `json:"updateCheckAutoUpdate,omitempty"`
	ImagePruneEnabled     *bool   `json:"imagePruneEnabled,omitempty"`
}

type environmentResponse struct {
	ID                    int64    `json:"id"`
	Name                  string   `json:"name"`
	ConnectionType        string   `json:"connectionType"`
	Host                  *string  `json:"host"`
	Port                  int64    `json:"port"`
	Protocol              string   `json:"protocol"`
	SocketPath            *string  `json:"socketPath"`
	TLSSkipVerify         bool     `json:"tlsSkipVerify"`
	Icon                  string   `json:"icon"`
	CollectActivity       bool     `json:"collectActivity"`
	CollectMetrics        bool     `json:"collectMetrics"`
	HighlightChanges      bool     `json:"highlightChanges"`
	Timezone              *string  `json:"timezone"`
	UpdateCheckEnabled    bool     `json:"updateCheckEnabled"`
	UpdateCheckAutoUpdate bool     `json:"updateCheckAutoUpdate"`
	ImagePruneEnabled     bool     `json:"imagePruneEnabled"`
	CreatedAt             *string  `json:"createdAt"`
	UpdatedAt             *string  `json:"updatedAt"`
	Labels                []string `json:"labels"`
}

type authSettingsResponse struct {
	ID              int64   `json:"id"`
	AuthEnabled     bool    `json:"authEnabled"`
	DefaultProvider string  `json:"defaultProvider"`
	SessionTimeout  int64   `json:"sessionTimeout"`
	CreatedAt       *string `json:"createdAt"`
	UpdatedAt       *string `json:"updatedAt"`
}

type authSettingsPayload struct {
	AuthEnabled     bool   `json:"authEnabled"`
	DefaultProvider string `json:"defaultProvider"`
	SessionTimeout  int64  `json:"sessionTimeout"`
}

type authProviderItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type authProvidersResponse struct {
	DefaultProvider string             `json:"defaultProvider"`
	Providers       []authProviderItem `json:"providers"`
}

type generalSettings struct {
	ConfirmDestructive        bool     `json:"confirmDestructive"`
	DarkTheme                 string   `json:"darkTheme"`
	DateFormat                string   `json:"dateFormat"`
	DefaultGrypeArgs          string   `json:"defaultGrypeArgs"`
	DefaultTimezone           string   `json:"defaultTimezone"`
	DefaultTrivyArgs          string   `json:"defaultTrivyArgs"`
	DownloadFormat            string   `json:"downloadFormat"`
	EditorFont                string   `json:"editorFont"`
	EventCleanupCron          string   `json:"eventCleanupCron"`
	EventCleanupEnabled       bool     `json:"eventCleanupEnabled"`
	EventCollectionMode       string   `json:"eventCollectionMode"`
	EventPollInterval         int64    `json:"eventPollInterval"`
	EventRetentionDays        int64    `json:"eventRetentionDays"`
	ExternalStackPaths        []string `json:"externalStackPaths"`
	Font                      string   `json:"font"`
	FontSize                  string   `json:"fontSize"`
	GridFontSize              string   `json:"gridFontSize"`
	HighlightUpdates          bool     `json:"highlightUpdates"`
	LightTheme                string   `json:"lightTheme"`
	LogBufferSizeKb           int64    `json:"logBufferSizeKb"`
	MetricsCollectionInterval int64    `json:"metricsCollectionInterval"`
	PrimaryStackLocation      *string  `json:"primaryStackLocation"`
	ScheduleCleanupCron       string   `json:"scheduleCleanupCron"`
	ScheduleCleanupEnabled    bool     `json:"scheduleCleanupEnabled"`
	ScheduleRetentionDays     int64    `json:"scheduleRetentionDays"`
	ShowStoppedContainers     bool     `json:"showStoppedContainers"`
	TerminalFont              string   `json:"terminalFont"`
	TimeFormat                string   `json:"timeFormat"`
}

func NewClient(endpoint string, sessionCookie string, defaultEnv string, insecure bool) (*Client, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if sessionCookie == "" {
		return nil, fmt.Errorf("session cookie is required")
	}

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecure,
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Client{
		baseURL: parsed,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		sessionCookie: sessionCookie,
		defaultEnv:    defaultEnv,
	}, nil
}

func (c *Client) GetGeneralSettings(ctx context.Context) (*generalSettings, int, error) {
	var out generalSettings
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/settings/general", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateGeneralSettings(ctx context.Context, payload generalSettings) (*generalSettings, int, error) {
	var out generalSettings
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/settings/general", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) ListRegistries(ctx context.Context) ([]registryResponse, int, error) {
	var out []registryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/registries", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetRegistry(ctx context.Context, id string) (*registryResponse, int, error) {
	var out registryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/registries/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateRegistry(ctx context.Context, payload map[string]any) (*registryResponse, int, error) {
	var out registryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/registries", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateRegistry(ctx context.Context, id string, payload map[string]any) (*registryResponse, int, error) {
	var out registryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/registries/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteRegistry(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/registries/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) ListGitCredentials(ctx context.Context) ([]gitCredentialResponse, int, error) {
	var out []gitCredentialResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/git/credentials", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetGitCredential(ctx context.Context, id string) (*gitCredentialResponse, int, error) {
	var out gitCredentialResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/git/credentials/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateGitCredential(ctx context.Context, payload gitCredentialPayload) (*gitCredentialResponse, int, error) {
	var out gitCredentialResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/git/credentials", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateGitCredential(ctx context.Context, id string, payload gitCredentialPayload) (*gitCredentialResponse, int, error) {
	var out gitCredentialResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/git/credentials/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteGitCredential(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/git/credentials/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) ListGitRepositories(ctx context.Context) ([]gitRepositoryResponse, int, error) {
	var out []gitRepositoryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/git/repositories", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetGitRepository(ctx context.Context, id string) (*gitRepositoryResponse, int, error) {
	var out gitRepositoryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/git/repositories/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateGitRepository(ctx context.Context, payload gitRepositoryPayload) (*gitRepositoryResponse, int, error) {
	var out gitRepositoryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/git/repositories", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateGitRepository(ctx context.Context, id string, payload gitRepositoryPayload) (*gitRepositoryResponse, int, error) {
	var out gitRepositoryResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/git/repositories/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteGitRepository(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/git/repositories/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) ListConfigSets(ctx context.Context) ([]configSetResponse, int, error) {
	var out []configSetResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/config-sets", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetConfigSet(ctx context.Context, id string) (*configSetResponse, int, error) {
	var out configSetResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/config-sets/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateConfigSet(ctx context.Context, payload configSetPayload) (*configSetResponse, int, error) {
	var out configSetResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/config-sets", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateConfigSet(ctx context.Context, id string, payload configSetPayload) (*configSetResponse, int, error) {
	var out configSetResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/config-sets/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteConfigSet(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/config-sets/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) ListNotifications(ctx context.Context) ([]notificationResponse, int, error) {
	var out []notificationResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/notifications", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetNotification(ctx context.Context, id string) (*notificationResponse, int, error) {
	var out notificationResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/notifications/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateNotification(ctx context.Context, payload notificationPayload) (*notificationResponse, int, error) {
	var out notificationResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/notifications", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateNotification(ctx context.Context, id string, payload notificationPayload) (*notificationResponse, int, error) {
	var out notificationResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/notifications/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteNotification(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/notifications/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) ListEnvironments(ctx context.Context) ([]environmentResponse, int, error) {
	var out []environmentResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/environments", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return out, status, nil
}

func (c *Client) GetEnvironment(ctx context.Context, id string) (*environmentResponse, int, error) {
	var out environmentResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/environments/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateEnvironment(ctx context.Context, payload environmentPayload) (*environmentResponse, int, error) {
	var out environmentResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/environments", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateEnvironment(ctx context.Context, id string, payload environmentPayload) (*environmentResponse, int, error) {
	var out environmentResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/environments/"+url.PathEscape(id), nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) DeleteEnvironment(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/environments/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) GetAuthSettings(ctx context.Context) (*authSettingsResponse, int, error) {
	var out authSettingsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/auth/settings", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateAuthSettings(ctx context.Context, payload authSettingsPayload) (*authSettingsResponse, int, error) {
	var out authSettingsResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/auth/settings", nil, payload, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) GetAuthProviders(ctx context.Context) (*authProvidersResponse, int, error) {
	var out authProvidersResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/auth/providers", nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) CreateStack(ctx context.Context, env string, payload stackPayload) error {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	if _, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/stacks", query, payload, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) ListStacks(ctx context.Context, env string) ([]stackResponse, int, error) {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	var raw json.RawMessage
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/stacks", query, nil, &raw)
	if err != nil {
		return nil, status, err
	}

	stacks, parseErr := parseStacks(raw)
	if parseErr != nil {
		return nil, status, parseErr
	}

	return stacks, status, nil
}

func (c *Client) GetStackByName(ctx context.Context, env string, name string) (*stackResponse, bool, error) {
	stacks, _, err := c.ListStacks(ctx, env)
	if err != nil {
		return nil, false, err
	}

	for i := range stacks {
		if stacks[i].Name == name {
			return &stacks[i], true, nil
		}
	}

	return nil, false, nil
}

func (c *Client) StartStack(ctx context.Context, env string, name string) error {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	if _, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/stacks/"+url.PathEscape(name)+"/start", query, nil, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) StopStack(ctx context.Context, env string, name string) error {
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	if _, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/stacks/"+url.PathEscape(name)+"/stop", query, nil, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteStack(ctx context.Context, env string, name string) (int, error) {
	query := map[string]string{
		"force": "true",
	}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/stacks/"+url.PathEscape(name), query, nil, nil)
}

func (c *Client) Health(ctx context.Context, env string) (*healthResponse, error) {
	// Dockhand docs do not expose a dedicated health endpoint.
	// We treat a successful dashboard stats request as API health.
	query := map[string]string{}
	if resolvedEnv := c.resolveEnv(env); resolvedEnv != "" {
		query["env"] = resolvedEnv
	}

	if _, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/dashboard/stats", query, nil, nil); err != nil {
		return nil, err
	}

	return &healthResponse{Status: "ok"}, nil
}

func (c *Client) CreateUser(ctx context.Context, payload userPayload) (*userResponse, error) {
	var out userResponse
	if _, err := c.doJSONWithStatus(ctx, http.MethodPost, "/api/users", nil, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*userResponse, int, error) {
	var out userResponse
	status, err := c.doJSONWithStatus(ctx, http.MethodGet, "/api/users/"+url.PathEscape(id), nil, nil, &out)
	if err != nil {
		return nil, status, err
	}
	return &out, status, nil
}

func (c *Client) UpdateUser(ctx context.Context, id string, payload userPayload) (*userResponse, error) {
	var out userResponse
	if _, err := c.doJSONWithStatus(ctx, http.MethodPut, "/api/users/"+url.PathEscape(id), nil, payload, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteUser(ctx context.Context, id string) (int, error) {
	return c.doJSONWithStatus(ctx, http.MethodDelete, "/api/users/"+url.PathEscape(id), nil, nil, nil)
}

func (c *Client) doJSONWithStatus(ctx context.Context, method string, path string, query map[string]string, in any, out any) (int, error) {
	var payloadBytes []byte
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return 0, err
		}
		payloadBytes = data
	}

	// Build the URL once; the request itself may be retried.
	ref := &url.URL{Path: path}
	if len(query) > 0 {
		values := url.Values{}
		for k, v := range query {
			if v != "" {
				values.Set(k, v)
			}
		}
		ref.RawQuery = values.Encode()
	}
	fullURL := c.baseURL.ResolveReference(ref).String()

	var lastStatus int
	var responseBody []byte

	for attempt := 0; attempt < 3; attempt++ {
		var body io.Reader
		if payloadBytes != nil {
			body = bytes.NewReader(payloadBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
		if err != nil {
			return 0, err
		}

		req.Header.Set("Accept", "application/json")
		if payloadBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.sessionCookie != "" {
			req.Header.Set("Cookie", c.sessionCookie)
		}

		res, err := c.httpClient.Do(req)
		if err != nil {
			if shouldRetry(method, 0, err) && attempt < 2 {
				if sleepErr := sleepBackoff(ctx, attempt); sleepErr != nil {
					return 0, err
				}
				continue
			}
			return 0, err
		}

		lastStatus = res.StatusCode

		// On errors, keep the body very small to avoid huge allocations in diagnostics.
		limit := int64(10 << 20) // 10 MiB
		if res.StatusCode < 200 || res.StatusCode > 299 {
			limit = 64 << 10 // 64 KiB
		}

		responseBody, err = io.ReadAll(io.LimitReader(res.Body, limit))
		res.Body.Close()
		if err != nil {
			if shouldRetry(method, lastStatus, err) && attempt < 2 {
				if sleepErr := sleepBackoff(ctx, attempt); sleepErr != nil {
					return lastStatus, err
				}
				continue
			}
			return lastStatus, err
		}

		if shouldRetry(method, lastStatus, nil) && attempt < 2 {
			if sleepErr := sleepBackoff(ctx, attempt); sleepErr != nil {
				break
			}
			continue
		}

		break
	}

	if lastStatus < 200 || lastStatus > 299 {
		if len(responseBody) == 0 {
			return lastStatus, fmt.Errorf("dockhand api returned status %d", lastStatus)
		}
		return lastStatus, fmt.Errorf("dockhand api returned status %d: %s", lastStatus, strings.TrimSpace(string(responseBody)))
	}

	if out != nil && len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, out); err != nil {
			return lastStatus, err
		}
	}

	return lastStatus, nil
}

func (c *Client) resolveEnv(value string) string {
	if value != "" {
		return value
	}
	return c.defaultEnv
}

func parseStacks(raw json.RawMessage) ([]stackResponse, error) {
	var asArray []map[string]any
	if err := json.Unmarshal(raw, &asArray); err == nil {
		return mapsToStacks(asArray), nil
	}

	var asObject map[string]json.RawMessage
	if err := json.Unmarshal(raw, &asObject); err != nil {
		return nil, err
	}

	if stacksRaw, ok := asObject["stacks"]; ok {
		if err := json.Unmarshal(stacksRaw, &asArray); err != nil {
			return nil, err
		}
		return mapsToStacks(asArray), nil
	}

	return nil, fmt.Errorf("unexpected stack list response shape")
}

func shouldRetry(method string, status int, err error) bool {
	switch method {
	case http.MethodGet, http.MethodDelete:
	default:
		return false
	}

	if err != nil {
		// Don't retry if the context is already cancelled.
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return true
		}
		// Retry other transient network errors (e.g. connection reset).
		return true
	}

	switch status {
	case 429, 502, 503, 504:
		return true
	default:
		return false
	}
}

func sleepBackoff(ctx context.Context, attempt int) error {
	delay := 200 * time.Millisecond
	if attempt == 1 {
		delay = 500 * time.Millisecond
	}
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func mapsToStacks(input []map[string]any) []stackResponse {
	output := make([]stackResponse, 0, len(input))

	for _, item := range input {
		name := firstString(item, "name", "stack", "stack_name")
		compose := firstString(item, "compose", "manifest")
		if name == "" {
			continue
		}
		output = append(output, stackResponse{
			Name:    name,
			Compose: compose,
		})
	}

	return output
}

func firstString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}

		switch v := value.(type) {
		case string:
			return v
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		}
	}

	return ""
}

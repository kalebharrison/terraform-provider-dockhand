package provider

import (
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL              *url.URL
	httpClient           *http.Client
	sessionCookie        string
	authHeader           string
	defaultEnv           string
	requestRetryAttempts int
	requestRetryMinDelay time.Duration
	requestRetryMaxDelay time.Duration
}

type stackPayload struct {
	Name    string `json:"name"`
	Compose string `json:"compose"`
}

type stackBasePathResponse struct {
	BasePath string `json:"basePath"`
}

type stackDefaultPathResponse struct {
	StackDir    string `json:"stackDir"`
	ComposePath string `json:"composePath"`
	EnvPath     string `json:"envPath"`
	Source      string `json:"source"`
}

type stackAdoptItemPayload struct {
	Name        string `json:"name"`
	ComposePath string `json:"composePath"`
}

type stackAdoptPayload struct {
	EnvironmentID int64                   `json:"environmentId"`
	Stacks        []stackAdoptItemPayload `json:"stacks"`
}

type stackContainerDetailResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Service      string `json:"service"`
	State        string `json:"state"`
	Status       string `json:"status"`
	Health       string `json:"health"`
	Image        string `json:"image"`
	RestartCount int64  `json:"restartCount"`
}

type stackResponse struct {
	Name             string                         `json:"name"`
	Compose          string                         `json:"compose"`
	Status           string                         `json:"status"`
	Containers       []string                       `json:"containers"`
	ContainerDetails []stackContainerDetailResponse `json:"containerDetails"`
}

type containerPortPayload struct {
	ContainerPort int64  `json:"containerPort"`
	HostPort      int64  `json:"hostPort"`
	Protocol      string `json:"protocol,omitempty"`
}

type containerPayload struct {
	Name          string                 `json:"name"`
	Image         string                 `json:"image"`
	Command       *string                `json:"command,omitempty"`
	Env           []string               `json:"env,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	Ports         []containerPortPayload `json:"ports,omitempty"`
	NetworkMode   *string                `json:"networkMode,omitempty"`
	RestartPolicy *string                `json:"restartPolicy,omitempty"`
	Privileged    *bool                  `json:"privileged,omitempty"`
	TTY           *bool                  `json:"tty,omitempty"`
	Memory        *int64                 `json:"memory,omitempty"`
	NanoCPUs      *int64                 `json:"nanoCpus,omitempty"`
	CapAdd        []string               `json:"capAdd,omitempty"`
}

type containerCreateResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id"`
}

type containerResponse struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	State        string            `json:"state"`
	Status       string            `json:"status"`
	Health       string            `json:"health"`
	RestartCount int64             `json:"restartCount"`
	Labels       map[string]string `json:"labels"`
	Command      *string           `json:"command"`
}

type containerLogsResponse struct {
	Logs string `json:"logs"`
}

type containerTopResponse struct {
	Titles    []string   `json:"Titles"`
	Processes [][]string `json:"Processes"`
	Error     *string    `json:"error"`
}

type containerShellOptionResponse struct {
	Path      string `json:"path"`
	Label     string `json:"label"`
	Available bool   `json:"available"`
}

type containerShellsResponse struct {
	Shells       []string                       `json:"shells"`
	DefaultShell *string                        `json:"defaultShell"`
	AllShells    []containerShellOptionResponse `json:"allShells"`
}

type containerStatsResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	CPUPercent  float64 `json:"cpuPercent"`
	MemoryUsage int64   `json:"memoryUsage"`
	MemoryRaw   int64   `json:"memoryRaw"`
	MemoryCache int64   `json:"memoryCache"`
	MemoryLimit int64   `json:"memoryLimit"`
	MemoryPct   float64 `json:"memoryPercent"`
	NetworkRX   int64   `json:"networkRx"`
	NetworkTX   int64   `json:"networkTx"`
	BlockRead   int64   `json:"blockRead"`
	BlockWrite  int64   `json:"blockWrite"`
}

type containerUpdateCheckResult struct {
	ContainerID   string  `json:"containerId"`
	ContainerName string  `json:"containerName"`
	ImageName     string  `json:"imageName"`
	HasUpdate     bool    `json:"hasUpdate"`
	CurrentDigest *string `json:"currentDigest"`
	LatestDigest  *string `json:"latestDigest"`
}

type containerUpdateCheckResponse struct {
	Total        int64                        `json:"total"`
	UpdatesFound int64                        `json:"updatesFound"`
	Results      []containerUpdateCheckResult `json:"results"`
}

type containerPendingUpdatesResponse struct {
	EnvironmentID  int64            `json:"environmentId"`
	PendingUpdates []map[string]any `json:"pendingUpdates"`
}

type activityEventResponse struct {
	ID              int64             `json:"id"`
	EnvironmentID   *int64            `json:"environmentId"`
	ContainerID     *string           `json:"containerId"`
	ContainerName   *string           `json:"containerName"`
	Image           *string           `json:"image"`
	Action          string            `json:"action"`
	ActorAttributes map[string]string `json:"actorAttributes"`
	Timestamp       *string           `json:"timestamp"`
	Type            *string           `json:"type"`
	Status          *string           `json:"status"`
	Details         map[string]any    `json:"details"`
}

type activityResponse struct {
	Events []activityEventResponse `json:"events"`
}

type imageScanPayload struct {
	ImageName string `json:"imageName"`
}

type imagePushPayload struct {
	ImageID    string `json:"imageId"`
	RegistryID int64  `json:"registryId"`
}

type networkContainerPayload struct {
	ContainerID string `json:"containerId"`
}

type volumeClonePayload struct {
	Name string `json:"name"`
}

type hawserConnectStatus struct {
	Status            string `json:"status"`
	Message           string `json:"message"`
	Protocol          string `json:"protocol"`
	ActiveConnections int64  `json:"activeConnections"`
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

type batchItemPayload struct {
	ID string `json:"id"`
}

type batchRequestPayload struct {
	EntityType string             `json:"entityType"`
	Operation  string             `json:"operation"`
	Items      []batchItemPayload `json:"items"`
}

type batchResponse struct {
	JobID  string         `json:"jobId"`
	Status string         `json:"status"`
	Result map[string]any `json:"result"`
}

type jobLineResponse struct {
	Data map[string]any `json:"data"`
}

type jobResponse struct {
	ID     string            `json:"id"`
	Status string            `json:"status"`
	Lines  []jobLineResponse `json:"lines"`
	Result map[string]any    `json:"result"`
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

type registryTagResponse struct {
	Name        string  `json:"name"`
	Size        int64   `json:"size"`
	LastUpdated *string `json:"lastUpdated"`
	Digest      *string `json:"digest"`
}

type registryTagsResponse struct {
	Tags     []registryTagResponse `json:"tags"`
	Total    int64                 `json:"total"`
	Page     int64                 `json:"page"`
	PageSize int64                 `json:"pageSize"`
	HasNext  bool                  `json:"hasNext"`
	HasPrev  bool                  `json:"hasPrev"`
}

type gitCredentialPayload struct {
	Name     string  `json:"name"`
	AuthType string  `json:"authType"`
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	SSHKey   *string `json:"sshPrivateKey,omitempty"`
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

type gitRepositoryTestPayload struct {
	URL          string  `json:"url"`
	Branch       *string `json:"branch,omitempty"`
	CredentialID *int64  `json:"credentialId,omitempty"`
	ComposePath  *string `json:"composePath,omitempty"`
}

type gitRepositoryTestResponse struct {
	Success    bool    `json:"success"`
	Error      *string `json:"error"`
	Branch     *string `json:"branch"`
	LastCommit *string `json:"lastCommit"`
}

type gitPreviewEnvPayload struct {
	RepositoryID *int64  `json:"repositoryId,omitempty"`
	URL          *string `json:"url,omitempty"`
	Branch       *string `json:"branch,omitempty"`
	ComposePath  string  `json:"composePath"`
	CredentialID *int64  `json:"credentialId,omitempty"`
}

type gitPreviewEnvResponse struct {
	Vars    map[string]any `json:"vars"`
	Sources map[string]any `json:"sources"`
	Error   *string        `json:"error"`
}

type gitStackEnvVarPayload struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

type gitStackPayload struct {
	StackName         string                  `json:"stackName"`
	EnvironmentID     *int64                  `json:"environmentId,omitempty"`
	RepositoryID      *int64                  `json:"repositoryId,omitempty"`
	RepoName          *string                 `json:"repoName,omitempty"`
	URL               *string                 `json:"url,omitempty"`
	Branch            *string                 `json:"branch,omitempty"`
	CredentialID      *int64                  `json:"credentialId,omitempty"`
	ComposePath       string                  `json:"composePath"`
	ContextDir        *string                 `json:"contextDir,omitempty"`
	EnvFilePath       *string                 `json:"envFilePath,omitempty"`
	AutoUpdateEnabled bool                    `json:"autoUpdateEnabled"`
	AutoUpdate        *bool                   `json:"autoUpdate,omitempty"`
	AutoUpdateCron    string                  `json:"autoUpdateCron"`
	WebhookEnabled    bool                    `json:"webhookEnabled"`
	WebhookSecret     *string                 `json:"webhookSecret,omitempty"`
	DeployNow         bool                    `json:"deployNow"`
	BuildOnDeploy     bool                    `json:"buildOnDeploy"`
	RepullImages      bool                    `json:"repullImages"`
	ForceRedeploy     bool                    `json:"forceRedeploy"`
	EnvVars           []gitStackEnvVarPayload `json:"envVars,omitempty"`
}

type gitStackRepositoryResponse struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	URL          string  `json:"url"`
	Branch       *string `json:"branch"`
	CredentialID *int64  `json:"credentialId"`
}

type gitStackResponse struct {
	ID                 int64                       `json:"id"`
	StackName          string                      `json:"stackName"`
	EnvironmentID      *int64                      `json:"environmentId"`
	RepositoryID       *int64                      `json:"repositoryId"`
	ComposePath        *string                     `json:"composePath"`
	ContextDir         *string                     `json:"contextDir"`
	EnvFilePath        *string                     `json:"envFilePath"`
	AutoUpdate         bool                        `json:"autoUpdate"`
	AutoUpdateEnabled  *bool                       `json:"autoUpdateEnabled"`
	AutoUpdateSchedule *string                     `json:"autoUpdateSchedule"`
	AutoUpdateCron     *string                     `json:"autoUpdateCron"`
	WebhookEnabled     bool                        `json:"webhookEnabled"`
	WebhookSecret      *string                     `json:"webhookSecret"`
	BuildOnDeploy      *bool                       `json:"buildOnDeploy"`
	RepullImages       *bool                       `json:"repullImages"`
	ForceRedeploy      *bool                       `json:"forceRedeploy"`
	LastSync           *string                     `json:"lastSync"`
	LastCommit         *string                     `json:"lastCommit"`
	SyncStatus         *string                     `json:"syncStatus"`
	SyncError          *string                     `json:"syncError"`
	CreatedAt          *string                     `json:"createdAt"`
	UpdatedAt          *string                     `json:"updatedAt"`
	Repository         *gitStackRepositoryResponse `json:"repository"`
}

type stackEnvVariable struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"isSecret"`
}

type stackEnvResponse struct {
	Variables []stackEnvVariable `json:"variables"`
}

type stackEnvRawResponse struct {
	Content string `json:"content"`
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

type notificationTestPayload struct {
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
}

type notificationTestResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

type environmentPayload struct {
	Name                  string  `json:"name"`
	ConnectionType        string  `json:"connectionType"`
	HawserToken           *string `json:"hawserToken,omitempty"`
	Host                  *string `json:"host,omitempty"`
	Port                  *int64  `json:"port,omitempty"`
	Protocol              *string `json:"protocol,omitempty"`
	SocketPath            *string `json:"socketPath,omitempty"`
	TLSSkipVerify         *bool   `json:"tlsSkipVerify,omitempty"`
	CACert                *string `json:"tlsCa,omitempty"`
	ClientCert            *string `json:"tlsCert,omitempty"`
	ClientKey             *string `json:"tlsKey,omitempty"`
	Icon                  *string `json:"icon,omitempty"`
	CollectActivity       *bool   `json:"collectActivity,omitempty"`
	CollectMetrics        *bool   `json:"collectMetrics,omitempty"`
	HighlightChanges      *bool   `json:"highlightChanges,omitempty"`
	Timezone              *string `json:"timezone,omitempty"`
	UpdateCheckEnabled    *bool   `json:"updateCheckEnabled,omitempty"`
	UpdateCheckAutoUpdate *bool   `json:"updateCheckAutoUpdate,omitempty"`
	ImagePruneEnabled     *bool   `json:"imagePruneEnabled,omitempty"`
	PublicIP              *string `json:"publicIp,omitempty"`
}

type environmentResponse struct {
	ID                    int64    `json:"id"`
	Name                  string   `json:"name"`
	ConnectionType        string   `json:"connectionType"`
	HawserToken           *string  `json:"hawserToken"`
	Host                  *string  `json:"host"`
	Port                  int64    `json:"port"`
	Protocol              string   `json:"protocol"`
	SocketPath            *string  `json:"socketPath"`
	TLSSkipVerify         bool     `json:"tlsSkipVerify"`
	CACert                *string  `json:"tlsCa"`
	ClientCert            *string  `json:"tlsCert"`
	ClientKey             *string  `json:"tlsKey"`
	Icon                  string   `json:"icon"`
	CollectActivity       bool     `json:"collectActivity"`
	CollectMetrics        bool     `json:"collectMetrics"`
	HighlightChanges      bool     `json:"highlightChanges"`
	Timezone              *string  `json:"timezone"`
	UpdateCheckEnabled    *bool    `json:"updateCheckEnabled"`
	UpdateCheckAutoUpdate *bool    `json:"updateCheckAutoUpdate"`
	ImagePruneEnabled     *bool    `json:"imagePruneEnabled"`
	CreatedAt             *string  `json:"createdAt"`
	UpdatedAt             *string  `json:"updatedAt"`
	PublicIP              *string  `json:"publicIp"`
	Labels                []string `json:"labels"`
}

type environmentTestResponse struct {
	Success bool           `json:"success"`
	Error   string         `json:"error"`
	Info    map[string]any `json:"info"`
	Hawser  map[string]any `json:"hawser"`
}

type environmentDetectSocketResponse struct {
	Sockets []any  `json:"sockets"`
	HomeDir string `json:"homedir"`
}

type hawserTokenPayload struct {
	Name          string  `json:"name"`
	EnvironmentID int64   `json:"environmentId"`
	RawToken      *string `json:"rawToken,omitempty"`
}

type hawserTokenResponse struct {
	Token   string `json:"token"`
	TokenID int64  `json:"tokenId"`
	Message string `json:"message"`
}

type environmentTimezoneResponse struct {
	Timezone string `json:"timezone"`
}

type environmentTimezonePayload struct {
	Timezone string `json:"timezone"`
}

type environmentUpdateCheckSettings struct {
	Enabled               bool   `json:"enabled"`
	Cron                  string `json:"cron"`
	AutoUpdate            bool   `json:"autoUpdate"`
	VulnerabilityCriteria string `json:"vulnerabilityCriteria"`
}

type environmentUpdateCheckResponse struct {
	Settings *environmentUpdateCheckSettings `json:"settings"`
}

type environmentUpdateCheckPayload struct {
	Enabled               bool   `json:"enabled"`
	Cron                  string `json:"cron"`
	AutoUpdate            bool   `json:"autoUpdate"`
	VulnerabilityCriteria string `json:"vulnerabilityCriteria"`
}

type environmentImagePruneSettings struct {
	Enabled        bool           `json:"enabled"`
	CronExpression string         `json:"cronExpression"`
	PruneMode      string         `json:"pruneMode"`
	LastPruned     *string        `json:"lastPruned"`
	LastResult     map[string]any `json:"lastResult"`
}

type environmentImagePruneResponse struct {
	Settings *environmentImagePruneSettings `json:"settings"`
}

type environmentImagePrunePayload struct {
	Enabled        bool   `json:"enabled"`
	CronExpression string `json:"cronExpression"`
	PruneMode      string `json:"pruneMode"`
}

type scannerSettingsInner struct {
	Scanner string `json:"scanner"`
}

type scannerSettingsResponse struct {
	Settings     *scannerSettingsInner `json:"settings"`
	Availability map[string]bool       `json:"availability"`
	Versions     map[string]any        `json:"versions"`
}

type scannerSettingsPayload struct {
	Scanner string `json:"scanner"`
	EnvID   int64  `json:"envId"`
}

type scannerUpdateInfo struct {
	HasUpdate bool `json:"hasUpdate"`
}

type scannerCheckUpdatesResponse struct {
	Updates map[string]scannerUpdateInfo `json:"updates"`
}

type simpleSuccessResponse struct {
	Success bool `json:"success"`
}

type networkPayload struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	Options    map[string]string `json:"options,omitempty"`
}

type networkResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Driver     string  `json:"driver"`
	Internal   bool    `json:"internal"`
	Attachable bool    `json:"attachable"`
	Scope      *string `json:"scope"`
	CreatedAt  *string `json:"createdAt"`
}

type networkInspectResponse struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	Scope      *string           `json:"scope"`
	CreatedAt  *string           `json:"createdAt"`
	Options    map[string]string `json:"options"`
	Labels     map[string]string `json:"labels"`
}

type volumePayload struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	DriverOpts map[string]string `json:"driverOpts,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

type volumeResponse struct {
	Name       string             `json:"name"`
	Driver     string             `json:"driver"`
	Mountpoint *string            `json:"mountpoint"`
	Scope      *string            `json:"scope"`
	CreatedAt  *string            `json:"createdAt"`
	Labels     map[string]string  `json:"labels"`
	Options    map[string]any     `json:"options"`
	Status     map[string]any     `json:"status"`
	UsageData  map[string]float64 `json:"usageData"`
}

type imagePullPayload struct {
	Image         string `json:"image"`
	ScanAfterPull bool   `json:"scanAfterPull"`
}

type imageResponse struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	Size    int64    `json:"size"`
	Created int64    `json:"created"`
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

type licensePayload struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type licenseResponse struct {
	Valid    bool    `json:"valid"`
	Active   bool    `json:"active"`
	Hostname *string `json:"hostname"`
}

type scheduleExecutionResponse struct {
	ID          int64   `json:"id"`
	Status      string  `json:"status"`
	TriggeredAt *string `json:"triggeredAt"`
	CompletedAt *string `json:"completedAt"`
}

type scheduleResponse struct {
	ID              int64                      `json:"id"`
	Type            string                     `json:"type"`
	Name            string                     `json:"name"`
	EntityName      *string                    `json:"entityName"`
	Description     *string                    `json:"description"`
	EnvironmentID   *int64                     `json:"environmentId"`
	EnvironmentName *string                    `json:"environmentName"`
	Enabled         bool                       `json:"enabled"`
	ScheduleType    *string                    `json:"scheduleType"`
	CronExpression  *string                    `json:"cronExpression"`
	NextRun         *string                    `json:"nextRun"`
	IsSystem        bool                       `json:"isSystem"`
	LastExecution   *scheduleExecutionResponse `json:"lastExecution"`
}

type schedulesListResponse struct {
	Schedules []scheduleResponse `json:"schedules"`
}

type scheduleExecutionItemResponse struct {
	ID            int64          `json:"id"`
	ScheduleType  string         `json:"scheduleType"`
	ScheduleID    int64          `json:"scheduleId"`
	EnvironmentID *int64         `json:"environmentId"`
	EntityName    *string        `json:"entityName"`
	TriggeredBy   *string        `json:"triggeredBy"`
	TriggeredAt   *string        `json:"triggeredAt"`
	StartedAt     *string        `json:"startedAt"`
	CompletedAt   *string        `json:"completedAt"`
	Duration      *int64         `json:"duration"`
	Status        *string        `json:"status"`
	ErrorMessage  *string        `json:"errorMessage"`
	Details       map[string]any `json:"details"`
	CreatedAt     *string        `json:"createdAt"`
	Logs          *string        `json:"logs"`
}

type schedulesExecutionsResponse struct {
	Executions []scheduleExecutionItemResponse `json:"executions"`
	Total      int64                           `json:"total"`
	Limit      int64                           `json:"limit"`
	Offset     int64                           `json:"offset"`
}

type scheduleSettingsResponse struct {
	HideSystemJobs bool `json:"hideSystemJobs"`
}

type scheduleSettingsPayload struct {
	HideSystemJobs bool `json:"hideSystemJobs"`
}

type scheduleStreamEvent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type systemFilesEntryResponse struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Type  string `json:"type"`
	Size  int64  `json:"size"`
	Mtime string `json:"mtime"`
	Mode  string `json:"mode"`
}

type systemFilesResponse struct {
	Path    *string                    `json:"path"`
	Parent  *string                    `json:"parent"`
	Entries []systemFilesEntryResponse `json:"entries"`
}

type systemFileContentResponse struct {
	Path    *string `json:"path"`
	Content *string `json:"content"`
	Size    *int64  `json:"size"`
	Mtime   *string `json:"mtime"`
}

type stackScanResponse struct {
	Discovered []map[string]any `json:"discovered"`
	Adopted    []map[string]any `json:"adopted"`
	Skipped    []map[string]any `json:"skipped"`
	Errors     []map[string]any `json:"errors"`
}

type stackSourceRepositoryResponse struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	URL                string  `json:"url"`
	Branch             *string `json:"branch"`
	CredentialID       *int64  `json:"credentialId"`
	ComposePath        *string `json:"composePath"`
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

type stackSourceResponse struct {
	SourceType  string                         `json:"sourceType"`
	ComposePath *string                        `json:"composePath"`
	Repository  *stackSourceRepositoryResponse `json:"repository"`
}

type stackAdoptResponse struct {
	Adopted []string `json:"adopted"`
	Failed  []string `json:"failed"`
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

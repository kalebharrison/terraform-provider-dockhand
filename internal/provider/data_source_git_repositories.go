package provider

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*gitRepositoriesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*gitRepositoriesDataSource)(nil)
)

func NewGitRepositoriesDataSource() datasource.DataSource {
	return &gitRepositoriesDataSource{}
}

type gitRepositoriesDataSource struct {
	client *Client
}

type gitRepositoriesDataSourceRepositoryModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	URL                types.String `tfsdk:"url"`
	Branch             types.String `tfsdk:"branch"`
	ComposePath        types.String `tfsdk:"compose_path"`
	CredentialID       types.String `tfsdk:"credential_id"`
	EnvironmentID      types.String `tfsdk:"environment_id"`
	AutoUpdate         types.Bool   `tfsdk:"auto_update"`
	AutoUpdateSchedule types.String `tfsdk:"auto_update_schedule"`
	AutoUpdateCron     types.String `tfsdk:"auto_update_cron"`
	WebhookEnabled     types.Bool   `tfsdk:"webhook_enabled"`
	WebhookSecret      types.String `tfsdk:"webhook_secret"`
	LastSync           types.String `tfsdk:"last_sync"`
	LastCommit         types.String `tfsdk:"last_commit"`
	SyncStatus         types.String `tfsdk:"sync_status"`
	SyncError          types.String `tfsdk:"sync_error"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

type gitRepositoriesDataSourceModel struct {
	Repositories []gitRepositoriesDataSourceRepositoryModel `tfsdk:"repositories"`
	Names        types.List                                 `tfsdk:"names"`
	IDs          types.List                                 `tfsdk:"ids"`
}

func (d *gitRepositoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_repositories"
}

func (d *gitRepositoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Git repositories from Dockhand `/api/git/repositories`.",
		Attributes: map[string]schema.Attribute{
			"repositories": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                   schema.StringAttribute{Computed: true},
						"name":                 schema.StringAttribute{Computed: true},
						"url":                  schema.StringAttribute{Computed: true},
						"branch":               schema.StringAttribute{Computed: true},
						"compose_path":         schema.StringAttribute{Computed: true},
						"credential_id":        schema.StringAttribute{Computed: true},
						"environment_id":       schema.StringAttribute{Computed: true},
						"auto_update":          schema.BoolAttribute{Computed: true},
						"auto_update_schedule": schema.StringAttribute{Computed: true},
						"auto_update_cron":     schema.StringAttribute{Computed: true},
						"webhook_enabled":      schema.BoolAttribute{Computed: true},
						"webhook_secret":       schema.StringAttribute{Computed: true, Sensitive: true},
						"last_sync":            schema.StringAttribute{Computed: true},
						"last_commit":          schema.StringAttribute{Computed: true},
						"sync_status":          schema.StringAttribute{Computed: true},
						"sync_error":           schema.StringAttribute{Computed: true},
						"created_at":           schema.StringAttribute{Computed: true},
						"updated_at":           schema.StringAttribute{Computed: true},
					},
				},
			},
			"names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *gitRepositoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *Client, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *gitRepositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data gitRepositoriesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListGitRepositories(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand git repositories", err.Error())
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	out := make([]gitRepositoriesDataSourceRepositoryModel, 0, len(items))
	names := make([]string, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		id := strconv.FormatInt(item.ID, 10)
		out = append(out, gitRepositoriesDataSourceRepositoryModel{
			ID:                 types.StringValue(id),
			Name:               types.StringValue(item.Name),
			URL:                types.StringValue(item.URL),
			Branch:             stringValueOrNull(item.Branch),
			ComposePath:        stringValueOrNull(item.ComposePath),
			CredentialID:       int64StringValueOrNull(item.CredentialID),
			EnvironmentID:      int64StringValueOrNull(item.EnvironmentID),
			AutoUpdate:         types.BoolValue(item.AutoUpdate),
			AutoUpdateSchedule: stringValueOrNull(item.AutoUpdateSchedule),
			AutoUpdateCron:     stringValueOrNull(item.AutoUpdateCron),
			WebhookEnabled:     types.BoolValue(item.WebhookEnabled),
			WebhookSecret:      stringValueOrNull(item.WebhookSecret),
			LastSync:           stringValueOrNull(item.LastSync),
			LastCommit:         stringValueOrNull(item.LastCommit),
			SyncStatus:         stringValueOrNull(item.SyncStatus),
			SyncError:          stringValueOrNull(item.SyncError),
			CreatedAt:          stringValueOrNull(item.CreatedAt),
			UpdatedAt:          stringValueOrNull(item.UpdatedAt),
		})
		names = append(names, item.Name)
		ids = append(ids, id)
	}

	namesVal, diags := types.ListValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	idsVal, diags := types.ListValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Repositories = out
	data.Names = namesVal
	data.IDs = idsVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

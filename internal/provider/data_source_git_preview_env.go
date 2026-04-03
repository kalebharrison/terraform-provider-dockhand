package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*gitPreviewEnvDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*gitPreviewEnvDataSource)(nil)
)

func NewGitPreviewEnvDataSource() datasource.DataSource {
	return &gitPreviewEnvDataSource{}
}

type gitPreviewEnvDataSource struct {
	client *Client
}

type gitPreviewEnvDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	RepositoryID types.String `tfsdk:"repository_id"`
	URL          types.String `tfsdk:"url"`
	Branch       types.String `tfsdk:"branch"`
	CredentialID types.String `tfsdk:"credential_id"`
	ComposePath  types.String `tfsdk:"compose_path"`

	VarsJSON      types.String `tfsdk:"vars_json"`
	SourcesJSON   types.String `tfsdk:"sources_json"`
	VariableNames types.List   `tfsdk:"variable_names"`
	SourceNames   types.List   `tfsdk:"source_names"`
}

func (d *gitPreviewEnvDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_preview_env"
}

func (d *gitPreviewEnvDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Previews Git-backed stack environment variables via `POST /api/git/preview-env`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"repository_id": schema.StringAttribute{
				MarkdownDescription: "Existing Dockhand Git repository ID to preview. Mutually exclusive with `url`.",
				Optional:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Git clone URL to preview when `repository_id` is not used.",
				Optional:            true,
			},
			"branch": schema.StringAttribute{
				MarkdownDescription: "Git branch to preview when `url` is used.",
				Optional:            true,
			},
			"credential_id": schema.StringAttribute{
				MarkdownDescription: "Optional Dockhand Git credential ID used when `url` is set.",
				Optional:            true,
			},
			"compose_path": schema.StringAttribute{
				MarkdownDescription: "Repository-relative compose file path to inspect.",
				Required:            true,
			},
			"vars_json": schema.StringAttribute{
				Computed: true,
			},
			"sources_json": schema.StringAttribute{
				Computed: true,
			},
			"variable_names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"source_names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *gitPreviewEnvDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *gitPreviewEnvDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data gitPreviewEnvDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, id, err := buildGitPreviewEnvPayload(data)
	if err != nil {
		resp.Diagnostics.AddError("Invalid git preview env configuration", err.Error())
		return
	}

	out, _, err := d.client.PreviewGitEnv(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error previewing Dockhand git environment variables", err.Error())
		return
	}
	if msg := strings.TrimSpace(valueOrEmpty(out.Error)); msg != "" {
		resp.Diagnostics.AddError("Dockhand git preview env failed", msg)
		return
	}

	variableNames, diags := types.ListValueFrom(ctx, types.StringType, sortedMapKeys(out.Vars))
	resp.Diagnostics.Append(diags...)
	sourceNames, diags := types.ListValueFrom(ctx, types.StringType, sortedMapKeys(out.Sources))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(id)
	data.RepositoryID = normalizedOptionalString(data.RepositoryID)
	data.URL = normalizedOptionalString(data.URL)
	data.Branch = normalizedOptionalString(data.Branch)
	data.CredentialID = normalizedOptionalString(data.CredentialID)
	data.ComposePath = types.StringValue(strings.TrimSpace(data.ComposePath.ValueString()))
	data.VarsJSON = types.StringValue(mustJSON(out.Vars))
	data.SourcesJSON = types.StringValue(mustJSON(out.Sources))
	data.VariableNames = variableNames
	data.SourceNames = sourceNames

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func buildGitPreviewEnvPayload(data gitPreviewEnvDataSourceModel) (gitPreviewEnvPayload, string, error) {
	composePath := strings.TrimSpace(data.ComposePath.ValueString())
	if composePath == "" {
		return gitPreviewEnvPayload{}, "", fmt.Errorf("compose_path is required")
	}

	repositoryID := strings.TrimSpace(data.RepositoryID.ValueString())
	urlValue := strings.TrimSpace(data.URL.ValueString())
	switch {
	case repositoryID != "" && urlValue != "":
		return gitPreviewEnvPayload{}, "", fmt.Errorf("set either `repository_id` or `url`, not both")
	case repositoryID == "" && urlValue == "":
		return gitPreviewEnvPayload{}, "", fmt.Errorf("set `repository_id` or `url`")
	}

	payload := gitPreviewEnvPayload{
		ComposePath: composePath,
	}
	if repositoryID != "" {
		if !data.Branch.IsNull() || !data.CredentialID.IsNull() {
			return gitPreviewEnvPayload{}, "", fmt.Errorf("`branch` and `credential_id` cannot be set when `repository_id` is used")
		}
		repoID, err := optionalInt64StringValue("repository_id", data.RepositoryID)
		if err != nil {
			return gitPreviewEnvPayload{}, "", err
		}
		payload.RepositoryID = repoID
		return payload, "repo:" + repositoryID + ":" + composePath, nil
	}

	credentialID, err := optionalInt64StringValue("credential_id", data.CredentialID)
	if err != nil {
		return gitPreviewEnvPayload{}, "", err
	}
	payload.URL = &urlValue
	payload.Branch = optionalStringValue(data.Branch)
	payload.CredentialID = credentialID
	return payload, "url:" + urlValue + ":" + composePath, nil
}

func normalizedOptionalString(v types.String) types.String {
	if v.IsNull() || v.IsUnknown() {
		return types.StringNull()
	}
	value := strings.TrimSpace(v.ValueString())
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

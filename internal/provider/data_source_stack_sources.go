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
	_ datasource.DataSource              = (*stackSourcesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*stackSourcesDataSource)(nil)
)

func NewStackSourcesDataSource() datasource.DataSource {
	return &stackSourcesDataSource{}
}

type stackSourcesDataSource struct {
	client *Client
}

type stackSourceItemModel struct {
	StackName               types.String `tfsdk:"stack_name"`
	SourceType              types.String `tfsdk:"source_type"`
	ComposePath             types.String `tfsdk:"compose_path"`
	RepositoryID            types.String `tfsdk:"repository_id"`
	RepositoryName          types.String `tfsdk:"repository_name"`
	RepositoryURL           types.String `tfsdk:"repository_url"`
	RepositoryBranch        types.String `tfsdk:"repository_branch"`
	RepositoryComposePath   types.String `tfsdk:"repository_compose_path"`
	RepositoryEnvironmentID types.String `tfsdk:"repository_environment_id"`
	RepositorySyncStatus    types.String `tfsdk:"repository_sync_status"`
}

type stackSourcesDataSourceModel struct {
	ID      types.String           `tfsdk:"id"`
	Sources []stackSourceItemModel `tfsdk:"sources"`
}

func (d *stackSourcesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_sources"
}

func (d *stackSourcesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads stack source mappings from `/api/stacks/sources`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"sources": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"stack_name":                schema.StringAttribute{Computed: true},
					"source_type":               schema.StringAttribute{Computed: true},
					"compose_path":              schema.StringAttribute{Computed: true},
					"repository_id":             schema.StringAttribute{Computed: true},
					"repository_name":           schema.StringAttribute{Computed: true},
					"repository_url":            schema.StringAttribute{Computed: true},
					"repository_branch":         schema.StringAttribute{Computed: true},
					"repository_compose_path":   schema.StringAttribute{Computed: true},
					"repository_environment_id": schema.StringAttribute{Computed: true},
					"repository_sync_status":    schema.StringAttribute{Computed: true},
				}},
			},
		},
	}
}

func (d *stackSourcesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *stackSourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	out, _, err := d.client.GetStackSources(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading stack sources", err.Error())
		return
	}

	names := make([]string, 0, len(out))
	for k := range out {
		names = append(names, k)
	}
	sort.Strings(names)

	sources := make([]stackSourceItemModel, 0, len(names))
	for _, name := range names {
		s := out[name]
		item := stackSourceItemModel{StackName: types.StringValue(name), SourceType: types.StringValue(s.SourceType)}
		if s.ComposePath != nil {
			item.ComposePath = types.StringValue(*s.ComposePath)
		} else {
			item.ComposePath = types.StringNull()
		}
		if s.Repository != nil {
			item.RepositoryID = types.StringValue(strconv.FormatInt(s.Repository.ID, 10))
			item.RepositoryName = types.StringValue(s.Repository.Name)
			item.RepositoryURL = types.StringValue(s.Repository.URL)
			if s.Repository.Branch != nil {
				item.RepositoryBranch = types.StringValue(*s.Repository.Branch)
			} else {
				item.RepositoryBranch = types.StringNull()
			}
			if s.Repository.ComposePath != nil {
				item.RepositoryComposePath = types.StringValue(*s.Repository.ComposePath)
			} else {
				item.RepositoryComposePath = types.StringNull()
			}
			if s.Repository.EnvironmentID != nil {
				item.RepositoryEnvironmentID = types.StringValue(strconv.FormatInt(*s.Repository.EnvironmentID, 10))
			} else {
				item.RepositoryEnvironmentID = types.StringNull()
			}
			if s.Repository.SyncStatus != nil {
				item.RepositorySyncStatus = types.StringValue(*s.Repository.SyncStatus)
			} else {
				item.RepositorySyncStatus = types.StringNull()
			}
		} else {
			item.RepositoryID = types.StringNull()
			item.RepositoryName = types.StringNull()
			item.RepositoryURL = types.StringNull()
			item.RepositoryBranch = types.StringNull()
			item.RepositoryComposePath = types.StringNull()
			item.RepositoryEnvironmentID = types.StringNull()
			item.RepositorySyncStatus = types.StringNull()
		}
		sources = append(sources, item)
	}

	state := stackSourcesDataSourceModel{ID: types.StringValue("stack-sources"), Sources: sources}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

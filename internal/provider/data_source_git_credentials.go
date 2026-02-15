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
	_ datasource.DataSource              = (*gitCredentialsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*gitCredentialsDataSource)(nil)
)

func NewGitCredentialsDataSource() datasource.DataSource {
	return &gitCredentialsDataSource{}
}

type gitCredentialsDataSource struct {
	client *Client
}

type gitCredentialsDataSourceCredentialModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	AuthType    types.String `tfsdk:"auth_type"`
	Username    types.String `tfsdk:"username"`
	HasPassword types.Bool   `tfsdk:"has_password"`
	HasSSHKey   types.Bool   `tfsdk:"has_ssh_key"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

type gitCredentialsDataSourceModel struct {
	Credentials []gitCredentialsDataSourceCredentialModel `tfsdk:"credentials"`
	Names       types.List                                `tfsdk:"names"`
	IDs         types.List                                `tfsdk:"ids"`
}

func (d *gitCredentialsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_credentials"
}

func (d *gitCredentialsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists Git credentials from Dockhand `/api/git/credentials`.",
		Attributes: map[string]schema.Attribute{
			"credentials": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"name":         schema.StringAttribute{Computed: true},
						"auth_type":    schema.StringAttribute{Computed: true},
						"username":     schema.StringAttribute{Computed: true},
						"has_password": schema.BoolAttribute{Computed: true},
						"has_ssh_key":  schema.BoolAttribute{Computed: true},
						"created_at":   schema.StringAttribute{Computed: true},
						"updated_at":   schema.StringAttribute{Computed: true},
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

func (d *gitCredentialsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *gitCredentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data gitCredentialsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListGitCredentials(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand git credentials", err.Error())
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	out := make([]gitCredentialsDataSourceCredentialModel, 0, len(items))
	names := make([]string, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		id := strconv.FormatInt(item.ID, 10)
		out = append(out, gitCredentialsDataSourceCredentialModel{
			ID:          types.StringValue(id),
			Name:        types.StringValue(item.Name),
			AuthType:    types.StringValue(item.AuthType),
			Username:    stringValueOrNull(item.Username),
			HasPassword: types.BoolValue(item.HasPassword),
			HasSSHKey:   types.BoolValue(item.HasSSHKey),
			CreatedAt:   stringValueOrNull(item.CreatedAt),
			UpdatedAt:   stringValueOrNull(item.UpdatedAt),
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

	data.Credentials = out
	data.Names = namesVal
	data.IDs = idsVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

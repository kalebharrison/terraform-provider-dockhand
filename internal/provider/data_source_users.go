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
	_ datasource.DataSource              = (*usersDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*usersDataSource)(nil)
)

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

type usersDataSource struct {
	client *Client
}

type usersDataSourceUserModel struct {
	ID          types.String `tfsdk:"id"`
	Username    types.String `tfsdk:"username"`
	Email       types.String `tfsdk:"email"`
	DisplayName types.String `tfsdk:"display_name"`
	MFAEnabled  types.Bool   `tfsdk:"mfa_enabled"`
	IsAdmin     types.Bool   `tfsdk:"is_admin"`
	IsActive    types.Bool   `tfsdk:"is_active"`
	LastLogin   types.String `tfsdk:"last_login"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

type usersDataSourceModel struct {
	Users     []usersDataSourceUserModel `tfsdk:"users"`
	Usernames types.List                 `tfsdk:"usernames"`
	IDs       types.List                 `tfsdk:"ids"`
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists users from Dockhand `/api/users`.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"username":     schema.StringAttribute{Computed: true},
						"email":        schema.StringAttribute{Computed: true},
						"display_name": schema.StringAttribute{Computed: true},
						"mfa_enabled":  schema.BoolAttribute{Computed: true},
						"is_admin":     schema.BoolAttribute{Computed: true},
						"is_active":    schema.BoolAttribute{Computed: true},
						"last_login":   schema.StringAttribute{Computed: true},
						"created_at":   schema.StringAttribute{Computed: true},
						"updated_at":   schema.StringAttribute{Computed: true},
					},
				},
			},
			"usernames": schema.ListAttribute{
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

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var data usersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, _, err := d.client.ListUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing Dockhand users", err.Error())
		return
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Username < items[j].Username
	})

	out := make([]usersDataSourceUserModel, 0, len(items))
	usernames := make([]string, 0, len(items))
	ids := make([]string, 0, len(items))

	for _, item := range items {
		id := strconv.FormatInt(item.ID, 10)
		out = append(out, usersDataSourceUserModel{
			ID:          types.StringValue(id),
			Username:    types.StringValue(item.Username),
			Email:       stringValueOrNull(item.Email),
			DisplayName: stringValueOrNull(item.DisplayName),
			MFAEnabled:  types.BoolValue(item.MFAEnabled),
			IsAdmin:     types.BoolValue(item.IsAdmin),
			IsActive:    types.BoolValue(item.IsActive),
			LastLogin:   stringValueOrNull(item.LastLogin),
			CreatedAt:   stringValueOrNull(item.CreatedAt),
			UpdatedAt:   stringValueOrNull(item.UpdatedAt),
		})
		usernames = append(usernames, item.Username)
		ids = append(ids, id)
	}

	usernamesVal, diags := types.ListValueFrom(ctx, types.StringType, usernames)
	resp.Diagnostics.Append(diags...)
	idsVal, diags := types.ListValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Users = out
	data.Usernames = usernamesVal
	data.IDs = idsVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

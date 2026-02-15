package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*gitStackEnvFileResource)(nil)
	_ resource.ResourceWithConfigure   = (*gitStackEnvFileResource)(nil)
	_ resource.ResourceWithImportState = (*gitStackEnvFileResource)(nil)
)

func NewGitStackEnvFileResource() resource.Resource {
	return &gitStackEnvFileResource{}
}

type gitStackEnvFileResource struct {
	client *Client
}

type gitStackEnvFileResourceModel struct {
	ID        types.String `tfsdk:"id"`
	StackID   types.String `tfsdk:"stack_id"`
	Path      types.String `tfsdk:"path"`
	Trigger   types.String `tfsdk:"trigger"`
	VarsJSON  types.String `tfsdk:"vars_json"`
	FilePaths types.List   `tfsdk:"file_paths"`
}

func (r *gitStackEnvFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_stack_env_file"
}

func (r *gitStackEnvFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads env vars for a selected git stack env file via `/api/git/stacks/{id}/env-files`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"stack_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"path": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"trigger": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vars_json": schema.StringAttribute{Computed: true},
			"file_paths": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *gitStackEnvFileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func mapToSortedJSON(m map[string]string) (string, error) {
	if m == nil {
		m = map[string]string{}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ordered := make(map[string]string, len(m))
	for _, k := range keys {
		ordered[k] = m[k]
	}
	b, err := json.Marshal(ordered)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func stringSliceToList(stringsIn []string) types.List {
	elems := make([]types.String, 0, len(stringsIn))
	for _, s := range stringsIn {
		elems = append(elems, types.StringValue(s))
	}
	out, diags := types.ListValueFrom(context.Background(), types.StringType, elems)
	if diags.HasError() {
		return types.ListNull(types.StringType)
	}
	return out
}

func (r *gitStackEnvFileResource) refresh(ctx context.Context, state *gitStackEnvFileResourceModel) error {
	stackID := strings.TrimSpace(state.StackID.ValueString())
	path := strings.TrimSpace(state.Path.ValueString())
	if stackID == "" || path == "" {
		return fmt.Errorf("stack_id and path are required")
	}

	files, status, err := r.client.ListGitStackEnvFiles(ctx, stackID)
	if err != nil {
		return err
	}
	if status < 200 || status > 299 {
		return fmt.Errorf("list env files returned status %d", status)
	}
	sort.Strings(files)

	vars, status, err := r.client.GetGitStackEnvFileVars(ctx, stackID, path)
	if err != nil {
		return err
	}
	if status < 200 || status > 299 {
		return fmt.Errorf("read env file vars returned status %d", status)
	}

	jsonBody, err := mapToSortedJSON(vars)
	if err != nil {
		return err
	}

	state.FilePaths = stringSliceToList(files)
	state.VarsJSON = types.StringValue(jsonBody)
	state.ID = types.StringValue(fmt.Sprintf("%s:%s:%s", stackID, path, state.Trigger.ValueString()))
	return nil
}

func (r *gitStackEnvFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}
	var plan gitStackEnvFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.refresh(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Error reading git stack env file", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gitStackEnvFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}
	var state gitStackEnvFileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.refresh(ctx, &state); err != nil {
		resp.Diagnostics.AddError("Error reading git stack env file", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gitStackEnvFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan gitStackEnvFileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gitStackEnvFileResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *gitStackEnvFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<stack_id>:<path>:<trigger>`.")
		return
	}
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) < 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<stack_id>:<path>` or `<stack_id>:<path>:<trigger>`.")
		return
	}
	stackID := strings.TrimSpace(parts[0])
	pathValue := strings.TrimSpace(parts[1])
	trigger := ""
	if len(parts) == 3 {
		trigger = strings.TrimSpace(parts[2])
	}
	if stackID == "" || pathValue == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Both `stack_id` and `path` must be non-empty.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("stack_id"), stackID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("path"), pathValue)...)
	if trigger != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("trigger"), trigger)...)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = (*gitRepositoryTestActionResource)(nil)
	_ resource.ResourceWithConfigure = (*gitRepositoryTestActionResource)(nil)
)

func NewGitRepositoryTestActionResource() resource.Resource {
	return &gitRepositoryTestActionResource{}
}

type gitRepositoryTestActionResource struct {
	client *Client
}

type gitRepositoryTestActionModel struct {
	ID           types.String `tfsdk:"id"`
	RepositoryID types.String `tfsdk:"repository_id"`
	URL          types.String `tfsdk:"url"`
	Branch       types.String `tfsdk:"branch"`
	CredentialID types.String `tfsdk:"credential_id"`
	ComposePath  types.String `tfsdk:"compose_path"`
	FailOnError  types.Bool   `tfsdk:"fail_on_error"`
	Trigger      types.String `tfsdk:"trigger"`

	Success        types.Bool   `tfsdk:"success"`
	Error          types.String `tfsdk:"error"`
	ResolvedBranch types.String `tfsdk:"resolved_branch"`
	LastCommit     types.String `tfsdk:"last_commit"`
	ResultJSON     types.String `tfsdk:"result_json"`
}

func (r *gitRepositoryTestActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_repository_test_action"
}

func (r *gitRepositoryTestActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot Git repository connectivity test via `POST /api/git/repositories/test`. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"repository_id": schema.StringAttribute{
				MarkdownDescription: "Existing Dockhand Git repository ID to test. Mutually exclusive with `url`.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Git clone URL to test when `repository_id` is not used.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"branch": schema.StringAttribute{
				MarkdownDescription: "Git branch to test when `url` is used.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"credential_id": schema.StringAttribute{
				MarkdownDescription: "Optional Dockhand Git credential ID to test with `url`.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compose_path": schema.StringAttribute{
				MarkdownDescription: "Optional compose path included in the test payload when `url` is used.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fail_on_error": schema.BoolAttribute{
				MarkdownDescription: "If true (default), apply fails when Dockhand returns `success = false`.",
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
			"resolved_branch": schema.StringAttribute{
				Computed: true,
			},
			"last_commit": schema.StringAttribute{
				Computed: true,
			},
			"result_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *gitRepositoryTestActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gitRepositoryTestActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan gitRepositoryTestActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, sourceLabel, err := r.resolveTestPayload(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Invalid git repository test configuration", err.Error())
		return
	}

	out, _, err := r.client.TestGitRepository(ctx, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error testing Dockhand git repository", err.Error())
		return
	}

	plan.Success = types.BoolValue(out.Success)
	plan.Error = stringValueOrNull(out.Error)
	plan.ResolvedBranch = stringValueOrNull(out.Branch)
	plan.LastCommit = stringValueOrNull(out.LastCommit)
	plan.ResultJSON = types.StringValue(mustJSON(map[string]any{
		"success":    out.Success,
		"error":      strings.TrimSpace(valueOrEmpty(out.Error)),
		"branch":     strings.TrimSpace(valueOrEmpty(out.Branch)),
		"lastCommit": strings.TrimSpace(valueOrEmpty(out.LastCommit)),
	}))
	plan.ID = types.StringValue(fmt.Sprintf("git-repo-test:%s:%s", sourceLabel, strings.TrimSpace(plan.Trigger.ValueString())))

	failOnError := true
	if !plan.FailOnError.IsNull() && !plan.FailOnError.IsUnknown() {
		failOnError = plan.FailOnError.ValueBool()
	}
	if failOnError && !out.Success {
		msg := strings.TrimSpace(valueOrEmpty(out.Error))
		if msg == "" {
			msg = "Dockhand git repository test returned success=false"
		}
		resp.Diagnostics.AddError("Dockhand git repository test failed", msg)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gitRepositoryTestActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state gitRepositoryTestActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *gitRepositoryTestActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan gitRepositoryTestActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gitRepositoryTestActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// No-op one-shot action.
}

func (r *gitRepositoryTestActionResource) resolveTestPayload(ctx context.Context, plan gitRepositoryTestActionModel) (gitRepositoryTestPayload, string, error) {
	repositoryID := strings.TrimSpace(plan.RepositoryID.ValueString())
	urlValue := strings.TrimSpace(plan.URL.ValueString())

	switch {
	case repositoryID != "" && urlValue != "":
		return gitRepositoryTestPayload{}, "", fmt.Errorf("set either `repository_id` or `url`, not both")
	case repositoryID == "" && urlValue == "":
		return gitRepositoryTestPayload{}, "", fmt.Errorf("set `repository_id` or `url`")
	}

	if repositoryID != "" {
		if !plan.Branch.IsNull() || !plan.CredentialID.IsNull() || !plan.ComposePath.IsNull() {
			return gitRepositoryTestPayload{}, "", fmt.Errorf("`branch`, `credential_id`, and `compose_path` cannot be set when `repository_id` is used")
		}

		repo, _, err := r.client.GetGitRepository(ctx, repositoryID)
		if err != nil {
			return gitRepositoryTestPayload{}, "", err
		}
		if strings.TrimSpace(repo.URL) == "" {
			return gitRepositoryTestPayload{}, "", fmt.Errorf("git repository %s has empty url", repositoryID)
		}

		payload := gitRepositoryTestPayload{
			URL:          strings.TrimSpace(repo.URL),
			Branch:       repo.Branch,
			CredentialID: repo.CredentialID,
			ComposePath:  repo.ComposePath,
		}
		return payload, "id-" + repositoryID, nil
	}

	credentialID, err := optionalInt64StringValue("credential_id", plan.CredentialID)
	if err != nil {
		return gitRepositoryTestPayload{}, "", err
	}

	payload := gitRepositoryTestPayload{
		URL:          urlValue,
		Branch:       optionalStringValue(plan.Branch),
		CredentialID: credentialID,
		ComposePath:  optionalStringValue(plan.ComposePath),
	}
	return payload, "url-" + urlValue, nil
}

func valueOrEmpty(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

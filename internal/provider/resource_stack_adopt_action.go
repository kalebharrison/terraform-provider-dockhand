package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*stackAdoptActionResource)(nil)
	_ resource.ResourceWithConfigure   = (*stackAdoptActionResource)(nil)
	_ resource.ResourceWithImportState = (*stackAdoptActionResource)(nil)
)

func NewStackAdoptActionResource() resource.Resource {
	return &stackAdoptActionResource{}
}

type stackAdoptActionResource struct {
	client *Client
}

type stackAdoptItemModel struct {
	Name        types.String `tfsdk:"name"`
	ComposePath types.String `tfsdk:"compose_path"`
}

type stackAdoptActionModel struct {
	ID            types.String          `tfsdk:"id"`
	EnvironmentID types.Int64           `tfsdk:"environment_id"`
	Stacks        []stackAdoptItemModel `tfsdk:"stacks"`
	Trigger       types.String          `tfsdk:"trigger"`
	Adopted       types.List            `tfsdk:"adopted"`
	Failed        types.List            `tfsdk:"failed"`
}

func (r *stackAdoptActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_adopt_action"
}

func (r *stackAdoptActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Runs a one-shot stack adopt operation via `/api/stacks/adopt`. Change `trigger` to run it again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"environment_id": schema.Int64Attribute{
				Required:      true,
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"stacks": schema.ListNestedAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"name":         schema.StringAttribute{Required: true},
					"compose_path": schema.StringAttribute{Required: true},
				}},
			},
			"trigger": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"adopted": schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"failed":  schema.ListAttribute{Computed: true, ElementType: types.StringType},
		},
	}
}

func (r *stackAdoptActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *stackAdoptActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var plan stackAdoptActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items := make([]stackAdoptItemPayload, 0, len(plan.Stacks))
	for _, s := range plan.Stacks {
		name := strings.TrimSpace(s.Name.ValueString())
		composePath := strings.TrimSpace(s.ComposePath.ValueString())
		if name == "" || composePath == "" {
			resp.Diagnostics.AddError("Invalid stack adopt item", "Each `stacks` item requires non-empty `name` and `compose_path`.")
			return
		}
		items = append(items, stackAdoptItemPayload{Name: name, ComposePath: composePath})
	}

	out, status, err := r.client.AdoptStacks(ctx, stackAdoptPayload{
		EnvironmentID: plan.EnvironmentID.ValueInt64(),
		Stacks:        items,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error adopting stacks", err.Error())
		return
	}
	if status < 200 || status > 299 {
		resp.Diagnostics.AddError("Error adopting stacks", fmt.Sprintf("Dockhand returned status %d", status))
		return
	}

	adopted := make([]attr.Value, 0, len(out.Adopted))
	for _, s := range out.Adopted {
		adopted = append(adopted, types.StringValue(s))
	}
	failed := make([]attr.Value, 0, len(out.Failed))
	for _, s := range out.Failed {
		failed = append(failed, types.StringValue(s))
	}
	plan.Adopted = types.ListValueMust(types.StringType, adopted)
	plan.Failed = types.ListValueMust(types.StringType, failed)
	plan.ID = types.StringValue(fmt.Sprintf("%d:%s", plan.EnvironmentID.ValueInt64(), plan.Trigger.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackAdoptActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state stackAdoptActionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
}

func (r *stackAdoptActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan stackAdoptActionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *stackAdoptActionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
}

func (r *stackAdoptActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	raw := strings.TrimSpace(req.ID)
	if raw == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected `<environment_id>:<trigger>`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), raw)...)
}

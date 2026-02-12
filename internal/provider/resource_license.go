package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*licenseResource)(nil)
	_ resource.ResourceWithConfigure   = (*licenseResource)(nil)
	_ resource.ResourceWithImportState = (*licenseResource)(nil)
)

func NewLicenseResource() resource.Resource {
	return &licenseResource{}
}

type licenseResource struct {
	client *Client
}

type licenseModel struct {
	ID types.String `tfsdk:"id"`

	Name types.String `tfsdk:"name"`
	Key  types.String `tfsdk:"key"`

	Valid    types.Bool   `tfsdk:"valid"`
	Active   types.Bool   `tfsdk:"active"`
	Hostname types.String `tfsdk:"hostname"`
}

func (r *licenseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

func (r *licenseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Dockhand licensing via `/api/license`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Singleton ID. Always `license`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Licensee name used when setting/updating a license.",
				Optional:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "License key used when setting/updating a license.",
				Optional:            true,
				Sensitive:           true,
			},
			"valid": schema.BoolAttribute{
				MarkdownDescription: "Whether the current license is valid.",
				Computed:            true,
			},
			"active": schema.BoolAttribute{
				MarkdownDescription: "Whether the current license is active.",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Hostname the license is associated with, when provided by Dockhand.",
				Computed:            true,
			},
		},
	}
}

func (r *licenseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *licenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan licenseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, diags := r.applyPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *licenseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	var prior licenseModel
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, status, err := r.client.GetLicense(ctx)
	if err != nil {
		if status == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Dockhand license", err.Error())
		return
	}

	state := modelFromLicenseResponse(prior, current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *licenseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan licenseModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, diags := r.applyPlan(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *licenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
		return
	}

	status, err := r.client.DeleteLicense(ctx)
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Error deleting Dockhand license", err.Error())
		return
	}
}

func (r *licenseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "license")...)
}

func (r *licenseResource) applyPlan(ctx context.Context, plan licenseModel) (licenseModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if r.client == nil {
		diags.AddError("Unconfigured client", "The provider client was not configured.")
		return licenseModel{}, diags
	}

	licenseName := ""
	licenseKey := ""
	hasName := !plan.Name.IsNull() && !plan.Name.IsUnknown()
	hasKey := !plan.Key.IsNull() && !plan.Key.IsUnknown()
	if hasName {
		licenseName = strings.TrimSpace(plan.Name.ValueString())
	}
	if hasKey {
		licenseKey = strings.TrimSpace(plan.Key.ValueString())
	}

	if hasName || hasKey {
		if licenseName == "" || licenseKey == "" {
			diags.AddError("Invalid license configuration", "To set/update a Dockhand license, both `name` and `key` must be set and non-empty.")
			return licenseModel{}, diags
		}

		if _, _, err := r.client.SetLicense(ctx, licensePayload{Name: licenseName, Key: licenseKey}); err != nil {
			diags.AddError("Error updating Dockhand license", err.Error())
			return licenseModel{}, diags
		}
	}

	current, _, err := r.client.GetLicense(ctx)
	if err != nil {
		diags.AddError("Error reading Dockhand license", err.Error())
		return licenseModel{}, diags
	}

	return modelFromLicenseResponse(plan, current), diags
}

func modelFromLicenseResponse(prior licenseModel, in *licenseResponse) licenseModel {
	out := licenseModel{
		ID:     types.StringValue("license"),
		Name:   prior.Name,
		Key:    prior.Key,
		Valid:  types.BoolValue(in.Valid),
		Active: types.BoolValue(in.Active),
	}

	if in.Hostname != nil && *in.Hostname != "" {
		out.Hostname = types.StringValue(*in.Hostname)
	} else {
		out.Hostname = types.StringNull()
	}

	return out
}

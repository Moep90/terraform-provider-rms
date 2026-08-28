package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CompanyResource{}
var _ resource.ResourceWithImportState = &CompanyResource{}
var _ resource.ResourceWithConfigure = &CompanyResource{}

// NewCompanyResource creates a new company resource.
func NewCompanyResource() resource.Resource {
	return &CompanyResource{}
}

// CompanyResource defines the company resource implementation.
type CompanyResource struct {
	client *api.Client
}

// CompanyResourceModel describes the resource data model.
type CompanyResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	CompanyName types.String `tfsdk:"company_name"`
	ParentID    types.Int64  `tfsdk:"parent_id"`
	DeviceCount types.Int64  `tfsdk:"device_count"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

// Metadata returns the resource type name.
func (r *CompanyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_company"
}

// Schema defines the schema for the resource.
func (r *CompanyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS Company. Companies can be hierarchical with parent-child relationships.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the company.",
			},
			"company_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the company.",
			},
			"parent_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The parent company ID. If set, this company becomes a subsidiary of the parent.",
			},
			"device_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of devices associated with this company.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the company was created.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *CompanyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CompanyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"company_name": data.CompanyName.ValueString(),
	}

	if !data.ParentID.IsNull() {
		createReq["parent_id"] = data.ParentID.ValueInt64()
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/companies", createReq, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error creating company",
			fmt.Sprintf("Could not create company: %s", err),
		)
		return
	}

	id, ok := result["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing company ID",
			"Could not parse company ID from API response",
		)
		return
	}
	data.ID = types.Int64Value(int64(id))
	companyName, ok := result["company_name"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing company name",
			"Could not parse company name from API response",
		)
		return
	}
	data.CompanyName = types.StringValue(companyName)

	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}

	if deviceCount, ok := result["device_count"].(float64); ok {
		data.DeviceCount = types.Int64Value(int64(deviceCount))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *CompanyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CompanyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/companies/%d", data.ID.ValueInt64()), nil, &result); err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading company",
			fmt.Sprintf("Could not read company %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}

	if companyName, ok := result["company_name"].(string); ok {
		data.CompanyName = types.StringValue(companyName)
	}

	if parentID, ok := result["parent_id"].(float64); ok {
		data.ParentID = types.Int64Value(int64(parentID))
	}

	if deviceCount, ok := result["device_count"].(float64); ok {
		data.DeviceCount = types.Int64Value(int64(deviceCount))
	}

	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *CompanyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CompanyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{
		"company_name": data.CompanyName.ValueString(),
	}

	if !data.ParentID.IsNull() {
		updateReq["parent_id"] = data.ParentID.ValueInt64()
	}

	var result map[string]interface{}
	if err := r.client.Put(ctx, fmt.Sprintf("/companies/%d", data.ID.ValueInt64()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error updating company",
			fmt.Sprintf("Could not update company %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}

	var updatedResult map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/companies/%d", data.ID.ValueInt64()), nil, &updatedResult); err != nil {
		resp.Diagnostics.AddError(
			"Error reading company",
			fmt.Sprintf("Could not read company %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}

	if deviceCount, ok := updatedResult["device_count"].(float64); ok {
		data.DeviceCount = types.Int64Value(int64(deviceCount))
	}

	if createdAt, ok := updatedResult["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state.
func (r *CompanyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CompanyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, fmt.Sprintf("/companies/%d", data.ID.ValueInt64()), nil); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting company",
			fmt.Sprintf("Could not delete company %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *CompanyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Configure sets the client for the resource.
func (r *CompanyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *api.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

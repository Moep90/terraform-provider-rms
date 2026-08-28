package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

var _ resource.Resource = &TagResource{}
var _ resource.ResourceWithImportState = &TagResource{}

func NewTagResource() resource.Resource {
	return &TagResource{}
}

type TagResource struct {
	client *api.Client
}

type TagResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Color       types.String `tfsdk:"color"`
	CompanyID   types.Int64  `tfsdk:"company_id"`
	DeviceCount types.Int64  `tfsdk:"device_count"`
}

func (r *TagResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *TagResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS Tag. Tags can be used to organize and filter devices.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the tag.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the tag.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"color": schema.StringAttribute{
				Optional:    true,
				Description: "The color of the tag (hex code).",
			},
			"company_id": schema.Int64Attribute{
				Required:    true,
				Description: "The company ID to assign the tag to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"device_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of devices assigned to this tag.",
			},
		},
	}
}

func (r *TagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TagResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name":       data.Name.ValueString(),
		"company_id": data.CompanyID.ValueInt64(),
	}

	if !data.Color.IsNull() {
		createReq["color"] = data.Color.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/tags", createReq, &result); err != nil {
		resp.Diagnostics.AddError("Error creating tag", fmt.Sprintf("Could not create tag: %s", err))
		return
	}

	id, ok := result["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse ID from API response",
		)
		return
	}
	data.ID = types.Int64Value(int64(id))
	name, ok := result["name"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing name",
			"Could not parse name from API response",
		)
		return
	}
	data.Name = types.StringValue(name)

	if deviceCount, ok := result["device_count"].(float64); ok {
		data.DeviceCount = types.Int64Value(int64(deviceCount))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TagResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/tags/%d", data.ID.ValueInt64()), nil, &result); err != nil {
		resp.Diagnostics.AddError("Error reading tag", fmt.Sprintf("Could not read tag %d: %s", data.ID.ValueInt64(), err))
		return
	}

	name, ok := result["name"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing name",
			"Could not parse name from API response",
		)
		return
	}
	data.Name = types.StringValue(name)

	if color, ok := result["color"].(string); ok {
		data.Color = types.StringValue(color)
	}

	if companyID, ok := result["company_id"].(float64); ok {
		data.CompanyID = types.Int64Value(int64(companyID))
	}

	if deviceCount, ok := result["device_count"].(float64); ok {
		data.DeviceCount = types.Int64Value(int64(deviceCount))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TagResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{
		"name": data.Name.ValueString(),
	}

	if !data.Color.IsNull() {
		updateReq["color"] = data.Color.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Put(ctx, fmt.Sprintf("/tags/%d", data.ID.ValueInt64()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError("Error updating tag", fmt.Sprintf("Could not update tag %d: %s", data.ID.ValueInt64(), err))
		return
	}

	var updatedResult map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/tags/%d", data.ID.ValueInt64()), nil, &updatedResult); err != nil {
		resp.Diagnostics.AddError("Error reading tag", fmt.Sprintf("Could not read tag %d: %s", data.ID.ValueInt64(), err))
		return
	}

	if deviceCount, ok := updatedResult["device_count"].(float64); ok {
		data.DeviceCount = types.Int64Value(int64(deviceCount))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TagResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Delete(ctx, fmt.Sprintf("/tags/%d", data.ID.ValueInt64()), nil); err != nil {
		resp.Diagnostics.AddError("Error deleting tag", fmt.Sprintf("Could not delete tag %d: %s", data.ID.ValueInt64(), err))
		return
	}
}

func (r *TagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Configure sets the client for the resource.
func (r *TagResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

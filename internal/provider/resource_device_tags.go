package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &DeviceTagsResource{}
var _ resource.ResourceWithImportState = &DeviceTagsResource{}

func NewDeviceTagsResource() resource.Resource {
	return &DeviceTagsResource{}
}

type DeviceTagsResource struct {
	client *api.Client
}

type DeviceTagsResourceModel struct {
	ID       types.Int64 `tfsdk:"id"`
	DeviceID types.Int64 `tfsdk:"device_id"`
	TagIDs   types.Set   `tfsdk:"tag_ids"`
}

func (r *DeviceTagsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "rms_device_tags"
}

func (r *DeviceTagsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages tag assignments for a Teltonika RMS Device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for this assignment.",
			},
			"device_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the device.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"tag_ids": schema.SetAttribute{
				ElementType: types.Int64Type,
				Required:    true,
				Description: "The list of tag IDs to assign to the device.",
			},
		},
	}
}

// tagAssignment builds the RMS selector body shared by
// PUT /devices/tags/overwrite and PUT /devices/tags/unassign.
func tagAssignment(deviceID int64, tagIDs []int64) map[string]interface{} {
	return map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"device_id": deviceID,
				"tag_id":    tagIDs,
			},
		},
	}
}

// planTagIDs reads the tag_ids set out of a model.
func planTagIDs(ctx context.Context, set types.Set, diags *diag.Diagnostics) []int64 {
	tagIDs := make([]int64, 0, len(set.Elements()))
	diags.Append(set.ElementsAs(ctx, &tagIDs, false)...)
	return tagIDs
}

func (r *DeviceTagsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DeviceTagsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagIDs := planTagIDs(ctx, plan.TagIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// RMS exposes no /devices/{id}/tags collection; tag assignment goes through
	// PUT /devices/tags/overwrite.
	if err := r.client.Put(ctx, "/devices/tags/overwrite", tagAssignment(plan.DeviceID.ValueInt64(), tagIDs), nil); err != nil {
		resp.Diagnostics.AddError(
			"Error assigning tags to device",
			fmt.Sprintf("Could not assign tags to device %d: %s", plan.DeviceID.ValueInt64(), err),
		)
		return
	}

	plan.ID = plan.DeviceID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceTagsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DeviceTagsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The device record carries its tags; there is no per-device tag endpoint.
	var device map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/devices/%d", state.DeviceID.ValueInt64()), nil, &device); err != nil {
		if errors.Is(err, api.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading device tags",
			fmt.Sprintf("Could not read tags for device %d: %s", state.DeviceID.ValueInt64(), err),
		)
		return
	}

	tagIDs := make([]int64, 0)
	if tags, ok := device["tags"].([]interface{}); ok {
		for _, tag := range tags {
			tagMap, ok := tag.(map[string]interface{})
			if !ok {
				continue
			}
			if id, ok := tagMap["id"].(float64); ok {
				tagIDs = append(tagIDs, int64(id))
			}
		}
	}

	newSet, diags := types.SetValueFrom(ctx, types.Int64Type, tagIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.TagIDs = newSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DeviceTagsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DeviceTagsResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagIDs := planTagIDs(ctx, plan.TagIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Put(ctx, "/devices/tags/overwrite", tagAssignment(plan.DeviceID.ValueInt64(), tagIDs), nil); err != nil {
		resp.Diagnostics.AddError(
			"Error updating device tags",
			fmt.Sprintf("Could not update tags for device %d: %s", plan.DeviceID.ValueInt64(), err),
		)
		return
	}

	plan.ID = plan.DeviceID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeviceTagsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DeviceTagsResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagIDs := planTagIDs(ctx, state.TagIDs, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Put(ctx, "/devices/tags/unassign", tagAssignment(state.DeviceID.ValueInt64(), tagIDs), nil); err != nil {
		resp.Diagnostics.AddError(
			"Error clearing device tags",
			fmt.Sprintf("Could not clear tags for device %d: %s", state.DeviceID.ValueInt64(), err),
		)
		return
	}
}

func (r *DeviceTagsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("device_id"), req, resp)
}

func (r *DeviceTagsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

package provider

import (
	"context"
	"fmt"

	"github.com/Moep90/terraform-provider-rms/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &TaskGroupResource{}
var _ resource.ResourceWithImportState = &TaskGroupResource{}

func NewTaskGroupResource() resource.Resource {
	return &TaskGroupResource{}
}

type TaskGroupResource struct {
	client *api.Client
}

type TaskGroupResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CompanyID   types.Int64  `tfsdk:"company_id"`
	Status      types.String `tfsdk:"status"`
	TaskCount   types.Int64  `tfsdk:"task_count"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *TaskGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "rms_task_group"
}

func (r *TaskGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Teltonika RMS Task Group. Task groups organize related tasks for batch operations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Description: "The unique identifier for the task group.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the task group.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the task group.",
			},
			"company_id": schema.Int64Attribute{
				Required:    true,
				Description: "The company ID that owns this task group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the task group (e.g., 'active', 'paused', 'completed').",
			},
			"task_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of tasks in this group.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp in ISO8601 format.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Last update timestamp in ISO8601 format.",
			},
		},
	}
}

func (r *TaskGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *api.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *TaskGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name":       data.Name.ValueString(),
		"company_id": data.CompanyID.ValueInt64(),
	}

	if !data.Description.IsNull() {
		createReq["description"] = data.Description.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Post(ctx, "/task-groups", createReq, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error creating task group",
			fmt.Sprintf("Could not create task group: %s", err),
		)
		return
	}

	id, ok := result["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError(
			"Error parsing task group ID",
			"Could not parse task group ID from API response",
		)
		return
	}
	data.ID = types.Int64Value(int64(id))

	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
	if description, ok := result["description"].(string); ok {
		data.Description = types.StringValue(description)
	}
	if status, ok := result["status"].(string); ok {
		data.Status = types.StringValue(status)
	} else {
		data.Status = types.StringValue("active")
	}
	if companyID, ok := result["company_id"].(float64); ok {
		data.CompanyID = types.Int64Value(int64(companyID))
	}
	if taskCount, ok := result["task_count"].(float64); ok {
		data.TaskCount = types.Int64Value(int64(taskCount))
	} else {
		data.TaskCount = types.Int64Value(0)
	}
	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	} else {
		data.CreatedAt = types.StringValue("2024-01-01T00:00:00Z")
	}
	if updatedAt, ok := result["updated_at"].(string); ok {
		data.UpdatedAt = types.StringValue(updatedAt)
	} else {
		data.UpdatedAt = types.StringValue("2024-01-01T00:00:00Z")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Get(ctx, fmt.Sprintf("/task-groups/%d", data.ID.ValueInt64()), nil, &result); err != nil {
		if err.Error() == "API error 404: 404 Not Found" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading task group",
			fmt.Sprintf("Could not read task group %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}

	if id, ok := result["id"].(float64); ok {
		data.ID = types.Int64Value(int64(id))
	}
	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
	if description, ok := result["description"].(string); ok {
		data.Description = types.StringValue(description)
	}
	if status, ok := result["status"].(string); ok {
		data.Status = types.StringValue(status)
	}
	if companyID, ok := result["company_id"].(float64); ok {
		data.CompanyID = types.Int64Value(int64(companyID))
	}
	if taskCount, ok := result["task_count"].(float64); ok {
		data.TaskCount = types.Int64Value(int64(taskCount))
	}
	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}
	if updatedAt, ok := result["updated_at"].(string); ok {
		data.UpdatedAt = types.StringValue(updatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Description.IsNull() {
		updateReq["description"] = data.Description.ValueString()
	}
	if !data.Name.IsNull() {
		updateReq["name"] = data.Name.ValueString()
	}

	var result map[string]interface{}
	if err := r.client.Put(ctx, fmt.Sprintf("/task-groups/%d", data.ID.ValueInt64()), updateReq, &result); err != nil {
		resp.Diagnostics.AddError(
			"Error updating task group",
			fmt.Sprintf("Could not update task group %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}

	if status, ok := result["status"].(string); ok {
		data.Status = types.StringValue(status)
	} else {
		data.Status = types.StringValue("active")
	}
	if taskCount, ok := result["task_count"].(float64); ok {
		data.TaskCount = types.Int64Value(int64(taskCount))
	} else {
		data.TaskCount = types.Int64Value(0)
	}
	if createdAt, ok := result["created_at"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	} else {
		data.CreatedAt = types.StringValue("2024-01-01T00:00:00Z")
	}
	if updatedAt, ok := result["updated_at"].(string); ok {
		data.UpdatedAt = types.StringValue(updatedAt)
	} else {
		data.UpdatedAt = types.StringValue("2024-01-01T00:00:00Z")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result map[string]interface{}
	if err := r.client.Delete(ctx, fmt.Sprintf("/task-groups/%d", data.ID.ValueInt64()), &result); err != nil {
		if err.Error() == "API error 404: 404 Not Found" {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting task group",
			fmt.Sprintf("Could not delete task group %d: %s", data.ID.ValueInt64(), err),
		)
		return
	}
}

func (r *TaskGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

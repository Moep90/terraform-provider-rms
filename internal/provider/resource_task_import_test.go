package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestTaskResourceImportState guards the advice in the rms_task create error,
// which tells the user to import an existing task. ImportStatePassthroughID
// writes the raw string ID, which an Int64 "id" attribute rejects.
func TestTaskResourceImportState(t *testing.T) {
	ctx := context.Background()
	r, ok := NewTaskResource().(*TaskResource)
	if !ok {
		t.Fatal("NewTaskResource did not return *TaskResource")
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema: %v", schemaResp.Diagnostics)
	}

	resp := &resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "42"}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("import of task 42 failed: %v", resp.Diagnostics)
	}

	var id types.Int64
	if diags := resp.State.GetAttribute(ctx, path.Root("id"), &id); diags.HasError() {
		t.Fatalf("reading imported id: %v", diags)
	}
	if id.ValueInt64() != 42 {
		t.Fatalf("imported id = %d, want 42", id.ValueInt64())
	}
}

func TestTaskResourceImportStateRejectsNonNumeric(t *testing.T) {
	ctx := context.Background()
	r, ok := NewTaskResource().(*TaskResource)
	if !ok {
		t.Fatal("NewTaskResource did not return *TaskResource")
	}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	resp := &resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: schemaResp.Schema,
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
		},
	}
	r.ImportState(ctx, resource.ImportStateRequest{ID: "not-a-number"}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error for a non-numeric import ID, got none")
	}
}

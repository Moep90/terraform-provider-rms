package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

func TestCompanyResourceUpdatePopulatesComputed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/companies/1" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"company_name": "Updated Company",
				"created_at":   "2024-01-01T00:00:00Z",
				"device_count": 5,
				"parent_id":    42,
			})
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/companies/1" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"company_name": "Updated Company",
				"created_at":   "2024-01-01T00:00:00Z",
				"device_count": 5,
				"parent_id":    42,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	r := &CompanyResource{}
	r.client = api.NewClientWithBaseURL(server.URL, "")

	var schResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schResp)
	sch := schResp.Schema

	// Create state with empty computed values
	stateObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"id":           basetypes.Int64Type{},
			"company_name": basetypes.StringType{},
			"parent_id":    basetypes.Int64Type{},
			"device_count": basetypes.Int64Type{},
			"created_at":   basetypes.StringType{},
		},
		map[string]attr.Value{
			"id":           types.Int64Value(1),
			"company_name": types.StringValue("Old Company"),
			"parent_id":    types.Int64Value(42),
			"device_count": types.Int64Null(),
			"created_at":   types.StringNull(),
		},
	)
	if diags.HasError() {
		t.Fatalf("Failed to create state object: %v", diags)
	}
	stateVal, err := stateObj.ToTerraformValue(context.Background())
	if err != nil {
		t.Fatalf("Failed to convert state to terraform value: %v", err)
	}
	state := tfsdk.State{
		Raw:    stateVal,
		Schema: sch,
	}

	// Create plan with updated values (without the computed fields)
	planObj, diags := types.ObjectValue(
		map[string]attr.Type{
			"id":           basetypes.Int64Type{},
			"company_name": basetypes.StringType{},
			"parent_id":    basetypes.Int64Type{},
			"device_count": basetypes.Int64Type{},
			"created_at":   basetypes.StringType{},
		},
		map[string]attr.Value{
			"id":           types.Int64Value(1),
			"company_name": types.StringValue("Updated Company"),
			"parent_id":    types.Int64Value(42),
			"device_count": types.Int64Null(),
			"created_at":   types.StringNull(),
		},
	)
	if diags.HasError() {
		t.Fatalf("Failed to create plan object: %v", diags)
	}
	planVal, err := planObj.ToTerraformValue(context.Background())
	if err != nil {
		t.Fatalf("Failed to convert plan to terraform value: %v", err)
	}
	plan := tfsdk.Plan{
		Raw:    planVal,
		Schema: sch,
	}

	updateResp := &resource.UpdateResponse{State: state}
	updateReq := resource.UpdateRequest{
		State: state,
		Plan:  plan,
	}

	r.Update(context.Background(), updateReq, updateResp)

	if updateResp.Diagnostics.HasError() {
		t.Fatalf("Update failed: %v", updateResp.Diagnostics)
	}

	var result CompanyResourceModel
	if diag := updateResp.State.Get(context.Background(), &result); diag.HasError() {
		t.Fatalf("Failed to get state: %v", diag)
	}

	if result.CreatedAt.IsNull() {
		t.Error("Expected created_at to be populated after update")
	}
	if result.DeviceCount.IsNull() {
		t.Error("Expected device_count to be populated after update")
	}
}

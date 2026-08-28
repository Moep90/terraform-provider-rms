package acceptance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/teltonika-rms/terraform-provider-teltonika-rms/internal/api"
)

func TestMockAPI_CompanyUpdate(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1 for acceptance tests")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v3/companies":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"company_name": req["company_name"],
				"created_at":   "2024-01-01T00:00:00Z",
				"device_count": 0,
			})

		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/companies/1":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"company_name": "Test Company",
				"created_at":   "2024-01-01T00:00:00Z",
				"device_count": 0,
			})

		case r.Method == http.MethodPut && r.URL.Path == "/api/v3/companies/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":           1,
				"company_name": req["company_name"],
				"parent_id":    req["parent_id"],
				"created_at":   "2024-01-01T00:00:00Z",
				"device_count": 0,
			})

		case r.Method == http.MethodDelete && r.URL.Path == "/api/v3/companies/1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := api.NewClientWithBaseURL(server.URL+"/api/v3", "test-token")

	var result map[string]interface{}
	err := client.Post(context.Background(), "/companies", map[string]interface{}{
		"company_name": "Test Company",
	}, &result)
	if err != nil {
		t.Fatalf("Failed to create company: %s", err)
	}

	if result["id"] == nil {
		t.Fatal("Expected company ID in response")
	}

	var updateResult map[string]interface{}
	err = client.Put(context.Background(), "/companies/1", map[string]interface{}{
		"company_name": "Test Company",
		"parent_id":    float64(42),
	}, &updateResult)
	if err != nil {
		t.Fatalf("Failed to update company: %s", err)
	}

	if updateResult["created_at"] == nil {
		t.Error("Expected created_at in update response - N1 failure")
	}

	if updateResult["device_count"] == nil {
		t.Error("Expected device_count in update response - N1 failure")
	}

	err = client.Delete(context.Background(), "/companies/1", nil)
	if err != nil {
		t.Fatalf("Failed to delete company: %s - N2 failure", err)
	}
}

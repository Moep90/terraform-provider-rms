package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientUpdatePopulatesComputed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/companies/1" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
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

	client := NewClient(context.Background(), "test-token")
	client.baseURL = server.URL

	// Verify the PUT response includes computed fields
	var result map[string]interface{}
	err := client.Put(context.Background(), "/companies/1", map[string]interface{}{
		"company_name": "Updated Company",
	}, &result)
	if err != nil {
		t.Fatalf("PUT failed: %s", err)
	}

	if result["created_at"] == nil {
		t.Error("Expected created_at in PUT response")
	}
	if result["device_count"] == nil {
		t.Error("Expected device_count in PUT response")
	}
	if result["parent_id"] == nil {
		t.Error("Expected parent_id in PUT response")
	}
}

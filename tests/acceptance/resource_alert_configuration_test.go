package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAlertConfigurationResource_CRUD(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	alertState := make(map[int]map[string]interface{})
	nextID := 1

	mux.HandleFunc("/alerts-configurations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}
			
			currentID := nextID
			nextID++
			
			if dataArr, ok := req["data"].([]interface{}); ok && len(dataArr) > 0 {
				if data, ok := dataArr[0].(map[string]interface{}); ok {
					alertState[currentID] = map[string]interface{}{
						"id":                float64(currentID),
						"device_id":         data["device_id"],
						"alert_type_id":     data["alert_type_id"],
						"alert_subtype_id":  data["alert_subtype_id"],
						"action":            data["action"],
						"subject":           data["subject"],
						"message":           data["message"],
						"email":             data["email"],
					}
				}
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"id": float64(currentID)}); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/alerts-configurations/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			id, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if r.Method == http.MethodGet {
				if cfg, ok := alertState[id]; ok {
					w.Header().Set("Content-Type", "application/json")
					resp := map[string]interface{}{
						"data": []map[string]interface{}{cfg},
					}
					if err := json.NewEncoder(w).Encode(resp); err != nil {
						t.Logf("error encoding response: %v", err)
					}
					return
				}
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if r.Method == http.MethodPut {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Logf("error decoding request: %v", err)
				}
				if cfg, ok := alertState[id]; ok {
					if action, ok := req["action"].(float64); ok {
						cfg["action"] = fmt.Sprintf("%d", int64(action))
					}
					if subject, ok := req["subject"].(string); ok {
						cfg["subject"] = subject
					}
					if message, ok := req["message"].(string); ok {
						cfg["message"] = message
					}
					if email, ok := req["email"].(string); ok {
						cfg["email"] = email
					}
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(cfg); err != nil {
						t.Logf("error encoding response: %v", err)
					}
					return
				}
				w.WriteHeader(http.StatusNotFound)
				return
			}

			if r.Method == http.MethodDelete {
				delete(alertState, id)
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAlertConfigConfig(server.URL, 123, 5, 3, 1, "Test Alert", "Test message", "test@example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "device_id", "123"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "alert_type_id", "5"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "alert_subtype_id", "3"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "action", "1"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "subject", "Test Alert"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "message", "Test message"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "email", "test@example.com"),
				),
			},
			{
				Config: testAlertConfigConfig(server.URL, 123, 5, 3, 2, "Updated Alert", "Updated message", "updated@example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "action", "2"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "subject", "Updated Alert"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "message", "Updated message"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "email", "updated@example.com"),
				),
			},
		},
	})
}

func testAlertConfigConfig(baseURL string, deviceID, alertTypeID, alertSubtypeID, action int64, subject, message, email string) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_alert_configuration" "test" {
  device_id          = %d
  alert_type_id      = %d
  alert_subtype_id   = %d
  action             = %d
  subject            = "%s"
  message            = "%s"
  email              = "%s"
}
`, baseURL, deviceID, alertTypeID, alertSubtypeID, action, subject, message, email)
}

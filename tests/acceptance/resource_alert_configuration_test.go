package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// alertConfigurationServer serves the alert configuration operations RMS
// defines: POST and GET on /devices/alerts-configurations, GET and PUT on
// /alerts-configurations/{id} and DELETE on
// /devices/{device_id}/alerts-configurations/{alert_id}. Every other path 404s,
// so a call to an undefined route fails the test.
//
// deleted reports whether the configuration was removed, either by Terraform or
// out of band through the returned function.
func alertConfigurationServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	var mu sync.Mutex
	store := map[string]interface{}{
		"id":               float64(1),
		"device_id":        float64(123),
		"alert_type_id":    float64(5),
		"alert_subtype_id": float64(3),
		"action":           float64(1),
		"subject":          "Test Alert",
		"message":          "Test message",
		"email":            "test@example.com",
		"created_at":       "2024-01-01T00:00:00Z",
		"updated_at":       "2024-01-01T00:00:00Z",
	}
	deleted := false

	snapshot := func() map[string]interface{} {
		out := make(map[string]interface{}, len(store))
		for k, v := range store {
			out[k] = v
		}
		return out
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		// Create. RMS answers {"success": true} with no id, so the provider
		// resolves the id from the collection afterwards.
		case r.Method == http.MethodPost && r.URL.Path == "/devices/alerts-configurations":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			for _, field := range []string{"device_id", "alert_type_id", "alert_subtype_id", "action", "subject", "message", "email"} {
				if v, ok := req[field]; ok {
					store[field] = v
				}
			}
			deleted = false
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})

		case r.Method == http.MethodGet && r.URL.Path == "/devices/alerts-configurations":
			items := []interface{}{}
			if !deleted {
				items = append(items, snapshot())
			}
			writeRMSList(t, w, items)

		// Read by id. RMS serves this off /alerts-configurations, not
		// /devices/alerts-configurations.
		case r.Method == http.MethodGet && r.URL.Path == "/alerts-configurations/1":
			if deleted {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    snapshot(),
			})

		case r.Method == http.MethodPut && r.URL.Path == "/alerts-configurations/1":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)
			for _, field := range []string{"action", "subject", "message", "email"} {
				if v, ok := req[field]; ok {
					store[field] = v
				}
			}
			store["updated_at"] = "2024-01-02T00:00:00Z"
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"data":    snapshot(),
			})

		// Delete is scoped to the owning device.
		case r.Method == http.MethodDelete && r.URL.Path == "/devices/123/alerts-configurations/1":
			deleted = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	deleteOutOfBand := func() {
		mu.Lock()
		defer mu.Unlock()
		deleted = true
	}

	return server, deleteOutOfBand
}

func TestAccAlertConfiguration(t *testing.T) {
	server, _ := alertConfigurationServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
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
					resource.TestCheckResourceAttrSet("rms_alert_configuration.test", "id"),
				),
			},
			{
				Config: testAlertConfigConfig(server.URL, 123, 5, 3, 2, "Updated Alert", "Updated message", "updated@example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "device_id", "123"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "alert_type_id", "5"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "action", "2"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "subject", "Updated Alert"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "message", "Updated message"),
					resource.TestCheckResourceAttr("rms_alert_configuration.test", "email", "updated@example.com"),
				),
			},
		},
	})
}

// TestAccAlertConfiguration_ReadRemovesDeletedConfig covers the 404-on-read
// rule: a configuration deleted out of band must leave state and the next plan
// must show a create.
func TestAccAlertConfiguration_ReadRemovesDeletedConfig(t *testing.T) {
	server, deleteOutOfBand := alertConfigurationServer(t)
	defer server.Close()

	config := testAlertConfigConfig(server.URL, 123, 5, 3, 1, "Test Alert", "Test message", "test@example.com")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr("rms_alert_configuration.test", "subject", "Test Alert"),
			},
			{
				PreConfig: deleteOutOfBand,
				Config:    config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_alert_configuration.test", plancheck.ResourceActionCreate),
					},
				},
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

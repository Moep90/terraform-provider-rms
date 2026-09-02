package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// emailConfigurationMux serves the operations RMS defines: GET and POST on
// /email-configurations and PUT and DELETE on /email-configurations/{id}.
// There is no read-by-id, so that path 404s on GET.
func emailConfigurationMux(t *testing.T, emailState map[int]map[string]interface{}) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/email-configurations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			configs := make([]interface{}, 0, len(emailState))
			for _, cfg := range emailState {
				configs = append(configs, cfg)
			}
			writeRMSList(t, w, configs)
			return
		}

		if r.Method == http.MethodPost {
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("error decoding request: %v", err)
			}
			newID := 2
			for id := range emailState {
				if id >= newID {
					newID = id + 1
				}
			}
			emailState[newID] = map[string]interface{}{
				"id":       float64(newID),
				"name":     req["name"],
				"host":     req["host"],
				"port":     req["port"],
				"email":    req["email"],
				"username": req["username"],
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"id": newID}); err != nil {
				t.Logf("error encoding response: %v", err)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc("/email-configurations/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > 22 {
			var id int
			if _, err := fmt.Sscanf(path[22:], "%d", &id); err != nil {
				t.Logf("error parsing ID: %v", err)
			}

			if r.Method == http.MethodPut {
				var req map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Logf("error decoding request: %v", err)
				}
				if cfg, ok := emailState[id]; ok {
					if name, ok := req["name"].(string); ok {
						cfg["name"] = name
					}
					if host, ok := req["host"].(string); ok {
						cfg["host"] = host
					}
					if port, ok := req["port"].(float64); ok {
						cfg["port"] = int64(port)
					}
					if email, ok := req["email"].(string); ok {
						cfg["email"] = email
					}
					if username, ok := req["username"].(string); ok {
						cfg["username"] = username
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
				delete(emailState, id)
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	return mux
}

func TestEmailConfigurationResource_CRUD(t *testing.T) {
	emailState := newEmailConfigState()
	server := httptest.NewServer(emailConfigurationMux(t, emailState))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testEmailConfigConfig(server.URL, "Test SMTP", "smtp.example.com", 587, "test@example.com", "testuser", "secret"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_email_configuration.test", "name", "Test SMTP"),
					resource.TestCheckResourceAttr("rms_email_configuration.test", "host", "smtp.example.com"),
					resource.TestCheckResourceAttr("rms_email_configuration.test", "port", "587"),
					resource.TestCheckResourceAttr("rms_email_configuration.test", "email", "test@example.com"),
					resource.TestCheckResourceAttr("rms_email_configuration.test", "username", "testuser"),
				),
			},
			{
				Config: testEmailConfigConfig(server.URL, "Updated SMTP", "smtp.updated.com", 465, "updated@example.com", "updateduser", "newsecret"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_email_configuration.test", "name", "Updated SMTP"),
					resource.TestCheckResourceAttr("rms_email_configuration.test", "host", "smtp.updated.com"),
					resource.TestCheckResourceAttr("rms_email_configuration.test", "port", "465"),
					resource.TestCheckResourceAttr("rms_email_configuration.test", "email", "updated@example.com"),
					resource.TestCheckResourceAttr("rms_email_configuration.test", "username", "updateduser"),
				),
			},
		},
	})
}

func testEmailConfigConfig(baseURL string, name, host string, port int64, email, username, password string) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_email_configuration" "test" {
  name     = "%s"
  host     = "%s"
  port     = %d
  email    = "%s"
  username = "%s"
  password = "%s"
}
`, baseURL, name, host, port, email, username, password)
}

func newEmailConfigState() map[int]map[string]interface{} {
	return map[int]map[string]interface{}{1: {
		"id":       float64(1),
		"name":     "Test SMTP",
		"host":     "smtp.example.com",
		"port":     int64(587),
		"email":    "test@example.com",
		"username": "testuser",
	}}
}

// TestEmailConfigurationResource_ReadRemovesDeletedConfig covers the
// read-from-list rule: a configuration deleted out of band drops out of
// /email-configurations, so it must leave state and the next plan must show a
// create.
func TestEmailConfigurationResource_ReadRemovesDeletedConfig(t *testing.T) {
	emailState := newEmailConfigState()
	server := httptest.NewServer(emailConfigurationMux(t, emailState))
	defer server.Close()

	config := testEmailConfigConfig(server.URL, "Test SMTP", "smtp.example.com", 587, "test@example.com", "testuser", "secret")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr("rms_email_configuration.test", "name", "Test SMTP"),
			},
			{
				PreConfig: func() {
					for id := range emailState {
						delete(emailState, id)
					}
				},
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_email_configuration.test", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

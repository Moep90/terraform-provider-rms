package acceptance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// invitationServer serves the invitation operations RMS defines: POST
// /users/invite, GET /users/invitations and DELETE /users/invitations/{id}.
// There is no read-by-id, so GET /users/invitations/{id} 404s the way RMS does.
//
// The returned function drops the invitation from the collection to simulate an
// out-of-band delete.
func invitationServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	var mu sync.Mutex
	invitations := map[int]map[string]interface{}{}
	nextID := 1

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/users/invite":
			var req map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&req)

			id := nextID
			nextID++
			invitations[id] = map[string]interface{}{
				"id":         float64(id),
				"email":      req["email"],
				"role":       req["role"],
				"company_id": req["company_id"],
				"created_at": "2024-01-01T00:00:00Z",
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         float64(id),
				"created_at": "2024-01-01T00:00:00Z",
			})

		case r.Method == http.MethodGet && r.URL.Path == "/users/invitations":
			items := make([]interface{}, 0, len(invitations))
			for _, invitation := range invitations {
				items = append(items, invitation)
			}
			writeRMSList(t, w, items)

		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/users/invitations/"):
			id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/users/invitations/"))
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			delete(invitations, id)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	deleteOutOfBand := func() {
		mu.Lock()
		defer mu.Unlock()
		for id := range invitations {
			delete(invitations, id)
		}
	}

	return server, deleteOutOfBand
}

func TestAccInvitation(t *testing.T) {
	server, _ := invitationServer(t)
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testInvitationConfig(server.URL, "invitee@example.com", "end_user", 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_invitation.test", "email", "invitee@example.com"),
					resource.TestCheckResourceAttr("rms_invitation.test", "role", "end_user"),
					resource.TestCheckResourceAttr("rms_invitation.test", "company_id", "1"),
					resource.TestCheckResourceAttrSet("rms_invitation.test", "id"),
				),
			},
		},
	})
}

// TestAccInvitation_ReadRemovesDeletedInvitation covers the read-from-list
// rule: an invitation revoked out of band drops out of /users/invitations, so
// it must leave state and the next plan must show a create.
func TestAccInvitation_ReadRemovesDeletedInvitation(t *testing.T) {
	server, deleteOutOfBand := invitationServer(t)
	defer server.Close()

	config := testInvitationConfig(server.URL, "invitee@example.com", "end_user", 1)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr("rms_invitation.test", "email", "invitee@example.com"),
			},
			{
				PreConfig: deleteOutOfBand,
				Config:    config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_invitation.test", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testInvitationConfig(baseURL, email, role string, companyID int64) string {
	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_invitation" "test" {
  email      = "%s"
  role       = "%s"
  company_id = %d
}
`, baseURL, email, role, companyID)
}

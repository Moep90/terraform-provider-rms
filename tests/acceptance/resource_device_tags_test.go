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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// deviceTagsMux serves the tag operations RMS defines: PUT on
// /devices/tags/overwrite and /devices/tags/unassign, with the assignment read
// back off the device record at GET /devices/{id}. There is no
// /devices/{id}/tags collection, so that path 404s.
func deviceTagsMux(t *testing.T, deviceTagState map[int][]int) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()

	// selector pulls the single device_id / tag_id pair out of an assignment body.
	selector := func(w http.ResponseWriter, r *http.Request) (int, []int, bool) {
		var req struct {
			Data []struct {
				DeviceID int   `json:"device_id"`
				TagID    []int `json:"tag_id"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("%s %s must carry a JSON body: %v", r.Method, r.URL.Path, err)
			w.WriteHeader(http.StatusBadRequest)
			return 0, nil, false
		}
		if len(req.Data) != 1 {
			t.Errorf("%s %s expects exactly one data entry, got %d", r.Method, r.URL.Path, len(req.Data))
			w.WriteHeader(http.StatusBadRequest)
			return 0, nil, false
		}
		return req.Data[0].DeviceID, req.Data[0].TagID, true
	}

	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			resp := map[string]interface{}{
				"id":                123,
				"name":              "Test Device",
				"device_series":     "rut",
				"serial":            "SN123",
				"company_id":        float64(1),
				"mac":               "00:11:22:33:44:55",
				"status":            "online",
				"firmware":          "v1.0",
				"created_at":        "2024-01-01T00:00:00Z",
				"monitoring_enable": false,
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				t.Logf("error encoding response: %v", err)
			}

		case http.MethodDelete:
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("DELETE /devices must carry a JSON body: %v", err)
			}
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	mux.HandleFunc("/devices/tags/overwrite", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		deviceID, tagIDs, ok := selector(w, r)
		if !ok {
			return
		}
		deviceTagState[deviceID] = tagIDs
		writeSuccess(t, w)
	})

	mux.HandleFunc("/devices/tags/unassign", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		deviceID, tagIDs, ok := selector(w, r)
		if !ok {
			return
		}

		remove := make(map[int]bool, len(tagIDs))
		for _, id := range tagIDs {
			remove[id] = true
		}
		kept := make([]int, 0, len(deviceTagState[deviceID]))
		for _, id := range deviceTagState[deviceID] {
			if !remove[id] {
				kept = append(kept, id)
			}
		}
		deviceTagState[deviceID] = kept
		writeSuccess(t, w)
	})

	mux.HandleFunc("/devices/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
		deviceID, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodPut {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		tagIDs := deviceTagState[deviceID]
		tags := make([]interface{}, 0, len(tagIDs))
		for _, id := range tagIDs {
			tags = append(tags, map[string]interface{}{
				"id":   float64(id),
				"name": fmt.Sprintf("Tag %d", id),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":                deviceID,
			"name":              "Test Device",
			"device_series":     "rut",
			"status":            "online",
			"firmware":          "v1.0",
			"created_at":        "2024-01-01T00:00:00Z",
			"monitoring_enable": false,
			"tags":              tags,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Logf("error encoding response: %v", err)
		}
	})

	return mux
}

func writeSuccess(t *testing.T, w http.ResponseWriter) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"success": true}); err != nil {
		t.Logf("error encoding response: %v", err)
	}
}

func TestDeviceTagsResource_Assign(t *testing.T) {
	deviceTagState := map[int][]int{123: {10, 20}}
	server := httptest.NewServer(deviceTagsMux(t, deviceTagState))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testDeviceTagsConfig(server.URL, []int{10, 20}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_device_tags.assignment", "device_id", "123"),
					resource.TestCheckTypeSetElemAttr("rms_device_tags.assignment", "tag_ids.*", "10"),
					resource.TestCheckTypeSetElemAttr("rms_device_tags.assignment", "tag_ids.*", "20"),
				),
			},
			{
				Config: testDeviceTagsConfig(server.URL, []int{10, 20, 30}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_device_tags.assignment", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_device_tags.assignment", "device_id", "123"),
					resource.TestCheckTypeSetElemAttr("rms_device_tags.assignment", "tag_ids.*", "10"),
					resource.TestCheckTypeSetElemAttr("rms_device_tags.assignment", "tag_ids.*", "20"),
					resource.TestCheckTypeSetElemAttr("rms_device_tags.assignment", "tag_ids.*", "30"),
				),
			},
			{
				Config: testDeviceTagsConfig(server.URL, []int{}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rms_device_tags.assignment", "device_id", "123"),
					resource.TestCheckResourceAttr("rms_device_tags.assignment", "tag_ids.#", "0"),
				),
			},
		},
	})
}

// TestDeviceTagsResource_ReadRemovesDeletedDevice covers the 404-on-read rule:
// once the device is gone the assignment has nothing to describe, so it must
// leave state.
func TestDeviceTagsResource_ReadRemovesDeletedDevice(t *testing.T) {
	deviceTagState := map[int][]int{123: {10}}
	deviceGone := false

	mux := deviceTagsMux(t, deviceTagState)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if deviceGone && r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/devices/123") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		mux.ServeHTTP(w, r)
	}))
	defer server.Close()

	// The assignment stands alone here so the refresh only exercises
	// rms_device_tags.
	config := fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_device_tags" "assignment" {
  device_id = 123
  tag_ids   = [10]
}
`, server.URL)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr("rms_device_tags.assignment", "device_id", "123"),
			},
			{
				PreConfig: func() { deviceGone = true },
				Config:    config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("rms_device_tags.assignment", plancheck.ResourceActionCreate),
					},
				},
				// The device stays absent, so the post-apply refresh drops the
				// assignment again.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testDeviceTagsConfig(baseURL string, tagIDs []int) string {
	tagIDsStr := "["
	for i, id := range tagIDs {
		if i > 0 {
			tagIDsStr += ", "
		}
		tagIDsStr += fmt.Sprintf("%d", id)
	}
	tagIDsStr += "]"

	return fmt.Sprintf(`
provider "rms" {
  token    = "test-token"
  base_url = "%s"
}

resource "rms_device" "test" {
  name              = "Test Device"
  device_series     = "rut"
  serial            = "SN123"
  company_id        = 1
  mac               = "00:11:22:33:44:55"
  monitoring_enable = false
}

resource "rms_device_tags" "assignment" {
  device_id = rms_device.test.id
  tag_ids   = %s
}
`, baseURL, tagIDsStr)
}

package provider

import (
	"context"
	"strconv"

	"github.com/Moep90/terraform-provider-rms/internal/api"
)

// listPageSize is the page size used when scanning an RMS collection. RMS
// defaults to 10 results per request, which would hide most objects.
const listPageSize = 100

// findInList pages through an RMS collection endpoint and returns the item
// whose "id" equals id. RMS exposes no read-by-id operation for tasks, email
// configurations, invitations or VPN hubs, so their Read has to scan the
// collection instead.
//
// It returns api.ErrNotFound when the id is absent from the whole collection,
// which lets callers reuse the existing state-removal handling.
func findInList(ctx context.Context, client *api.Client, path string, id int64) (map[string]interface{}, error) {
	for offset := 0; ; offset += listPageSize {
		params := map[string]string{
			"limit":  strconv.Itoa(listPageSize),
			"offset": strconv.Itoa(offset),
		}

		var items []map[string]interface{}
		if err := client.Get(ctx, path, params, &items); err != nil {
			return nil, err
		}

		for _, item := range items {
			if v, ok := item["id"].(float64); ok && int64(v) == id {
				return item, nil
			}
		}

		if len(items) < listPageSize {
			return nil, api.ErrNotFound
		}
	}
}

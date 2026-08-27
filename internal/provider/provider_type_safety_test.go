package provider

import (
	"os"
	"strings"
	"testing"
)

// TestResourceTypeSafety verifies resources handle type assertion failures gracefully
func TestResourceTypeSafety(t *testing.T) {
	files := []string{
		"../../internal/provider/resource_company.go",
		"../../internal/provider/resource_device.go",
		"../../internal/provider/resource_tag.go",
		"../../internal/provider/resource_user.go",
		"../../internal/provider/resource_invitation.go",
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read %s: %s", file, err)
		}

		// Check for unsafe type assertions: .(string) or .(float64) not preceded by "ok :"
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "//") {
				continue
			}
			// Look for .(string) or .(float64) that doesn't have ", ok" pattern nearby
			if strings.Contains(line, ".(string)") || strings.Contains(line, ".(float64)") {
				// Check if this is part of a safe comma-ok pattern
				if !strings.Contains(line, ", ok") && !strings.Contains(line, "if ") {
					t.Errorf("%s:%d contains unsafe type assertion: %s", file, i+1, trimmed)
				}
			}
		}
	}
}

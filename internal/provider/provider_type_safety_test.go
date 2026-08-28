package provider

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResourceTypeSafety verifies resources handle type assertion failures gracefully
func TestResourceTypeSafety(t *testing.T) {
	// Cover every resource and data source rather than a hand-maintained list,
	// which silently excluded each newly added file.
	var files []string
	for _, pattern := range []string{"resource_*.go", "datasource_*.go"} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("Failed to glob %s: %s", pattern, err)
		}
		for _, match := range matches {
			if strings.HasSuffix(match, "_test.go") {
				continue
			}
			files = append(files, match)
		}
	}

	if len(files) == 0 {
		t.Fatal("No provider files matched; the type safety check would pass vacuously")
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

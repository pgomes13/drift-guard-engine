package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"drift-guard-diff-engine/internal/parser/openapi/helpers"
	"drift-guard-diff-engine/pkg/schema"

	"gopkg.in/yaml.v3"
)

// Parse reads and parses an OpenAPI 3.x file (JSON or YAML) into a normalized Schema.
func Parse(path string) (*schema.Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	var raw helpers.RawOpenAPI
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing YAML %s: %w", path, err)
		}
	case ".json":
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("parsing JSON %s: %w", path, err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &raw); err != nil {
			if err2 := json.Unmarshal(data, &raw); err2 != nil {
				return nil, fmt.Errorf("unsupported format for %s", path)
			}
		}
	}

	return helpers.Normalize(&raw), nil
}

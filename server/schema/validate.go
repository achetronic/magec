package schema

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

// Validate validates data against a JSON Schema defined as map[string]interface{}.
// Returns nil if validation passes.
func Validate(schemaMap map[string]interface{}, data map[string]interface{}) error {
	if schemaMap == nil {
		return nil
	}

	raw, err := json.Marshal(schemaMap)
	if err != nil {
		return fmt.Errorf("marshaling schema: %w", err)
	}

	var s jsonschema.Schema
	if err := json.Unmarshal(raw, &s); err != nil {
		return fmt.Errorf("parsing schema: %w", err)
	}

	resolved, err := s.Resolve(nil)
	if err != nil {
		return fmt.Errorf("resolving schema: %w", err)
	}

	return resolved.Validate(data)
}

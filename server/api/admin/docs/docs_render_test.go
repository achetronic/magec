package docs

import (
	"encoding/json"
	"testing"

	"github.com/swaggo/swag"
)

// TestReadDocRendersValidJSON is a canary for a subtle swaggo failure mode:
// when a Go doc comment copied into the swagger template contains literal
// template delimiters (e.g. the `{{ input }}` flow placeholders documented
// on store.FlowNode), template parsing fails and swag silently serves the
// RAW template, showing `{{.Version}}` and friends in the Swagger UI.
// Rendering must produce valid JSON with the info block actually expanded.
func TestReadDocRendersValidJSON(t *testing.T) {
	doc, err := swag.ReadDoc(SwaggerInfo.InstanceName())
	if err != nil {
		t.Fatalf("ReadDoc(%q) failed: %v", SwaggerInfo.InstanceName(), err)
	}

	var rendered struct {
		Info struct {
			Version string `json:"version"`
			Title   string `json:"title"`
		} `json:"info"`
	}
	if err := json.Unmarshal([]byte(doc), &rendered); err != nil {
		t.Fatalf("rendered swagger doc is not valid JSON (raw template served?): %v", err)
	}
	if rendered.Info.Version != SwaggerInfo.Version {
		t.Fatalf("info.version = %q, want %q (placeholder not expanded)", rendered.Info.Version, SwaggerInfo.Version)
	}
	if rendered.Info.Title != SwaggerInfo.Title {
		t.Fatalf("info.title = %q, want %q (placeholder not expanded)", rendered.Info.Title, SwaggerInfo.Title)
	}
}

package openapi

import (
	"strings"
	"testing"

	"cloudpact/parser/grammar"
)

func TestGenerate(t *testing.T) {
	src := `model User {
    id: Int
    name: String
}`
	f, err := grammar.ParseString(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	yaml, err := Generate(f)
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}
	checks := []string{
		"openapi: \"3.0.0\"",
		"title: \"Cloudpact API\"",
		"User:",
		"type: \"integer\"",
		"type: \"string\"",
	}
	for _, c := range checks {
		if !strings.Contains(yaml, c) {
			t.Fatalf("expected YAML to contain %q\n%s", c, yaml)
		}
	}
}

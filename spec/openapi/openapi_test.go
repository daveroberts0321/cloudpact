package openapi

import (
	"strings"
	"testing"

	"github.com/daveroberts0321/cloudpact/parser/grammar"
)

func TestGenerate(t *testing.T) {
	src := `define record Person
    first: text
    last: text

function hello(name: text) returns text
    why: "Greets a user"
    do:
        return "hi"`
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
		"title: \"CloudPact API\"",
		"/hello:",
	}
	for _, c := range checks {
		if !strings.Contains(yaml, c) {
			t.Fatalf("expected YAML to contain %q\n%s", c, yaml)
		}
	}
}

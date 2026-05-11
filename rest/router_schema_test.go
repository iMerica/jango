package rest_test

import (
	"testing"

	"github.com/iMerica/jango/examples/blogapi"
	"github.com/iMerica/jango/orm"
	"github.com/iMerica/jango/rest"
)

func TestDefaultRouterAndSchema(t *testing.T) {
	router := rest.NewDefaultRouter()
	router.Register("posts", "post", blogapi.PostViewSet())
	patterns := router.URLPatterns()
	if len(patterns) != 2 {
		t.Fatalf("expected collection and detail patterns, got %d", len(patterns))
	}
	if patterns[0].Name != "post-list" || patterns[1].Name != "post-detail" {
		t.Fatalf("unexpected pattern names: %q %q", patterns[0].Name, patterns[1].Name)
	}

	schema := rest.SchemaGenerator{Title: "Blog API", Patterns: patterns, Models: []*orm.ModelMeta{blogapi.PostMeta}}.OpenAPI()
	if schema["openapi"] != "3.0.3" {
		t.Fatalf("unexpected openapi version: %#v", schema["openapi"])
	}
	paths := schema["paths"].(map[string]interface{})
	if _, ok := paths["/posts/"]; !ok {
		t.Fatalf("expected posts path in schema: %#v", paths)
	}
}

func TestSimpleMetadata(t *testing.T) {
	data := rest.SimpleMetadata{}.DetermineMetadata(nil, nil, blogapi.PostMeta)
	fields := data["fields"].(map[string]interface{})
	if _, ok := fields["title"]; !ok {
		t.Fatalf("expected title metadata, got %#v", fields)
	}
}

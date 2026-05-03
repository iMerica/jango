package urls

import (
	"net/http"
	"testing"
)

func dummyView(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func TestResolverSimplePattern(t *testing.T) {
	patterns := []Pattern{
		Path("/", http.HandlerFunc(dummyView), "index"),
		Path("/about/", http.HandlerFunc(dummyView), "about"),
	}
	resolver := NewResolver(patterns)

	match, err := resolver.Resolve("/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Name != "index" {
		t.Errorf("expected name=index, got %s", match.Name)
	}

	match, err = resolver.Resolve("/about/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Name != "about" {
		t.Errorf("expected name=about, got %s", match.Name)
	}
}

func TestResolverPathWithParams(t *testing.T) {
	patterns := []Pattern{
		Path("/articles/<int:year>/", http.HandlerFunc(dummyView), "article-year"),
	}
	resolver := NewResolver(patterns)

	match, err := resolver.Resolve("/articles/2024/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Params["year"] != "2024" {
		t.Errorf("expected year=2024, got %s", match.Params["year"])
	}
}

func TestResolverPathWithMultipleParams(t *testing.T) {
	patterns := []Pattern{
		Path("/articles/<int:year>/<slug:slug>/", http.HandlerFunc(dummyView), "article-detail"),
	}
	resolver := NewResolver(patterns)

	match, err := resolver.Resolve("/articles/2024/hello-world/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Params["year"] != "2024" {
		t.Errorf("expected year=2024, got %s", match.Params["year"])
	}
	if match.Params["slug"] != "hello-world" {
		t.Errorf("expected slug=hello-world, got %s", match.Params["slug"])
	}
}

func TestResolverPathWithDefaultConverter(t *testing.T) {
	patterns := []Pattern{
		Path("/users/<username>/", http.HandlerFunc(dummyView), "user-profile"),
	}
	resolver := NewResolver(patterns)

	match, err := resolver.Resolve("/users/johndoe/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Params["username"] != "johndoe" {
		t.Errorf("expected username=johndoe, got %s", match.Params["username"])
	}
}

func TestResolverUUID(t *testing.T) {
	patterns := []Pattern{
		Path("/items/<uuid:id>/", http.HandlerFunc(dummyView), "item-detail"),
	}
	resolver := NewResolver(patterns)

	match, err := resolver.Resolve("/items/550e8400-e29b-41d4-a716-446655440000/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Params["id"] != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected uuid match, got %s", match.Params["id"])
	}
}

func TestResolverInvalidIntConverter(t *testing.T) {
	patterns := []Pattern{
		Path("/articles/<int:year>/", http.HandlerFunc(dummyView), "article-year"),
	}
	resolver := NewResolver(patterns)

	_, err := resolver.Resolve("/articles/notayear/")
	if err == nil {
		t.Error("expected error for non-integer path, got nil")
	}
}

func TestResolverRePath(t *testing.T) {
	patterns := []Pattern{
		RePath(`/articles/(?P<year>\d{4})/`, http.HandlerFunc(dummyView), "article-re"),
	}
	resolver := NewResolver(patterns)

	match, err := resolver.Resolve("/articles/2024/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Params["year"] != "2024" {
		t.Errorf("expected year=2024, got %s", match.Params["year"])
	}
}

func TestResolverNotFound(t *testing.T) {
	patterns := []Pattern{
		Path("/exists/", http.HandlerFunc(dummyView), "exists"),
	}
	resolver := NewResolver(patterns)

	_, err := resolver.Resolve("/nonexistent/")
	if err == nil {
		t.Error("expected error for unresolved path")
	}
}

func TestResolverInclude(t *testing.T) {
	subPatterns := []Pattern{
		Path("/", http.HandlerFunc(dummyView), "api-root"),
		Path("/items/", http.HandlerFunc(dummyView), "api-items"),
		Path("/items/<int:id>/", http.HandlerFunc(dummyView), "api-item-detail"),
	}
	patterns := []Pattern{
		Include("/api", subPatterns, "api"),
		Path("/", http.HandlerFunc(dummyView), "index"),
	}
	resolver := NewResolver(patterns)

	match, err := resolver.Resolve("/api/items/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Name != "api:api-items" {
		t.Errorf("expected name=api:api-items, got %s", match.Name)
	}

	match, err = resolver.Resolve("/api/items/42/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Params["id"] != "42" {
		t.Errorf("expected id=42, got %s", match.Params["id"])
	}
}

func TestResolverNestedInclude(t *testing.T) {
	innerPatterns := []Pattern{
		Path("/", http.HandlerFunc(dummyView), "admin-dashboard"),
	}
	outerPatterns := []Pattern{
		Include("/admin", innerPatterns, "admin"),
	}
	patterns := []Pattern{
		Include("/api", outerPatterns, "api"),
	}
	resolver := NewResolver(patterns)

	match, err := resolver.Resolve("/api/admin/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Name != "api:admin:admin-dashboard" {
		t.Errorf("expected name=api:admin:admin-dashboard, got %s", match.Name)
	}
}

func TestResolverIncludeNoNamespace(t *testing.T) {
	subPatterns := []Pattern{
		Path("/list/", http.HandlerFunc(dummyView), "list"),
	}
	patterns := []Pattern{
		Include("/items", subPatterns, ""),
	}
	resolver := NewResolver(patterns)

	match, err := resolver.Resolve("/items/list/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Name != "list" {
		t.Errorf("expected name=list, got %s", match.Name)
	}
}

func TestResolverReverse(t *testing.T) {
	patterns := []Pattern{
		Path("/", http.HandlerFunc(dummyView), "index"),
		Path("/about/", http.HandlerFunc(dummyView), "about"),
		Path("/articles/<int:year>/", http.HandlerFunc(dummyView), "article-year"),
	}
	resolver := NewResolver(patterns)

	url, err := resolver.Reverse("index", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "/" {
		t.Errorf("expected /, got %s", url)
	}

	url, err = resolver.Reverse("about", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "/about/" {
		t.Errorf("expected /about/, got %s", url)
	}

	url, err = resolver.Reverse("article-year", map[string]string{"year": "2024"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "/articles/2024/" {
		t.Errorf("expected /articles/2024/, got %s", url)
	}
}

func TestResolverReverseMissingParam(t *testing.T) {
	patterns := []Pattern{
		Path("/articles/<int:year>/", http.HandlerFunc(dummyView), "article-year"),
	}
	resolver := NewResolver(patterns)

	_, err := resolver.Reverse("article-year", nil)
	if err == nil {
		t.Error("expected error for missing parameter")
	}
}

func TestResolverReverseUnknownName(t *testing.T) {
	patterns := []Pattern{
		Path("/", http.HandlerFunc(dummyView), "index"),
	}
	resolver := NewResolver(patterns)

	_, err := resolver.Reverse("nonexistent", nil)
	if err == nil {
		t.Error("expected error for unknown name")
	}
}

func TestResolverReverseWithNamespace(t *testing.T) {
	subPatterns := []Pattern{
		Path("/", http.HandlerFunc(dummyView), "root"),
		Path("/items/", http.HandlerFunc(dummyView), "items"),
	}
	patterns := []Pattern{
		Include("/api", subPatterns, "api"),
	}
	resolver := NewResolver(patterns)

	url, err := resolver.Reverse("api:root", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "/api/" {
		t.Errorf("expected /api/, got %s", url)
	}

	url, err = resolver.Reverse("api:items", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "/api/items/" {
		t.Errorf("expected /api/items/, got %s", url)
	}
}

func TestResolverRegisterPattern(t *testing.T) {
	resolver := NewResolver([]Pattern{})
	resolver.Register(Path("/new/", http.HandlerFunc(dummyView), "new"))

	match, err := resolver.Resolve("/new/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Name != "new" {
		t.Errorf("expected name=new, got %s", match.Name)
	}
}

func TestResolverMultiplePatternsOrdering(t *testing.T) {
	patterns := []Pattern{
		Path("/articles/<int:id>/", http.HandlerFunc(dummyView), "article-detail"),
		Path("/articles/special/", http.HandlerFunc(dummyView), "article-special"),
	}
	resolver := NewResolver(patterns)

	// First matching pattern wins
	match, err := resolver.Resolve("/articles/42/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Name != "article-detail" {
		t.Errorf("expected first matching pattern, got %s", match.Name)
	}
}

func TestResolverPatternPriority(t *testing.T) {
	patterns := []Pattern{
		Path("/articles/special/", http.HandlerFunc(dummyView), "special"),
		Path("/articles/<slug:slug>/", http.HandlerFunc(dummyView), "article-slug"),
	}
	resolver := NewResolver(patterns)

	// Static path should match first
	match, err := resolver.Resolve("/articles/special/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Name != "special" {
		t.Errorf("expected 'special' to match first, got %s", match.Name)
	}

	// Dynamic path
	match, err = resolver.Resolve("/articles/my-article/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Name != "article-slug" {
		t.Errorf("expected 'article-slug' for dynamic path, got %s", match.Name)
	}
}
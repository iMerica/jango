package urls

import (
	"net/http"
	"testing"
)

func TestPathSimple(t *testing.T) {
	pattern := Path("/hello/", nil, "hello")
	if pattern.Name != "hello" {
		t.Errorf("expected name=hello, got %s", pattern.Name)
	}
	if pattern.Handler != nil {
		t.Error("expected nil handler")
	}
	if pattern.Regex == nil {
		t.Error("expected regex to be compiled")
	}
}

func TestPathWithParams(t *testing.T) {
	pattern := Path("/articles/<int:year>/", nil, "article-year")
	if pattern.Name != "article-year" {
		t.Errorf("expected name=article-year, got %s", pattern.Name)
	}
	if len(pattern.paramSpecs) != 1 {
		t.Fatalf("expected 1 param spec, got %d", len(pattern.paramSpecs))
	}
	if pattern.paramSpecs[0].Name != "year" {
		t.Errorf("expected param name=year, got %s", pattern.paramSpecs[0].Name)
	}
	if pattern.paramSpecs[0].Converter != IntConverter {
		t.Errorf("expected int converter, got %s", pattern.paramSpecs[0].Converter)
	}
}

func TestPathWithMultipleParams(t *testing.T) {
	pattern := Path("/articles/<int:year>/<slug:slug>/", nil, "article-detail")
	if len(pattern.paramSpecs) != 2 {
		t.Fatalf("expected 2 param specs, got %d", len(pattern.paramSpecs))
	}
	if pattern.paramSpecs[0].Name != "year" {
		t.Errorf("expected param 0 name=year, got %s", pattern.paramSpecs[0].Name)
	}
	if pattern.paramSpecs[1].Name != "slug" {
		t.Errorf("expected param 1 name=slug, got %s", pattern.paramSpecs[1].Name)
	}
}

func TestPathDefaultConverter(t *testing.T) {
	pattern := Path("/users/<username>/", nil, "user-profile")
	if len(pattern.paramSpecs) != 1 {
		t.Fatalf("expected 1 param spec, got %d", len(pattern.paramSpecs))
	}
	if pattern.paramSpecs[0].Converter != StringConverter {
		t.Errorf("expected str converter for default, got %s", pattern.paramSpecs[0].Converter)
	}
}

func TestRePath(t *testing.T) {
	pattern := RePath(`^articles/(?P<year>\d{4})/$`, nil, "article-re")
	if pattern.Name != "article-re" {
		t.Errorf("expected name=article-re, got %s", pattern.Name)
	}
	if pattern.Regex == nil {
		t.Error("expected regex to be compiled")
	}
	if len(pattern.paramSpecs) != 1 {
		t.Fatalf("expected 1 param spec, got %d", len(pattern.paramSpecs))
	}
	if pattern.paramSpecs[0].Name != "year" {
		t.Errorf("expected param name=year, got %s", pattern.paramSpecs[0].Name)
	}
}

func TestInclude(t *testing.T) {
	subPatterns := []Pattern{
		Path("/list/", nil, "sub-list"),
		Path("/<int:id>/", nil, "sub-detail"),
	}
	pattern := Include("/api", subPatterns, "api")
	if !pattern.IsInclude {
		t.Error("expected pattern to be include")
	}
	if pattern.Namespace != "api" {
		t.Errorf("expected namespace=api, got %s", pattern.Namespace)
	}
	if len(pattern.SubPatterns) != 2 {
		t.Fatalf("expected 2 sub-patterns, got %d", len(pattern.SubPatterns))
	}
}

func TestURLPatterns(t *testing.T) {
	patterns := URLPatterns(
		Path("/", nil, "index"),
		Path("/about/", nil, "about"),
	)
	if len(patterns) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(patterns))
	}
	if patterns[0].Name != "index" {
		t.Errorf("expected first pattern name=index, got %s", patterns[0].Name)
	}
}

func TestConverterValidation(t *testing.T) {
	tests := []struct {
		converter Converter
		value     string
		valid     bool
	}{
		{IntConverter, "123", true},
		{IntConverter, "0", true},
		{IntConverter, "01", false},
		{IntConverter, "abc", false},
		{IntConverter, "", false},
		{SlugConverter, "hello-world", true},
		{SlugConverter, "hello_world", true},
		{SlugConverter, "hello world", false},
		{StringConverter, "anything", true},
		{StringConverter, "", true},
		{PathConverter, "a/b/c", true},
	}
	for _, tt := range tests {
		t.Run(string(tt.converter)+"_"+tt.value, func(t *testing.T) {
			convDef := Converters[tt.converter]
			_, err := convDef.Converter(tt.value)
			if tt.valid && err != nil {
				t.Errorf("expected %q to be valid for %s, got error: %v", tt.value, tt.converter, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected %q to be invalid for %s", tt.value, tt.converter)
			}
		})
	}
}

func TestCompilePathPatternStatic(t *testing.T) {
	regex, specs := compilePathPattern("/hello/world/")
	if len(specs) != 0 {
		t.Errorf("expected no param specs for static path, got %d", len(specs))
	}
	if !regex.MatchString("/hello/world/") {
		t.Error("expected /hello/world/ to match")
	}
	if regex.MatchString("/hello/world/extra") {
		t.Error("expected /hello/world/extra NOT to match (should be exact)")
	}
}

func TestCompilePathPatternWithParams(t *testing.T) {
	regex, specs := compilePathPattern("/articles/<int:year>/")
	if len(specs) != 1 {
		t.Fatalf("expected 1 param spec, got %d", len(specs))
	}
	if specs[0].Name != "year" {
		t.Errorf("expected param name=year, got %s", specs[0].Name)
	}
	if !regex.MatchString("/articles/2024/") {
		t.Error("expected /articles/2024/ to match")
	}
	if regex.MatchString("/articles/abc/") {
		t.Error("expected /articles/abc/ NOT to match (int converter)")
	}
}

func TestExtractNamedGroups(t *testing.T) {
	specs := extractNamedGroups(`^users/(?P<id>\d+)/details/(?P<slug>\w+)$`)
	if len(specs) != 2 {
		t.Fatalf("expected 2 named groups, got %d", len(specs))
	}
	if specs[0].Name != "id" {
		t.Errorf("expected first group name=id, got %s", specs[0].Name)
	}
	if specs[1].Name != "slug" {
		t.Errorf("expected second group name=slug, got %s", specs[1].Name)
	}
}

// Helper handler for resolver tests
func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func TestHandlePattern(t *testing.T) {
	handler := http.HandlerFunc(dummyHandler)
	pattern := Path("/test/", handler, "test")
	
	// The Handle function exists but we just test that patterns can hold handlers
	if pattern.Handler == nil {
		t.Error("expected handler to be set")
	}
}
package rest_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iMerica/jango/rest"
)

func TestParsersAndNegotiation(t *testing.T) {
	req := rest.WrapRequest(httptest.NewRequest(http.MethodPost, "/posts.json", strings.NewReader(`{"title":"Hello"}`)))
	req.Header.Set("Content-Type", "application/json")
	data, err := rest.JSONParser{}.Parse(req)
	if err != nil {
		t.Fatalf("JSONParser returned error: %v", err)
	}
	if data["title"] != "Hello" {
		t.Fatalf("expected title, got %#v", data["title"])
	}

	formReq := rest.WrapRequest(httptest.NewRequest(http.MethodPost, "/posts/", strings.NewReader("title=Form")))
	formReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	form, err := rest.FormParser{}.Parse(formReq)
	if err != nil {
		t.Fatalf("FormParser returned error: %v", err)
	}
	if form["title"] != "Form" {
		t.Fatalf("expected form title, got %#v", form["title"])
	}

	renderer, format, err := rest.DefaultContentNegotiator{}.SelectRenderer(req, []rest.Renderer{rest.JSONRenderer{}, rest.BrowsableAPIRenderer{}})
	if err != nil {
		t.Fatalf("SelectRenderer returned error: %v", err)
	}
	if format != "json" || renderer.ContentType() != "application/json" {
		t.Fatalf("expected json renderer, got %s %s", format, renderer.ContentType())
	}
}

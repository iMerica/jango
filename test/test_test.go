package test_test

import (
	"net/http"
	"testing"

	jangotest "github.com/iMerica/jango/test"
)

func TestClientRequestsAndCookies(t *testing.T) {
	client := jangotest.NewHandlerClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "sessionid", Value: "abc"})
		w.Header().Set("X-Seen", r.Header.Get("X-Test"))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))

	client.SetHeader("X-Test", "yes")
	resp, err := client.Get("/")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if resp.Headers.Get("X-Seen") != "yes" {
		t.Fatalf("expected header propagation")
	}
	if len(client.Cookies) != 1 || client.Cookies[0].Name != "sessionid" {
		t.Fatalf("expected persisted cookie, got %#v", client.Cookies)
	}
}

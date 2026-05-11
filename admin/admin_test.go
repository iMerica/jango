package admin_test

import (
	"net/http"
	"testing"

	"github.com/iMerica/jango/admin"
	"github.com/iMerica/jango/auth"
	"github.com/iMerica/jango/examples/blogapi"
	jangohttp "github.com/iMerica/jango/http"
)

func TestSiteRegistryAndAccess(t *testing.T) {
	site := admin.NewSite("admin")
	if err := site.Register(blogapi.PostMeta, &admin.ModelAdmin{ListDisplay: []string{"title"}}); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if _, ok := site.Get(blogapi.PostMeta); !ok {
		t.Fatal("expected registered model admin")
	}

	req := jangohttp.WrapRequest(mustRequest(t))
	req.User = &auth.User{ID: 1, Username: "staff", IsActive: true, IsStaff: true}
	resp := site.Index(req)
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected staff access, got %d", resp.StatusCode())
	}

	req.User = auth.Anonymous
	resp = site.Index(req)
	if resp.StatusCode() != http.StatusForbidden {
		t.Fatalf("expected anonymous denial, got %d", resp.StatusCode())
	}
}

func mustRequest(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "/admin/", nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

package rest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iMerica/jango/cache"
	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/rest"
)

func TestRouterExtraActionsAndDispatch(t *testing.T) {
	viewset := rest.ModelViewSet[struct{}]{
		ExtraActions: []rest.ExtraAction[struct{}]{
			rest.CollectionAction[struct{}]("publish", func(req *rest.APIRequest) jangohttp.Response {
				return rest.NewAPIResponse(map[string]interface{}{"action": "publish"}, http.StatusOK)
			}, "post"),
			rest.DetailAction[struct{}]("archive", func(req *rest.APIRequest) jangohttp.Response {
				return rest.NewAPIResponse(map[string]interface{}{"id": req.URLParam("id")}, http.StatusOK)
			}),
		},
	}
	router := rest.NewDefaultRouter()
	router.Register("posts", "post", viewset)
	patterns := router.URLPatterns()
	if len(patterns) != 4 {
		t.Fatalf("expected default plus extra action patterns, got %d", len(patterns))
	}
	if patterns[2].Path != "/posts/publish/" || patterns[2].Name != "post-publish" {
		t.Fatalf("unexpected collection action route: %#v", patterns[2])
	}
	if patterns[3].Path != "/posts/<int:id>/archive/" || patterns[3].Name != "post-detail-archive" {
		t.Fatalf("unexpected detail action route: %#v", patterns[3])
	}

	publish := patterns[2].Handler.(jangohttp.ViewFunc)
	resp := publish(jangohttp.WrapRequest(httptest.NewRequest(http.MethodPost, "/posts/publish/", nil)))
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected publish 200, got %d", resp.StatusCode())
	}
	body := writeJSONResponse(t, resp)
	if body["action"] != "publish" {
		t.Fatalf("unexpected publish body: %#v", body)
	}

	resp = publish(jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/posts/publish/", nil)))
	if resp.StatusCode() != http.StatusMethodNotAllowed {
		t.Fatalf("expected unsupported method 405, got %d", resp.StatusCode())
	}
	rec := httptest.NewRecorder()
	if err := resp.WriteTo(rec, httptest.NewRequest(http.MethodGet, "/", nil)); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	if rec.Header().Get("Allow") != "POST, OPTIONS" {
		t.Fatalf("unexpected Allow header %q", rec.Header().Get("Allow"))
	}

	archive := patterns[3].Handler.(jangohttp.ViewFunc)
	req := jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/posts/7/archive/", nil))
	req.SetParam("id", "7")
	resp = archive(req)
	body = writeJSONResponse(t, resp)
	if body["id"] != "7" {
		t.Fatalf("detail extra action did not receive URL params: %#v", body)
	}
}

func TestAPIViewVersioningStrategies(t *testing.T) {
	api := rest.APIView{
		Versioning: rest.QueryParameterVersioning{
			ParamName:       "api_version",
			DefaultVersion:  "v1",
			AllowedVersions: []string{"v1", "v2"},
		},
	}
	view := api.AsView(map[string]rest.APIHandler{
		http.MethodGet: func(req *rest.APIRequest) jangohttp.Response {
			return rest.NewAPIResponse(map[string]interface{}{"version": req.Version}, http.StatusOK)
		},
	})
	resp := view(jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/things/?api_version=v2", nil)))
	body := writeJSONResponse(t, resp)
	if body["version"] != "v2" {
		t.Fatalf("expected query version v2, got %#v", body)
	}
	resp = view(jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/things/?api_version=v3", nil)))
	if resp.StatusCode() != http.StatusBadRequest {
		t.Fatalf("expected unsupported version 400, got %d", resp.StatusCode())
	}
	resp = view(jangohttp.WrapRequest(httptest.NewRequest(http.MethodOptions, "/things/", nil)))
	body = writeJSONResponse(t, resp)
	versioning := body["versioning"].(map[string]interface{})
	if versioning["param"] != "api_version" {
		t.Fatalf("expected OPTIONS version metadata, got %#v", versioning)
	}

	headerAPI := rest.APIView{Versioning: rest.HeaderVersioning{HeaderName: "X-Version", DefaultVersion: "2026"}}
	headerView := headerAPI.AsView(map[string]rest.APIHandler{
		http.MethodGet: func(req *rest.APIRequest) jangohttp.Response {
			return rest.NewAPIResponse(map[string]interface{}{"version": req.Version}, http.StatusOK)
		},
	})
	req := jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/things/", nil))
	req.Header.Set("X-Version", "2027")
	body = writeJSONResponse(t, headerView(req))
	if body["version"] != "2027" {
		t.Fatalf("expected header version, got %#v", body)
	}

	pathAPI := rest.APIView{Versioning: rest.URLPathVersioning{ParamName: "version"}}
	pathView := pathAPI.AsView(map[string]rest.APIHandler{
		http.MethodGet: func(req *rest.APIRequest) jangohttp.Response {
			return rest.NewAPIResponse(map[string]interface{}{"version": req.Version}, http.StatusOK)
		},
	})
	pathReq := jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/v1/things/", nil))
	pathReq.SetParam("version", "v1")
	body = writeJSONResponse(t, pathView(pathReq))
	if body["version"] != "v1" {
		t.Fatalf("expected path version, got %#v", body)
	}

	unsetView := rest.APIView{}.AsView(map[string]rest.APIHandler{
		http.MethodGet: func(req *rest.APIRequest) jangohttp.Response {
			return rest.NewAPIResponse(map[string]interface{}{"version": req.Version}, http.StatusOK)
		},
	})
	body = writeJSONResponse(t, unsetView(jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/things/", nil))))
	if body["version"] != "" {
		t.Fatalf("expected no default versioning, got %#v", body)
	}
}

func TestCacheBackedThrottles(t *testing.T) {
	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	c := cache.NewMemoryCache()
	api := rest.APIView{
		Throttles: []rest.Throttle{
			rest.AnonRateThrottle{RateThrottle: rest.RateThrottle{
				Rate:  "2/min",
				Cache: c,
				Now:   func() time.Time { return now },
			}},
		},
	}
	view := api.AsView(map[string]rest.APIHandler{
		http.MethodGet: func(req *rest.APIRequest) jangohttp.Response {
			return rest.NewAPIResponse(map[string]interface{}{"ok": true}, http.StatusOK)
		},
	})
	for i := 0; i < 2; i++ {
		req := jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/limited/", nil))
		req.Header.Set("X-Forwarded-For", "198.51.100.7")
		if resp := view(req); resp.StatusCode() != http.StatusOK {
			t.Fatalf("request %d expected 200, got %d", i+1, resp.StatusCode())
		}
	}
	req := jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/limited/", nil))
	req.Header.Set("X-Forwarded-For", "198.51.100.7")
	resp := view(req)
	if resp.StatusCode() != http.StatusTooManyRequests {
		t.Fatalf("expected throttle 429, got %d", resp.StatusCode())
	}
	rec := httptest.NewRecorder()
	if err := resp.WriteTo(rec, httptest.NewRequest(http.MethodGet, "/", nil)); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	if rec.Header().Get("Retry-After") != "60" {
		t.Fatalf("unexpected Retry-After %q", rec.Header().Get("Retry-After"))
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("throttle body was not JSON: %v", err)
	}
	if body["detail"] != "request was throttled" {
		t.Fatalf("unexpected throttle detail: %#v", body)
	}

	now = now.Add(61 * time.Second)
	req = jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/limited/", nil))
	req.Header.Set("X-Forwarded-For", "198.51.100.7")
	if resp := view(req); resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected throttle window to expire, got %d", resp.StatusCode())
	}
}

func TestRateParsingAndScopedThrottle(t *testing.T) {
	limit, window, err := rest.ParseRate("1000/hour")
	if err != nil {
		t.Fatalf("ParseRate returned error: %v", err)
	}
	if limit != 1000 || window != time.Hour {
		t.Fatalf("unexpected parsed rate: %d %s", limit, window)
	}
	if _, _, err := rest.ParseRate("bad"); err == nil {
		t.Fatal("expected invalid rate error")
	}

	now := time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC)
	api := rest.APIView{
		ThrottleScope: "uploads",
		Throttles: []rest.Throttle{
			rest.ScopedRateThrottle{RateThrottle: rest.RateThrottle{
				Cache: cache.NewMemoryCache(),
				Now:   func() time.Time { return now },
			}, Rates: map[string]string{"uploads": "1/s"}},
		},
	}
	view := api.AsView(map[string]rest.APIHandler{
		http.MethodPost: func(req *rest.APIRequest) jangohttp.Response {
			return rest.NewAPIResponse(map[string]interface{}{"ok": true}, http.StatusOK)
		},
	})
	req := jangohttp.WrapRequest(httptest.NewRequest(http.MethodPost, "/uploads/", nil))
	req.Header.Set("X-Forwarded-For", "198.51.100.8")
	if resp := view(req); resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected first scoped request 200, got %d", resp.StatusCode())
	}
	req = jangohttp.WrapRequest(httptest.NewRequest(http.MethodPost, "/uploads/", nil))
	req.Header.Set("X-Forwarded-For", "198.51.100.8")
	if resp := view(req); resp.StatusCode() != http.StatusTooManyRequests {
		t.Fatalf("expected scoped throttle 429, got %d", resp.StatusCode())
	}
}

package rest_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/iMerica/jango/examples/blogapi"
	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/orm"
	"github.com/iMerica/jango/rest"
)

func setupRESTTestDB(t *testing.T) *orm.DB {
	t.Helper()
	if os.Getenv("DATABASE_URL") == "" {
		t.Setenv("DATABASE_URL", "postgres://jango:password@localhost:5432/jango_test?sslmode=disable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := orm.OpenDB(ctx, orm.DefaultDBConfig())
	if err != nil {
		t.Skipf("Skipping REST integration test; unable to connect to DB: %v", err)
	}

	queries := []string{
		`DROP TABLE IF EXISTS blogapi_post CASCADE`,
		`DROP TABLE IF EXISTS blogapi_category CASCADE`,
		`DROP TABLE IF EXISTS accounts_user CASCADE`,
		`CREATE TABLE accounts_user (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(150) NOT NULL UNIQUE,
			email VARCHAR(254) NOT NULL UNIQUE,
			password VARCHAR(128) NOT NULL
		)`,
		`CREATE TABLE blogapi_category (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			slug VARCHAR(100) NOT NULL UNIQUE,
			description VARCHAR(500)
		)`,
		`CREATE TABLE blogapi_post (
			id BIGSERIAL PRIMARY KEY,
			title VARCHAR(200) NOT NULL,
			slug VARCHAR(200) NOT NULL UNIQUE,
			body TEXT NOT NULL,
			metadata JSONB,
			author_id BIGINT NOT NULL REFERENCES accounts_user(id),
			category_id BIGINT REFERENCES blogapi_category(id),
			published_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL,
			is_published BOOLEAN NOT NULL DEFAULT false,
			search_vector TEXT
		)`,
		`INSERT INTO accounts_user (id, username, email, password) VALUES
			(1, 'author', 'author@example.com', 'x')`,
		`INSERT INTO blogapi_category (id, name, slug, description) VALUES
			(1, 'Go', 'go', 'Go posts')`,
		`INSERT INTO blogapi_post
			(id, title, slug, body, metadata, author_id, category_id, published_at, created_at, updated_at, is_published, search_vector)
			VALUES
			(1, 'Alpha', 'alpha', 'A', '{"rank":1}', 1, 1, '2026-01-01T10:00:00Z', '2026-01-01T09:00:00Z', '2026-01-01T09:30:00Z', true, 'alpha'),
			(2, 'Beta', 'beta', 'B', '{"rank":2}', 1, 1, '2026-01-02T10:00:00Z', '2026-01-02T09:00:00Z', '2026-01-02T09:30:00Z', false, 'beta'),
			(3, 'Gamma', 'gamma', 'C', '{"rank":3}', 1, 1, '2026-01-03T10:00:00Z', '2026-01-03T09:00:00Z', '2026-01-03T09:30:00Z', true, 'gamma')`,
	}
	for _, query := range queries {
		if _, err := conn.Exec(context.Background(), query); err != nil {
			conn.Close()
			t.Fatalf("setup query failed: %v\nSQL: %s", err, query)
		}
	}
	orm.SetDefaultDB(conn)
	return conn
}

func postSerializer() rest.Serializer[blogapi.Post] {
	return rest.NewModelSerializer[blogapi.Post](
		blogapi.PostMeta,
		rest.Fields("id", "title", "slug", "category_id", "published_at", "created_at", "is_published"),
	)
}

func TestListAPIViewReturnsPaginatedEnvelope(t *testing.T) {
	conn := setupRESTTestDB(t)
	defer conn.Close()

	view := rest.ListAPIView[blogapi.Post]{
		QuerySet:        orm.Objects[blogapi.Post]("blogapi", "Post"),
		Serializer:      postSerializer(),
		FilterFields:    []string{"slug", "is_published", "category_id", "title"},
		OrderingFields:  []string{"published_at", "created_at", "title"},
		DefaultPageSize: 2,
	}
	req := jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/api/posts/?ordering=title&limit=2&offset=1", nil))

	resp := view.AsView()(req)
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode())
	}
	body := writeJSONResponse(t, resp)
	if body["count"] != float64(3) {
		t.Fatalf("expected count 3, got %#v", body["count"])
	}
	if body["limit"] != float64(2) || body["offset"] != float64(1) {
		t.Fatalf("unexpected pagination values: %#v", body)
	}
	results := body["results"].([]interface{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	first := results[0].(map[string]interface{})
	if first["slug"] != "beta" {
		t.Fatalf("expected offset result beta, got %#v", first["slug"])
	}
}

func TestListAPIViewFilteringAndUnsupportedFields(t *testing.T) {
	conn := setupRESTTestDB(t)
	defer conn.Close()

	view := rest.ListAPIView[blogapi.Post]{
		QuerySet:        orm.Objects[blogapi.Post]("blogapi", "Post"),
		Serializer:      postSerializer(),
		FilterFields:    []string{"slug", "is_published", "category_id", "title"},
		OrderingFields:  []string{"published_at", "created_at", "title"},
		DefaultPageSize: 10,
	}
	req := jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/api/posts/?is_published=true&title__icontains=a&ordering=-published_at", nil))
	resp := view.AsView()(req)
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode())
	}
	body := writeJSONResponse(t, resp)
	if body["count"] != float64(2) {
		t.Fatalf("expected two published posts, got %#v", body["count"])
	}
	results := body["results"].([]interface{})
	first := results[0].(map[string]interface{})
	if first["slug"] != "gamma" {
		t.Fatalf("expected descending published_at order, got %#v", first["slug"])
	}

	req = jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/api/posts/?body=A", nil))
	resp = view.AsView()(req)
	if resp.StatusCode() != http.StatusBadRequest {
		t.Fatalf("expected unsupported filter 400, got %d", resp.StatusCode())
	}

	req = jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/api/posts/?ordering=body", nil))
	resp = view.AsView()(req)
	if resp.StatusCode() != http.StatusBadRequest {
		t.Fatalf("expected unsupported ordering 400, got %d", resp.StatusCode())
	}
}

func TestDetailAPIView(t *testing.T) {
	conn := setupRESTTestDB(t)
	defer conn.Close()

	view := rest.DetailAPIView[blogapi.Post]{
		QuerySet:    orm.Objects[blogapi.Post]("blogapi", "Post"),
		Serializer:  postSerializer(),
		LookupField: "id",
	}
	req := jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/api/posts/2/", nil))
	req.SetParam("id", "2")

	resp := view.AsView()(req)
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode())
	}
	body := writeJSONResponse(t, resp)
	if body["slug"] != "beta" {
		t.Fatalf("expected beta detail, got %#v", body["slug"])
	}

	req = jangohttp.WrapRequest(httptest.NewRequest(http.MethodGet, "/api/posts/99/", nil))
	req.SetParam("id", "99")
	resp = view.AsView()(req)
	if resp.StatusCode() != http.StatusNotFound {
		t.Fatalf("expected missing detail 404, got %d", resp.StatusCode())
	}
}

func TestReadOnlyViewMethods(t *testing.T) {
	view := rest.ListAPIView[blogapi.Post]{
		QuerySet:   orm.Objects[blogapi.Post]("blogapi", "Post"),
		Serializer: postSerializer(),
	}
	req := jangohttp.WrapRequest(httptest.NewRequest(http.MethodOptions, "/api/posts/", nil))
	resp := view.AsView()(req)
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("expected OPTIONS 200, got %d", resp.StatusCode())
	}
	rec := httptest.NewRecorder()
	if err := resp.WriteTo(rec, req.Request); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	if rec.Header().Get("Allow") != "GET, HEAD, OPTIONS" {
		t.Fatalf("unexpected Allow header %q", rec.Header().Get("Allow"))
	}

	req = jangohttp.WrapRequest(httptest.NewRequest(http.MethodPost, "/api/posts/", nil))
	resp = view.AsView()(req)
	if resp.StatusCode() != http.StatusMethodNotAllowed {
		t.Fatalf("expected POST 405, got %d", resp.StatusCode())
	}
}

func writeJSONResponse(t *testing.T, resp jangohttp.Response) map[string]interface{} {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := resp.WriteTo(rec, req); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body was not JSON: %v\n%s", err, rec.Body.String())
	}
	return body
}

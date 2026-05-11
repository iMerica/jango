package rest_test

import (
	"testing"
	"time"

	"github.com/iMerica/jango/examples/blogapi"
	"github.com/iMerica/jango/rest"
)

func TestModelSerializerSelectedFields(t *testing.T) {
	publishedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	categoryID := int64(7)
	post := blogapi.Post{
		ID:          12,
		Title:       "REST foundations",
		Slug:        "rest-foundations",
		CategoryID:  &categoryID,
		PublishedAt: publishedAt,
		IsPublished: true,
	}
	serializer := rest.NewModelSerializer[blogapi.Post](
		blogapi.PostMeta,
		rest.Fields("id", "title", "slug", "category_id", "published_at", "is_published"),
	)

	data, err := serializer.Serialize(&post)
	if err != nil {
		t.Fatalf("Serialize returned error: %v", err)
	}
	if data["id"] != int64(12) {
		t.Fatalf("expected id=12, got %#v", data["id"])
	}
	if data["title"] != "REST foundations" {
		t.Fatalf("expected title, got %#v", data["title"])
	}
	if data["category_id"] != &categoryID {
		t.Fatalf("expected category ID pointer, got %#v", data["category_id"])
	}
	if data["published_at"] != publishedAt {
		t.Fatalf("expected published_at, got %#v", data["published_at"])
	}
	if data["is_published"] != true {
		t.Fatalf("expected is_published=true, got %#v", data["is_published"])
	}
}

func TestModelSerializerDefaultExcludesManyToMany(t *testing.T) {
	serializer := rest.NewModelSerializer[blogapi.Post](blogapi.PostMeta)
	data, err := serializer.Serialize(blogapi.Post{ID: 1, Title: "Default"})
	if err != nil {
		t.Fatalf("Serialize returned error: %v", err)
	}
	if _, ok := data["tags"]; ok {
		t.Fatal("expected many-to-many field tags to be excluded")
	}
	if _, ok := data["id"]; !ok {
		t.Fatal("expected concrete field id to be serialized")
	}
}

func TestModelSerializerUnknownOptionsFailClearly(t *testing.T) {
	serializer := rest.NewModelSerializer[blogapi.Post](blogapi.PostMeta, rest.Fields("missing"))
	if _, err := serializer.Serialize(blogapi.Post{}); err == nil {
		t.Fatal("expected unknown field error")
	}

	serializer = rest.NewModelSerializer[blogapi.Post](blogapi.PostMeta, rest.Exclude("missing"))
	if _, err := serializer.Serialize(blogapi.Post{}); err == nil {
		t.Fatal("expected unknown exclude field error")
	}

	serializer = rest.NewModelSerializer[blogapi.Post](blogapi.PostMeta, rest.Fields("tags"))
	if _, err := serializer.Serialize(blogapi.Post{}); err == nil {
		t.Fatal("expected many-to-many field error")
	}
}

func TestModelSerializerPointerAndValueInstances(t *testing.T) {
	serializer := rest.NewModelSerializer[blogapi.Post](
		blogapi.PostMeta,
		rest.Fields("id", "title"),
	)

	valueData, err := serializer.Serialize(blogapi.Post{ID: 1, Title: "Value"})
	if err != nil {
		t.Fatalf("Serialize value returned error: %v", err)
	}
	if valueData["title"] != "Value" {
		t.Fatalf("expected value title, got %#v", valueData["title"])
	}

	pointerData, err := serializer.Serialize(&blogapi.Post{ID: 2, Title: "Pointer"})
	if err != nil {
		t.Fatalf("Serialize pointer returned error: %v", err)
	}
	if pointerData["title"] != "Pointer" {
		t.Fatalf("expected pointer title, got %#v", pointerData["title"])
	}
}

func TestModelSerializerBindValidationAndPartial(t *testing.T) {
	serializer := rest.NewModelSerializer[blogapi.Post](
		blogapi.PostMeta,
		rest.Fields("id", "title", "slug", "body", "author_id", "is_published"),
		rest.Field("id", rest.ReadOnly()),
		rest.ValidateField("title", func(value interface{}) error {
			if value == "bad" {
				return assertError("bad title")
			}
			return nil
		}),
	)

	if err := serializer.Bind(map[string]interface{}{
		"title":        "Hello",
		"slug":         "hello",
		"body":         "Body",
		"author_id":    "42",
		"is_published": "true",
	}); err != nil {
		t.Fatalf("Bind returned error: %v", err)
	}
	data := serializer.ValidatedData()
	if data["AuthorID"] != int64(42) {
		t.Fatalf("expected coerced author ID, got %#v", data["AuthorID"])
	}
	if data["IsPublished"] != true {
		t.Fatalf("expected coerced boolean, got %#v", data["IsPublished"])
	}

	if err := serializer.Bind(map[string]interface{}{"title": "bad"}); err == nil {
		t.Fatal("expected validation error")
	}
	if len(serializer.Errors()["title"]) == 0 {
		t.Fatalf("expected title error, got %#v", serializer.Errors())
	}

	if err := serializer.BindPartial(map[string]interface{}{"title": "Only title"}); err != nil {
		t.Fatalf("partial bind should not require missing fields: %v", err)
	}
}

type assertError string

func (e assertError) Error() string { return string(e) }

package forms_test

import (
	"testing"

	"github.com/iMerica/jango/examples/blogapi"
	"github.com/iMerica/jango/forms"
)

func TestFormBindAndModelForm(t *testing.T) {
	form := forms.NewForm(map[string]forms.Field{
		"title": forms.CharField{BaseField: forms.BaseField{FieldName: "title", Required: true}, MaxLength: 10},
		"age":   forms.IntegerField{BaseField: forms.BaseField{FieldName: "age", Required: true}},
	})
	if !form.Bind(map[string]interface{}{"title": "Hello", "age": "7"}) {
		t.Fatalf("expected valid form, got errors %#v", form.Errors)
	}
	if form.CleanedData["age"] != int64(7) {
		t.Fatalf("expected coerced age, got %#v", form.CleanedData["age"])
	}

	modelForm := forms.NewModelForm(blogapi.PostMeta, "title", "author_id", "is_published")
	if !modelForm.Bind(map[string]interface{}{"title": "Post", "author_id": "1", "is_published": "true"}) {
		t.Fatalf("expected valid model form, got errors %#v", modelForm.Errors)
	}
	if modelForm.CleanedData["author_id"] != int64(1) {
		t.Fatalf("expected coerced author id, got %#v", modelForm.CleanedData["author_id"])
	}
}

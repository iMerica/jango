package rest_test

import (
	"testing"

	"github.com/iMerica/jango/rest"
)

func TestPlainSerializerValidationDefaultsAndPartial(t *testing.T) {
	serializer := rest.NewPlainSerializer("LoginRequest", []rest.PlainField{
		{Name: "username", Type: "string", Required: true, MaxLength: 20},
		{Name: "password", Type: "string", Required: true, WriteOnly: true},
		{Name: "remember", Type: "boolean", Default: false},
		{Name: "token", Type: "string", ReadOnly: true},
		{Name: "age", Type: "integer", Validators: []func(interface{}) error{
			func(value interface{}) error {
				if value.(int64) < 18 {
					return assertError("must be at least 18")
				}
				return nil
			},
		}},
	}, rest.PlainObjectValidator(func(data map[string]interface{}) error {
		if data["username"] == "root" {
			return assertError("reserved username")
		}
		return nil
	}))

	err := serializer.Bind(map[string]interface{}{
		"username": "alice",
		"password": "secret",
		"age":      "21",
	})
	if err != nil {
		t.Fatalf("Bind returned error: %v", err)
	}
	data := serializer.ValidatedData()
	if data["remember"] != false {
		t.Fatalf("expected default remember=false, got %#v", data["remember"])
	}
	if data["age"] != int64(21) {
		t.Fatalf("expected coerced age, got %#v", data["age"])
	}
	if _, ok := data["token"]; ok {
		t.Fatalf("read-only token should not be validated: %#v", data)
	}

	if err := serializer.Bind(map[string]interface{}{"username": "bob"}); err == nil {
		t.Fatal("expected missing password error")
	}
	if len(serializer.Errors()["password"]) == 0 {
		t.Fatalf("expected password error, got %#v", serializer.Errors())
	}

	if err := serializer.BindPartial(map[string]interface{}{"username": "carol"}); err != nil {
		t.Fatalf("partial bind should not require password: %v", err)
	}

	if err := serializer.Bind(map[string]interface{}{
		"username": "root",
		"password": "secret",
	}); err == nil {
		t.Fatal("expected object validator error")
	}
	if len(serializer.Errors()["non_field_errors"]) == 0 {
		t.Fatalf("expected non-field error, got %#v", serializer.Errors())
	}
}

func TestPlainSerializerSerializeAndSchema(t *testing.T) {
	serializer := rest.NewPlainSerializer("TokenResponse", []rest.PlainField{
		{Name: "access", Type: "string", Required: true},
		{Name: "refresh", Type: "string", WriteOnly: true},
	})

	data, err := serializer.Serialize(map[string]interface{}{
		"access":  "a",
		"refresh": "r",
	})
	if err != nil {
		t.Fatalf("Serialize returned error: %v", err)
	}
	if data["access"] != "a" {
		t.Fatalf("expected access value, got %#v", data)
	}
	if _, ok := data["refresh"]; ok {
		t.Fatalf("write-only refresh should not serialize: %#v", data)
	}

	fields := serializer.Fields()
	if len(fields) != 1 || fields[0] != "access" {
		t.Fatalf("unexpected public fields: %#v", fields)
	}
	schemaFields := serializer.SchemaFields()
	if len(schemaFields) != 2 || !schemaFields[1].WriteOnly {
		t.Fatalf("expected schema to include write-only metadata: %#v", schemaFields)
	}
}

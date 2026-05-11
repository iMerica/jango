package rest_test

import (
	"net/http"
	"testing"

	"github.com/iMerica/jango/examples/blogapi"
	jangohttp "github.com/iMerica/jango/http"
	"github.com/iMerica/jango/orm"
	"github.com/iMerica/jango/rest"
	"github.com/iMerica/jango/urls"
)

func TestDefaultRouterAndSchema(t *testing.T) {
	router := rest.NewDefaultRouter()
	router.Register("posts", "post", blogapi.PostViewSet())
	patterns := router.URLPatterns()
	if len(patterns) != 2 {
		t.Fatalf("expected collection and detail patterns, got %d", len(patterns))
	}
	if patterns[0].Name != "post-list" || patterns[1].Name != "post-detail" {
		t.Fatalf("unexpected pattern names: %q %q", patterns[0].Name, patterns[1].Name)
	}

	schema := rest.SchemaGenerator{Title: "Blog API", Patterns: patterns, Models: []*orm.ModelMeta{blogapi.PostMeta}}.OpenAPI()
	if schema["openapi"] != "3.0.3" {
		t.Fatalf("unexpected openapi version: %#v", schema["openapi"])
	}
	paths := schema["paths"].(map[string]interface{})
	if _, ok := paths["/posts/"]; !ok {
		t.Fatalf("expected posts path in schema: %#v", paths)
	}
	postPath := paths["/posts/"].(map[string]interface{})
	if _, ok := postPath["get"]; !ok {
		t.Fatalf("expected GET operation: %#v", postPath)
	}
	if _, ok := postPath["put"]; ok {
		t.Fatalf("collection route should not advertise PUT: %#v", postPath)
	}
	getOperation := postPath["get"].(map[string]interface{})
	if getOperation["operationId"] != "post_list" {
		t.Fatalf("unexpected operation ID: %#v", getOperation["operationId"])
	}
	responses := getOperation["responses"].(map[string]interface{})
	content := responses["200"].(map[string]interface{})["content"]
	if content == nil {
		t.Fatalf("expected response schema content: %#v", responses)
	}
}

func TestSimpleMetadata(t *testing.T) {
	data := rest.SimpleMetadata{}.DetermineMetadata(nil, nil, blogapi.PostMeta)
	fields := data["fields"].(map[string]interface{})
	if _, ok := fields["title"]; !ok {
		t.Fatalf("expected title metadata, got %#v", fields)
	}
}

func TestSchemaIncludesVersionThrottleAndPlainSerializer(t *testing.T) {
	plain := rest.NewPlainSerializer("EchoPayload", []rest.PlainField{
		{Name: "message", Type: "string", Required: true},
		{Name: "secret", Type: "string", WriteOnly: true},
	})
	api := rest.APIView{
		Throttles: []rest.Throttle{rest.AnonRateThrottle{RateThrottle: rest.RateThrottle{Rate: "1/s"}}},
		Versioning: rest.QueryParameterVersioning{
			ParamName: "version",
		},
	}
	view := api.AsView(map[string]rest.APIHandler{
		http.MethodPost: func(req *rest.APIRequest) jangohttp.Response {
			return rest.NewAPIResponse(nil, http.StatusOK)
		},
	})
	schema := rest.SchemaGenerator{
		Patterns: []urls.Pattern{
			urls.Path("/echo/", view, "echo"),
		},
		Serializers: []rest.SerializerSchemaProvider{plain},
	}.OpenAPI()
	paths := schema["paths"].(map[string]interface{})
	echo := paths["/echo/"].(map[string]interface{})
	if _, ok := echo["post"]; !ok {
		t.Fatalf("expected POST operation only, got %#v", echo)
	}
	if _, ok := echo["get"]; ok {
		t.Fatalf("unexpected GET operation: %#v", echo)
	}
	post := echo["post"].(map[string]interface{})
	responses := post["responses"].(map[string]interface{})
	if _, ok := responses["429"]; !ok {
		t.Fatalf("expected throttled operation to include 429: %#v", responses)
	}
	params := post["parameters"].([]map[string]interface{})
	foundVersion := false
	for _, param := range params {
		if param["name"] == "version" && param["in"] == "query" {
			foundVersion = true
		}
	}
	if !foundVersion {
		t.Fatalf("expected version query parameter: %#v", params)
	}
	components := schema["components"].(map[string]interface{})
	schemas := components["schemas"].(map[string]interface{})
	echoSchema := schemas["EchoPayload"].(map[string]interface{})
	props := echoSchema["properties"].(map[string]interface{})
	if props["secret"].(map[string]interface{})["writeOnly"] != true {
		t.Fatalf("expected writeOnly plain serializer field: %#v", props["secret"])
	}
}

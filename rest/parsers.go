package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

type Parser interface {
	Parse(req *APIRequest) (map[string]interface{}, error)
	ContentTypes() []string
}

type JSONParser struct{}

func (p JSONParser) ContentTypes() []string {
	return []string{"application/json"}
}

func (p JSONParser) Parse(req *APIRequest) (map[string]interface{}, error) {
	body, err := req.Body()
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return map[string]interface{}{}, nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

type FormParser struct{}

func (p FormParser) ContentTypes() []string {
	return []string{"application/x-www-form-urlencoded", "multipart/form-data"}
}

func (p FormParser) Parse(req *APIRequest) (map[string]interface{}, error) {
	form, err := req.Form()
	if err != nil {
		return nil, err
	}
	return valuesToMap(form), nil
}

func valuesToMap(values url.Values) map[string]interface{} {
	data := make(map[string]interface{}, len(values))
	for key, vals := range values {
		if len(vals) == 1 {
			data[key] = vals[0]
		} else {
			items := make([]interface{}, len(vals))
			for i, val := range vals {
				items[i] = val
			}
			data[key] = items
		}
	}
	return data
}

func parserForContentType(parsers []Parser, contentType string) Parser {
	if contentType == "" {
		return JSONParser{}
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.Split(contentType, ";")[0]
	}
	mediaType = strings.TrimSpace(strings.ToLower(mediaType))
	for _, parser := range parsers {
		for _, candidate := range parser.ContentTypes() {
			if mediaType == strings.ToLower(candidate) {
				return parser
			}
		}
	}
	return nil
}

func parseRequestData(req *APIRequest, parsers []Parser) error {
	if req.Method == http.MethodGet || req.Method == http.MethodHead || req.Method == http.MethodOptions {
		req.Data = map[string]interface{}{}
		return nil
	}
	parser := parserForContentType(parsers, req.Header.Get("Content-Type"))
	if parser == nil {
		return fmt.Errorf("unsupported media type")
	}
	data, err := parser.Parse(req)
	if err != nil {
		if err == io.EOF {
			req.Data = map[string]interface{}{}
			return nil
		}
		return err
	}
	req.Data = data
	return nil
}

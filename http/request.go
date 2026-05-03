package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Request struct {
	*http.Request

	Params map[string]string

	User             interface{}
	Session          interface{}
	AcceptedRenderer interface{}

	queryParsed bool
	queryParams url.Values

	formParsed bool
	formData   url.Values

	bodyParsed bool
	bodyBytes  []byte
	bodyError  error
}

func WrapRequest(r *http.Request) *Request {
	return &Request{
		Request: r,
		Params:   make(map[string]string),
	}
}

func (r *Request) Context() context.Context {
	return r.Request.Context()
}

func (r *Request) WithContext(ctx context.Context) *Request {
	newReq := r.Request.WithContext(ctx)
	return &Request{
		Request:          newReq,
		Params:           r.Params,
		User:             r.User,
		Session:          r.Session,
		AcceptedRenderer: r.AcceptedRenderer,
		queryParsed:      r.queryParsed,
		queryParams:      r.queryParams,
		formParsed:       r.formParsed,
		formData:         r.formData,
		bodyParsed:       r.bodyParsed,
		bodyBytes:        r.bodyBytes,
		bodyError:        r.bodyError,
	}
}

func (r *Request) Query() url.Values {
	if !r.queryParsed {
		r.queryParams = r.Request.URL.Query()
		r.queryParsed = true
	}
	return r.queryParams
}

func (r *Request) Form() (url.Values, error) {
	if !r.formParsed {
		contentType := r.Request.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") || strings.HasPrefix(contentType, "multipart/form-data") {
			if err := r.Request.ParseForm(); err != nil {
				return nil, err
			}
			if err := r.Request.ParseMultipartForm(32 << 20); err != nil && err != http.ErrNotMultipart {
				return nil, err
			}
			r.formData = r.Request.Form
		} else {
			r.formData = make(url.Values)
		}
		r.formParsed = true
	}
	return r.formData, nil
}

func (r *Request) Body() ([]byte, error) {
	if !r.bodyParsed {
		r.bodyBytes, r.bodyError = io.ReadAll(r.Request.Body)
		r.bodyParsed = true
		r.Request.Body = io.NopCloser(strings.NewReader(string(r.bodyBytes)))
	}
	return r.bodyBytes, r.bodyError
}

func (r *Request) ParseBody(v interface{}) error {
	data, err := r.Body()
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	ct := r.Request.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		return json.Unmarshal(data, v)
	}
	if strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		form, err := r.Form()
		if err != nil {
			return err
		}
		r.assignFormToStruct(form, v)
		return nil
	}
	return json.Unmarshal(data, v)
}

func (r *Request) Param(key string) string {
	if v, ok := r.Params[key]; ok {
		return v
	}
	return ""
}

func (r *Request) SetParam(key, value string) *Request {
	r.Params[key] = value
	return r
}

func (r *Request) SetUser(user interface{}) *Request {
	r.User = user
	return r
}

func (r *Request) SetSession(session interface{}) *Request {
	r.Session = session
	return r
}

func (r *Request) Cookies() []*http.Cookie {
	return r.Request.Cookies()
}

func (r *Request) assignFormToStruct(form url.Values, v interface{}) {
	//best-effort assignment; JSON parsing is the primary path
}
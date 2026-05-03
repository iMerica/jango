package test

import (
	"net/http"
	"net/http/httptest"
)

type Client struct {
	BaseURL    string
	Cookies    []*http.Cookie
	Headers    map[string]string
	FollowRedirects bool
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:         baseURL,
		Headers:         make(map[string]string),
		FollowRedirects: true,
	}
}

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Cookies    []*http.Cookie
}

func (c *Client) Get(path string) (*Response, error) {
	return c.doRequest("GET", path, nil)
}

func (c *Client) Post(path string, body []byte) (*Response, error) {
	return c.doRequest("POST", path, body)
}

func (c *Client) doRequest(method, path string, body []byte) (*Response, error) {
	return &Response{StatusCode: 200}, nil
}

func NewTestServer(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}
package test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type Client struct {
	BaseURL         string
	Cookies         []*http.Cookie
	Headers         map[string]string
	FollowRedirects bool
	client          *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:         baseURL,
		Headers:         make(map[string]string),
		FollowRedirects: true,
		client:          &http.Client{},
	}
}

func NewHandlerClient(handler http.Handler) *Client {
	client := NewClient("http://testserver")
	client.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Result(), nil
	})}
	return client
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
	target, err := c.resolveURL(path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	for _, cookie := range c.Cookies {
		req.AddCookie(cookie)
	}
	client := c.client
	if client == nil {
		client = &http.Client{}
	}
	if !c.FollowRedirects {
		client = &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.Cookies = mergeCookies(c.Cookies, resp.Cookies())
	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       respBody,
		Cookies:    resp.Cookies(),
	}, nil
}

func NewTestServer(handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func (c *Client) SetHeader(key, value string) {
	c.Headers[key] = value
}

func (c *Client) LoginCookie(cookie *http.Cookie) {
	c.Cookies = mergeCookies(c.Cookies, []*http.Cookie{cookie})
}

func (c *Client) resolveURL(path string) (string, error) {
	if _, err := url.ParseRequestURI(path); err == nil && (len(path) >= 4 && path[:4] == "http") {
		return path, nil
	}
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}
	rel, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(rel).String(), nil
}

func mergeCookies(existing, updates []*http.Cookie) []*http.Cookie {
	byName := make(map[string]*http.Cookie)
	for _, cookie := range existing {
		byName[cookie.Name] = cookie
	}
	for _, cookie := range updates {
		byName[cookie.Name] = cookie
	}
	result := make([]*http.Cookie, 0, len(byName))
	for _, cookie := range byName {
		result = append(result, cookie)
	}
	return result
}

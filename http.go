package yiigo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// defaultHTTPTimeout default http request timeout
const defaultHTTPTimeout = 10 * time.Second

// httpOptions http request options
type httpOptions struct {
	headers map[string]string
	cookies []*http.Cookie
	close   bool
	timeout time.Duration
}

// HTTPOption configures how we set up the http request
type HTTPOption interface {
	apply(*httpOptions)
}

// funcHTTPOption implements request option
type funcHTTPOption struct {
	f func(*httpOptions)
}

func (fo *funcHTTPOption) apply(o *httpOptions) {
	fo.f(o)
}

func newFuncHTTPOption(f func(*httpOptions)) *funcHTTPOption {
	return &funcHTTPOption{f: f}
}

// WithHTTPHeader specifies the header to http request.
func WithHTTPHeader(key, value string) HTTPOption {
	return newFuncHTTPOption(func(o *httpOptions) {
		o.headers[key] = value
	})
}

// WithHTTPCookies specifies the cookies to http request.
func WithHTTPCookies(cookies ...*http.Cookie) HTTPOption {
	return newFuncHTTPOption(func(o *httpOptions) {
		o.cookies = cookies
	})
}

// WithHTTPClose specifies close the connection after
// replying to this request (for servers) or after sending this
// request and reading its response (for clients).
func WithHTTPClose() HTTPOption {
	return newFuncHTTPOption(func(o *httpOptions) {
		o.close = true
	})
}

// WithHTTPTimeout specifies the timeout to http request.
func WithHTTPTimeout(d time.Duration) HTTPOption {
	return newFuncHTTPOption(func(o *httpOptions) {
		o.timeout = d
	})
}

// HTTPClient http client
type HTTPClient struct {
	client  *http.Client
	timeout time.Duration
}

// Get http get request
func (h *HTTPClient) Get(reqURL string, options ...HTTPOption) ([]byte, error) {
	o := &httpOptions{
		headers: make(map[string]string),
		timeout: h.timeout,
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	req, err := http.NewRequest("GET", reqURL, nil)

	if err != nil {
		return nil, err
	}

	// headers
	if len(o.headers) > 0 {
		for k, v := range o.headers {
			req.Header.Set(k, v)
		}
	}

	// cookies
	if len(o.cookies) > 0 {
		for _, v := range o.cookies {
			req.AddCookie(v)
		}
	}

	if o.close {
		req.Close = true
	}

	// timeout
	ctx, cancel := context.WithTimeout(req.Context(), o.timeout)

	defer cancel()

	resp, err := h.client.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body)

		return nil, fmt.Errorf("error http code: %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// Post http post request
func (h *HTTPClient) Post(reqURL string, body []byte, options ...HTTPOption) ([]byte, error) {
	o := &httpOptions{
		headers: make(map[string]string),
		timeout: h.timeout,
	}

	if len(options) > 0 {
		for _, option := range options {
			option.apply(o)
		}
	}

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	// headers
	if len(o.headers) > 0 {
		for k, v := range o.headers {
			req.Header.Set(k, v)
		}
	}

	// cookies
	if len(o.cookies) > 0 {
		for _, v := range o.cookies {
			req.AddCookie(v)
		}
	}

	if o.close {
		req.Close = true
	}

	// timeout
	ctx, cancel := context.WithTimeout(req.Context(), o.timeout)

	defer cancel()

	resp, err := h.client.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body)

		return nil, fmt.Errorf("error http code: %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// defaultHTTPClient default http client
var defaultHTTPClient = &HTTPClient{
	client: &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			MaxIdleConns:          0,
			MaxIdleConnsPerHost:   1000,
			MaxConnsPerHost:       1000,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	},
	timeout: defaultHTTPTimeout,
}

// NewHTTPClient returns a new http client
func NewHTTPClient(client *http.Client, defaultTimeout ...time.Duration) *HTTPClient {
	c := &HTTPClient{
		client:  client,
		timeout: defaultHTTPTimeout,
	}

	if len(defaultTimeout) != 0 {
		c.timeout = defaultTimeout[0]
	}

	return c
}

// HTTPGet http get request
func HTTPGet(reqURL string, options ...HTTPOption) ([]byte, error) {
	return defaultHTTPClient.Get(reqURL, options...)
}

// HTTPPost http post request
func HTTPPost(reqURL string, body []byte, options ...HTTPOption) ([]byte, error) {
	return defaultHTTPClient.Post(reqURL, body, options...)
}

package transferwise

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

const (
	LiveURL    = "https://api.transferwise.com"
	SandboxURL = "https://api.sandbox.transferwise.tech"
	userAgent  = "github.com/billglover/transferwise"
)

// Client is the TransferWise client
type Client struct {
	baseURL   *url.URL
	userAgent string
	client    *http.Client
}

// NewClient returns a new Starling API client. If a nil httpClient is
// provided, http.DefaultClient will be used. To use API methods which require
// authentication, provide an http.Client that will perform the authentication
// e.g. the golang.org/x/oauth2 library.
func NewClient(c *http.Client) *Client {
	if c == nil {
		c = http.DefaultClient
	}

	baseURL, _ := url.Parse(SandboxURL)

	return &Client{baseURL: baseURL, userAgent: userAgent, client: c}
}

// ClientOptions provides a set of options that can be used to configure the Client.
type ClientOptions struct {
	baseURL   *url.URL
	userAgent string
}

// NewClientWithOptions takes ClientOptions, configures and returns a new client.
func NewClientWithOptions(c *http.Client, opts ClientOptions) *Client {
	client := NewClient(c)

	client.baseURL = opts.baseURL
	client.userAgent = opts.userAgent

	return client
}

// NewRequest takes a method, path and optional body. It returns an *http.Request.
// If a non-nil body is provided it will be JSON encoded and included in the request.
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	return req, nil
}

// Do sends a request to the API and returns the response. An error is returned
// if the request cannot be sent or if the API returns an error. If a response is
// received, the body is decoded and stored in the value pointed to by v.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	req = req.WithContext(ctx)
	resp, err := c.client.Do(req)

	if err != nil {
		select {
		case <-ctx.Done():
			return nil, errors.Wrap(err, ctx.Err().Error())
		default:
			return nil, err
		}
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read body")
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "unable to close body")
	}

	if v != nil && len(data) != 0 {
		err = json.Unmarshal(data, v)

		switch err {
		case nil:
		case io.EOF:
			err = nil
		default:
			err = errors.Wrap(err, "unable to parse API response")
		}
	}

	return resp, err
}

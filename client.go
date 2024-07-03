package go_moonshot

import (
	"context"
	"encoding/json"
	"errors"
	gomoonshot "go-moonshot/internal"
	"io"
	"net/http"
)

var client = map[string]*Client{}

type Client struct {
	config         ClientConfig
	requestBuilder gomoonshot.RequestBuilder
}

func NewClient(apiKey string) (*Client, error) {
	config := DefaultConfig(apiKey)
	return NewClientWithConfig(config)
}

func NewClientWithConfig(config ClientConfig) (*Client, error) {
	if config.ApiKey == "" {
		return nil, errors.New("config is error")
	}
	if c, ok := client[config.ApiKey]; ok {
		return c, nil
	}
	c := &Client{
		config:         config,
		requestBuilder: gomoonshot.NewRequestBuilder(),
	}
	client[config.ApiKey] = c
	return client[config.ApiKey], nil
}

type requestOptions struct {
	body   any
	header http.Header
}

type requestOption func(*requestOptions)

func withBody(body any) requestOption {
	return func(o *requestOptions) {
		o.body = body
	}
}

func withHeader(header http.Header) requestOption {
	return func(o *requestOptions) {
		o.header = header
	}
}

func (c *Client) newRequest(ctx context.Context, method, url string, setters ...requestOption) (*http.Request, error) {
	args := &requestOptions{
		body:   nil,
		header: make(http.Header),
	}
	for _, setter := range setters {
		setter(args)
	}
	req, err := c.requestBuilder.Build(ctx, method, url, args.body, args.header)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *Client) sendRequest(req *http.Request, v any) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	res, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return c.handleErrorResp(res)
	}

	return decodeResponse(res.Body, v)
}

func decodeResponse(body io.Reader, v any) error {
	if v == nil {
		return nil
	}
	if result, ok := v.(*string); ok {
		return decodeStringResponse(body, result)
	}
	return json.NewDecoder(body).Decode(v)
}

func decodeStringResponse(body io.Reader, output *string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	*output = string(data)
	return nil
}

func (c *Client) handleErrorResp(res *http.Response) error {
	return errors.New("handleErrorResp error " + res.Status)
}

func (c *Client) fullUrl(suffix string) string {
	return c.config.BaseUrl + apiPrefix + suffix
}

func sendRequestStream[T streamable](client *Client, req *http.Request) (*streamReader[T], error) {
	req.Header.Set("Authorization", "Bearer "+client.config.ApiKey)
	stream := &streamReader[T]{
		isFinished:  false,
		response:    nil,
		scanner:     nil,
		unmarshaler: &gomoonshot.JSONUnmarshaler{},
	}
	resp, err := client.config.HTTPClient.Do(req)
	if err != nil {
		stream.isFinished = true
		return stream, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		stream.isFinished = true
		return stream, client.handleErrorResp(resp)
	}
	stream.response = resp
	return stream, nil
}

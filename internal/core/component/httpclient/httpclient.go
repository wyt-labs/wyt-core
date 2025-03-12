package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type config struct {
	Timeout time.Duration
	BaseURL string
}

type Option func(*config)

func WithBaseURL(url string) Option {
	return func(c *config) {
		c.BaseURL = url
	}
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewHttpClient(opts ...Option) (*Client, error) {
	conf := &config{
		Timeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(conf)
	}
	var hclient = &http.Client{}

	client := &Client{
		httpClient: hclient,
		baseURL:    conf.BaseURL,
	}
	return client, nil
}

func (c *Client) DoRequest(method, path string, body []byte, queries, headers map[string]string) ([]byte, error) {
	ul, err := c.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, ul, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

func (c *Client) DoRequestV2(method, parsedUrl string, body []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, parsedUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

func (c *Client) GetV2(parsedUrl string, headers map[string]string) ([]byte, error) {
	return c.DoRequestV2("GET", parsedUrl, nil, headers)
}

func (c *Client) Get(path string, headers, queries map[string]string) ([]byte, error) {
	return c.DoRequest("GET", path, nil, queries, headers)
}

func (c *Client) Post(path string, body []byte, headers, queries map[string]string) ([]byte, error) {
	return c.DoRequest("POST", path, body, queries, headers)
}

func (c *Client) ParseURL(path string, queries map[string]string) (string, error) {
	if c.baseURL == "" {
		return "", errors.New("base url is empty")
	}
	ul := c.baseURL + path
	parsedURL, err := url.Parse(ul)
	if queries != nil {
		if err != nil {
			return "", err
		}
		queryParams := url.Values{}
		for key, value := range queries {
			queryParams.Add(key, value)
		}
		parsedURL.RawQuery = queryParams.Encode()
	}
	return parsedURL.String(), nil
}

func (c *Client) PostV2(path string, payload *strings.Reader, headers, queries map[string]string) ([]byte, error) {
	ul, err := c.ParseURL(path, queries)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", ul, payload)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

func (c *Client) ParseJSONResponse(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (c *Client) GetBaseURL() string {
	return c.baseURL
}

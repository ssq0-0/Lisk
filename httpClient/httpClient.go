package httpClient

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lisk/logger"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"golang.org/x/net/http2"
)

type HttpClient struct {
	Client *http.Client
}

func NewHttpClient(proxyURL string) (*HttpClient, error) {
	transport := &http.Transport{}
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %v", err)
		}
		transport.Proxy = http.ProxyURL(proxy)
	}

	if err := http2.ConfigureTransport(transport); err != nil {
		return nil, fmt.Errorf("failed to configure HTTP/2 transport: %v", err)
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %v", err)
	}

	client := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   30 * time.Second,
	}

	return &HttpClient{Client: client}, nil
}

func (h *HttpClient) SendJSONRequest(urlRequest, method string, reqBody, respBody interface{}) error {
	req, err := h.createRequest(urlRequest, method, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	return h.executeWithRetries(req, respBody)
}

func (h *HttpClient) createRequest(urlRequest, method string, reqBody interface{}) (*http.Request, error) {
	var body io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, urlRequest, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	h.setHeaders(req)
	return req, nil
}

func (h *HttpClient) executeWithRetries(req *http.Request, respBody interface{}) error {
	const maxRetries = 3
	const retryDelay = 1500 * time.Millisecond

	for attempts := 0; attempts < maxRetries; attempts++ {
		resp, err := h.Client.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "unexpected EOF") {
				logger.GlobalLogger.Warn("Unexpected EOF encountered. Retrying... Attempt %d", attempts+1)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("request error: %v", err)
		}
		defer resp.Body.Close()

		if err := h.parseResponse(resp, respBody); err != nil {
			if resp.StatusCode == http.StatusTooManyRequests {
				logger.GlobalLogger.Warn("Rate limit reached. Retrying... Attempt %d", attempts+1)
				time.Sleep(retryDelay)
				continue
			}
			return err
		}

		return nil
	}

	return fmt.Errorf("request failed after %d retries", maxRetries)
}

func (h *HttpClient) parseResponse(resp *http.Response, respBody interface{}) error {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body) // Ignoring read error to avoid masking original status code
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	reader := io.ReadCloser(resp.Body)
	defer reader.Close()

	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if respBody != nil {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("failed to parse response JSON: %v", err)
		}
	}
	return nil
}

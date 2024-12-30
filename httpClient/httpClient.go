package httpClient

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lisk/globals"
	"lisk/logger"
	"math/rand"
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
	var req *http.Request
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}

		req, err = http.NewRequestWithContext(ctx, method, urlRequest, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, urlRequest, nil)
		if err != nil {
			return err
		}
	}

	h.setHeaders(req)

	for attempts := 0; attempts < 3; attempts++ {
		resp, err := h.Client.Do(req)
		if err != nil {
			return err
		}
		bodyErr := h.checkAndParseResp(resp, respBody)
		_ = resp.Body.Close()

		if bodyErr == nil {
			return nil
		}

		if resp.StatusCode == 429 {
			logger.GlobalLogger.Warn("Rate limit reached. Retrying... Attempt %d", attempts+1)
			time.Sleep(time.Millisecond * 1500)
			continue
		}

		return err
	}

	return fmt.Errorf("failed after multiple retries")
}

func (h *HttpClient) checkAndParseResp(resp *http.Response, respBody interface{}) error {
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected status code: %d, and failed to read body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var reader io.ReadCloser = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if respBody != nil {
		err := json.Unmarshal(bodyBytes, respBody)
		if err != nil {
			return fmt.Errorf("failed to parse JSON: %w\nBody: %s", err, string(bodyBytes))
		}
	}

	return nil
}

func (h *HttpClient) setHeaders(req *http.Request) {
	userAgent := h.getRandomUserAgent()
	secChUa, platform := h.getSecChUa(userAgent)

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9,"+fmt.Sprintf("q=%.1f", 0.5+rand.Float32()/2))
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", secChUa)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", platform)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "cross-site")
	req.Header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
}

func (h *HttpClient) getRandomUserAgent() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return globals.UserAgents[r.Intn(len(globals.UserAgents))]
}

func (h *HttpClient) getSecChUa(userAgent string) (string, string) {
	if strings.Contains(userAgent, "Macintosh") {
		return globals.SecChUa["Macintosh"], globals.Platforms["Macintosh"]
	} else if strings.Contains(userAgent, "Windows") {
		return globals.SecChUa["Windows"], globals.Platforms["Windows"]
	} else if strings.Contains(userAgent, "Linux") {
		return globals.SecChUa["Linux"], globals.Platforms["Linux"]
	}
	return globals.SecChUa["Unknown"], `"Unknown"`
}

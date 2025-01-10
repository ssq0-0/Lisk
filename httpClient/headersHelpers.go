package httpClient

import (
	"fmt"
	"lisk/globals"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

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

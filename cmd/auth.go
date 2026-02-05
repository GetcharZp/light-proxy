package cmd

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

var proxyAuthToken string

// authenticate 鉴权中间件
func authenticate(w http.ResponseWriter, r *http.Request) bool {
	if proxyAuthToken == "" {
		return true
	}

	auth := r.Header.Get("Proxy-Authorization")
	if auth == "" {
		w.Header().Set("Proxy-Authenticate", `Basic realm="Proxy"`)
		w.WriteHeader(http.StatusProxyAuthRequired)
		return false
	}

	// 解析 Basic Auth
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		w.WriteHeader(http.StatusProxyAuthRequired)
		return false
	}

	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		w.WriteHeader(http.StatusProxyAuthRequired)
		return false
	}

	actualToken := strings.TrimSuffix(string(payload), ":")

	if subtle.ConstantTimeCompare([]byte(actualToken), []byte(proxyAuthToken)) != 1 {
		w.WriteHeader(http.StatusProxyAuthRequired)
		return false
	}

	return true
}

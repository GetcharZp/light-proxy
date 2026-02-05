package cmd

import (
	"crypto/tls"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// 全局 Transport
var globalTransport = &http.Transport{
	TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func NewRelayCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relay",
		Short: "Use relay CLI to forward requests to the real network.",
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetString("port")
			cert, _ := cmd.Flags().GetString("cert")
			key, _ := cmd.Flags().GetString("key")
			proxyAuthToken, _ = cmd.Flags().GetString("auth")
			relay(port, cert, key)
		},
	}

	cmd.Flags().StringP("port", "p", "8080", "proxy port")
	cmd.Flags().String("cert", "", "path to cert file (enable HTTPS proxy)")
	cmd.Flags().String("key", "", "path to key file (enable HTTPS proxy)")
	cmd.Flags().String("auth", "", "single token for proxy authentication")

	return cmd
}

func relay(port, cert, key string) {
	server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authenticate(w, r) {
				log.Printf("[warn] auth failed from %s \n", r.RemoteAddr)
				return
			}

			if r.Method == http.MethodConnect {
				handleHttps(w, r)
			} else {
				handleHttp(w, r)
			}
		}),
	}

	showLocalIpv4s()

	if cert != "" && key != "" {
		log.Printf("[sys] TLS proxy server mode enabled on port:%s \n", port)
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		log.Fatal(server.ListenAndServeTLS(cert, key))
	} else {
		log.Printf("[sys] HTTP proxy server start on port:%s \n", port)
		log.Fatal(server.ListenAndServe())
	}
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	// 补全 URL
	if r.URL.Host == "" {
		r.URL.Host = r.Host
	}
	if r.URL.Scheme == "" {
		r.URL.Scheme = "http"
	}

	// 清除 RequestURI
	r.RequestURI = ""

	resp, err := globalTransport.RoundTrip(r)
	if err != nil {
		log.Printf("[sys] transport error: %s \n", err.Error())
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制 Header
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func handleHttps(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		log.Printf("[sys] dial error: %s \n", r.Host)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		destConn.Close()
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		destConn.Close()
		return
	}

	// 统一关闭
	done := make(chan bool, 2)
	go func() {
		defer destConn.Close()
		defer clientConn.Close()
		transfer(destConn, clientConn, done)
	}()
	go func() {
		defer destConn.Close()
		defer clientConn.Close()
		transfer(clientConn, destConn, done)
	}()
}

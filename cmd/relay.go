package cmd

import (
	"github.com/spf13/cobra"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func NewRelayCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relay",
		Short: "Use relay CLI to forward requests to the real network.",
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetString("port")
			relay(port)
		},
	}

	cmd.Flags().StringP("port", "p", "8080", "proxy port, default 8080")

	return cmd
}

func relay(port string) {
	server := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				handleHttps(w, r)
			} else {
				handleHttp(w, r)
			}
		}),
	}

	showLocalIpv4s()
	log.Printf("[sys] proxy server start success port:%s \n", port)
	log.Fatal(server.ListenAndServe())
}

func handleHttp(w http.ResponseWriter, r *http.Request) {
	transport := &http.Transport{}
	log.Printf("[sys] get proxy request address: %s \n", r.URL.String())
	outReq, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		log.Printf("[sys] create request error:%s \n", err.Error())
		http.Error(w, "[http] failed to create request", http.StatusInternalServerError)
		return
	}

	outReq.Header = r.Header
	resp, err := transport.RoundTrip(outReq)
	if err != nil {
		log.Printf("[sys] transport round trip error:%s \n", err.Error())
		http.Error(w, "[http] transport round trip error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func handleHttps(w http.ResponseWriter, r *http.Request) {
	log.Printf("[sys] get proxy request address: %s \n", r.Host)
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		log.Printf("[sys] dial timeout error:%s \n", err.Error())
		http.Error(w, "[net] dial timeout error:"+err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("[sys] hijacker not supported \n")
		http.Error(w, "[http] hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("[sys] hijack error:%s \n", err.Error())
		http.Error(w, "[http] hijack error:"+err.Error(), http.StatusServiceUnavailable)
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

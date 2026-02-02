package cmd

import (
	"github.com/up-zero/gotool/netutil"
	"io"
	"log"
	"strings"
)

func transfer(destination io.Writer, source io.Reader, done chan<- bool) {
	io.Copy(destination, source)
	done <- true
}

func showLocalIpv4s() {
	ips, err := netutil.Ipv4sLocal()
	if err == nil {
		log.Printf("[sys] local ipv4: %s \n", strings.Join(ips, ";"))
	}
}

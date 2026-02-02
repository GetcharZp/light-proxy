package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"net"
	"time"
)

var (
	bridgePort   string
	relayAddress string
)

func NewBridgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge",
		Short: "Use bridge CLI to forward requests to the relay PC.",
		Run: func(cmd *cobra.Command, args []string) {
			bridgePort, _ = cmd.Flags().GetString("port")
			relayAddress, _ = cmd.Flags().GetString("relay")
			bridge()
		},
	}

	cmd.Flags().StringP("port", "p", "8080", "proxy port, default 8080")
	cmd.Flags().StringP("relay", "r", "", "relay address (required)")
	cmd.MarkFlagRequired("relay")

	return cmd
}

func bridge() {
	ln, err := net.Listen("tcp", ":"+bridgePort)
	if err != nil {
		log.Printf("[sys] listen error:%s \n", err.Error())
		return
	}
	showLocalIpv4s()
	log.Printf("[sys] proxy server start success port:%s \n", bridgePort)

	for {
		clientConn, err := ln.Accept()
		if err != nil {
			log.Printf("[sys] accept error:%s \n", err.Error())
			continue
		}
		go bridgePipe(clientConn)
	}
}

func bridgePipe(clientConn net.Conn) {
	log.Printf("[sys] get proxy request address: %s\n", clientConn.RemoteAddr().String())
	serverConn, err := net.DialTimeout("tcp", relayAddress, 10*time.Second)
	if err != nil {
		log.Printf("[sys] net dial error:%s \n", err.Error())
		return
	}

	// 统一关闭
	done := make(chan bool, 2)
	go func() {
		defer serverConn.Close()
		defer clientConn.Close()
		transfer(serverConn, clientConn, done)
	}()
	go func() {
		defer serverConn.Close()
		defer clientConn.Close()
		transfer(clientConn, serverConn, done)
	}()
}

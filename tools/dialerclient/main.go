package main

import (
	"context"
	"flag"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
	"net/http"
)

var serverAddr string
var clientID string
var debug bool

func auth(string, string) bool {
	return true
}

func main() {
	flag.StringVar(&serverAddr, "connect", "ws://192.168.0.105:8080/api/dialer/connect", "Address to connect to")
	flag.StringVar(&clientID, "id", "docker", "Client ID")
	flag.BoolVar(&debug, "debug", true, "Debug logging")
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	headers := http.Header{
		"X-Tunnel-ID": []string{clientID},
	}

	remotedialer.ClientConnect(context.Background(), serverAddr, headers, nil, auth, nil)
}

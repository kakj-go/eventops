package main

import (
	"context"
	"flag"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
	"net/http"
	"tiggerops/pkg/dialer"
)

var serverAddr string
var clientID string
var token string
var debug bool

func auth(string, string) bool {
	return true
}

func main() {
	flag.StringVar(&serverAddr, "connect", "ws://192.168.0.105:8080/api/dialer/connect", "Address to connect to")
	flag.StringVar(&clientID, "id", "docker", "Client ID")
	flag.StringVar(&token, "token", "123456", "Client Token")
	flag.BoolVar(&debug, "debug", true, "Debug logging")
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	headers := http.Header{
		dialer.IdHeader:   []string{clientID},
		dialer.AuthHeader: []string{token},
	}

	remotedialer.ClientConnect(context.Background(), serverAddr, headers, nil, auth, nil)
}

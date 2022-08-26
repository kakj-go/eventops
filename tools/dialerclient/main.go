package main

import (
	"context"
	"eventops/internal/core/dialer"
	"flag"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var serverAddr string
var clientID string
var token string
var user string
var debug bool

func auth(string, string) bool {
	return true
}

func main() {
	flag.StringVar(&serverAddr, "connect", "ws://192.168.0.105:8080/api/dialer/connect", "Address to connect to")
	flag.StringVar(&clientID, "id", "runner_tunnel", "Client ID")
	flag.StringVar(&token, "token", "123456", "Client Token")
	flag.StringVar(&user, "user", "kakj", "Client User")
	flag.BoolVar(&debug, "debug", true, "Debug logging")
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	headers := http.Header{
		dialer.IdHeader:   []string{clientID},
		dialer.AuthHeader: []string{token},
		dialer.UserHeader: []string{user},
	}

	for {
		remotedialer.ClientConnect(context.Background(), serverAddr, headers, nil, auth, nil)
		logrus.Errorf("failed to client server: %v, reclient after 5 second", serverAddr)
		time.Sleep(5 * time.Second)
	}
}

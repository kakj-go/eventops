package main

import (
	"eventops/conf"
	log "github.com/sirupsen/logrus"
)

func main() {
	conf.LoadConf()
	initLog()

	server, err := newServer()
	if err != nil {
		panic(err)
	}
	server.run()
}

func initLog() {
	if conf.IsDebug() {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}

package main

import (
	log "github.com/sirupsen/logrus"
	"tiggerops/conf"
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

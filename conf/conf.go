package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

func LoadConf() {
	parseFlag()
	parseConf()
}

var conf Config

type Config struct {
	Port            string   `yaml:"port"`
	Debug           bool     `yaml:"debug"`
	CallbackAddress string   `yaml:"callbackAddress"`
	Mysql           *Mysql   `yaml:"mysql"`
	Uc              Uc       `yaml:"uc"`
	Event           Event    `yaml:"event"`
	Actuator        Actuator `yaml:"actuator"`
	Minio           Minio    `yaml:"minio"`
}

type Minio struct {
	Server          string `yaml:"server"`
	AccessKeyId     string `yaml:"accessKeyId"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	Ssl             bool   `yaml:"ssl"`
	BasePath        string `yaml:"basePath"`
}

type Mysql struct {
	Post     string `default:"3306" yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Address  string `yaml:"address"`
	Db       string `default:"eventops" yaml:"db"`
}

func GetPort() string {
	return conf.Port
}

func GetCallbackAddress() string {
	return conf.CallbackAddress
}

func GetLoginTokenSignature() string {
	return conf.Uc.LoginTokenSignature
}

func GetMysql() *Mysql {
	return conf.Mysql
}

func IsDebug() bool {
	return conf.Debug
}

func GetUc() Uc {
	return conf.Uc
}

func GetEvent() Event {
	return conf.Event
}

func GetActuator() Actuator {
	return conf.Actuator
}

func GetMinio() Minio {
	return conf.Minio
}

func GetLoginTokenExpiresTime() time.Duration {
	return time.Second * time.Duration(conf.Uc.LoginTokenExpiresTime)
}

func readConf(filename string) (Config, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}
	var conf Config
	err = yaml.Unmarshal(buf, &conf)
	if err != nil {
		return Config{}, fmt.Errorf("yaml Unmarshal file %q err %v", filename, err)
	}
	return conf, nil
}

func parseConf() {
	if configPath == nil {
		panic("config file path was empty")
	}

	var err error
	conf, err = readConf(*configPath)
	if err != nil {
		panic(fmt.Errorf("parse error %v", err))
	}
}

/*
 * Copyright 2022 The kakj-go Authors.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package conf

import (
	"github.com/jinzhu/configor"
	"time"
)

func LoadConf() {
	parseFlag()
	parseConf()
}

var conf Config

type Config struct {
	Port            string   `default:"8080" env:"EVENTOPS_PORT" yaml:"port"`
	Debug           bool     `default:"false" env:"EVENTOPS_DEBUG" yaml:"debug"`
	CallbackAddress string   `required:"true" env:"EVENTOPS_CALLBACK_ADDRESS" yaml:"callbackAddress"`
	Mysql           Mysql    `required:"true" yaml:"mysql"`
	Uc              Uc       `yaml:"uc"`
	Event           Event    `yaml:"event"`
	Actuator        Actuator `yaml:"actuator"`
	Minio           Minio    `yaml:"minio"`
}

type Uc struct {
	Auth                  Auth   `yaml:"auth"`
	LoginTokenExpiresTime int64  `default:"315360000" env:"EVENTOPS_LOGIN_TOKEN_EXPIRES_TIME" yaml:"loginTokenExpiresTime"`
	LoginTokenSignature   string `default:"signature" env:"EVENTOPS_LOGIN_TOKEN_SIGNATURE" yaml:"loginTokenSignature"`
}

type Auth struct {
	WhiteUrlList []string
}

type Actuator struct {
	PrintTunnelData bool `default:"false" env:"EVENTOPS_PRINT_TUNNEL_DATA" yaml:"printTunnelData"`
}

type Event struct {
	Process Process `yaml:"process"`
}

type Process struct {
	BufferSize         int64 `default:"500" env:"EVENTOPS_PROCESS_BUFFER_SIZE" yaml:"bufferSize"`
	WorkNum            int64 `default:"5" env:"EVENTOPS_PROCESS_WORK_NUM" yaml:"workNum"`
	ProcessingOverTime int64 `default:"120" env:"EVENTOPS_PROCESS_OVER_TIME" yaml:"processingOverTime"`

	TriggerCacheSize      int `default:"10000" env:"EVENTOPS_PROCESS_TRIGGER_CACHE_SIZE" yaml:"triggerCacheSize"`
	LoopLoadEventInterval int `default:"300" env:"EVENTOPS_PROCESS_LOOP_INTERVAL" yaml:"loopLoadEventInterval"`
}

type Minio struct {
	Server          string `env:"MINIO_SERVER" yaml:"server"`
	AccessKeyId     string `env:"MINIO_ACCESS_KEY" yaml:"accessKeyId" yaml:"accessKeyId"`
	SecretAccessKey string `env:"MINIO_SECRET_KEY" yaml:"secretAccessKey" yaml:"secretAccessKey"`
	Ssl             bool   `default:"false" env:"MINIO_SSL" yaml:"ssl"`
	BasePath        string `default:"eventops" env:"MINIO_BASE_PATH" yaml:"basePath"`
}

type Mysql struct {
	Post     string `default:"3306" env:"MYSQL_PORT" yaml:"post"`
	User     string `required:"true" env:"MYSQL_USER" yaml:"user"`
	Password string `required:"true" env:"MYSQL_PASSWORD" yaml:"password"`
	Address  string `required:"true" env:"MYSQL_ADDRESS" yaml:"address"`
	Db       string `default:"eventops" env:"MYSQL_DB" yaml:"db"`
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

func GetMysql() Mysql {
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

func parseConf() {
	if configPath == nil {
		panic("config file path was empty")
	}

	err := configor.Load(&conf, *configPath)
	if err != nil {
		panic(err)
	}
	conf.Uc.Auth.WhiteUrlList = []string{
		"/api/user/register",
		"/api/user/login",
		"/api/dialer/connect",
		"/api/pipeline/callback",
	}
}

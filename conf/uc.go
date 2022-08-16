package conf

type Uc struct {
	Auth                  Auth   `yaml:"auth"`
	LoginTokenExpiresTime int64  `yaml:"loginTokenExpiresTime"`
	LoginTokenSignature   string `yaml:"loginTokenSignature"`
}

type Auth struct {
	WhiteUrlList []string `yaml:"whiteUrlList"`
}

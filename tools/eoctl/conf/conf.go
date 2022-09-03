package conf

import (
	"eventops/apistructs"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

func NewConfig() (*Config, error) {
	return readConf(CfgFile)
}

type Config struct {
	UserInfos []*UserInfo `yaml:"userInfos"`
}

func (c *Config) SetUser(user apistructs.User, server string) error {
	defaultUser := c.GetDefaultUser()
	if defaultUser == nil {
		c.UserInfos = append(c.UserInfos, &UserInfo{
			Server:   server,
			Username: user.Name,
			Token:    user.Token,
			Default:  true,
		})
		return c.save()
	}

	userInfo := c.GetUser(user.Name, server)
	if userInfo == nil {
		c.UserInfos = append(c.UserInfos, &UserInfo{
			Server:   server,
			Username: user.Name,
			Token:    user.Token,
			Default:  true,
		})
		return c.save()
	}

	for index, userInfo := range c.UserInfos {
		if userInfo.Username == user.Name && userInfo.Server == server {
			c.UserInfos[index] = &UserInfo{
				Server:   userInfo.Server,
				Username: userInfo.Username,
				Token:    user.Token,
				Default:  userInfo.Default,
			}
		}
		return c.save()
	}

	return nil
}

func (c *Config) GetUser(username string, server string) *UserInfo {
	if username == "" {
		return c.GetDefaultUser()
	}

	for _, user := range c.UserInfos {
		if user.Username == username && server == user.Server {
			return user
		}
	}
	return nil
}

func (c *Config) GetDefaultUser() *UserInfo {
	for _, user := range c.UserInfos {
		if user.Default {
			return user
		}
	}
	return nil
}

func (c *Config) save() error {
	f, err := os.OpenFile(CfgFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	out, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	_, err = f.Write(out)
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func readConf(filename string) (*Config, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return &Config{}, err
	}

	var conf Config
	err = yaml.Unmarshal(buf, &conf)
	if err != nil {
		return &Config{}, fmt.Errorf("yaml Unmarshal file %q err %v", filename, err)
	}
	return &conf, nil
}

type UserInfo struct {
	Server   string `yaml:"server"`
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
	Default  bool   `yaml:"default"`
}

var CfgFile string

func InitConfig() {
	if CfgFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println("get home path error: ", err)
			os.Exit(1)
		}

		CfgFile = fmt.Sprintf("%s/%s", home, ".eoctl.yaml")
		exist, err := pathExists(CfgFile)
		if err != nil {
			fmt.Println("get config file error: ", err)
			os.Exit(1)
		}

		if !exist {
			_, err := os.Create(CfgFile)
			if err != nil {
				fmt.Printf("default config file(path: %v) create error: %v \n ", CfgFile, err)
				os.Exit(1)
			}
		}
	}

	exist, err := pathExists(CfgFile)
	if err != nil {
		fmt.Println("get config file error: ", err)
		os.Exit(1)
	}
	if !exist {
		fmt.Printf("config file(path: %v) not exist", CfgFile)
		os.Exit(1)
	}
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

var configFile = "config.yaml"
var appConfig *Config

//Config the app config
type Config struct {
	IsDevelop bool `yaml:"isDevelop"`
	HTTP      struct {
		Addr string `yaml:"addr"`
	}

	UploadDir string `yaml:"uploadDir"`
	SqliteDB  string `yaml:"sqliteDB"`

	AutoCert struct {
		Domains  []string
		CacheDir string `yaml:"cacheDir"`
	}

	Redis struct {
		Addr     string
		Password string `yaml:"password"`
		DB       int
	}

	Token struct {
		Keeper   string
		Expire   time.Duration `yaml:"expire"`
		AuthName string        `yaml:"authName"`
		Length   int
	}

	Logger struct {
		File     string
		Level    string
		MaxAge   time.Duration
		Compress bool
	}

	User struct {
		CacheTTL time.Duration `yaml:"cacheTTL"`
		//是否可以多个终端登陆同一个账户
		AllowMultiLogin bool `yaml:"allowMultiLogin"`
	}

	VCode struct {
		CacheTTL time.Duration `yaml:"cacheTTL"`
		Length   int           `yaml:"length"`
	}
}

func readFile(name string) ([]byte, error) {
	if !filepath.IsAbs(name) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("os.Getwd error:%v", err)
		}

		confPath, found := SearchPath(wd, name)
		if !found {
			return nil, fmt.Errorf("%s was not found", name)
		}
		name = confPath
	}

	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("read %s error:%v", name, err)
	}
	return data, nil
}

//ParseConfig parse yaml config
func ParseConfig(data []byte) (*Config, error) {
	var conf Config
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

//SaveConfig save config to filePath
func SaveConfig(config *Config, filePath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, data, 0660)
}

//SetConfigFile set app config
func SetConfigFile(f string) {
	configFile = f
	appConfig = nil
}

//AppConfig get singleton config instance
func AppConfig() *Config {
	if appConfig == nil {
		data, err := readFile(configFile)
		if err != nil {
			log.Fatalf(err.Error())
		}
		appConfig, err = ParseConfig(data)
	}
	return appConfig
}

//TestConfig returns config instance for testing
func TestConfig() *Config {
	data, err := readFile("config_test.yaml")
	if err != nil {
		log.Fatalf(err.Error())
	}
	config, err := ParseConfig(data)
	return config
}

package common

import (
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var Config *LocalConfig = nil

// 手动调用这个"config.yaml"
func InitConfig(filename string) error {
	Logger.Info("load config", zap.String("name", filename))

	var err error
	Config, err = LoadConfig(filename)
	return err
}

type RedisConf struct {
	RedisHost string `yaml:"redis_host" `
	RedisPwd  string `yaml:"redis_pwd"`
}

type MongoConf struct {
	MongoHost string `yaml:"mongo_host" `
	DbName    string `yaml:"db_name"`
}

type ServerConf struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	QueLen   int    `yaml:"queue_len"`
	Workers  int    `yaml:"workers"`
	Schema   string `yaml:"schema"`
	CertFile string `yaml:"cert"`
	KeyFile  string `yaml:"key"`
}

type LocalConfig struct {
	Redis   RedisConf  `yaml:"redis"`
	Server  ServerConf `yaml:"server"`
	MongoDb MongoConf  `yaml:"mongo"`
}

// "config.yaml"
func LoadConfig(fileName string) (*LocalConfig, error) {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, fileName)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err.Error())
	} // 将读取的yaml文件解析为响应的 struct
	config := LocalConfig{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err.Error())
	}
	if config.Server.QueLen < 100 {
		config.Server.QueLen = 100
	}
	if config.Server.Workers < 1 {
		config.Server.Workers = 1
	}
	return &config, err
}

func SaveConfig(conf *LocalConfig) bool {
	data, err := yaml.Marshal(conf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(string(data))
	err = ioutil.WriteFile("./config.yaml", data, 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return true
}

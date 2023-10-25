package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type RedisConf struct {
	RedisHost string `yaml:"redis_host" `
	RedisPwd  string `yaml:"redis_pwd"`
}

type ServerConf struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	QueLen  int    `yaml:"queue_len"`
	Workers int    `yaml:"workers"`
}

type Config struct {
	Redis  RedisConf  `yaml:"redis"`
	Server ServerConf `yaml:"server"`
}

func LoadConfig() (conf *Config) {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return nil
	}
	path := filepath.Join(dir, "config.yaml")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err.Error())
	} // 将读取的yaml文件解析为响应的 struct
	config := Config{}
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
	return &config
}

func SaveConfig(conf *Config) bool {
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

package main

import (
	"encoding/json"
	"fmt"
	"github.com/cihub/seelog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type EngineConfig struct {
	EmailConfig    EmailConfig    `yaml:"email-config"`
	DatabaseConfig DatabaseConfig `yaml:"database-config"`
	RedisConfig    RedisConfig    `yaml:"redis-config"`
}

type EmailConfig struct {
	SmtpHost     string `yaml:"smtp-host"`
	SmtpPort     int    `yaml:"smtp-port"`
	SmtpUser     string `yaml:"smtp-user"`
	SmtpPassword string `yaml:"smtp-password"`
	SmtpTls      bool   `yaml:"smtp-tls"`
}

type DatabaseConfig struct {
	DBHost     string `yaml:"db-host"`
	DBPort     int    `yaml:"db-port"`
	DBUser     string `yaml:"db-user"`
	DBPassword string `yaml:"db-password"`
}

type RedisConfig struct {
	RedisHost string `yaml:"redis-host"`
	RedisPort int    `yaml:"redis-port"`
	RedisDB   int    `yaml:"redis-db"`
}

var Config EngineConfig

func LoadConfig(configPath string) {
	buffer, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	yaml.Unmarshal(buffer, &Config)
	configJson, _ := json.Marshal(Config)
	seelog.Debug(string(configJson))
}

package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var cf *Config

type Config struct {
	General struct {
		BotToken          string `yaml:"bot_token"`
		Language          string `yaml:"language"`
		WorkerNum         int    `yaml:"worker_num"`
		DownloadWorkerNum int    `yaml:"download_worker_num"`
		AdminUID          int64  `yaml:"admin_uid"`
		UserDailyLimit    int    `yaml:"user_daily_limit"`
	} `yaml:"general"`
	Logger struct {
		Report         bool   `yaml:"report"`
		ReportUrl      string `yaml:"report_url"`
		ReportQueryKey string `yaml:"report_query_key"`
	} `yaml:"logger"`
	Redis struct {
		Server   string `yaml:"server"`
		Port     string `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
}

func Init() {
	_, err := os.Stat("./config.yaml")
	if err != nil && os.IsNotExist(err) {
		log.Fatalln("config.yaml not exist！！")
	}

	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalln("failed to load config.yaml", err)
	}

	cf = new(Config)
	err = yaml.Unmarshal(data, cf)
	if err != nil {
		log.Fatalln("failed to parse config.yaml", err)
	}
}

func Get() *Config {
	return cf
}

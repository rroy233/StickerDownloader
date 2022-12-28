package config

import (
	"github.com/rroy233/tg-stickers-dl/utils"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var cf *Config

type Config struct {
	General struct {
		BotToken          string `yaml:"bot_token"`
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
	if utils.IsExist("./config.yaml") == false {
		log.Fatalln("配置文件config.yaml不存在！！")
	}

	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalln("读取配置文件失败！！", err)
	}

	cf = new(Config)
	err = yaml.Unmarshal(data, cf)
	if err != nil {
		log.Fatalln("解析配置文件失败！！", err)
	}
}

func Get() *Config {
	return cf
}

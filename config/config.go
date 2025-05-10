package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

var cf *Config

type Config struct {
	General struct {
		BotToken                string `yaml:"bot_token"`
		Language                string `yaml:"language"`
		WorkerNum               int    `yaml:"worker_num"`
		DownloadWorkerNum       int    `yaml:"download_worker_num"`
		AdminUID                int64  `yaml:"admin_uid"`
		UserDailyLimit          int    `yaml:"user_daily_limit"`
		ProcessWaitQueueMaxSize int    `yaml:"process_wait_queue_max_size"`
		ProcessTimeout          int    `yaml:"process_timeout"`
		SupportTGSFile          bool   `yaml:"support_tgs_file"`
		MaxAmountPerReq         int    `yaml:"max_amount_per_req"`
	} `yaml:"general"`
	Community struct {
		Enable          bool `yaml:"enable"`
		ForceChannelSub bool `yaml:"force_channel_sub"`
		RewardOnSub     bool `yaml:"reward_on_sub"`
		Channel         struct {
			Username string `yaml:"username"`
		} `yaml:"channel"`
		Reward struct {
			ExtraDownloadTimes int `yaml:"extra_download_times"`
		} `yaml:"reward"`
	} `yaml:"community"`
	Cache struct {
		Enabled            bool   `yaml:"enabled"`
		StorageDir         string `yaml:"storage_dir"`
		MaxDiskUsage       int    `yaml:"max_disk_usage"`
		CacheExpire        int    `yaml:"cache_expire"`
		CacheCleanInterval int    `yaml:"cache_clean_interval"`
	} `json:"cache"`
	Logger struct {
		Report         bool   `yaml:"report"`
		ReportUrl      string `yaml:"report_url"`
		ReportQueryKey string `yaml:"report_query_key"`
	} `yaml:"logger"`
	Redis struct {
		Server   string `yaml:"server"`
		Port     string `yaml:"port"`
		TLS      bool   `yaml:"tls"`
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

	//validate config
	if cf.General.MaxAmountPerReq == 0 {
		log.Fatalln("General.MaxAmountPerReq should NOT be 0")
	}

	//community
	if cf.Community.Enable {
		if cf.Community.Channel.Username == "" || cf.Community.Channel.Username == "@your_channel" {
			log.Fatalln("You have enabled the community setting, but you did not provide the correct channel username")
		}
		if !strings.HasPrefix(cf.Community.Channel.Username, "@") {
			log.Fatalln("The channel name should start with @, for example @your_channel")
		}
		if cf.Community.RewardOnSub && cf.Community.Reward.ExtraDownloadTimes == 0 {
			log.Println("[WARN] You have enabled channel subscription rewards, but the number of rewards you set is 0")
		}
	}
}

func Get() *Config {
	return cf
}

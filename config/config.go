package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

var cf *Config

type Config struct {
	General struct {
		BotToken                string `yaml:"bot_token"                env:"BOT_TOKEN,required"`
		Language                string `yaml:"language"                 env:"LANGUAGE"           envDefault:"zh-hans"`
		WorkerNum               int    `yaml:"worker_num"               env:"WORKER_NUM"         envDefault:"2"`
		DownloadWorkerNum       int    `yaml:"download_worker_num"      env:"DOWNLOAD_WORKER_NUM" envDefault:"3"`
		AdminUID                int64  `yaml:"admin_uid"                env:"ADMIN_UID"          envDefault:"0"`
		UserDailyLimit          int    `yaml:"user_daily_limit"         env:"USER_DAILY_LIMIT"   envDefault:"10"`
		ProcessWaitQueueMaxSize int    `yaml:"process_wait_queue_max_size" env:"PROCESS_WAIT_QUEUE_MAX_SIZE" envDefault:"50"`
		ProcessTimeout          int    `yaml:"process_timeout"          env:"PROCESS_TIMEOUT"    envDefault:"60"`
		SupportTGSFile          bool   `yaml:"support_tgs_file"         env:"SUPPORT_TGS_FILE"   envDefault:"false"`
		MaxAmountPerReq         int    `yaml:"max_amount_per_req"       env:"MAX_AMOUNT_PER_REQ" envDefault:"100"`
	} `yaml:"general" envPrefix:"GENERAL_"`

	Community struct {
		Enable          bool `yaml:"enable"            env:"ENABLE"            envDefault:"true"`
		ForceChannelSub bool `yaml:"force_channel_sub" env:"FORCE_CHANNEL_SUB" envDefault:"true"`
		RewardOnSub     bool `yaml:"reward_on_sub"     env:"REWARD_ON_SUB"     envDefault:"true"`

		Channel struct {
			Username string `yaml:"username" env:"USERNAME"`
		} `yaml:"channel" envPrefix:"CHANNEL_"`

		Reward struct {
			ExtraDownloadTimes int `yaml:"extra_download_times" env:"EXTRA_DOWNLOAD_TIMES" envDefault:"3"`
		} `yaml:"reward" envPrefix:"REWARD_"`
	} `yaml:"community" envPrefix:"COMMUNITY_"`

	Cache struct {
		Enabled            bool   `yaml:"enabled"              env:"ENABLED"              envDefault:"false"`
		StorageDir         string `yaml:"storage_dir"          env:"STORAGE_DIR"          envDefault:"./storage/cache"`
		MaxDiskUsage       int    `yaml:"max_disk_usage"       env:"MAX_DISK_USAGE"       envDefault:"1024"`
		CacheExpire        int    `yaml:"cache_expire"         env:"CACHE_EXPIRE"         envDefault:"86400"`
		CacheCleanInterval int    `yaml:"cache_clean_interval" env:"CACHE_CLEAN_INTERVAL" envDefault:"1800"`
	} `yaml:"cache" envPrefix:"CACHE_"`

	Logger struct {
		Report         bool   `yaml:"report"           env:"REPORT"            envDefault:"false"`
		ReportUrl      string `yaml:"report_url"       env:"REPORT_URL"` // 若遵循 GoLint 可改名为 ReportURL
		ReportQueryKey string `yaml:"report_query_key" env:"REPORT_QUERY_KEY"`
	} `yaml:"logger" envPrefix:"LOGGER_"`

	Redis struct {
		Server   string `yaml:"server"   env:"SERVER"   envDefault:"localhost"`
		Port     string `yaml:"port"     env:"PORT"     envDefault:"6379"`
		TLS      bool   `yaml:"tls"      env:"TLS"      envDefault:"false"`
		Password string `yaml:"password" env:"PASSWORD"`
		DB       int    `yaml:"db"       env:"DB"       envDefault:"0"`
	} `yaml:"redis" envPrefix:"REDIS_"`
}

func Init() {
	if !isExist("./config.yaml") {
		if !isExist("./.env") {
			log.Fatalln("config.yaml and .env not exist！！")
		}
		//env
		log.Println("Loading configuration from: .env")
		_ = godotenv.Load()
		cf = new(Config)
		if err := env.Parse(cf); err != nil {
			log.Fatalln("failed to parse .env", err)
		}
	} else {
		//yaml
		log.Println("Loading configuration from: config.yaml")
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

func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

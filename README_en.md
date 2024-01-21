# StickerDownloader
<p>
   <a href="https://github.com/rroy233/StickerDownloader">
      <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/rroy233/StickerDownloader?style=flat-square">
   </a>
   <a href="https://github.com/rroy233/StickerDownloader/releases">
      <img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/rroy233/StickerDownloader?style=flat-square">
   </a>
   <a href="https://github.com/rroy233/StickerDownloader/blob/main/LICENSE">
      <img alt="GitHub license" src="https://img.shields.io/github/license/rroy233/StickerDownloader?style=flat-square">
   </a>
   <a href="https://github.com/rroy233/StickerDownloader/commits/main">
      <img alt="GitHub last commit" src="https://img.shields.io/github/last-commit/rroy233/StickerDownloader?style=flat-square">
   </a>
    <a href="https://t.me/stickers_download_bot">
      <img alt="GitHub last commit" src="https://img.shields.io/badge/demo-%40stickers__download__bot-green?style=flat-square">
   </a>
</p>

> A Telegram Stickers Download Bot.

[中文](README.md) | EN

### Feature

* Send stickers or sticker links to the bot, and it will convert them into easily savable GIF files for you.
* Supports the conversion of Telegram's official stickers (tgs) to GIFs.
* Forward GIFs to the bot, and it will send them back to you in file form for easy saving.
* Download single sticker.
* Download whole sticker set.

![cover](docs/demo.gif)

### Requirement

- Redis
- ffmpeg
- [lottie2gif](https://github.com/rroy233/lottie2gif) (Optional)

### Usage

#### Download

1. Clone Repo

   ```shell
   git clone https://github.com/rroy233/StickerDownloader.git
   ```

2. Get executable

    1. Compile it yourself

       ```shell
       cd StickerDownloader/
       # Go version：Go 1.19+
       go build
       ```
    2. Download from release
       download from [releases](https://github.com/rroy233/StickerDownloader/releases), rename executable to `StickerDownloader`, and then put it into project root folder.


#### Create Bot

use BotFather to create bot.

get `bot_token` , and finish command setting.

```
help - Help
getlimit - Get remaining usage times
admin - Get admin commands
```

#### Configuration

copy `config.example.yaml` to `config.yaml`.

```yaml
general:
  bot_token: "xxx" # Obtained from BotFather
  language: "zh-hans" # Default language (corresponding to the filename in the /languages folder)
  worker_num: 2 # Number of threads for message processing
  download_worker_num: 3 # Number of threads for downloading and file transcoding
  admin_uid: 0 # Admin UID
  user_daily_limit: 10 # Daily usage limit
  process_wait_queue_max_size: 50 # Maximum length of the wait queue
  process_timeout: 60 # Processing timeout (s)
  support_tgs_file: false # Whether to enable tgs stickers support
  max_amount_per_req: 100 # Maximum number of stickers allowed when downloading the whole set

cache:
  enabled: false # Whether to enable file caching (requires Redis)
  storage_dir: "./storage/cache" # Location for storing file cache
  max_disk_usage: 1024 # Maximum disk usage (MB)
  cache_expire: 86400 # File cache validity period (s)
  cache_clean_interval: 1800 # Expiration file check interval (s)

logger:
  report: false # Whether to enable remote log reporting (requires a custom receiver, see https://github.com/rroy233/logger)
  report_url: "" # Remote log reporting URL (POST)
  report_query_key: "" # Query parameter for remote log reporting

redis:
  server: "localhost" # Redis server address
  port: "6379" # Redis server port
  tls: false # Whether to enable TLS for Redis
  password: "" # Redis password
  db: 0 # Redis database number
```

#### Download ffmpeg

If ffmpeg is already installed, you can skip this step.

Download ffmpeg from [official website](https://ffmpeg.org/),  rename it to `ffmpeg` or `ffmpeg.exe`, and put it into `./ffmpeg` folder.

#### Integrate lottie2gif

If you wish to enable the conversion of TGS format stickers, you will need to integrate [lottie2gif](https://github.com/rroy233/lottie2gif) into the StickerDownloader and make the following changes to the configuration file:

```yaml
  support_tgs_file: true
```

#### Launch Script

```shell
# build and run
bash ./buildrun.sh 

# run
bash ./run.sh 
```

### LICENSE
GPL-3.0 license
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

> 一个可以帮你下载表情包的telegram机器人

中文 | [EN](README_en.md)

### 功能

* 发送表情、表情链接给bot，bot为您转换为便于保存的gif文件.
* 支持将Telegram官方出品的表情(tgs)格式转换为gif.
* 转发gif图给bot，bot会以文件形式发送回给你以便保存.
* 下载单个表情.
* 下载整个表情包.

![cover](docs/demo.gif)

### 运行要求

- Redis
- ffmpeg
- [lottie2gif](https://github.com/rroy233/lottie2gif) (可选)

### 使用方法

#### 下载

1. 克隆仓库

   ```shell
   git clone https://github.com/rroy233/StickerDownloader.git
   ```

2. 获取可执行文件

    1. 自行编译

       ```shell
       cd StickerDownloader/
       # 自行编译
       # go版本要求：go1.19+
       go build
       ```
       
    2. 前往release下载

       下载已编译的[可执行文件](https://github.com/rroy233/StickerDownloader/releases)，重新命名为`StickerDownloader`，放于项目文件夹内

#### 找 BotFather 创建Bot

获得`bot_token`,然后设置命令列表

```
help - 帮助
getlimit - 获取当日使用限额
admin - 查看管理员指令
```

#### 配置

复制`config.example.yaml`为`config.yaml`

```yaml
general:
  bot_token: "xxx" # 从BotFather获得
  language: "zh-hans" # 默认语言(对应/languages文件夹中的文件名)
  worker_num: 2 # 消息处理的线程数
  download_worker_num: 3 # 下载、文件转码工作线程数
  admin_uid: 0 # 管理员UID
  user_daily_limit: 10 # 每日使用次数限制
  process_wait_queue_max_size: 50 # 等待队列最大长度
  process_timeout: 60 # 处理超时时间(s)
  support_tgs_file: false # 是否开启tgs表情支持
  max_amount_per_req: 100 # 下载整套表情包时允许的最大数量

cache:
  enabled: false # 是否启用文件缓存(需要使用Redis)
  storage_dir: "./storage/cache" # 文件缓存存放位置
  max_disk_usage: 1024 # 最大磁盘占用(MB)
  cache_expire: 86400 # 文件缓存有效期(s)
  cache_clean_interval: 1800 # 过期文件检查周期(s)

logger:
  report: false # 是否启用远程日志上报(需要自行设计接收端，参考https://github.com/rroy233/logger)
  report_url: "" # 远程日志上报url(POST)
  report_query_key: "" # 远程日志上报query参数

redis:
  server: "localhost" # redis服务器地址
  port: "6379" # redis服务器端口
  tls: false # redis是否启用tls
  password: "" # redis密码
  db: 0 # redis数据库编号
```


#### 下载ffmpeg

若已安装ffmpeg可跳过该步骤。

下载对应平台的[ffmpeg](https://ffmpeg.org/)的可执行文件，命名格式为`ffmpeg`或`ffmpeg.exe`，复制到`./ffmpeg`文件夹。

#### lottie2gif集成

若需要支持tgs格式表情转换，需要为StickerDownloader集成[lottie2gif](https://github.com/rroy233/lottie2gif).

并更改配置文件：

```yaml
  support_tgs_file: true
```

#### 后台运行脚本

```shell
# 编译并运行
bash ./buildrun.sh 

# 直接运行
bash ./run.sh 
```

### LICENSE
GPL-3.0 license

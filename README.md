# tg-stickers-dl

> 一个可以帮你下载表情包的telegram机器人

### 功能

* 发送表情给bot，bot为您转换为便于保存的gif文件
* 下载单个表情
* 下载整个表情包

### 运行要求

- Redis
- ffmpeg

### 使用方法

#### 下载

1. 克隆仓库

   ```shell
   git clone https://github.com/rroy233/tg-stickers-dl.git
   ```

2. 获取可执行文件

    1. 自行编译

       ```shell
       cd tg-stickers-dl/
       # 自行编译
       # go版本要求：go1.19+
       go build
       ```
       
    2. 前往release下载

       下载已编译的[可执行文件](https://github.com/rroy233/tg-stickers-dl/releases)，重新命名为`tg-stickers-dl`，放于项目文件夹内

#### 找 BotFather 创建Bot

获得`bot_token`,然后设置命令列表

```
help - 帮助
getlimit - 获取当日使用限额
```

#### 创建配置文件

复制`config.example.yaml`为`config.yaml`

#### 下载ffmpeg

下载对应平台的[ffmpeg](https://ffmpeg.org/)的可执行文件，命名格式为`ffmpeg-{GOOS}-{GOARCH}`，复制到`./ffmpeg`文件夹

#### 运行程序

```shell
# 编译并运行
bash ./buildrun.sh 

# 直接运行
bash ./run.sh 
```


# tg-stickers-dl
<p>
   <a href="https://github.com/rroy233/tg-stickers-dl">
      <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/rroy233/tg-stickers-dl?style=flat-square">
   </a>
   <a href="https://github.com/rroy233/tg-stickers-dl/releases">
      <img alt="GitHub release (latest by date)" src="https://img.shields.io/github/v/release/rroy233/tg-stickers-dl?style=flat-square">
   </a>
   <a href="https://github.com/rroy233/tg-stickers-dl/blob/main/LICENSE">
      <img alt="GitHub license" src="https://img.shields.io/github/license/rroy233/tg-stickers-dl?style=flat-square">
   </a>
   <a href="https://github.com/rroy233/tg-stickers-dl/commits/main">
      <img alt="GitHub last commit" src="https://img.shields.io/github/last-commit/rroy233/tg-stickers-dl?style=flat-square">
   </a>
    <a href="https://t.me/stickers_download_bot">
      <img alt="GitHub last commit" src="https://img.shields.io/badge/demo-%40stickers__download__bot-green?style=flat-square">
   </a>
</p>

> A Telegram Stickes Download Bot.

[中文](README.md) | EN

### Feature

* Send sticker to Bot, so it will help you convert into gif file.
* Download single sticker.
* Download whole sticker set.

![cover](docs/demo.gif)

### Requirement

- Redis
- ffmpeg

### Usage

#### Download

1. Clone Repo

   ```shell
   git clone https://github.com/rroy233/tg-stickers-dl.git
   ```

2. Get executable

    1. Compile it yourself

       ```shell
       cd tg-stickers-dl/
       # Go version：Go 1.19+
       go build
       ```
    
2. Download from release
   
   download from [releases](https://github.com/rroy233/tg-stickers-dl/releases), rename executable to `tg-stickers-dl`, and then put it into project root folder.

#### Create Bot

use BotFather to create bot.

get `bot_token` , and finish command setting.

```
help - Help
getlimit - Get remaining usage times
```

#### Create Config File

copy `config.example.yaml` to `config.yaml`.

#### Download ffmpeg

download ffmpeg from [official website](https://ffmpeg.org/),  rename it to `ffmpeg-{GOOS}-{GOARCH}`, and put it into `./ffmpeg` folder.

#### Run

```shell
# build and run
bash ./buildrun.sh 

# run
bash ./run.sh 
```

### LICENSE
GPL-3.0 license
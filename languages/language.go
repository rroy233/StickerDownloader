package languages

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/config"
	"gopkg.in/rroy233/logger.v2"
	"os"
	"strings"
)

type LanguageStruct struct {
	System struct {
		Running            string `json:"running"`
		StopRunning        string `json:"stop_running"`
		DbRedisStartFailed string `json:"db_redis_start_failed"`
		DbRedisConnected   string `json:"db_redis_connected"`
		FfmpegNotExist     string `json:"ffmpeg_not_exist"`
		RlottieNotExist    string `json:"rlottie_not_exist"`
	} `json:"system"`
	BotMsg struct {
		Processing                    string `json:"processing"`
		StickersSetInfoFromUrl        string `json:"stickers_set_info_from_url"`
		DownloadingWithProgress       string `json:"downloading_with_progress"`
		UploadedThirdParty            string `json:"uploaded_third_party"`
		UploadedTelegram              string `json:"uploaded_telegram"`
		GetLimitCommand               string `json:"get_limit_command"`
		StartCommand                  string `json:"start_command"`
		HelpCommand                   string `json:"help_command"`
		ConvertCompleted              string `json:"convert_completed"`
		ConvertedWaitingUpload        string `json:"converted_waiting_upload"`
		DownloadStickerSet            string `json:"download_sticker_set"`
		ReloadConfigSuccess           string `json:"reload_config_success"`
		QueueAbortBtn                 string `json:"queue_abort_btn"`
		QueueAborted                  string `json:"queue_aborted"`
		QueueProcess                  string `json:"queue_process"`
		ErrRateReachLimit             string `json:"err_rate_reach_limit"`
		ErrSysBusy                    string `json:"err_sys_busy"`
		ErrNoPermission               string `json:"err_no_permission"`
		ErrReachLimit                 string `json:"err_reach_limit"`
		ErrFailedToDownload           string `json:"err_failed_to_download"`
		ErrSysFailureOccurred         string `json:"err_sys_failure_occurred"`
		ErrFailed                     string `json:"err_failed"`
		ErrTimeout                    string `json:"err_timeout"`
		ErrUploadFailed               string `json:"err_upload_failed"`
		ErrStickerNotSupport          string `json:"err_sticker_not_support"`
		ErrConvertFailed              string `json:"err_convert_failed"`
		ErrSendFileFailed             string `json:"err_send_file_failed"`
		ErrStickerSetAmountReachLimit string `json:"err_sticker_set_amount_reach_limit"`
	} `json:"bot_msg"`
}

var lang map[string]*LanguageStruct

func Init() {
	dir, err := os.ReadDir("./languages")
	if err != nil {
		logger.Error.Fatalln(fmt.Sprintf("failed to read language folder! \n"))
	}

	lang = make(map[string]*LanguageStruct)
	for _, entry := range dir {
		if strings.HasSuffix(entry.Name(), ".json") != true {
			continue
		}
		namePart := strings.Split(entry.Name(), ".")
		fileData, err := os.ReadFile(fmt.Sprintf("./languages/%s", entry.Name()))
		if err != nil {
			logger.Error.Fatalln(fmt.Sprintf("failed to load language pack! "))
		}

		langItem := new(LanguageStruct)
		err = json.Unmarshal(fileData, langItem)
		if err != nil {
			logger.Error.Fatalln(fmt.Sprintf("failed to parse language pack! "))
		}

		logger.Info.Printf("Loaded language <%s>", namePart[0])
		lang[namePart[0]] = langItem
	}
	if len(lang) == 0 {
		logger.Error.Fatalln(fmt.Sprintf("NO language config has been loaded! "))
	}
	//check default language
	if lang[config.Get().General.Language] == nil {
		logger.Error.Fatalln(fmt.Sprintf("default language config NOT exist! "))
	}

	return
}

// Get return language config depending on user's language code
//
// if pass a nil, then it will return default language config
func Get(update *tgbotapi.Update) *LanguageStruct {
	if update == nil {
		return lang[config.Get().General.Language]
	}

	languageCode := ""
	if update.Message != nil && lang[update.Message.From.LanguageCode] != nil {
		languageCode = update.Message.From.LanguageCode
	} else if update.CallbackQuery != nil && lang[update.CallbackQuery.From.LanguageCode] != nil {
		languageCode = update.CallbackQuery.From.LanguageCode
	} else {
		//no matched language, return default language
		languageCode = config.Get().General.Language
	}
	return lang[languageCode]
}

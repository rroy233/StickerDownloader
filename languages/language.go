package languages

import (
	"encoding/json"
	"fmt"
	"github.com/rroy233/logger"
	"os"
)

type LanguageStruct struct {
	System struct {
		Running            string `json:"running"`
		StopRunning        string `json:"stop_running"`
		DbRedisStartFailed string `json:"db_redis_start_failed"`
		DbRedisConnected   string `json:"db_redis_connected"`
		FfmpegNotExist     string `json:"ffmpeg_not_exist"`
	} `json:"system"`
	BotMsg struct {
		Processing              string `json:"processing"`
		DownloadingWithProgress string `json:"downloading_with_progress"`
		UploadedThirdParty      string `json:"uploaded_third_party"`
		UploadedTelegram        string `json:"uploaded_telegram"`
		GetLimitCommand         string `json:"get_limit_command"`
		StartCommand            string `json:"start_command"`
		HelpCommand             string `json:"help_command"`
		ConvertCompleted        string `json:"convert_completed"`
		ConvertedWaitingUpload  string `json:"converted_waiting_upload"`
		DownloadStickerSet      string `json:"download_sticker_set"`
		ReloadConfigSuccess     string `json:"reload_config_success"`
		ErrNoPermission         string `json:"err_no_permission"`
		ErrReachLimit           string `json:"err_reach_limit"`
		ErrFailedToDownload     string `json:"err_failed_to_download"`
		ErrSysFailureOccurred   string `json:"err_sys_failure_occurred"`
		ErrFailed               string `json:"err_failed"`
		ErrTimeout              string `json:"err_timeout"`
		ErrUploadFailed         string `json:"err_upload_failed"`
		ErrStickerNotSupport    string `json:"err_sticker_not_support"`
		ErrConvertFailed        string `json:"err_convert_failed"`
		ErrSendFileFailed       string `json:"err_send_file_failed"`
	} `json:"bot_msg"`
}

var lang *LanguageStruct

func Init(languageCode string) {
	_, err := os.Stat(fmt.Sprintf("./languages/%s.json", languageCode))
	if err != nil && os.IsNotExist(err) {
		logger.Error.Fatalln(fmt.Sprintf("failed to init language pack! %s.json not exist!\n", languageCode))
	}

	fileData, err := os.ReadFile(fmt.Sprintf("./languages/%s.json", languageCode))
	if err != nil {
		logger.Error.Fatalln(fmt.Sprintf("failed to load language pack! \n"))
	}

	lang = new(LanguageStruct)
	err = json.Unmarshal(fileData, lang)
	if err != nil {
		logger.Error.Fatalln(fmt.Sprintf("failed to parse language pack! \n"))
	}
	return
}

func Get() *LanguageStruct {
	return lang
}

package db

import (
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/utils"
	"os"
	"strings"
	"time"
)

var cacheEnabled bool
var cacheDir string
var cacheMaxUsage int64

var CacheExpire = 24 * time.Hour
var CacheCleanInterval = 30 * time.Minute

var (
	CacheErrorDisabled     = errors.New("CacheErrorDisabled")
	CacheErrorNotExist     = errors.New("CacheErrorNotExist")
	CacheErrorVerifyFailed = errors.New("CacheErrorVerifyFailed")
)

var cacheLocalDiskUsage int64

func initCache() {
	if config.Get().Cache.Enabled == false {
		return
	}
	loggerPrefix := "[CacheInit]"
	cacheEnabled = true
	if utils.IsExist(config.Get().Cache.StorageDir) == false {
		err := os.Mkdir(config.Get().Cache.StorageDir, 0755)
		if err != nil {
			logger.FATAL.Println(loggerPrefix+"Failed to create cache folder!!", err)
		}
	}
	cacheDir = config.Get().Cache.StorageDir

	cacheMaxUsage = int64(config.Get().Cache.MaxDiskUsage << 20)
	err := cacheGetLocalDiskUsage()
	if err != nil {
		logger.FATAL.Println(loggerPrefix+"Failed to calculate local storage usage:", err)
	}

	go cacheCleaner()

	return
}

func cacheGetLocalDiskUsage() error {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	sum := int64(0)
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".gif") == false {
			continue
		}
		info, _ := entry.Info()
		sum += info.Size()
	}

	cacheLocalDiskUsage = sum

	return nil
}

// FindStickerCache 查询是否有缓存
// 返回本地文件地址
//
// 若不存在，则返回CacheErrorNotExist
func FindStickerCache(uniqueID string) (string, error) {
	if cacheEnabled == false {
		return "", CacheErrorDisabled
	}
	data := rdb.Get(ctx, fmt.Sprintf("%s:Sticker_Cache:%s", ServicePrefix, utils.MD5Short(uniqueID))).Val()
	if data == "" {
		return "", CacheErrorNotExist
	}

	item := new(stickerItem)
	err := json.Unmarshal([]byte(data), item)
	if err != nil {
		return "", CacheErrorNotExist
	}

	if utils.IsExist(item.SavePath) == false {
		return "", CacheErrorNotExist
	}

	//校验
	fileMd5, err := utils.MD5File(item.SavePath)
	if err != nil {
		return "", CacheErrorNotExist
	}
	if fileMd5 != item.MD5 {
		return "", CacheErrorVerifyFailed
	}

	//复制一份到tmp
	newFilePath := fmt.Sprintf("./storage/tmp/convert_%s.gif", utils.RandString())
	err = utils.CopyFile(item.SavePath, newFilePath)
	if err != nil {
		return "", err
	}

	return newFilePath, nil
}

func CacheSticker(sticker tgbotapi.Sticker, convertedFilePath string) {
	if cacheEnabled == false {
		return
	}
	loggerPrefix := "[CacheSticker]"
	data := rdb.Get(ctx, fmt.Sprintf("%s:Sticker_Cache:%s", ServicePrefix, utils.MD5Short(sticker.FileUniqueID))).Val()
	if data != "" {
		return
	}

	item := new(stickerItem)
	item.Info = sticker
	item.SaveTimeStamp = time.Now().Unix()
	item.SavePath = fmt.Sprintf("%s/%d_%s.gif", cacheDir, item.SaveTimeStamp, utils.MD5(sticker.FileUniqueID))

	stat, _ := os.Stat(convertedFilePath)
	item.Size = stat.Size()

	fileMd5, err := utils.MD5File(convertedFilePath)
	if err != nil {
		logger.Error.Println(loggerPrefix+"utils.MD5File(convertedFilePath) error:", err)
	}
	item.MD5 = fileMd5

	err = utils.CopyFile(convertedFilePath, item.SavePath)
	if err != nil {
		logger.Error.Println(loggerPrefix+"Failed to copy file:", err)
		return
	}

	out, err := json.Marshal(item)
	if err != nil {
		logger.Error.Println(loggerPrefix+" json.Marshal(item) error:", err)
		return
	}

	err = rdb.Set(ctx, fmt.Sprintf("%s:Sticker_Cache:%s", ServicePrefix, utils.MD5Short(sticker.FileUniqueID)), string(out), CacheExpire).Err()
	if err != nil {
		logger.Error.Println(loggerPrefix+"Failed to store redis:", err)
		return
	}

	cacheLocalDiskUsage += item.Size
	return
}

func cacheCleaner() {
	for true {
		cacheDoClean()
		logger.Info.Printf("Cache Usage %dMB/%dMB", cacheLocalDiskUsage>>20, cacheMaxUsage>>20)
		time.Sleep(CacheCleanInterval)
	}
}

func cacheDoClean() {
	loggerPrefix := "[cacheDoClean]"
	keys := rdb.Keys(ctx, fmt.Sprintf("%s:Sticker_Cache:*", ServicePrefix)).Val()

	//本地文件名后缀的哈希表，若checkMap[localFilenamePrefix]!=0则本地文件有效，无需因过期而清除
	checkMap := make(map[string]int, len(keys))
	//本地文件名到stickerItem的映射
	itemMapByFilename := make(map[string]*stickerItem)
	for _, key := range keys {
		item := new(stickerItem)
		err := json.Unmarshal([]byte(rdb.Get(ctx, key).Val()), item)
		if err != nil {
			logger.Error.Println(loggerPrefix+"Failed to parse cache:", err)
			continue
		}
		checkMap[utils.MD5(item.Info.FileUniqueID)+".gif"]++

		//查看本地文件是否存在，若不存在，则删除redis对应记录
		if utils.IsExist(fmt.Sprintf("%s/%d_%s.gif", cacheDir, item.SaveTimeStamp, utils.MD5(item.Info.FileUniqueID))) == true {
			//存在
			itemMapByFilename[fmt.Sprintf("%d_%s.gif", item.SaveTimeStamp, utils.MD5(item.Info.FileUniqueID))] = item
		} else {
			//不存在
			rdb.Del(ctx, key)
		}
	}

	//读取本地所有文件
	localEntries, err := os.ReadDir(cacheDir)
	if len(localEntries) == 0 {
		return
	}
	if err != nil {
		logger.Error.Println(loggerPrefix+"Failed to read local files:", err)
		return
	}

	//遍历所有本地文件，若在redis中不存在，则删除本地文件
	for _, entry := range localEntries {
		if strings.Contains(entry.Name(), "_") == false {
			continue
		}
		if checkMap[strings.Split(entry.Name(), "_")[1]] == 0 {
			utils.RemoveFile(fmt.Sprintf("%s/%s", cacheDir, entry.Name()))
			logger.Info.Printf("%s cache file [%s] expired", loggerPrefix, entry.Name())
			itemMapByFilename[entry.Name()] = nil
		}
	}

	//判断是否达到容量阈值，若达到则清除旧缓存文件，直至最大容量的75%
	err = cacheGetLocalDiskUsage()
	if err != nil {
		logger.Error.Println(loggerPrefix+"cacheGetLocalDiskUsage error", err)
		return
	}
	if cacheLocalDiskUsage > cacheMaxUsage {
		for i := 0; i < len(localEntries); i++ {
			if cacheLocalDiskUsage < int64(float64(cacheMaxUsage)*0.75) {
				break
			}
			item := itemMapByFilename[localEntries[i].Name()]
			if item != nil {
				utils.RemoveFile(item.SavePath)
				rdb.Del(ctx, fmt.Sprintf("%s:Sticker_Cache:%s", ServicePrefix, utils.MD5Short(item.Info.FileUniqueID)))
				cacheLocalDiskUsage -= item.Size
			}
		}
	}
	_ = cacheGetLocalDiskUsage()
	return
}

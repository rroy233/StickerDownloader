package statistics

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gopkg.in/rroy233/logger.v2"
	"sync"
	"sync/atomic"
	"time"
)

type statistics struct {
	//用户
	//使用的用户(脱敏处理)
	UserTotalNum      map[string]int `json:"user_total_num"`
	_userTotalNumLock sync.Mutex

	//命令
	//命令使用情况
	Command      map[string]int `json:"command"`
	_commandLock sync.Mutex

	//消息处理
	//已处理的update总数
	MsgHandleTotalTimes int32 `json:"handle_total_times"`
	//已处理的sticker类型消息
	MsgStickerNum int32 `json:"msg_sticker_num"`
	//已处理的animation类型消息
	MsgAnimationNum int32 `json:"msg_animation_num"`
	//已处理的下载整套表情包的请求数
	MsgStickerSet int32 `json:"msg_sticker_set"`
	//已处理的链接下载请求
	MsgStickerUrl int32 `json:"msg_sticker_url"`

	//存储
	//存储变化(B)
	StorageChange int64 `json:"storage_change"`

	//缓存
	//击中缓存次数
	CacheHit  int32 `json:"cache_hit"`
	CacheMiss int32 `json:"cache_miss"`

	//网络
	//上传(B)
	NetworkUpload   int64 `json:"network_upload"`
	NetworkDownload int64 `json:"network_download"`

	//自身属性
	lock      sync.Mutex
	fieldMap  map[string]*int32
	StartTime time.Time `json:"start_time"`
	SaveTime  time.Time `json:"save_time"`
	EndTime   time.Time `json:"end_time"`
}

var Statistics *statistics
var rdb *redis.Client

const ServicePrefix = "StickerDl"

// InitStatistic 初始化统计模块
func InitStatistic(db *redis.Client) {
	//redis
	rdb = db

	Statistics = new(statistics)

	//查看redis是否有保存
	rdbData := getStatistic()
	if rdbData == "" {
		Statistics.new()
	} else {
		//redis内有记录，解析
		err := json.Unmarshal([]byte(rdbData), Statistics)
		if err != nil {
			logger.FATAL.Println("Failed to Init Statistics(parse json):", err)
			return
		}
		if Statistics.EndTime.Unix() < time.Now().Unix() {
			//过期了
			Statistics.PrintToLog() //输出
			Statistics.Reset()      //重置
		}
	}

	//auto job
	go autoReset()
	go autoSave()

	return
}

func (s *statistics) new() {
	//init maps
	s.UserTotalNum = make(map[string]int)
	s._userTotalNumLock = sync.Mutex{}
	s.Command = make(map[string]int)
	s._commandLock = sync.Mutex{}

	//calculate time
	s.StartTime = time.Now()
	//s.EndTime为下周一0点
	offset := 1
	now := time.Now()
	if now.Weekday() > time.Sunday {
		offset = 8 - int(now.Weekday())
	}
	s.EndTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)

	//lock
	s.lock = sync.Mutex{}

	return
}

func (s *statistics) RecordUser(uidMd5Short string) {
	s._userTotalNumLock.Lock()
	defer s._userTotalNumLock.Unlock()
	s.UserTotalNum[uidMd5Short]++
}

func (s *statistics) Record(field string, delta int32) {
	loggerPrefix := "[statistics-Record]"

	var dest32 *int32
	switch field {
	case "MsgHandleTotalTimes":
		dest32 = &s.MsgHandleTotalTimes
	case "MsgStickerNum":
		dest32 = &s.MsgStickerNum
	case "MsgAnimationNum":
		dest32 = &s.MsgAnimationNum
	case "MsgStickerSet":
		dest32 = &s.MsgStickerSet
	case "MsgStickerUrl":
		dest32 = &s.MsgStickerUrl
	case "CacheHit":
		dest32 = &s.CacheHit
	case "CacheMiss":
		dest32 = &s.CacheMiss
	}
	if dest32 != nil {
		atomic.AddInt32(dest32, delta)
		return
	}

	var dest64 *int64
	switch field {
	case "NetworkUpload":
		dest64 = &s.NetworkUpload
	case "NetworkDownload":
		dest64 = &s.NetworkDownload
	case "StorageChange":
		dest64 = &s.StorageChange
	default:
		logger.Error.Println(loggerPrefix+"Failed to find field:", field)
		return
	}
	if dest64 != nil {
		atomic.AddInt64(dest64, int64(delta))
		return
	}
}

func (s *statistics) RecordCommand(commandName string) {
	s._commandLock.Lock()
	defer s._commandLock.Unlock()
	s.Command[commandName]++
}

func (s *statistics) Reset() {
	*s = statistics{}
	s.new()
	return
}

func (s *statistics) Save() {
	loggerPrefix := "[statistics-Save]"

	//lock all the locks
	s.lock.Lock()
	s._commandLock.Lock()
	s._userTotalNumLock.Lock()
	defer func() {
		//unlock all the locks
		s.lock.Unlock()
		s._userTotalNumLock.Unlock()
		s._commandLock.Unlock()
	}()

	s.SaveTime = time.Now()

	data, err := json.Marshal(s)
	if err != nil {
		logger.Error.Println(loggerPrefix+"json marshal failed:", err)
		return
	}

	err = storeStatistic(data)
	if err != nil {
		logger.Error.Println(loggerPrefix+"redis set failed:", err)
		return
	}

	return
}

func (s *statistics) PrintToLog() {
	s.lock.Lock()
	defer s.lock.Unlock()
	data, _ := json.Marshal(s)
	logger.Info.Printf(
		"Statistics(%s-%s)：\n%s",
		s.StartTime.Format("2006-01-02 15:04:05"),
		time.Now().Format("2006-01-02 15:04:05"),
		string(data))
	return
}

func (s *statistics) Printf() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	text := "Weekly Active Users [%d]\nHandled Requests [%d]\nHandled Messages:\n\tSticker [%d]\n\tAnimation [%d]\n\tSticker Url [%d]\n\tSticker Set [%d]\nStorage Changed [%d MB]\nCache:\n\tHit [%d]\n\tMiss [%d]\nNetwork:\n\tUploaded [%d MB]\n\tDownloaded [%d MB]\n"

	return fmt.Sprintf(text,
		len(s.UserTotalNum),
		s.MsgHandleTotalTimes,
		s.MsgStickerNum,
		s.MsgAnimationNum,
		s.MsgStickerUrl,
		s.MsgStickerSet,
		s.StorageChange>>20,
		s.CacheHit,
		s.CacheMiss,
		s.NetworkUpload>>20,
		s.NetworkDownload>>20,
	)
}

func storeStatistic(data []byte) error {
	return rdb.Set(context.Background(), fmt.Sprintf("%s:Statistics", ServicePrefix), data, redis.KeepTTL).Err()
}

func getStatistic() string {
	return rdb.Get(context.Background(), fmt.Sprintf("%s:Statistics", ServicePrefix)).Val()
}

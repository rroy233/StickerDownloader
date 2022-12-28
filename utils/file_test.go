package utils

import (
	"github.com/rroy233/logger"
	"testing"
)

func TestUpload(t *testing.T) {
	logger.New(&logger.Config{
		StdOutput: true,
	})
	filePath := "../ffmpeg/ffmpeg-linux-amd64"
	if IsExist(filePath) == false {
		t.Error("文件不存在")
		return
	}
	file := NewUploadFile(filePath)
	err := file.Upload2FileHost()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(file.InfoRes)
}

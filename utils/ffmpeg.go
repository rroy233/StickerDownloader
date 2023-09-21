package utils

import (
	"context"
	"fmt"
	"gopkg.in/rroy233/logger.v2"
	"os/exec"
	"runtime"
	"strings"
)

type logWriter struct{}

func ConvertToGif(ctx context.Context, inFile, outFile string) error {
	cmd := exec.CommandContext(ctx, getFFmpeg(), strings.Split(fmt.Sprintf("-y -i %s -vf scale=-1:-1 -r 20 %s", inFile, outFile), " ")...)
	cmd.Stdout = logWriter{}

	err := cmd.Run()
	if err != nil {
		return err
	}

	return err
}

func (w logWriter) Write(p []byte) (n int, err error) {
	logger.Error.Println("[ffmpeg]" + string(p))
	return len(p), nil
}

func getFFmpeg() string {
	if isSystemFFmpegExist == true {
		if runtime.GOOS == "windows" {
			return "ffmpeg.exe"
		}
		return "ffmpeg"
	}
	return "./ffmpeg/" + getFfmpegFilename()
}

func getFfmpegFilename() string {
	exeSuffix := ""
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}
	return fmt.Sprintf("ffmpeg-%s-%s"+exeSuffix, runtime.GOOS, runtime.GOARCH)
}

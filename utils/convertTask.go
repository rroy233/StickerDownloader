package utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/rroy233/StickerDownloader/config"
	"gopkg.in/rroy233/logger.v2"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type ConvertTask struct {
	//Input file for converting
	InputFilePath string

	//Input file extension
	//support: webp,webm,tgs
	InputExtension string

	//Output file for converting
	OutputFilePath string
}

func (task *ConvertTask) Run(ctx context.Context) error {
	var cmd *exec.Cmd
	if task.InputExtension == "tgs" {
		if config.Get().General.SupportTGSFile == false {
			return errors.New("SupportTGSFile is disabled")
		}
		//tgs gzip decode
		if err := task.tgsDecode(); err != nil {
			return err
		}
		//rename to xxx.tgs.json
		if err := os.Rename(task.InputFilePath, task.InputFilePath+".json"); err != nil {
			return err
		}
		task.InputFilePath = task.InputFilePath + ".json"
		//handle it to rlottie
		cmd = exec.CommandContext(ctx, rlottieExcutablePath, strings.Split(fmt.Sprintf("%s 200x200", task.InputFilePath), " ")...)
		//remember to delete xxx.tgs.json
		defer func() {
			if err := os.Remove(task.InputFilePath); err != nil {
				logger.Warn.Println("failed to remove", task.InputFilePath)
			}
		}()
	} else {
		cmd = exec.CommandContext(ctx, ffmpegExecutablePath, strings.Split(fmt.Sprintf("-y -i %s -vf scale=-1:-1 -r 20 %s", task.InputFilePath, task.OutputFilePath), " ")...)
	}

	//cmd.Stderr = logWriter{}

	err := cmd.Run()
	if err != nil {
		return err
	}

	//postprocessing
	if task.InputExtension == "tgs" {
		//mv to OutputFilePath
		err = os.Rename(task.InputFilePath+".gif", task.OutputFilePath)
		if err != nil {
			return err
		}
	}

	return err
}

type logWriter struct{}

func (w logWriter) Write(p []byte) (n int, err error) {
	logger.Error.Println("[ConvertTask]" + string(p))
	return len(p), nil
}

func getRlottieFilename() string {
	exeSuffix := ""
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}
	return fmt.Sprintf("lottie2gif" + exeSuffix)
}

func getFfmpegFilename(simplify bool) string {
	exeSuffix := ""

	if simplify == false {
		exeSuffix += fmt.Sprintf("-%s-%s", runtime.GOOS, runtime.GOARCH)
	}

	//windows
	if runtime.GOOS == "windows" {
		exeSuffix += ".exe"
	}

	return "ffmpeg" + exeSuffix
}

func (task *ConvertTask) tgsDecode() error {
	file, err := os.OpenFile(task.InputFilePath, os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	r, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	buff := bytes.Buffer{}
	if _, err = buff.ReadFrom(r); err != nil {
		return err
	}

	_ = file.Truncate(0)
	file.Seek(0, 0)
	if _, err = file.Write(buff.Bytes()); err != nil {
		return err
	}

	return nil
}

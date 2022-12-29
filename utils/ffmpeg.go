package utils

import (
	"fmt"
	"github.com/rroy233/logger"
	"os/exec"
	"runtime"
	"strings"
)

type logWriter struct{}

func ConvertToGif(inFile, outFile string) error {
	cmd := exec.Command("./ffmpeg/"+getFfmpeg(), strings.Split(fmt.Sprintf("-y -i %s -vf scale=-1:-1 %s", inFile, outFile), " ")...)
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

func getFfmpeg() string {
	return fmt.Sprintf("ffmpeg-%s-%s", runtime.GOOS, runtime.GOARCH)
}

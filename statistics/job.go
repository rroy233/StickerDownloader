package statistics

import (
	"time"
)

func autoReset() {
	for {
		time.Sleep(Statistics.EndTime.Sub(time.Now()))
		Statistics.PrintToLog()
		Statistics.Reset()
	}
}

func autoSave() {
	for {
		time.Sleep(3 * time.Minute)
		Statistics.Save()
	}
}

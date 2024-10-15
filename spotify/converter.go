package spotify

import (
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"os"
	"os/exec"
)

var hasFFmpeg bool

func init() {
	_, err := exec.LookPath("ffmpeg")
	if err == nil {
		hasFFmpeg = true
	}
}

func (d *Downloader) convertMp3(inputFile string, outputFile string) (err error) {
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf(`input file [%s] not exists`, inputFile)
	}

	_, err = exec.LookPath("ffmpeg")
	if err != nil {
		log.Warnln("ffmpeg not found, skip converting format")
		return
	}

	log.Debugf("Converting [%s] to [%s]", inputFile, outputFile)

	bitrate := ""
	switch d.quality {
	case Quality96Vorbis:
		bitrate = "96"
	case Quality128MP4, Quality128MP4Dual:
		bitrate = "128"
	case Quality160Vorbis:
		bitrate = "160"
	case Quality256MP4, Quality256MP4Dual:
		bitrate = "256"
	case Quality320Vorbis:
		bitrate = "320"
	}
	log.Debugf("Set convertor bitrate: %sk", bitrate)

	err = ffmpeg.Input(inputFile).
		Output(outputFile, ffmpeg.KwArgs{
			"format":        "mp3",
			"audio_bitrate": bitrate + "k",
			//"acodec": "libmp3lame",
		}).
		OverWriteOutput().Silent(true).Run()
	if err != nil {
		return fmt.Errorf("error while converting to mp3: %v", err)
	}
	log.Debugln("Convert successfully")

	return
}

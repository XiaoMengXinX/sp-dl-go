package main

import (
	"flag"
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"github.com/XiaoMengXinX/sp-dl-go/spotify"
	"os"
)

func init() {
	log.Init(log.LevelInfo)
}

func main() {
	id := flag.String("id", "", "Spotify URL/URI/ID (required). Example usage: -id https://open.spotify.com/track/4jTrKMoc44RYZsoFsIlQev")
	quality := flag.String("quality", spotify.Quality128MP4Dual, "Quality level. Options: MP4_128, MP4_128_DUAL, MP4_256, MP4_256_DUAL, OGG_VORBIS_320, OGG_VORBIS_160, OGG_VORBIS_96")
	output := flag.String("output", "./output", "Output path (default: ./output). Specify where the downloaded files should be saved.")
	debug := flag.Bool("debug", false, "Print debug information. Use this to enable more detailed logging for troubleshooting.")

	flag.Parse()

	if *id == "" {
		fmt.Println("Error: -id is required")
		flag.Usage()
		os.Exit(1)
	}

	if *debug {
		log.Init(log.LevelDebug)
	}

	log.Infof("Set Output path: %s", *output)
	log.Infof("Set Quality level: %s", *quality)

	sp := spotify.NewDownloader().Initialize()
	sp.OutputFolder = *output
	sp.Quality = *quality
	err := sp.Download(*id)
	if err != nil {
		log.Fatalln(err)
	}
}

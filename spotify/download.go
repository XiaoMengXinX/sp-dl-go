package spotify

import (
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"github.com/XiaoMengXinX/sp-dl-go/playplay"
	widevine "github.com/iyear/gowidevine"
	"os"
)

func (d *Downloader) downloadContent(ID string, content IDType) (err error) {
	var name, artist, fileID, format string
	var metadata trackMetadata
	switch {
	case mp4FormatSet[d.quality]:
		format = "mp3"
	case oggFormatSet[d.quality]:
		format = "ogg"
	default:
		return fmt.Errorf("invalid quality: %s", d.quality)
	}
	switch content {
	case TRACK:
		name, artist, fileID, metadata, err = d.getTrackMetadata(ID)
		if err != nil {
			defer func(ID string, err *error) {
				if *err != nil {
					log.Errorf("Failed to get metadata of trackID [%s]: %v", ID, (*err).Error())
				}
			}(ID, &err)
			return fmt.Errorf("failed to get metadata of trackID [%s]: %v", ID, err)
		}
	case EPISODE:
		name, artist, fileID, err = d.getEpisodeMetadata(ID)
		if err != nil {
			defer func(ID string, err *error) {
				if *err != nil {
					log.Errorf("Failed to get metadata of episodeID [%s]: %v", ID, (*err).Error())
				}
			}(ID, &err)
			return fmt.Errorf("failed to get metadata of episodeID [%s]: %v", ID, err)
		}
	default:
		return fmt.Errorf("invalid content type")
	}
	fileName := cleanFilename(fmt.Sprintf("%s - %s", name, artist))
	tmpFileName := fmt.Sprintf("%s.%s.tmp", fileName, format)
	tmpFilePath := fmt.Sprintf("%s/%s", d.OutputFolder, tmpFileName)
	outFilePath := fmt.Sprintf("%s/%s.%s", d.OutputFolder, fileName, format)

	log.Infof("Downloading %s [%s]", content, fileName)

	defer func(filename string, filePath string, err *error) {
		if *err != nil {
			_ = os.Remove(filePath)
			log.Errorf("Failed to download [%s]: %v", filename, (*err).Error())
		}
	}(fileName, outFilePath, &err)

	cdnURL, err := d.requestCDNURL(fileID)
	if err != nil {
		return err
	}

	err = d.downloadURL(cdnURL, tmpFileName)
	defer os.Remove(tmpFilePath)
	if err != nil {
		return err
	}

	//TODO: Download cover for track
	_ = metadata

	tmpFile, err := os.OpenFile(tmpFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	outFile, err := os.Create(outFilePath) // 使用 os.Create 创建文件
	if err != nil {
		return err
	}
	defer outFile.Close()

	switch format {
	case "mp3":
		PSSH, err := requestPSSH(fileID)
		if err != nil {
			return err
		}
		log.Debugf("Request PSSH for [%s] successfully: %s", ID, PSSH)

		keys, err := d.getMp4Keys(PSSH)
		if err != nil {
			return fmt.Errorf("get decrypt key failed: %v", err)
		}
		log.Debugf("Get decrypt key for [%s] successfully", ID)

		err = widevine.DecryptMP4Auto(tmpFile, keys, outFile)
		if err != nil {
			return fmt.Errorf("failed to decrypt file: %v", err)
		}
	case "ogg":
		key, err := d.getOggKeys(fileID)
		if err != nil {
			return fmt.Errorf("get decrypt key failed: %v", err)
		}
		err = playplay.DecryptFileStream(key[:], tmpFile, outFile)
		if err != nil {
			return fmt.Errorf("failed to decrypt file: %v", err)
		}
	}

	//TODO: Add id3v2 for track

	log.Infof("Download %s [%s] successfully", content, fileName)
	return nil
}

func (d *Downloader) DownloadTrack(ID string) (err error) {
	return d.downloadContent(ID, TRACK)
}

func (d *Downloader) DownloadEpisode(ID string) (err error) {
	return d.downloadContent(ID, EPISODE)
}

func (d *Downloader) Download(url string) (err error) {
	tracks, err := d.GetTracks(url)
	if err != nil {
		return fmt.Errorf("failed to get tracks: %v", err)
	}
	if len(tracks) == 0 {
		log.Info("No tracks to download")
		return nil
	}
	log.Infof("Downloading %d track(s)", len(tracks))
	_, idType, _ := getIDType(url)
	for _, track := range tracks {
		switch idType {
		case TRACK:
			_ = d.DownloadTrack(track)
		case SHOW:
			_ = d.DownloadEpisode(track)
		}
	}
	return nil
}

package spotify

import (
	"bytes"
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"github.com/XiaoMengXinX/sp-dl-go/playplay"
	widevine "github.com/iyear/gowidevine"
	"io"
	"os"
)

func (d *Downloader) downloadContent(ID string, content IDType) (err error) {
	var name, artist, fileID, format string
	var metadata trackMetadata
	switch {
	case mp4FormatSet[d.Quality]:
		format = "mp3"
	case oggFormatSet[d.Quality]:
		format = "ogg"
	default:
		return fmt.Errorf("invalid quality: %s", d.Quality)
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
	filename := cleanFilename(fmt.Sprintf("%s - %s", name, artist))
	outFile := fmt.Sprintf("%s/%s.%s", d.OutputFolder, filename, format)

	log.Infof("Downloading %s [%s]", content, filename)

	defer func(filename string, outFile string, err *error) {
		if *err != nil {
			_ = os.Remove(outFile)
			log.Errorf("Failed to download [%s]: %v", filename, (*err).Error())
		}
	}(filename, outFile, &err)

	cdnURL, err := d.requestCDNURL(fileID)
	if err != nil {
		return err
	}

	err = downloadURL(cdnURL, d.OutputFolder, filename+"."+format)
	if err != nil {
		return err
	}

	//TODO: Download cover for track
	_ = metadata

	file, err := os.OpenFile(outFile, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, file)
	if err != nil {
		return err
	}
	decryptedBuf := new(bytes.Buffer)

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

		err = widevine.DecryptMP4Auto(buf, keys, decryptedBuf)
		if err != nil {
			return fmt.Errorf("failed to decrypt file: %v", err)
		}
	case "ogg":
		key, err := d.getOggKeys(fileID)
		if err != nil {
			return fmt.Errorf("get decrypt key failed: %v", err)
		}
		err = playplay.DecryptFileStream(key[:], buf, decryptedBuf)
		if err != nil {
			return fmt.Errorf("failed to decrypt file: %v", err)
		}
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}
	err = file.Truncate(0)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, decryptedBuf)
	if err != nil {
		return err
	}

	//TODO: Add id3v2 for track

	log.Infof("Download %s [%s] successfully", content, filename)
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

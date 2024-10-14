package spotify

import (
	"bytes"
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	widevine "github.com/iyear/gowidevine"
	"io"
	"os"
)

func (d *Downloader) DownloadTrack(ID string) (err error) {
	name, artist, fileID, err := d.getTrackMetadata(ID)
	if err != nil {
		return fmt.Errorf("failed to get metadata of trackID [%s]: %v", ID, err)
	}
	filename := cleanFilename(fmt.Sprintf("%s - %s", name, artist))
	outFile := fmt.Sprintf("%s/%s.mp3", d.OutputFolder, filename)

	defer func(filename string, err *error) {
		if *err != nil {
			log.Errorf("Failed to download [%s]: %v", filename, (*err).Error())
		}
	}(filename, &err)

	PSSH, err := requestPSSH(fileID)
	if err != nil {
		return err
	}
	log.Debugf("Request PSSH for [%s] successfully: %s", ID, PSSH)

	keys, err := d.getKeys(PSSH)
	log.Debugf("Get decrypt key for [%s] successfully", ID)

	cdnURL, err := d.requestCDNURL(fileID)
	if err != nil {
		return err
	}

	err = downloadURL(cdnURL, d.OutputFolder, filename+".mp3")
	if err != nil {
		return err
	}

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

	err = widevine.DecryptMP4Auto(buf, keys, decryptedBuf)
	if err != nil {
		return fmt.Errorf("failed to decrypt [%s]: %v", outFile, err)
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

	log.Infof("File [%s] downloaded successfully", filename)
	return nil
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
	for _, track := range tracks {
		_ = d.DownloadTrack(track)
	}
	return nil
}

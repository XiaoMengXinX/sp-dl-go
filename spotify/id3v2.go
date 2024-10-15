package spotify

import (
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"github.com/bogem/id3v2"
	"net/http"
	"os"
)

func (d *Downloader) AddMetadata(metadata trackMetadata, filePath string) (err error) {
	f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file [%s]: %v", filePath, err)
	}

	coverFileName, err := d.downloadCoverImage(metadata)
	defer os.Remove(d.OutputFolder + "/" + coverFileName)
	if err != nil {
		log.Warnf("Failed to download cover image: %v, skip adding front cover", err)
	} else {
		coverFile, err := os.ReadFile(d.OutputFolder + "/" + coverFileName)
		if err != nil {
			log.Warnf("Failed to read cover file: %v, skip adding front cover", err)
		}
		return addMp3Id3v2(f, coverFile, metadata)
	}
	return addMp3Id3v2(f, nil, metadata)
}

func addMp3Id3v2(musicFile *os.File, picFile []byte, metadata trackMetadata) (err error) {
	musicTag, _ := id3v2.ParseReader(musicFile, id3v2.Options{Parse: true})
	defer musicTag.Close()
	musicTag.SetDefaultEncoding(id3v2.EncodingUTF8)
	musicTag.SetTitle(metadata.Name)
	musicTag.SetArtist(formatArtistsStr(metadata))
	if metadata.Album.Name != "" {
		musicTag.SetAlbum(metadata.Album.Name)
	}
	if len(picFile) > 32 {
		mime := http.DetectContentType(picFile[:32])
		pic := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingISO,
			MimeType:    mime,
			PictureType: id3v2.PTFrontCover,
			Description: "Front cover",
			Picture:     picFile,
		}
		musicTag.AddAttachedPicture(pic)
	}
	if err := musicTag.Save(); err != nil {
		return fmt.Errorf("failed to save id3v2: %v ", err)
	}
	return nil
}

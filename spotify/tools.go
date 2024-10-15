package spotify

import (
	"encoding/hex"
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func buildQueryParams(params map[string]interface{}) string {
	values := url.Values{}

	for key, value := range params {
		switch v := value.(type) {
		case string:
			values.Add(key, v)
		case int, int64:
			values.Add(key, fmt.Sprintf("%d", v))
		case float64:
			values.Add(key, strconv.FormatFloat(v, 'f', -1, 64))
		case bool:
			values.Add(key, strconv.FormatBool(v))
		default:
			// 如果是其他类型，将其转换为字符串
			values.Add(key, fmt.Sprintf("%v", v))
		}
	}

	return values.Encode()
}

func IDToHex(spotifyID string) string {
	if len(spotifyID) != 22 {
		return ""
	}

	idBytes := []byte(spotifyID)
	alphabet := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	dictionary := make([]byte, 256)
	for i := 0; i < 62; i++ {
		dictionary[alphabet[i]] = byte(i & 0xff)
	}

	big := make([]byte, 22)
	for i := 0; i < 22; i++ {
		big[i] = dictionary[idBytes[i]]
	}

	var out []byte
	for len(big) > 0 {
		var quotient []byte
		remainder := 0

		for _, b := range big {
			accumulator := int(b) + remainder*62
			digit := (accumulator - (accumulator % 256)) / 256
			remainder = accumulator % 256

			if len(quotient) > 0 || digit > 0 {
				quotient = append(quotient, byte(digit))
			}
		}

		out = append(out, byte(remainder))
		big = quotient
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}

	return hex.EncodeToString(out)
}

func getAllFiles(metadata trackMetadata) []fileEntry {
	if len(metadata.File) > 0 {
		return metadata.File
	}
	if len(metadata.AltFile) > 0 {
		return metadata.AltFile[0].File
	}
	return nil
}

func checkDirExist(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create folder %q: %w", dirPath, err)
		}
	}
	return nil
}

func cleanFilename(filename string) string {
	osType := os.Getenv("GOOS")

	var illegalChars *regexp.Regexp
	switch osType {
	case "darwin": // Mac
		illegalChars = regexp.MustCompile(`[:\x00-\x1F]`)
	case "linux":
		illegalChars = regexp.MustCompile(`[\x00/]`)
	default:
		illegalChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	}

	cleaned := illegalChars.ReplaceAllString(filename, "")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "." || cleaned == ".." || cleaned == "" {
		return "null"
	}

	if len(cleaned) > 250 {
		cleaned = cleaned[:250]
	}

	return cleaned
}

func (e *fileEntry) testFileIDOrFileId() string {
	if e.FileID != "" {
		return e.FileID
	}
	return e.FileId
}

func (d *Downloader) isSupportedFormat(format string) bool {
	switch {
	case mp4FormatSet[d.quality]:
		return mp4FormatSet[format]
	case oggFormatSet[d.quality]:
		return oggFormatSet[format]
	default:
		return false
	}
}

func (d *Downloader) selectFromQuality(entries []fileEntry) (string, error) {
	for _, entry := range entries {
		if entry.Format == d.quality {
			return entry.testFileIDOrFileId(), nil
		}
	}
	log.Warn("Unable to find desired quality. Falling back to best.")

	for _, entry := range entries {
		if d.isSupportedFormat(entry.Format) {
			log.Debugf("Selected new format: %s", entry.Format)
			d.quality = entry.Format
			return entry.testFileIDOrFileId(), nil
		}
	}
	return "", fmt.Errorf("no valid audio format found")
}

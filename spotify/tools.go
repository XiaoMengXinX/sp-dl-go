package spotify

import (
	"encoding/hex"
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"math/big"
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
			values.Add(key, fmt.Sprintf("%v", v))
		}
	}

	return values.Encode()
}

func SpIDToHex(spotifyID string) string {
	base62Charset := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	dictionary := make(map[byte]int64, len(base62Charset))
	for i := 0; i < len(base62Charset); i++ {
		dictionary[base62Charset[i]] = int64(i)
	}

	base := big.NewInt(62)
	result := big.NewInt(0)
	for i := 0; i < len(spotifyID); i++ {
		value := big.NewInt(dictionary[spotifyID[i]])
		result.Mul(result, base).Add(result, value)
	}
	return fmt.Sprintf("%032s", hex.EncodeToString(result.Bytes()))
}

func SpHexToID(hexStr string) string {
	base62Charset := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	hexBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return ""
	}
	num := new(big.Int).SetBytes(hexBytes)
	if num.Cmp(big.NewInt(0)) == 0 {
		return string(base62Charset[0])
	}

	base := big.NewInt(62)
	result := ""
	for num.Cmp(big.NewInt(0)) > 0 {
		remainder := new(big.Int)
		num.DivMod(num, base, remainder)
		result = string(base62Charset[remainder.Int64()]) + result
	}
	return fmt.Sprintf("%022s", result)
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

func formatArtistsStr(artists []artistDataBasic) string {
	if len(artists) == 0 {
		return ""
	}

	artistNames := make([]string, len(artists))
	for i, ar := range artists {
		artistNames[i] = ar.Name
	}

	return strings.Join(artistNames, ", ")
}

func cleanFilename(filename string) string {
	osType := os.Getenv("GOOS")

	var illegalChars *regexp.Regexp
	switch osType {
	case "darwin":
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
	case mp4FormatSet[d.preferQuality]:
		return mp4FormatSet[format]
	case oggFormatSet[d.preferQuality]:
		return oggFormatSet[format]
	default:
		return false
	}
}

func (d *Downloader) selectFromQuality(entries []fileEntry) (string, error) {
	for _, entry := range entries {
		if entry.Format == d.preferQuality {
			d.quality = entry.Format
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

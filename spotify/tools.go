package spotify

import (
	"encoding/hex"
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func downloadURL(url, path, filename string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("DNT", "1")
	req.Header.Add("Origin", "https://open.spotify.com")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Referer", "https://open.spotify.com/")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("download failed with http code: %d", resp.StatusCode)
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	file, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	buffer := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %v", writeErr)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read response body: %v", err)
		}
	}
	return nil
}

// buildQueryParams 将 map 转换为 URL 参数字符串
func buildQueryParams(params map[string]interface{}) string {
	values := url.Values{}

	// 遍历 map，将每个键值对添加到 URL 参数中
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

// IDToHex 将 Spotify ID 转换为十六进制字符串。
func IDToHex(spotifyID string) string {
	// 检查 Spotify ID 长度是否为 22
	if len(spotifyID) != 22 {
		return ""
	}

	idBytes := []byte(spotifyID)
	alphabet := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	// 初始化字典
	dictionary := make([]byte, 256)
	for i := 0; i < 62; i++ {
		dictionary[alphabet[i]] = byte(i & 0xff)
	}

	// 转换 Spotify ID 为 big 数组
	big := make([]byte, 22)
	for i := 0; i < 22; i++ {
		big[i] = dictionary[idBytes[i]]
	}

	// 计算输出的字节数组
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

	// 反转输出的字节数组
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}

	// 将字节数组转换为十六进制字符串并返回
	return hex.EncodeToString(out)
}

// getAllFiles 获取所有可用文件
func getAllFiles(metadata trackMetadata) []fileEntry {
	if len(metadata.File) > 0 {
		return metadata.File
	}
	if len(metadata.AltFile) > 0 {
		return metadata.AltFile[0].File
	}
	return nil
}

// checkDirExist 检查文件夹是否存在，不存在则创建
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
	// 获取当前系统信息
	osType := os.Getenv("GOOS")

	// 根据系统定义非法字符的正则表达式
	var illegalChars *regexp.Regexp
	switch osType {
	case "darwin": // Mac
		illegalChars = regexp.MustCompile(`[:\x00-\x1F]`)
	case "linux":
		illegalChars = regexp.MustCompile(`[\x00/]`)
	default:
		illegalChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	}

	// 去除非法字符并修剪前后空格
	cleaned := illegalChars.ReplaceAllString(filename, "")
	cleaned = strings.TrimSpace(cleaned)

	// 如果文件名为空或为特殊名称，使用默认名称
	if cleaned == "." || cleaned == ".." || cleaned == "" {
		return "default_filename"
	}

	if len(cleaned) > 250 {
		cleaned = cleaned[:250]
	}

	return cleaned
}

// testFileIDOrFileId 获取文件ID，兼容两种命名
func (e *fileEntry) testFileIDOrFileId() string {
	if e.FileID != "" {
		return e.FileID
	}
	return e.FileId
}

// isSupportedFormat 检查是否为支持的格式
func (d *Downloader) isSupportedFormat(format string) bool {
	switch d.Quality {
	case Quality128MP4, Quality256MP4:
		for _, f := range Mp4Formats {
			if f == format {
				return true
			}
		}
	case Quality96Vorbis, Quality160Vorbis, Quality320Vorbis:
		for _, f := range OggFormats {
			if f == format {
				return true
			}
		}
	}
	return false
}

// selectFromQuality 根据质量选择音频文件
func (d *Downloader) selectFromQuality(entries []fileEntry) (string, error) {
	for _, entry := range entries {
		if entry.Format == d.Quality {
			return entry.testFileIDOrFileId(), nil
		}
	}
	log.Warn("Unable to find desired quality. Falling back to best.")

	for _, entry := range entries {
		if d.isSupportedFormat(entry.Format) {
			return entry.testFileIDOrFileId(), nil
		}
	}
	return "", fmt.Errorf("no valid audio format found")
}

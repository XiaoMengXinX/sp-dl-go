package spotify

import (
	"encoding/json"
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var cdmData []byte

// readCDMs 读取 CDM 文件
func readCDMs() []string {
	cdms, err := filepath.Glob(filepath.Join("cdm", "*.wvd"))
	if err != nil || len(cdms) == 0 {
		log.Fatal(`No CDMs found in "./cdm" folder`) // 使用自定义日志系统的Fatal函数
	}
	cdmData, err = os.ReadFile(cdms[0])
	if err != nil {
		log.Fatal("Failed to read CDM file")
	}
	return cdms
}

// requestClientBases 获取 Spotify 下载 cdn 链接
func requestClientBases() []string {
	resp, err := http.Get("https://apresolve.spotify.com?type=spclient")
	if err != nil {
		log.Errorf("Unable to request client bases: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Unable to request client bases (%d): %s", resp.StatusCode, body)
		return nil
	}

	// 定义 JSON 响应结构
	var response struct {
		SpClient []string `json:"spclient"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Errorf("Error decoding client bases response: %v", err)
		return nil
	}

	// 使用 map 转换 endpoints
	var formattedEndpoints []string
	for _, endpoint := range response.SpClient {
		formatted := formatEndpoint(endpoint)
		if formatted != "" {
			formattedEndpoints = append(formattedEndpoints, formatted)
		}
	}

	return formattedEndpoints
}

// buildLicenseURL 构建许可证 URL
func buildLicenseURL(clientBases []string) string {
	if len(clientBases) == 0 {
		log.Warn("No client bases available to build license URL")
		return ""
	}
	return fmt.Sprintf("%s/widevine-license/v1/audio/license", clientBases[0])
}

// 格式化端点地址
func formatEndpoint(endpoint string) string {
	parts := strings.Split(endpoint, ":")
	if len(parts) != 2 {
		log.Warnf("Invalid endpoint format: %s", endpoint)
		return ""
	}
	domain, port := parts[0], parts[1]

	switch port {
	case "80":
		return fmt.Sprintf("http://%s", domain)
	case "443":
		return fmt.Sprintf("https://%s", domain)
	default:
		log.Warnf("Unknown port: %s", port)
		return ""
	}
}

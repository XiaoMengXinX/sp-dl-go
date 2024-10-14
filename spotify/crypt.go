package spotify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	widevine "github.com/iyear/gowidevine"
	"github.com/iyear/gowidevine/widevinepb"
	"net/http"
)

// requestPSSH 请求 PSSH
func requestPSSH(fildID string) (pssh string, err error) {
	url := fmt.Sprintf("https://seektables.scdn.co/seektable/%s.json", fildID)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("faied to request PSSH: %v", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("faied to request PSSH with status [%d]", resp.StatusCode)
	}

	// 解析响应体中的 JSON 数据
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("faied to request PSSH: %v", err)
	}

	// 检查并返回 PSSH 字段
	pssh, ok := result["pssh"].(string)
	if !ok {
		log.Debugf("Failed to find PSSH: %+v", result)
		return "", fmt.Errorf("PSSH not found")
	}

	return pssh, nil
}

// getKeys 获取解密 key
func (d *Downloader) getKeys(psshStr string) ([]*widevine.Key, error) {
	device, err := widevine.NewDevice(
		widevine.FromWVD(bytes.NewReader(cdmData)),
	)
	if err != nil {
		return nil, fmt.Errorf("create device: %w", err)
	}
	cdm := widevine.NewCDM(device)

	psshByte, _ := base64.StdEncoding.DecodeString(psshStr)
	pssh, err := widevine.NewPSSH(psshByte)
	if err != nil {
		return nil, fmt.Errorf("parse pssh: %w", err)
	}

	challenge, parseLicense, err := cdm.GetLicenseChallenge(pssh, widevinepb.LicenseType_AUTOMATIC, false)
	if err != nil {
		return nil, fmt.Errorf("get license challenge: %w", err)
	}

	license, err := d.makeRequest("POST", d.licenseURL, challenge)

	if err != nil {
		return nil, fmt.Errorf("request license: %w", err)
	}

	keys, err := parseLicense(license)
	if err != nil {
		return nil, fmt.Errorf("parse license: %w", err)
	}

	return keys, nil
}

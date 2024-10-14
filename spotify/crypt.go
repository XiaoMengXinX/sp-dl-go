package spotify

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/XiaoMengXinX/sp-dl-go/config"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"github.com/XiaoMengXinX/sp-dl-go/playplay"
	widevine "github.com/iyear/gowidevine"
	"github.com/iyear/gowidevine/widevinepb"
	"google.golang.org/protobuf/proto"
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

// getMp4Keys 获取 MP4 解密 key
func (d *Downloader) getMp4Keys(psshStr string) ([]*widevine.Key, error) {
	device, err := widevine.NewDevice(
		widevine.FromWVD(bytes.NewReader(cdmData)),
	)
	if err != nil {
		return nil, fmt.Errorf("create device failed: %w", err)
	}
	cdm := widevine.NewCDM(device)

	psshByte, _ := base64.StdEncoding.DecodeString(psshStr)
	pssh, err := widevine.NewPSSH(psshByte)
	if err != nil {
		return nil, fmt.Errorf("parse pssh failed: %w", err)
	}

	challenge, parseLicense, err := cdm.GetLicenseChallenge(pssh, widevinepb.LicenseType_AUTOMATIC, false)
	if err != nil {
		return nil, fmt.Errorf("get license challenge failed: %w", err)
	}

	license, err := d.makeRequest("POST", d.licenseURL, challenge)

	if err != nil {
		return nil, fmt.Errorf("request license failed: %w", err)
	}

	keys, err := parseLicense(license)
	if err != nil {
		return nil, fmt.Errorf("parse license failed: %w", err)
	}

	return keys, nil
}

func (d *Downloader) getOggKeys(fileID string) (key [16]byte, err error) {
	protoInt32 := func(i int32) *int32 { return &i }
	protoInteractivity := func(i playplay.Interactivity) *playplay.Interactivity { return &i }
	protoContentType := func(c playplay.ContentType) *playplay.ContentType { return &c }

	ppToken := config.CM.Get().PlayPlayToken
	if ppToken == "" {
		ppToken = config.CM.GetDefault().PlayPlayToken
	}
	reqToken, _ := hex.DecodeString(ppToken)
	req := &playplay.PlayPlayLicenseRequest{
		Version:       protoInt32(2),
		Token:         reqToken,
		Interactivity: protoInteractivity(playplay.Interactivity_DOWNLOAD),
		ContentType:   protoContentType(playplay.ContentType_AUDIO_TRACK),
	}
	body, err := proto.Marshal(req)
	if err != nil {
		return key, fmt.Errorf("serialize request failed: %v", err)
	}

	url := fmt.Sprintf("https://spclient.wg.spotify.com/playplay/v1/key/%s", fileID)
	resp, err := d.makeRequest("POST", url, body)
	if err != nil {
		return key, fmt.Errorf("request license failed: %w", err)
	}

	var playplayResponse playplay.PlayPlayLicenseResponse
	if err = proto.Unmarshal(resp, &playplayResponse); err != nil {
		return key, fmt.Errorf("parse response failrd: %v", err)
	}

	hexFileID, _ := hex.DecodeString(fileID)
	obfuscatedKey := [16]byte(playplayResponse.GetObfuscatedKey()[:])

	fmt.Printf("ObfuscatedKey: %x\n", obfuscatedKey)
	fmt.Printf("hexFileID: %x\n", hexFileID)

	key = playplay.PlayPlayDecrypt(obfuscatedKey, [20]byte(hexFileID[:]))

	fmt.Printf("key: %x\n", key)
	return key, nil
}

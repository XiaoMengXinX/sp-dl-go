package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/XiaoMengXinX/sp-dl-go/logger"
	"io"
	"net/http"
	"time"

	"github.com/XiaoMengXinX/sp-dl-go/token"
)

const (
	Quality128MP4    = "MP4_128_DUAL"
	Quality256MP4    = "MP4_256_DUAL"
	Quality96Vorbis  = "OGG_VORBIS_96"
	Quality160Vorbis = "OGG_VORBIS_160"
	Quality320Vorbis = "OGG_VORBIS_320"
)

var (
	Mp4Formats = []string{Quality128MP4, Quality256MP4}
	OggFormats = []string{Quality96Vorbis, Quality160Vorbis, Quality320Vorbis}
)

// Downloader 结构体定义
type Downloader struct {
	TokenManager *token.Manager
	OutputFolder string
	Quality      string

	clientBases []string
	licenseURL  string
}

// NewDownloader 创建 Downloader 实例
func NewDownloader() *Downloader {
	return &Downloader{
		TokenManager: token.NewTokenManager(),
		Quality:      Quality128MP4,
		OutputFolder: "./output",
	}
}

func (d *Downloader) Initialize() *Downloader {
	d.TokenManager.QuerySpDc()
	d.clientBases = requestClientBases()
	d.licenseURL = buildLicenseURL(d.clientBases)
	_ = readCDMs()
	if err := checkDirExist(d.OutputFolder); err != nil {
		log.Fatalln(err)
	}
	return d
}

// GetTracks 获取专辑、播放列表或单曲中的所有音轨
func (d *Downloader) GetTracks(url string) ([]string, error) {
	url, idType, err := getIDType(url)
	if err != nil {
		log.Debugf("Get IDType Failed: %v", err)
		return nil, err
	}
	switch idType {
	case ALBUM:
		return d.fetchAlbumTracks(url, 0, []string{})
	case PLAYLIST:
		return d.fetchPlaylistTracks(url, 0, []string{})
	case SHOW:
		return d.fetchShowEpisodes(url, 0, []string{})
	default:
		return []string{url}, nil
	}
}

// makeRequest 执行 HTTP 请求 (GET 或 POST) 并返回响应数据
func (d *Downloader) makeRequest(method, url string, body []byte) ([]byte, error) {
	accessToken, _ := d.TokenManager.GetAccessToken()
	var requestBody io.Reader
	if body != nil {
		requestBody = bytes.NewBuffer(body)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	log.Debugf("Using [%s] to request: %s", method, url)
	log.Debugf("Request Headers: %+v", req.Header)

	// 执行请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 记录请求结果
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request to [%s] failed with status [%d]", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// fetchAlbumTracks 获取专辑中的所有音轨
func (d *Downloader) fetchAlbumTracks(albumID string, offset int, tracks []string) ([]string, error) {
	url := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks?offset=%d&limit=50", albumID, offset)
	data, err := d.makeRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Debugf("Fetch Album tracks Failed: %v", err)
		return nil, err
	}

	var albumData albumData
	if err := json.Unmarshal(data, &albumData); err != nil {
		return nil, fmt.Errorf("failed to decode album data: %w", err)
	}

	for _, item := range albumData.Items {
		tracks = append(tracks, item.Id)
	}

	if len(albumData.Items) >= 50 {
		return d.fetchAlbumTracks(albumID, offset+50, tracks)
	}
	return tracks, nil
}

// fetchPlaylistTracks 获取播放列表中的所有音轨
func (d *Downloader) fetchPlaylistTracks(playlistID string, offset int, tracks []string) ([]string, error) {
	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?offset=%d&limit=100", playlistID, offset)
	data, err := d.makeRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Debugf("Fetch Playlist tracks Failed: %v", err)
		return nil, err
	}

	var playlistData playlistData
	if err := json.Unmarshal(data, &playlistData); err != nil {
		return nil, fmt.Errorf("failed to decode playlist data: %w", err)
	}

	for _, item := range playlistData.Items {
		if item.Track.Id != "" {
			tracks = append(tracks, item.Track.Id)
		}
	}

	if len(playlistData.Items) >= 100 {
		return d.fetchPlaylistTracks(playlistID, offset+100, tracks)
	}
	return tracks, nil
}

// fetchShowEpisodes 获取播客中的所有集数
func (d *Downloader) fetchShowEpisodes(showID string, offset int, episodes []string) ([]string, error) {
	url := fmt.Sprintf("https://api.spotify.com/v1/shows/%s/episodes?offset=%d&limit=50", showID, offset)
	data, err := d.makeRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Debugf("Fetch Show episodes Failed: %v", err)
		return nil, err
	}

	var showData showData
	if err := json.Unmarshal(data, &showData); err != nil {
		return nil, fmt.Errorf("failed to decode show data: %w", err)
	}

	for _, item := range showData.Items {
		episodes = append(episodes, item.Id)
	}

	if len(showData.Items) >= 50 {
		return d.fetchShowEpisodes(showID, offset+50, episodes)
	}
	return episodes, nil
}

// getTrackMetadata 获取音轨的元数据
func (d *Downloader) getTrackMetadata(trackID string) (name string, artist string, fileID string, metadata trackMetadata, err error) {
	url := fmt.Sprintf("https://spclient.wg.spotify.com/metadata/4/track/%s", IDToHex(trackID))
	resp, err := d.makeRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Debugf("Fetch track metadata Failed: %v", err)
		return "", "", "", metadata, err
	}

	if err := json.Unmarshal(resp, &metadata); err != nil {
		return "", "", "", metadata, fmt.Errorf("failed to decode track metadata: %w", err)
	}

	if len(metadata.Artist) != 0 {
		artist = metadata.Artist[0].Name
	}
	fileID, err = d.selectFromQuality(getAllFiles(metadata))
	if err != nil {
		return "", "", "", metadata, err
	}

	return metadata.Name, artist, fileID, metadata, nil
}

// getEpisodeMetadata 获取播客集数的元数据
func (d *Downloader) getEpisodeMetadata(episodeID string) (name string, creator string, fileID string, err error) {
	url := "https://api-partner.spotify.com/pathfinder/v1/query"
	var paramsVar []byte
	paramsVar, _ = json.Marshal(map[string]string{
		"uri": fmt.Sprintf("spotify:episode:%s", episodeID),
	})
	var paramsExtensions []byte
	paramsExtensions, _ = json.Marshal(map[string]interface{}{
		"persistedQuery": map[string]interface{}{
			"version":    1,
			"sha256Hash": "9697538fe993af785c10725a40bb9265a20b998ccd2383bd6f586e01303824e9",
		},
	})
	params := map[string]interface{}{
		"operationName": "getEpisodeOrChapter",
		"variables":     string(paramsVar),
		"extensions":    string(paramsExtensions),
	}
	resp, err := d.makeRequest(http.MethodGet, url+"?"+buildQueryParams(params), nil)
	if err != nil {
		log.Debugf("Fetch episode metadata Failed: %v", err)
		return "", "", "", err
	}

	var metadata episodeMetadata
	if err := json.Unmarshal(resp, &metadata); err != nil {
		return "", "", "", fmt.Errorf("failed to decode episode metadata: %w", err)
	}

	episode := metadata.Data.Episode
	fileID, err = d.selectFromQuality(episode.Audio.Items)
	if err != nil {
		return "", "", "", err
	}

	if episode.Creator == "" {
		episode.Creator = episode.Podcast.Data.Name
	}

	return episode.Name, episode.Creator, fileID, err
}

// requestCDNURL 请求 Spotify CDN 获取音频文件 URL
func (d *Downloader) requestCDNURL(fileID string) (string, error) {
	url := fmt.Sprintf("https://gew4-spclient.spotify.com/storage-resolve/files/audio/interactive/%s", fileID)
	params := buildQueryParams(map[string]interface{}{"alt": "json"})

	respBody, err := d.makeRequest("GET", url+"?"+params, nil)
	if err != nil {
		log.Debugf("Fetch CDN URL Failed: %v", err)
		return "", err
	}

	// 解析 JSON 响应为 cdnURL 结构体
	var cdnResponse cdnURL
	_ = json.Unmarshal(respBody, &cdnResponse)

	// 检查并返回第一个 CDN URL
	if len(cdnResponse.Cdnurl) == 0 {
		return "", fmt.Errorf("no CDN URL found in response")
	}
	log.Debugf("Get CDN URL successfully: %v", cdnResponse.Cdnurl)

	return cdnResponse.Cdnurl[0], nil
}

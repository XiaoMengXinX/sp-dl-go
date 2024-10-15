package spotify

import (
	"errors"
	"net/url"
	"strings"
)

type IDType string

const (
	TRACK    IDType = "track"
	ALBUM    IDType = "album"
	PLAYLIST IDType = "playlist"
	SHOW     IDType = "show"
	EPISODE  IDType = "episode"
)

func getIDType(urlID string) (string, IDType, error) {
	if strings.HasPrefix(urlID, "http") {
		parsedURL, err := url.Parse(urlID)
		if err != nil {
			return "", "", err
		}

		if parsedURL.Host != "open.spotify.com" {
			return "", "", errors.New("invalid domain")
		}

		pathSegments := strings.Split(parsedURL.Path, "/")
		if len(pathSegments) < 3 {
			return "", "", errors.New("invalid URL path")
		}

		return pathSegments[2], IDType(pathSegments[1]), nil
	} else if strings.HasPrefix(urlID, "spotify:") {
		split := strings.Split(urlID, ":")
		if len(split) < 3 {
			return "", "", errors.New("invalid URI format")
		}
		return split[2], IDType(split[1]), nil
	}
	if urlID == "" {
		return "", "", errors.New("invalid URL or ID")
	}
	return urlID, TRACK, nil
}

package spotify

type albumTracksData struct {
	Items []struct {
		Id string `json:"id"`
	} `json:"items"`
	Total  int    `json:"total"`
	Offset int    `json:"offset"`
	Next   string `json:"next"`
}

type playlistTracksData struct {
	Items []struct {
		Track struct {
			Id string `json:"id"`
		} `json:"track"`
	} `json:"items"`
	Total  int    `json:"total"`
	Offset int    `json:"offset"`
	Next   string `json:"next"`
}

type showTracksData struct {
	Items []struct {
		Id string `json:"id"`
	} `json:"items"`
	Total  int    `json:"total"`
	Offset int    `json:"offset"`
	Next   string `json:"next"`
}

type albumData struct {
	Type                 string            `json:"album_type"`
	TotalTracks          int               `json:"total_tracks"`
	Name                 string            `json:"name"`
	ReleaseDate          string            `json:"release_date"`
	ReleaseDatePrecision string            `json:"release_date_precision"`
	Artists              []artistDataBasic `json:"artists"`
	Copyrights           []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"copyrights"`
	ExternalIds struct {
		ISRC string `json:"isrc"`
		EAN  string `json:"ean"`
		UPC  string `json:"upc"`
	} `json:"external_ids"`
	Genres []string `json:"genres"`
	Label  string   `json:"label"`
}

type trackData struct {
	Album struct {
		Type                 string            `json:"album_type"`
		TotalTracks          int               `json:"total_tracks"`
		ID                   string            `json:"id"`
		Name                 string            `json:"name"`
		ReleaseDate          string            `json:"release_date"`
		ReleaseDatePrecision string            `json:"release_date_precision"`
		Artists              []artistDataBasic `json:"artists"`
	} `json:"album"`
	Artists     []artistDataBasic `json:"artists"`
	DurationMs  int               `json:"duration_ms"`
	ExternalIDs struct {
		ISRC string `json:"isrc"`
		EAN  string `json:"ean"`
		UPC  string `json:"upc"`
	} `json:"external_ids"`
	Name        string `json:"name"`
	TrackNumber int    `json:"track_number"`
}

type trackMetadata struct {
	GID   string `json:"gid"`
	Name  string `json:"name"`
	Album struct {
		Name       string `json:"name"`
		CoverGroup struct {
			Image []struct {
				FileId string `json:"file_id"`
				Size   string `json:"size"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"image"`
		} `json:"cover_group"`
	} `json:"album"`
	Artists []artistDataBasic `json:"artist"`
	File    []fileEntry       `json:"file"`
	AltFile []struct {
		File []fileEntry `json:"file"`
	} `json:"alternative,omitempty"`
	CanonicalURI string `json:"canonical_uri"`
}

type episodeMetadata struct {
	Data struct {
		Episode struct {
			Name    string `json:"name"`
			Creator string `json:"creator"`
			Audio   struct {
				Items []fileEntry `json:"items"`
			} `json:"audio"`
			Podcast struct {
				Data struct {
					Name string `json:"name"`
				} `json:"data"`
			} `json:"podcastV2"`
		} `json:"episodeUnionV2"`
	} `json:"data"`
}

type artistDataBasic struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type cdnURL struct {
	Result string   `json:"result"`
	CdnURL []string `json:"cdnurl"`
	Fileid string   `json:"fileid"`
	Ttl    int      `json:"ttl"`
}

type fileEntry struct {
	Format string `json:"format"`
	FileID string `json:"file_id,omitempty"`
	FileId string `json:"fileId,omitempty"`
}

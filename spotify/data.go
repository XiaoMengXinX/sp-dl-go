package spotify

type albumData struct {
	Items []struct {
		Id string `json:"id"` // 音轨 ID
	} `json:"items"`
	Total  int    `json:"total"`  // 总音轨数
	Offset int    `json:"offset"` // 当前偏移量
	Next   string `json:"next"`   // 下一页链接
}

type playlistData struct {
	Items []struct {
		Track struct {
			Id string `json:"id"` // 音轨 ID
		} `json:"track"`
	} `json:"items"`
	Total  int    `json:"total"`  // 总音轨数
	Offset int    `json:"offset"` // 当前偏移量
	Next   string `json:"next"`   // 下一页链接
}

type showData struct {
	Items []struct {
		Id string `json:"id"` // 集数 ID
	} `json:"items"`
	Total  int    `json:"total"`  // 总集数
	Offset int    `json:"offset"` // 当前偏移量
	Next   string `json:"next"`   // 下一页链接
}

type trackMetadata struct {
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
	Artist []struct {
		Name string `json:"name"`
	} `json:"artist"`
	File    []fileEntry `json:"file"`
	AltFile []struct {
		File []fileEntry `json:"file"`
	} `json:"alternative,omitempty"`
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

type cdnURL struct {
	Result string   `json:"result"`
	Cdnurl []string `json:"cdnurl"`
	Fileid string   `json:"fileid"`
	Ttl    int      `json:"ttl"`
}

type fileEntry struct {
	Format string `json:"format"`
	FileID string `json:"file_id,omitempty"`
	FileId string `json:"fileId,omitempty"`
}

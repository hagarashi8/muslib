package mis

import (
	"encoding/json"
	"museff/internal/common"
	"net/http"
	"net/url"
)

type SongInfoDTO struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type MusicInfoService struct {
	host *url.URL
}

func NewMusicInfoService(host string) *MusicInfoService {
	hst := common.Must(url.Parse(host))
	return &MusicInfoService{host: hst}
}

func (mis *MusicInfoService) GetMusicInfo(group string, song string) (si SongInfoDTO, err error) {
	si = SongInfoDTO{}
	u := *mis.host.JoinPath("/info")

	q := make(url.Values)
	q.Add("group", group)
	q.Add("song", song)
	u.RawQuery = q.Encode()

	r, err := http.Get(u.String())
	if err != nil {
		return
	}

	jw := json.NewDecoder(r.Body)
	err = jw.Decode(&si)
	return
}

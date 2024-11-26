package common

import (
	"time"
)

type SongInfo struct {
	ID          int
	Song        string    `example:"Angel Song"`
	Group       string    `example:"Nothing More"`
	ReleaseDate time.Time `example:"2024-05-17"`
	Text        string    `example:"(Hey, hey, hey)\n(Hey, hey, hey)\nHey!\n\nYou can't deny you got that feeling in your bones\nReady to go\nAre you ready to go?"`
	Link        string    `example:"https://www.youtube.com/watch?v=WLJ9b6HIMHw"`
} // @name Song

type SongInfoDTO struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return val
}

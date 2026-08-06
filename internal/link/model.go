package link

import (
	"go-api/internal/database"
	"math/rand"
)

func NewLink(url string) *database.Link {
	return &database.Link{
		Url:  url,
		Hash: RandStringRunes(10),
	}
}

func RandStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

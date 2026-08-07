package link

import "errors"

var (
	ErrInvalidURL    = errors.New("invalid URL")
	ErrLinkNotFound  = errors.New("link not found")
	ErrHashCollision = errors.New("link hash collision")
)

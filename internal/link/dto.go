package link

type CreateLinkRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type UpdateLinkRequest struct {
	URL  string `json:"url" validate:"required,url"`
	Hash string `json:"hash"`
}

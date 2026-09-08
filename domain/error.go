package domain

import "errors"

var (
	ErrInvalidArticle = errors.New("invalid article")
	ErrAuthorNotFound = errors.New("author not found")
)

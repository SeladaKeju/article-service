package domain

import "errors"

var (
	// ErrInvalidArticle indicates that required article data is invalid.
	ErrInvalidArticle = errors.New("invalid article")
	// ErrAuthorNotFound indicates that an article references no existing author.
	ErrAuthorNotFound = errors.New("author not found")
	// ErrInvalidLimit indicates an out-of-range or non-numeric page limit.
	ErrInvalidLimit = errors.New("invalid limit")
	// ErrInvalidCursor indicates a malformed pagination cursor.
	ErrInvalidCursor = errors.New("invalid cursor")
)

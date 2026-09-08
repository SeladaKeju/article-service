package domain

import "time"

type Author struct {
	ID   string
	Name string
}

type Article struct {
	ID        string
	AuthorID  string
	Title     string
	Body      string
	CreatedAt time.Time
}

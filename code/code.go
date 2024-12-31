package code

import "time"

type CodeClient interface {
	FetchSubmissions() []Submission
}

type Question struct {
	Id              string
	Title           string
	TitleSlug       string
	LastSubmittedAt time.Time
}

type Submission struct {
	Id              string
	Title           string
	TitleSlug       string
	LastSubmittedAt time.Time
	Lang            string
	Code            string
}

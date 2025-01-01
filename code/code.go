package code

import "time"

const SubmissionFetchingError = "error while fetching submissions"
const QuestionFetchingError = "error while fetching submissions"

type CodeClient interface {
	FetchSubmissions() ([]Submission, error)
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

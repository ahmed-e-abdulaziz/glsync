// This package is responsible for fetching the submissions of any code challenges site
// It is currently implemented by [leetcode.go]
package code

import "time"

const SubmissionFetchingError = "error while fetching submissions"
const QuestionFetchingError = "error while fetching questions"

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

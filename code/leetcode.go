package code

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/config"
)

//go:embed leetcode-graphql/submission-details-query.json
var submissionDetailsQuery string

//go:embed leetcode-graphql/submission-list-query.json
var submissionListQuery string

//go:embed leetcode-graphql/user-progress-question-list-query.json
var userProgressQuestionListQuery string

type leetcode struct {
	cfg        config.Config
	graphqlUrl string
}

func NewLeetCode(cfg config.Config, leetcodeGraphqlUrl string) leetcode {
	return leetcode{cfg, leetcodeGraphqlUrl}
}

func (lc leetcode) FetchSubmissions() ([]Submission, error) {
	questions, err := lc.fetchQuestions()
	if err != nil {
		return nil, errors.New(QuestionFetchingError)
	}
	submissions := make([]Submission, len(questions))
	for idx, question := range questions {
		lcSubmission, err := lc.fetchSubmissionOverview(question.TitleSlug)
		if err != nil {
			return nil, errors.New(SubmissionFetchingError)
		}
		code, err := lc.fetchSubmissionCode(lcSubmission.Id)
		if err != nil {
			return nil, errors.New(SubmissionFetchingError)
		}
		submissions[idx] = Submission{question.FrontendId, question.Title, question.TitleSlug, question.LastSubmittedAt, lcSubmission.Lang, code}
	}
	return submissions, nil
}

func (lc leetcode) fetchQuestions() ([]LCQuestion, error) {
	bodyBytes, err := lc.queryLeetcode(userProgressQuestionListQuery)
	if err != nil {
		log.Println(err)
		return nil, errors.New("encountered an error while fetching user questions from leetcode")
	}
	body := &RequestBody[lcUserProgressQuestionListData]{}
	json.Unmarshal(bodyBytes, body)
	
	return body.Data.QuestionsList.Questions, nil
}

func (lc leetcode) fetchSubmissionOverview(titleSlug string) (lcSumbissionOverview, error) {
	bodyBytes, err := lc.queryLeetcode(fmt.Sprintf(submissionListQuery, titleSlug))
	if err != nil {
		return lcSumbissionOverview{}, errors.New("encountered an error while fetching submssion overview from leetcode")
	}
	body := &RequestBody[lcSubmissionListData]{}
	json.Unmarshal(bodyBytes, body)
	return body.Data.LCSubmissionList.LCSubmissions[0], nil // we only need the lastest submission
}

func (lc leetcode) fetchSubmissionCode(id string) (string, error) {
	bodyBytes, err := lc.queryLeetcode(fmt.Sprintf(submissionDetailsQuery, id))
	if err != nil {
		log.Println(err)
		return "", errors.New("encountered an error while fetching submssion code from leetcode")
	}
	body := &RequestBody[lcSubmissionDetailsData]{}
	json.Unmarshal(bodyBytes, body)
	return body.Data.Details.Code, nil
}

func (lc leetcode) queryLeetcode(query string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, lc.graphqlUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, err
	}
	cookie := &http.Cookie{
		Name:     "LEETCODE_SESSION",
		Value:    lc.cfg.LCookie,
		Path:     "/",
		Domain:   ".leetcode.com",
		HttpOnly: true,
		MaxAge:   1209600,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}
	req.AddCookie(cookie)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, _ := io.ReadAll(res.Body)
	return bodyBytes, nil
}

type RequestBody[T any] struct {
	Data T `json:"data"`
}

type lcUserProgressQuestionListData struct {
	QuestionsList lcUserProgressQuestionList `json:"userProgressQuestionList"`
}

type lcUserProgressQuestionList struct {
	Questions []LCQuestion `json:"questions"`
}

type LCQuestion struct {
	FrontendId      string    `json:"frontendId"`
	Title           string    `json:"title"`
	TitleSlug       string    `json:"titleSlug"`
	LastSubmittedAt time.Time `json:"lastSubmittedAt"`
	QuestionStatus  string    `json:"questionStatus"`
	LastResult      string    `json:"lastResult"`
}

type lcSubmissionListData struct {
	LCSubmissionList lcSubmissionList `json:"questionSubmissionList"`
}
type lcSubmissionList struct {
	LCSubmissions []lcSumbissionOverview `json:"submissions"`
}

type lcSumbissionOverview struct {
	Id   string `json:"id"`
	Lang string `json:"lang"`
}

type lcSubmissionDetailsData struct {
	Details lcSubmissionDetails `json:"submissionDetails"`
}

type lcSubmissionDetails struct {
	Code string `json:"code"`
}

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

	"github.com/ahmed-e-abdulaziz/glsync/config"
)

//go:embed leetcode-graphql/submission-details-query.json
var submissionDetailsQuery string

//go:embed leetcode-graphql/submission-list-query.json
var submissionListQuery string

//go:embed leetcode-graphql/user-progress-question-list-query.json
var userProgressQuestionListQuery string

// Implementation of CodeClient for LeetCode
type leetcode struct {
	cfg        config.Config
	graphqlUrl string
}

func NewLeetCode(cfg config.Config, leetcodeGraphqlUrl string) leetcode {
	return leetcode{cfg, leetcodeGraphqlUrl}
}

// Fetches sumbissions from LeetCode
//
// Requires cfg.LcCookie to be set correctly or will fail due to access errors
// Returns an array of [Submission] struct
func (lc leetcode) FetchSubmissions() ([]Submission, error) {
	log.Println("\n==============\nFetching submissions next")
	questions, err := lc.fetchQuestions()
	if err != nil {
		return nil, errors.New(QuestionFetchingError)
	}
	log.Printf("User has %v questions accepted on LeetCode, fetching code for each next \n", len(questions))
	submissions := make([]Submission, len(questions))
	for idx, question := range questions {
		log.Printf("\tFetching latest submission for question: %v %v\n", question.FrontendId, question.Title)
		lcSubmission, err := lc.fetchSubmissionOverview(question.TitleSlug)
		if err != nil {
			log.Println(err.Error())
			return nil, errors.New(SubmissionFetchingError)
		}
		code, err := lc.fetchSubmissionCode(lcSubmission.Id)
		if err != nil {
			log.Println(err.Error())
			return nil, errors.New(SubmissionFetchingError)
		}
		submissions[idx] = Submission{question.FrontendId, question.Title, question.TitleSlug, question.LastSubmittedAt, lcSubmission.Lang, code}
	}
	log.Print("Fetched submissions successfully\n==============\n")
	return submissions, nil
}

// Fetches question to extract required info for Submission struct
// Uses LC's GraphQl query that's called userProgressQuestionList
func (lc leetcode) fetchQuestions() ([]lcQuestion, error) {
	bodyBytes, _, err := lc.queryLeetcode(userProgressQuestionListQuery)
	if err != nil {
		log.Println(err)
		return nil, errors.New("encountered an error while fetching user questions from leetcode")
	}
	body := &RequestBody[lcUserProgressQuestionListData]{}
	err = json.Unmarshal(bodyBytes, body)
	if err != nil {
		log.Println(err)
		return nil, errors.New("encountered an error while parsing user questions response from leetcode")
	}
	return body.Data.QuestionsList.Questions, nil
}

// Fetches id and language of submission into lcSubmissionOverview struct
// Uses LC's GraphQl query that's called submissionList
//
// titleSlug is a no-whitespace representation of the question title, used to query submissions for a question
// Returns an error if it encounters one while querying and an nil lcSumbissionOverview
func (lc leetcode) fetchSubmissionOverview(titleSlug string) (lcSumbissionOverview, error) {
	bodyBytes, _, err := lc.queryLeetcode(fmt.Sprintf(submissionListQuery, titleSlug))
	if err != nil {
		return lcSumbissionOverview{}, errors.New("encountered an error while fetching submssion overview from leetcode")
	}
	body := &RequestBody[lcSubmissionListData]{}
	err = json.Unmarshal(bodyBytes, body)
	if err != nil {
		log.Println(err)
		return lcSumbissionOverview{}, errors.New("encountered an error while parsing submssion overview from leetcode")
	}
	if len(body.Data.LCSubmissionList.LCSubmissions) == 0 {
		return lcSumbissionOverview{}, errors.New("couldn't fetch any submissions for question with title slug: " + titleSlug)
	}
	return body.Data.LCSubmissionList.LCSubmissions[0], nil // we only need the lastest submission
}

// Fetches submission's code using the leetcode's submission id
// Uses LC's GraphQl query that's called submissionDetails
//
// id is usually a string of numbers fetched earlier by fetchSubmissionOverview
// Returns an error if it encounters one while querying and an empty string
func (lc leetcode) fetchSubmissionCode(id string) (string, error) {
	bodyBytes, headers, err := lc.queryLeetcode(fmt.Sprintf(submissionDetailsQuery, id))
	if err != nil {
		log.Println(err)
		return "", errors.New("encountered an error while fetching submssion code from leetcode")
	}
	body := &RequestBody[lcSubmissionDetailsData]{}
	err = json.Unmarshal(bodyBytes, body)
	if err != nil {
		log.Println(err)
		return "", errors.New("encountered an error while parsing submssion code from leetcode")
	}
	if len(body.Data.Details.Code) == 0 {
		log.Println("Recieved the following response body:\n" + string(bodyBytes) + "\n")
		log.Printf("Recieved the following response headers:\n%v\n", headers)
		return "", errors.New("couldn't fetch the code submissions with id: " + id)
	}
	return body.Data.Details.Code, nil
}

// queryLeetcode sends the query string to leetcode's GraphQL URL (https://leetcode.com/graphql)
//
// On success it returns the resulting bytes of the response body and a nil error
// Otherwise it will return nil and any error it faces while creating the request or while communicating with LC
func (lc leetcode) queryLeetcode(query string) ([]byte, map[string][]string, error) {
	req, err := http.NewRequest(http.MethodPost, lc.graphqlUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, nil, err
	}
	lc.addCookieAndHeaders(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	bodyBytes, _ := io.ReadAll(res.Body)
	return bodyBytes, res.Header, nil
}

// Adds cfg.LcCookie cookie and necessary headers to req
func (lc leetcode) addCookieAndHeaders(req *http.Request) {
	cookie := &http.Cookie{
		Name:     "LEETCODE_SESSION",
		Value:    lc.cfg.LcCookie,
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
}

type RequestBody[T any] struct {
	Data T `json:"data"`
}

type lcUserProgressQuestionListData struct {
	QuestionsList lcUserProgressQuestionList `json:"userProgressQuestionList"`
}

type lcUserProgressQuestionList struct {
	Questions []lcQuestion `json:"questions"`
}

type lcQuestion struct {
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

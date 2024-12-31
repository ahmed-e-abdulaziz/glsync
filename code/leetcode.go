package code

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/config"
)

type leetcode struct {
	cfg config.Config
}

func NewLeetCode(cfg config.Config) leetcode {
	return leetcode{cfg}
}

func (lc leetcode) FetchSubmissions() []Submission {
	questions := lc.fetchQuestions()
	submissions := make([]Submission, len(questions))
	for idx, question := range questions {
		lcSubmission := lc.fetchSubmissionOverview(question.TitleSlug)
		code := lc.fetchSubmissionCode(lcSubmission.Id)
		submissions[idx] = Submission{question.FrontendId, question.Title, question.TitleSlug, question.LastSubmittedAt, lcSubmission.Lang, code}
	}
	return submissions
}

func (lc leetcode) fetchQuestions() []LCQuestion {
	query := `{
		"query": "\n    query userProgressQuestionList($filters: UserProgressQuestionListInput) {\n  userProgressQuestionList(filters: $filters) {\n    questions {\n      frontendId\n      title\n      titleSlug\n      lastSubmittedAt\n      questionStatus\n      lastResult\n    }\n  }\n}\n    ",
		"variables": {
			"filters": {
				"questionStatus": "SOLVED",
				"skip": 0,
				"limit": 4000
			}
		},
		"operationName": "userProgressQuestionList"
	}`
	bodyBytes, err := lc.queryLeetcode(query)
	if err != nil {
		log.Fatal("Encountered an error while fetching user questions from leetcode", err)
	}
	body := &RequestBody[lcUserProgressQuestionListData]{}
	json.Unmarshal(bodyBytes, body)
	return body.Data.QuestionsList.Questions
}

func (lc leetcode) fetchSubmissionOverview(titleSlug string) lcSumbissionOverview {
	query := fmt.Sprintf(`{
		"query": "\n    query submissionList($offset: Int!, $limit: Int!, $lastKey: String, $questionSlug: String!, $lang: Int, $status: Int) {\n  questionSubmissionList(\n    offset: $offset\n    limit: $limit\n    lastKey: $lastKey\n    questionSlug: $questionSlug\n    lang: $lang\n    status: $status\n  ) {\n    lastKey\n    hasNext\n    submissions {\n      id\n      title\n      titleSlug\n      status\n      statusDisplay\n      lang\n      langName\n      runtime\n      timestamp\n      url\n      isPending\n      memory\n      hasNotes\n      notes\n      flagType\n      frontendId\n      topicTags {\n        id\n      }\n    }\n  }\n}\n    ",
		"variables": {
			"questionSlug": "%v",
			"offset": 0,
			"limit": 1,
			"lastKey": null
		},
		"operationName": "submissionList"
	}`, titleSlug)

	bodyBytes, err := lc.queryLeetcode(query)
	if err != nil {
		log.Fatal("Encountered an error while fetching submssion overview from leetcode", err)
	}
	body := &RequestBody[lcSubmissionListData]{}
	json.Unmarshal(bodyBytes, body)
	return body.Data.LCSubmissionList.LCSubmissions[0] // we only need the lastest submission
}

func (lc leetcode) fetchSubmissionCode(id string) string {
	query := fmt.Sprintf(`{
			"query": "\n    query submissionDetails($submissionId: Int!) {\n  submissionDetails(submissionId: $submissionId) {\n    runtime\n    runtimeDisplay\n    runtimePercentile\n    runtimeDistribution\n    memory\n    memoryDisplay\n    memoryPercentile\n    memoryDistribution\n    code\n    timestamp\n    statusCode\n    user {\n      username\n      profile {\n        realName\n        userAvatar\n      }\n    }\n    lang {\n      name\n      verboseName\n    }\n    question {\n      questionId\n      titleSlug\n      hasFrontendPreview\n    }\n    notes\n    flagType\n    topicTags {\n      tagId\n      slug\n      name\n    }\n    runtimeError\n    compileError\n    lastTestcase\n    codeOutput\n    expectedOutput\n    totalCorrect\n    totalTestcases\n    fullCodeOutput\n    testDescriptions\n    testBodies\n    testInfo\n    stdOutput\n  }\n}\n    ",
			"variables": {
				"submissionId": %v
			},
			"operationName": "submissionDetails"
		}`, id)
	bodyBytes, err := lc.queryLeetcode(query)
	if err != nil {
		log.Fatal("Encountered an error while fetching submssion code from leetcode", err)
	}
	body := &RequestBody[lcSubmissionDetailsData]{}
	json.Unmarshal(bodyBytes, body)
	return body.Data.Details.Code
}

func (lc leetcode) queryLeetcode(query string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, "https://leetcode.com/graphql/", bytes.NewBuffer([]byte(query)))
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
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
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

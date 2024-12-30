package leetcode

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

type LeetCode struct {
	cfg config.Config
}

func NewLeetCode(cfg config.Config) LeetCode {
	return LeetCode{cfg}
}

func (lc LeetCode) FetchQuestions() []Question {
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
	bodyBytes, err := lc.queryGraphql(query)
	if err != nil {
		log.Fatal("Encountered an error while fetching user questions", err)
	}
	body := &RequestBody[UserProgressQuestionListData]{}
	json.Unmarshal(bodyBytes, body)
	return body.Data.QuestionsList.Questions
}

func (lc LeetCode) FetchSubmissionOverview(titleSlug string) SumbissionOverview {
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

	bodyBytes, err := lc.queryGraphql(query)
	if err != nil {
		log.Fatal("Encountered an error while fetching submssion overview", err)
	}
	body := &RequestBody[SubmissionListData]{}
	json.Unmarshal(bodyBytes, body)
	return body.Data.SubmissionList.Submissions[0] // we only need the lastest submission
}

func (lc LeetCode) FetchSubmissionCode(id string) string {
	query := fmt.Sprintf(`{
			"query": "\n    query submissionDetails($submissionId: Int!) {\n  submissionDetails(submissionId: $submissionId) {\n    runtime\n    runtimeDisplay\n    runtimePercentile\n    runtimeDistribution\n    memory\n    memoryDisplay\n    memoryPercentile\n    memoryDistribution\n    code\n    timestamp\n    statusCode\n    user {\n      username\n      profile {\n        realName\n        userAvatar\n      }\n    }\n    lang {\n      name\n      verboseName\n    }\n    question {\n      questionId\n      titleSlug\n      hasFrontendPreview\n    }\n    notes\n    flagType\n    topicTags {\n      tagId\n      slug\n      name\n    }\n    runtimeError\n    compileError\n    lastTestcase\n    codeOutput\n    expectedOutput\n    totalCorrect\n    totalTestcases\n    fullCodeOutput\n    testDescriptions\n    testBodies\n    testInfo\n    stdOutput\n  }\n}\n    ",
			"variables": {
				"submissionId": %v
			},
			"operationName": "submissionDetails"
		}`, id)
	bodyBytes, err := lc.queryGraphql(query)
	if err != nil {
		log.Fatal("Encountered an error while fetching submssion overview", err)
	}
	body := &RequestBody[SubmissionDetailsData]{}
	json.Unmarshal(bodyBytes, body)
	return body.Data.Details.Code
}

func (lc LeetCode) queryGraphql(query string) ([]byte, error) {
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

type UserProgressQuestionListData struct {
	QuestionsList UserProgressQuestionList `json:"userProgressQuestionList"`
}

type UserProgressQuestionList struct {
	Questions []Question `json:"questions"`
}

type Question struct {
	FrontendId      string    `json:"frontendId"`
	Title           string    `json:"title"`
	TitleSlug       string    `json:"titleSlug"`
	LastSubmittedAt time.Time `json:"lastSubmittedAt"`
	QuestionStatus  string    `json:"questionStatus"`
	LastResult      string    `json:"lastResult"`
}

type SubmissionListData struct {
	SubmissionList SubmissionList `json:"questionSubmissionList"`
}
type SubmissionList struct {
	Submissions []SumbissionOverview `json:"submissions"`
}

type SumbissionOverview struct {
	Id   string `json:"id"`
	Lang string `json:"lang"`
}

type SubmissionDetailsData struct {
	Details SubmissionDetails `json:"submissionDetails"`
}

type SubmissionDetails struct {
	Code string `json:"code"`
}

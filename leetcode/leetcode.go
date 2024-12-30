package leetcode

import (
	"bytes"
	"encoding/json"
	"io"
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
	bodyBytes, _ := lc.queryGraphql(query)
	body := &RequestBody[UserProgressQuestionListData]{}
	json.Unmarshal(bodyBytes, body)
	return body.Data.QuestionsList.Questions
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

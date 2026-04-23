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
	"strings"
	"time"

	"github.com/ahmed-e-abdulaziz/glsync/config"
)

//go:embed leetcode-graphql/submission-details-query.json
var submissionDetailsQuery string

//go:embed leetcode-graphql/submission-list-query.json
var submissionListQuery string

//go:embed leetcode-graphql/submission-list-cn-query.json
var submissionListQueryCN string

//go:embed leetcode-graphql/submission-detail-cn-query.json
var submissionDetailQueryCN string

//go:embed leetcode-graphql/user-progress-question-list-query.json
var userProgressQuestionListQuery string

const (
	maxRetry         = 25               // LeetCode API can fail A LOT :( It requires a ton of retries when it fails
	backoffTime      = 1 * time.Second  // 1 second to avoid keep using LeetCode API when it fails
	// empirically the cn rate-limit cooldown is ~510s (observed: cleared after 17x30s waits).
	// We wait 520s once to clear it cleanly rather than retrying 17 times in 30s increments.
	rateLimitBackoff = 520 * time.Second
)

// Implementation of CodeClient for LeetCode
type leetcode struct {
	cfg          config.Config
	graphqlUrl   string
	cookieDomain string // e.g. ".leetcode.com" or ".leetcode.cn"
	siteOrigin   string // e.g. "https://leetcode.com" or "https://leetcode.cn"
}

func NewLeetCode(cfg config.Config, leetcodeGraphqlUrl string) leetcode {
	cookieDomain := ".leetcode.com"
	siteOrigin := "https://leetcode.com"
	if strings.Contains(leetcodeGraphqlUrl, "leetcode.cn") {
		cookieDomain = ".leetcode.cn"
		siteOrigin = "https://leetcode.cn"
	}
	return leetcode{cfg, leetcodeGraphqlUrl, cookieDomain, siteOrigin}
}

// Fetches submissions from LeetCode
//
// Requires cfg.LcCookie to be set correctly or will fail due to access errors
// Returns an array of [Submission] struct
func (lc leetcode) FetchSubmissions() ([]Submission, error) {
	log.Println("\n==============\nFetching submissions next")
	questions, err := lc.fetchQuestions()
	if err != nil {
		log.Printf("Error fetching questions: %v\n", err)
		return nil, errors.New("failed to fetch questions from LeetCode")
	}

	log.Printf("User has %v questions accepted on LeetCode, fetching code for each next\n", len(questions))
	submissions := make([]Submission, 0, len(questions)) // Changed to 0 initial length

	for _, question := range questions {
		log.Printf("\tFetching latest submission for question: %v %v\n", question.FrontendId, question.Title)
		submission, err := lc.fetchQuestionSubmission(question)
		if err != nil {
			log.Printf("Warning: Failed to fetch submission for question %s: %v\n", question.Title, err)
			continue // Skip this submission but continue with others
		}
		submissions = append(submissions, submission)
	}

	if len(submissions) == 0 {
		return nil, errors.New("failed to fetch any submissions successfully")
	}

	log.Printf("Fetched %d/%d submissions successfully\n==============\n", len(submissions), len(questions))
	return submissions, nil
}

// cnRequestDelay throttles submission detail fetches on leetcode.cn.
// Measured: 10-minute sliding window, quota of 60 requests (1 req/10s).
// Each HTTP round-trip takes ~1s, so a 9s sleep gives ~10s total cycle,
// staying at the safe boundary without wasting extra time.
// Total run time: ~560 * 10s = ~95 minutes for 560 questions.
const cnRequestDelay = 9 * time.Second

// New helper function to handle single question submission
func (lc leetcode) fetchQuestionSubmission(question lcQuestion) (Submission, error) {
	lcSubmission, err := lc.fetchSubmissionOverview(question.TitleSlug)
	if err != nil {
		log.Printf("Error fetching question submissions: %v\n", err)
		return Submission{}, errors.New("submission overview error")
	}

	// Throttle requests on CN to avoid triggering the rate limiter.
	if lc.cookieDomain == ".leetcode.cn" {
		time.Sleep(cnRequestDelay)
	}

	code, err := lc.fetchSubmissionCode(lcSubmission.Id, 0)
	if err != nil {
		log.Printf("Error fetching submission code: %v\n", err)
		return Submission{}, errors.New("submission code error")
	}

	return Submission{
		question.FrontendId,
		question.Title,
		question.TitleSlug,
		question.LastSubmittedAt,
		lcSubmission.Lang,
		code,
	}, nil
}

// Fetches question to extract required info for Submission struct
// Uses LC's GraphQl query that's called userProgressQuestionList
func (lc leetcode) fetchQuestions() ([]lcQuestion, error) {
	bodyBytes, err := lc.queryLeetcode(userProgressQuestionListQuery)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("error fetching user questions from leetcode: %w", err)
	}
	body := &RequestBody[lcUserProgressQuestionListData]{}
	err = json.Unmarshal(bodyBytes, body)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("error parsing user questions response from leetcode: %w", err)
	}
	return body.Data.QuestionsList.Questions, nil
}

// Fetches id and language of submission into lcSubmissionOverview struct
// Uses LC's GraphQl query that's called submissionList
//
// titleSlug is a no-whitespace representation of the question title, used to query submissions for a question
// Returns an error if it encounters one while querying and an nil lcSumbissionOverview
func (lc leetcode) fetchSubmissionOverview(titleSlug string) (lcSumbissionOverview, error) {
	var (
		bodyBytes   []byte
		err         error
		submissions []lcSumbissionOverview
	)

	if lc.cookieDomain == ".leetcode.cn" {
		// leetcode.cn uses "submissionList" field; leetcode.com uses "questionSubmissionList"
		bodyBytes, err = lc.queryLeetcode(fmt.Sprintf(submissionListQueryCN, titleSlug))
		if err != nil {
			return lcSumbissionOverview{}, fmt.Errorf("error fetching submission overview from leetcode: %w", err)
		}
		body := &RequestBody[lcSubmissionListDataCN]{}
		if err = json.Unmarshal(bodyBytes, body); err != nil {
			log.Println(err)
			return lcSumbissionOverview{}, fmt.Errorf("error parsing submission overview from leetcode: %w", err)
		}
		submissions = body.Data.LCSubmissionList.LCSubmissions
	} else {
		bodyBytes, err = lc.queryLeetcode(fmt.Sprintf(submissionListQuery, titleSlug))
		if err != nil {
			return lcSumbissionOverview{}, fmt.Errorf("error fetching submission overview from leetcode: %w", err)
		}
		body := &RequestBody[lcSubmissionListData]{}
		if err = json.Unmarshal(bodyBytes, body); err != nil {
			log.Println(err)
			return lcSumbissionOverview{}, fmt.Errorf("error parsing submission overview from leetcode: %w", err)
		}
		submissions = body.Data.LCSubmissionList.LCSubmissions
	}

	if len(submissions) == 0 {
		return lcSumbissionOverview{}, fmt.Errorf("no submissions found for question: %s", titleSlug)
	}
	return submissions[0], nil // we only need the latest accepted submission
}

// Fetches submission's code using the leetcode's submission id.
// On leetcode.cn uses submissionDetail (singular); on leetcode.com uses submissionDetails (plural).
// Returns an empty string and an error if it encounters one while querying.
func (lc leetcode) fetchSubmissionCode(id string, retry int) (string, error) {
	if lc.cookieDomain == ".leetcode.cn" {
		return lc.fetchSubmissionCodeCN(id, retry)
	}
	return lc.fetchSubmissionCodeCOM(id, retry)
}

func (lc leetcode) fetchSubmissionCodeCOM(id string, retry int) (string, error) {
	bodyBytes, err := lc.queryLeetcode(fmt.Sprintf(submissionDetailsQuery, id))
	if err != nil {
		if retry < maxRetry {
			log.Printf("Network error, retry %d/%d after %v\n", retry+1, maxRetry, backoffTime)
			time.Sleep(backoffTime)
			return lc.fetchSubmissionCodeCOM(id, retry+1)
		}
		return "", fmt.Errorf("max retries reached for network error: %w", err)
	}

	body := &RequestBody[lcSubmissionDetailsData]{}
	if err := json.Unmarshal(bodyBytes, body); err != nil {
		return "", fmt.Errorf("JSON parsing error: %w", err)
	}

	if body.Data.Details == nil {
		if retry < maxRetry {
			log.Printf("Null response, retry %d/%d after %v\n", retry+1, maxRetry, backoffTime)
			time.Sleep(backoffTime)
			return lc.fetchSubmissionCodeCOM(id, retry+1)
		}
		log.Printf("Warning: Max retries reached, consistently getting null response for submission %s", id)
		return "", fmt.Errorf("max retries reached for null response%s", id)
	}

	if len(body.Data.Details.Code) == 0 {
		if retry < maxRetry {
			log.Printf("Empty code, retry %d/%d after %v\n", retry+1, maxRetry, backoffTime)
			time.Sleep(backoffTime)
			return lc.fetchSubmissionCodeCOM(id, retry+1)
		}
		log.Printf("Warning: Max retries reached with empty code for submission %s", id)
		return "", fmt.Errorf("max retries reached for empty code")
	}

	return body.Data.Details.Code, nil
}

func (lc leetcode) fetchSubmissionCodeCN(id string, retry int) (string, error) {
	bodyBytes, err := lc.queryLeetcode(fmt.Sprintf(submissionDetailQueryCN, id))
	if err != nil {
		if retry < maxRetry {
			log.Printf("Network error, retry %d/%d after %v\n", retry+1, maxRetry, backoffTime)
			time.Sleep(backoffTime)
			return lc.fetchSubmissionCodeCN(id, retry+1)
		}
		return "", fmt.Errorf("max retries reached for network error: %w", err)
	}

	// leetcode.cn rate-limits with a JSON-escaped Chinese message. The raw response
	// body contains literal \uXXXX sequences, not decoded UTF-8, so we match the
	// escaped form. "超出访问限制" = \u8d85\u51fa\u8bbf\u95ee\u9650\u5236
	if strings.Contains(string(bodyBytes), `\u8d85\u51fa\u8bbf\u95ee\u9650\u5236`) {
		if retry < maxRetry {
			log.Printf("Rate limit hit, retry %d/%d after %v\n", retry+1, maxRetry, rateLimitBackoff)
			time.Sleep(rateLimitBackoff)
			return lc.fetchSubmissionCodeCN(id, retry+1)
		}
		return "", fmt.Errorf("max retries reached for CN rate limit for id=%s", id)
	}

	body := &RequestBody[lcSubmissionDetailDataCN]{}
	if err := json.Unmarshal(bodyBytes, body); err != nil {
		return "", fmt.Errorf("JSON parsing error: %w", err)
	}

	if body.Data.Detail == nil {
		if retry < maxRetry {
			log.Printf("Null response, retry %d/%d after %v\n", retry+1, maxRetry, backoffTime)
			time.Sleep(backoffTime)
			return lc.fetchSubmissionCodeCN(id, retry+1)
		}
		return "", fmt.Errorf("max retries reached for null CN submissionDetail response for id=%s", id)
	}

	if len(body.Data.Detail.Code) == 0 {
		if retry < maxRetry {
			log.Printf("Empty code, retry %d/%d after %v\n", retry+1, maxRetry, backoffTime)
			time.Sleep(backoffTime)
			return lc.fetchSubmissionCodeCN(id, retry+1)
		}
		return "", fmt.Errorf("max retries reached for empty CN code for submission %s", id)
	}

	return body.Data.Detail.Code, nil
}

// queryLeetcode sends the query string to leetcode's GraphQL URL.
//
// On success it returns the resulting bytes of the response body and a nil error.
// Otherwise it will return nil and any error it faces while creating the request
// or while communicating with LC.
func (lc leetcode) queryLeetcode(query string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, lc.graphqlUrl, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, err
	}
	lc.addCookieAndHeaders(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, _ := io.ReadAll(res.Body)
	// A non-JSON response (e.g. HTML error page) would cause confusing downstream
	// JSON parse errors; surface the HTTP status and a snippet here instead.
	if len(bodyBytes) > 0 && bodyBytes[0] != '{' && bodyBytes[0] != '[' {
		preview := string(bodyBytes)
		if len(preview) > 300 {
			preview = preview[:300] + "..."
		}
		return nil, fmt.Errorf("unexpected non-JSON response (HTTP %d) from %s: %s",
			res.StatusCode, lc.graphqlUrl, preview)
	}
	return bodyBytes, nil
}

// browserUserAgent is a standard Chrome UA sent with every request.
// Cloudflare and leetcode.cn both fingerprint the User-Agent; the Go default
// ("Go-http-client/2.0") is immediately flagged as a bot.
const browserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// Adds cfg.LcCookie cookie and necessary headers to req.
// For leetcode.cn, also attaches the csrftoken and cf_clearance cookies,
// x-csrftoken header, and the Referer/Origin headers required by Django's
// CSRF middleware and Cloudflare Bot Management.
func (lc leetcode) addCookieAndHeaders(req *http.Request) {
	req.AddCookie(&http.Cookie{
		Name:     "LEETCODE_SESSION",
		Value:    lc.cfg.LcCookie,
		Path:     "/",
		Domain:   lc.cookieDomain,
		HttpOnly: true,
		MaxAge:   1209600,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})
	if lc.cfg.LcCsrfToken != "" {
		req.AddCookie(&http.Cookie{
			Name:     "csrftoken",
			Value:    lc.cfg.LcCsrfToken,
			Path:     "/",
			Domain:   lc.cookieDomain,
			SameSite: http.SameSiteLaxMode,
			Secure:   true,
		})
		req.Header.Add("x-csrftoken", lc.cfg.LcCsrfToken)
	}
	if lc.cfg.LcCfClearance != "" {
		// Cloudflare sets cf_clearance after the browser solves its JS challenge.
		// It is bound to the IP + User-Agent that solved the challenge, so
		// browserUserAgent must match what your browser sent at that time.
		req.AddCookie(&http.Cookie{
			Name:   "cf_clearance",
			Value:  lc.cfg.LcCfClearance,
			Path:   "/",
			Domain: lc.cookieDomain,
			Secure: true,
		})
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", lc.cfg.LcCookie))
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-type", "application/json")
	// Django's CsrfViewMiddleware validates Referer against Origin for HTTPS
	// requests; without it the server returns a 403 HTML page instead of JSON.
	req.Header.Add("Referer", lc.siteOrigin+"/")
	req.Header.Add("Origin", lc.siteOrigin)
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

// lcSubmissionListDataCN is used for leetcode.cn whose GraphQL schema exposes
// the field as "submissionList" instead of "questionSubmissionList".
type lcSubmissionListDataCN struct {
	LCSubmissionList lcSubmissionList `json:"submissionList"`
}

type lcSubmissionList struct {
	LCSubmissions []lcSumbissionOverview `json:"submissions"`
}

type lcSumbissionOverview struct {
	Id   string `json:"id"`
	Lang string `json:"lang"`
}

type lcSubmissionDetailsData struct {
	Details *lcSubmissionDetails `json:"submissionDetails"`
}

type lcSubmissionDetails struct {
	Code string `json:"code"`
}

// lcSubmissionDetailDataCN is the response wrapper for leetcode.cn's
// submissionDetail (singular) GraphQL field.
type lcSubmissionDetailDataCN struct {
	Detail *lcSubmissionDetails `json:"submissionDetail"`
}

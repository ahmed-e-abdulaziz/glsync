package code

import (
	_ "embed"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/config"
	"github.com/stretchr/testify/assert"
)

//go:embed leetcode-testdata/leetcode-responses/question-submission-list-response.json
var questionSubmissionListResponse []byte

//go:embed leetcode-testdata/leetcode-responses/submission-details-response.json
var submissionDetailsResponse []byte

//go:embed leetcode-testdata/leetcode-responses/user-progress-question-list-response.json
var userProgressQuestionListResponse []byte

var submissionListCalled = false
var submissionDetailsCalled = false
var userProgressQuestionListCalled = false

var lc leetcode
var currentHandler func(w http.ResponseWriter, reqBody string)

func TestMain(m *testing.M) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := io.ReadAll(r.Body)
		currentHandler(w, string(reqBody))
	}))
	testUrl := "http://" + server.Listener.Addr().String()
	cfg := config.Config{LcCookie: "COOKIE", RepoUrl: "REPO_URL"}
	lc = NewLeetCode(cfg, testUrl)
	m.Run()
}

func TestFetchSubmissions(t *testing.T) {
	// Given
	currentHandler = func(w http.ResponseWriter, reqBody string) {
		if strings.Contains(reqBody, "userProgressQuestionList") {
			userProgressQuestionListCalled = true
			w.Write(userProgressQuestionListResponse)
		}
		if strings.Contains(reqBody, "submissionList") {
			submissionListCalled = true
			w.Write(questionSubmissionListResponse)
		}
		if strings.Contains(reqBody, "submissionDetails") {
			submissionDetailsCalled = true
			w.Write(submissionDetailsResponse)
		}
	}

	// When
	res, _ := lc.FetchSubmissions()
	submission := res[0]

	// Then
	assert.Equal(t, submission.Id, "128")
	assert.Equal(t, submission.Lang, "golang")
	assert.Equal(t, submission.Title, "Longest Consecutive Sequence")
	assert.True(t, userProgressQuestionListCalled)
	assert.True(t, submissionListCalled)
	assert.True(t, submissionDetailsCalled)
}

func TestFetchSubmissionsShouldReturnErrorWhenFetchQuestionsFails(t *testing.T) {
	// Given
	currentHandler = func(w http.ResponseWriter, reqBody string) {
		if strings.Contains(reqBody, "userProgressQuestionList") {
			panic("panicing so the method fetchSubmissionCode fails")
		}
	}

	// When
	_, err := lc.FetchSubmissions()

	// Then
	assert.Error(t, err)
}

func TestFetchSubmissionsShouldReturnErrorWhenFetchSubmissionOverviewFails(t *testing.T) {
	// Given
	currentHandler = func(w http.ResponseWriter, reqBody string) {
		if strings.Contains(reqBody, "userProgressQuestionList") {
			w.Write(userProgressQuestionListResponse)
		}
		if strings.Contains(reqBody, "submissionList") {
			panic("panicing so the method fetchSubmissionCode fails")
		}
	}

	// When
	_, err := lc.FetchSubmissions()

	// Then
	assert.Error(t, err)
}

func TestFetchSubmissionsShouldReturnErrorWhenFetchSubmissionCodeFails(t *testing.T) {
	// Given
	currentHandler = func(w http.ResponseWriter, reqBody string) {
		if strings.Contains(reqBody, "userProgressQuestionList") {
			w.Write(userProgressQuestionListResponse)
		}
		if strings.Contains(reqBody, "submissionList") {
			w.Write(questionSubmissionListResponse)
		}
		if strings.Contains(reqBody, "submissionDetails") {
			panic("panicing so the method fetchSubmissionCode fails")
		}
	}

	// When
	_, err := lc.FetchSubmissions()

	// Then
	assert.Error(t, err)
}

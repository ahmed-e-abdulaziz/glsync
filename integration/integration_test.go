package integrationtest

import (
	_ "embed"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var questionSubmissionListResponse, _ = os.ReadFile("../code/leetcode-testdata/leetcode-responses/question-submission-list-response.json")

var submissionDetailsResponse, _ = os.ReadFile("../code/leetcode-testdata/leetcode-responses/submission-details-response.json")

var userProgressQuestionListResponse, _ = os.ReadFile("../code/leetcode-testdata/leetcode-responses/user-progress-question-list-response.json")

var userProgressQuestionListCalled, submissionListCalled, submissionDetailsCalled bool

func TestLeetCodeGitIntegration(t *testing.T) {
	// Given
	mockLeetCodeUrl := initMockLeetCode()
	mockGitRepoUrl := initStubRepo(t)
	os.Args = append(os.Args, "-lc-cookie=eyJhbGciOiJIUzI1NiJ9.eyJuYW1lIjoiWW91IGZvdW5kIGEgc2VjcmV0ISJ9.bg7oA5pFvtjAn1cXW7RRVhl0MUpJqmb90AUiRjh5XHY") // Fake but valid JWT, like LeetCode cookie
	os.Args = append(os.Args, "-repo-url="+mockGitRepoUrl)
	defer os.RemoveAll("repo")

	// When
	cmd.Execute(mockLeetCodeUrl)

	// Then
	// Assert LeetCode called as expected
	assert.True(t, userProgressQuestionListCalled)
	assert.True(t, submissionListCalled)
	assert.True(t, submissionDetailsCalled)

	// Assert the Git push worked successfully
	expectedTimestamp, _ := time.Parse("2006-01-02 15:04:05 -0700", "2024-12-28 17:25:31 +0000")
	actualTimestamp, message := getCommitTimeAndMessage(t, mockGitRepoUrl)
	assert.Equal(t, expectedTimestamp, actualTimestamp)
	assert.Equal(t, "Code challenge submission for question: 128 Longest Consecutive Sequence", message)
}

func initMockLeetCode() string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := io.ReadAll(r.Body)
		if strings.Contains(string(reqBody), "userProgressQuestionList") {
			userProgressQuestionListCalled = true
			w.Write(userProgressQuestionListResponse)
		}
		if strings.Contains(string(reqBody), "submissionList") {
			submissionListCalled = true
			w.Write(questionSubmissionListResponse)
		}
		if strings.Contains(string(reqBody), "submissionDetails") {
			submissionDetailsCalled = true
			w.Write(submissionDetailsResponse)
		}
	}))
	testUrl := "http://" + server.Listener.Addr().String()
	return testUrl
}

// Creates a simple folder with git init inside to create a repo
func initStubRepo(t *testing.T) string {
	repoPath := "repo/test.git"
	err := os.MkdirAll(repoPath, os.ModePerm)
	require.NoError(t, err)
	err = exec.Command("git", "init", "--bare", repoPath).Run()
	require.NoError(t, err)
	return repoPath
}

func getCommitTimeAndMessage(t *testing.T, repoPath string) (time.Time, string) {
	// git log --pretty=format:'%ad|%s'" --date=iso
	logOutputBytes, err := exec.Command("git", "-C", repoPath, "log", "--pretty=format:'%ad|%s'", "--date=iso").CombinedOutput()
	if err != nil {
		t.Fatal("Failed to do git log command ", logOutputBytes, err)
	}
	logOutput := string(logOutputBytes)
	logOutput = logOutput[1 : len(logOutput)-1]
	commitMessageAndDate := strings.Split(string(logOutput), "|")
	actualTimestamp, _ := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(commitMessageAndDate[0]))
	actualCommitMessage := strings.TrimSpace(commitMessageAndDate[1])
	return actualTimestamp, actualCommitMessage
}

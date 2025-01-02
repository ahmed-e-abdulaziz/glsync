package handler

import (
	"errors"
	"testing"
	"time"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/code"
	"github.com/ahmed-e-abdulaziz/gh-leet-sync/mocks/mock_code"
	"github.com/ahmed-e-abdulaziz/gh-leet-sync/mocks/mock_git"
	"go.uber.org/mock/gomock"
)

func TestExecute(t *testing.T) {
	ctrl, mockCodeClient, mockGitClient := initMocks(t)
	defer ctrl.Finish()

	subs := stubSubmissions()
	gomock.InOrder(
		mockCodeClient.EXPECT().FetchSubmissions().Return(subs, nil).Times(1),
		mockGitClient.EXPECT().
			Commit("1 Two Sum", "1two-sum.go", subs[0].Code, "Code challenge submission for question: 1 Two Sum", subs[0].LastSubmittedAt).
			Return(nil).
			Times(1),
		mockGitClient.EXPECT().
			Commit("2 Add Two Numbers", "2add-two-numbers.go", subs[1].Code, "Code challenge submission for question: 2 Add Two Numbers", subs[1].LastSubmittedAt).
			Return(nil).
			Times(1),
		mockGitClient.EXPECT().Push().Return(nil).Times(1),
	)

	NewHandler(mockCodeClient, mockGitClient).Execute()
}

func stubSubmissions() []code.Submission {
	subs := []code.Submission{
		{
			Id:              "1",
			Title:           "Two Sum",
			TitleSlug:       "two-sum",
			LastSubmittedAt: parseRFC3339("2024-12-31T00:00:00+02:00"),
			Lang:            "golang",
			Code:            "package main\n",
		},
		{
			Id:              "2",
			Title:           "Add Two Numbers",
			TitleSlug:       "add-two-numbers",
			LastSubmittedAt: parseRFC3339("2024-12-15T00:00:00+02:00"),
			Lang:            "golang",
			Code:            "package main\n",
		},
	}
	return subs
}

func TestExecuteShouldPanicWhenFetchSubmissionFails(t *testing.T) {
	ctrl, mockCodeClient, mockGitClient := initMocks(t)
	defer ctrl.Finish()
	mockCodeClient.EXPECT().FetchSubmissions().Return(nil, errors.New("mock error")).Times(1)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic although fetch sumbissions failed")
		}
	}()
	NewHandler(mockCodeClient, mockGitClient).Execute()
}

func TestExecuteShouldContinueWhenACommitFails(t *testing.T) {
	ctrl, mockCodeClient, mockGitClient := initMocks(t)
	defer ctrl.Finish()

	subs := stubSubmissions()
	gomock.InOrder(
		mockCodeClient.EXPECT().FetchSubmissions().Return(subs, nil).Times(1),
		mockGitClient.EXPECT().
			Commit("1 Two Sum", "1two-sum.go", subs[0].Code, "Code challenge submission for question: 1 Two Sum", subs[0].LastSubmittedAt).
			Return(nil).
			Times(1),
		mockGitClient.EXPECT().
			Commit("2 Add Two Numbers", "2add-two-numbers.go", subs[1].Code, "Code challenge submission for question: 2 Add Two Numbers", subs[1].LastSubmittedAt).
			Return(errors.New("Second Commit Failed")). // Commit Failure
			Times(1),
		mockGitClient.EXPECT().Push().Return(nil).Times(1), // Push should happen regardless of failure
	)
	NewHandler(mockCodeClient, mockGitClient).Execute()
}

func TestExecuteShouldPanicWhenPushFails(t *testing.T) {
	ctrl, mockCodeClient, mockGitClient := initMocks(t)
	defer ctrl.Finish()

	subs := stubSubmissions()
	gomock.InOrder(
		mockCodeClient.EXPECT().FetchSubmissions().Return(subs, nil).Times(1),
		mockGitClient.EXPECT().
			Commit("1 Two Sum", "1two-sum.go", subs[0].Code, "Code challenge submission for question: 1 Two Sum", subs[0].LastSubmittedAt).
			Return(nil).
			Times(1),
		mockGitClient.EXPECT().
			Commit("2 Add Two Numbers", "2add-two-numbers.go", subs[1].Code, "Code challenge submission for question: 2 Add Two Numbers", subs[1].LastSubmittedAt).
			Return(nil).
			Times(1),
		mockGitClient.EXPECT().Push().Return(errors.New("Error happened while pushing")).Times(1), // git.Push() fails
	)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic although git push failed")
		}
	}()
	NewHandler(mockCodeClient, mockGitClient).Execute()
}

func initMocks(t *testing.T) (*gomock.Controller, *mock_code.MockCodeClient, *mock_git.MockGitClient) {
	ctrl := gomock.NewController(t)
	mockCodeClient := mock_code.NewMockCodeClient(ctrl)
	mockGitClient := mock_git.NewMockGitClient(ctrl)
	return ctrl, mockCodeClient, mockGitClient
}

func parseRFC3339(timeString string) time.Time {
	timestamp, _ := time.Parse(time.RFC3339, timeString)
	return timestamp
}

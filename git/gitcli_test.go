package git

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var g gitcli

func TestMain(m *testing.M) {
	testDir := createTestFolder()
	output, err := exec.Command("git", "init").CombinedOutput()
	if err != nil {
		log.Fatal(string(output)+" ", err)
	}
	g = gitcli{config.Config{LcCookie: "COOKIE", RepoUrl: "REPO_URL"}, testDir}
	m.Run()
	deleteTestFolder(testDir)
}

func deleteTestFolder(testDir string) {
	// Go out of the current test folder and delete it
	os.Chdir("..")
	os.RemoveAll(testDir)
}

func createTestFolder() string {
	// Go to home directory
	home, _ := os.UserHomeDir()
	os.Chdir(home)

	// Create and go into the test folder
	testDir := "test-gitcli-folder"
	os.Mkdir(testDir, os.ModePerm)
	os.Chdir(home + "/" + testDir)
	return testDir
}

func TestCommit(t *testing.T) {
	// Given
	codeFolderName, fileName, code, commitMessage, timestamp := "new-code-folder", "stub.go", "package main\n", "commit message", time.Now()
	defer os.RemoveAll("new-code-folder")

	// When
	g.Commit(codeFolderName, fileName, code, commitMessage, timestamp)

	// Then
	// Verify that the folder and file of the code exists
	assert.DirExists(t, codeFolderName)
	filePath := codeFolderName + "/" + fileName
	assert.FileExists(t, filePath)

	//Veriy the code is correct
	actualCode, _ := os.ReadFile(filePath)
	assert.Equal(t, code, string(actualCode))

	// Verify the date of the commit is correct
	actualTimestamp, actualCommitMessage := getCommitTimeAndMessage(t)
	assert.Equal(t, commitMessage, actualCommitMessage)
	assert.Equal(t, timestamp.Round(time.Minute), actualTimestamp.Round(time.Minute)) // Round to avoid partial second errors
}

func TestCommitShouldFailWhenFolderCreationFails(t *testing.T) {
	// Given
	os.Mkdir("alreadyexists", os.ModeDir)
	defer os.Remove("alreadyexists")
	invalidFolderName, fileName, code, commitMessage, timestamp := "alreadyexists", "stub.go", "package main\n", "commit message", time.Now()

	// When
	err := g.Commit(invalidFolderName, fileName, code, commitMessage, timestamp)

	// Then
	require.Error(t, err)
}

// This test will fail if you for some reason keep track of your temp folder using git
// and in that case please tell me why you did that
// I would genuinely love to know who does something like this
func TestCommitShouldFailWhenGitAddFails(t *testing.T) {
	// Given
	originalDir, _ := os.Getwd()
	os.Chdir(os.TempDir()) // Not a git repo, so git add should fail
	folderName, fileName, code, commitMessage, timestamp := "new-code-folder", "stub.go", "package main\n", "commit message", time.Now()

	// When
	err := g.Commit(folderName, fileName, code, commitMessage, timestamp)

	// Then
	require.Error(t, err)
	os.RemoveAll("new-code-folder")
	os.Chdir(originalDir)
}

func TestCommitShouldFailWhenGitCommitFails(t *testing.T) {
	// Given
	codeFolderName, fileName, code, emptyCommitMessage, timestamp := "new-code-folder", "stub.go", "package main\n", "", time.Now()
	defer os.RemoveAll("new-code-folder")

	// When
	err := g.Commit(codeFolderName, fileName, code, emptyCommitMessage, timestamp)

	// Then
	assert.Error(t, err)
}

func getCommitTimeAndMessage(t *testing.T) (time.Time, string) {
	logOutputBytes, err := exec.Command("git", "log", "--pretty=format:'%ad|%s'", "--date=iso").CombinedOutput()
	if err != nil {
		t.Fatal("Failed to do git log command ", string(logOutputBytes), err)
	}
	logOutput := string(logOutputBytes)
	logOutput = logOutput[1 : len(logOutput)-1]
	commitMessageAndDate := strings.Split(string(logOutput), "|")
	actualTimestamp, _ := time.Parse("2006-01-02 15:04:05 -0700", strings.TrimSpace(commitMessageAndDate[0]))
	actualCommitMessage := strings.TrimSpace(commitMessageAndDate[1])
	return actualTimestamp, actualCommitMessage
}

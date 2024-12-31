package git

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/config"
)

const commitDateEnvVar = "GIT_COMMITTER_DATE"

type github struct {
	cfg            config.Config
	repoFolderName string
}

func NewGithub(cfg config.Config) GitClient {
	gh := github{cfg: cfg}
	url := strings.Split(gh.cfg.RepoUrl, "/")
	gh.repoFolderName = strings.Split(url[len(url)-1], ".")[0]
	if _, err := os.Stat(gh.repoFolderName); err == nil {
		log.Printf(`Removing folder: [%v] as it is the same as the repo folder's name to be able to clone the repo`,
			gh.repoFolderName)
		os.RemoveAll(gh.repoFolderName)
	}

	if err := exec.Command("git", "clone", cfg.RepoUrl).Run(); err != nil {
		log.Fatalf(
			`Encountered an error while cloning the repo.
			Please create your repo on Github before using glsync, the repo: "%s" doesn't exist`,
			cfg.RepoUrl)
	}
	os.Chdir(gh.repoFolderName)
	return gh
}

func (g github) Commit(folderName, fileName, code, commitMessage string, timestamp time.Time) {
	g.createCodeFolderAndFile(folderName, fileName, code)
	exec.Command("git", "add", ".").Run()
	os.Setenv(commitDateEnvVar, g.toGitDate(timestamp))
	g.buildCommitCommand(commitMessage, timestamp).Run()
	os.Unsetenv(commitDateEnvVar)
}

func (g github) Push() {
	exec.Command("git", "push").Run()
	os.Chdir("..")
	os.RemoveAll(g.repoFolderName)
}

func (g github) createCodeFolderAndFile(folderName string, fileName string, code string) {
	filePath := folderName + "/" + fileName
	os.Mkdir(folderName, os.ModePerm)
	os.WriteFile(filePath, []byte(code), os.ModePerm)
}

func (g github) buildCommitCommand(commitMessage string, timestamp time.Time) *exec.Cmd {
	commitCommand := exec.Command("git", "commit", fmt.Sprintf("--date='%v'", g.toGitDate(timestamp)), fmt.Sprintf("-m %s", commitMessage))
	return commitCommand
}

func (g github) toGitDate(timestamp time.Time) string {
	_, offset := timestamp.Zone()
	return fmt.Sprintf("%v %+05d", timestamp.Unix(), offset)
}

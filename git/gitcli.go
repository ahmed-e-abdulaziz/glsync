package git

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ahmed-e-abdulaziz/glsync/config"
)

const commitDateEnvVar = "GIT_COMMITTER_DATE"

type gitcli struct {
	cfg            config.Config
	repoFolderName string
}

func NewGitCli(cfg config.Config) gitcli {
	gh := gitcli{cfg: cfg}
	url := strings.Split(gh.cfg.RepoUrl, "/")
	gh.repoFolderName = strings.Split(url[len(url)-1], ".")[0]
	if _, err := os.Stat(gh.repoFolderName); err == nil {
		log.Printf(`Removing folder: [%v] as it is the same as the repo folder's name to be able to clone the repo`,
			gh.repoFolderName)
		os.RemoveAll(gh.repoFolderName)
	}
	err := exec.Command("git", "clone", cfg.RepoUrl).Run()
	if err != nil {
		log.Panicf(
			`Encountered an error while cloning the repo.
			Please create your repo on Git before using glsync, the repo: "%s" doesn't exist`,
			cfg.RepoUrl)
	}
	err = os.Chdir(gh.repoFolderName)
	if err != nil {
		log.Panicf("Couldn't chdir into repo folder %s. Please check permissions and try again", gh.repoFolderName)
	}
	return gh
}

func (g gitcli) Commit(folderName, fileName, code, commitMessage string, timestamp time.Time) error {
	// TODO: Sanitize the folder and file names to make them valid, https://stackoverflow.com/questions/4814040/allowed-characters-in-filename
	err := g.createCodeFolderAndFile(folderName, fileName, code)
	if err != nil {
		return fmt.Errorf("encountered the following error while creating the code folder and file:\n%v", err)
	}
	out, err := exec.Command("git", "add", ".").CombinedOutput()
	if err != nil {
		return fmt.Errorf(`encountered an error while executing the command 'git add .' in folder %s.
			The error:%s with command output:%s`, g.repoFolderName, err, string(out))
	}
	os.Setenv(commitDateEnvVar, g.toGitDate(timestamp))
	defer os.Unsetenv(commitDateEnvVar)
	out, err = exec.Command("git", "commit", fmt.Sprintf("--date='%v'", g.toGitDate(timestamp)), fmt.Sprintf("-m %s", commitMessage)).CombinedOutput()
	if err != nil {
		return fmt.Errorf(`encountered an error while executing the command 'git commit --date='%s' -m %s' in folder %s.
			The error: %s 
			with command output: %s`,
			fmt.Sprintf("--date='%v'", g.toGitDate(timestamp)), fmt.Sprintf("-m %s", commitMessage),
			g.repoFolderName, err, string(out))
	}
	return nil
}

func (g gitcli) Push() error {
	err := exec.Command("git", "push").Run()
	if err != nil {
		return errors.New("encountered an error while doing the command 'git push' in the repo folder: " + g.repoFolderName)
	}
	err = os.Chdir("..")
	if err != nil {
		return errors.New("couldn't go back to the enclosing folder 'ch ..', could be a permissions issue")
	}
	err = os.RemoveAll(g.repoFolderName)
	if err != nil {
		return fmt.Errorf("couldn't delete the repo folder after pushing 'rm -rf %s', could be a permissions issue",
			g.repoFolderName)
	}
	return nil
}

func (g gitcli) createCodeFolderAndFile(folderName string, fileName string, code string) error {
	filePath := folderName + "/" + fileName
	err := os.Mkdir(folderName, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, []byte(code), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (g gitcli) toGitDate(timestamp time.Time) string {
	_, offset := timestamp.Zone()
	return fmt.Sprintf("%v %+05d", timestamp.Unix(), offset)
}

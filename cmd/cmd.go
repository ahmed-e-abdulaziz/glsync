package cmd

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/ahmed-e-abdulaziz/glsync/code"
	"github.com/ahmed-e-abdulaziz/glsync/config"
	"github.com/ahmed-e-abdulaziz/glsync/git"
	"github.com/ahmed-e-abdulaziz/glsync/handler"
)

const (
	lcCookieCmd = "lc-cookie"
	repoUrlCmd  = "repo-url"
)

func Execute(lcGraphQlUrl string) {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	initUsageFunc()
	cfg := initConfig()
	lc := code.NewLeetCode(cfg, lcGraphQlUrl)
	gh := git.NewGitCli(cfg)
	handler := handler.NewHandler(lc, gh)
	handler.Execute()
}

func initUsageFunc() {
	oldUsage := flag.Usage
	flag.Usage = func() {
		log.Print("CLI tool to sync all your LeetCode submissions to Github (And possibly any other git client)\n\n")
		oldUsage()
	}
}

func initConfig() config.Config {
	cfg := config.Config{}
	flag.StringVar(&cfg.LcCookie, lcCookieCmd, "", "The cookie of your LeetCode session, refer to the README.md for more info")
	flag.StringVar(&cfg.RepoUrl, repoUrlCmd, "", "The git repo's url to push LC submissions to")
	flag.Parse()
	if cfg.LcCookie == "" || !isValidCookie(cfg.LcCookie) {
		log.Panicf("Invalid leet code session cookie provided, use -%v option to provide your leetcode cookie", lcCookieCmd)
	}
	if cfg.RepoUrl == "" {
		log.Panicf("No git repo url was provided, use -%v option to provide your git repo url ", repoUrlCmd)
	}
	log.Println("Input parsed successfully.")
	return cfg
}

func isValidCookie(cookie string) bool {
	splittedCookie := strings.Split(cookie, ".")
	if len(splittedCookie) < 3 {
		log.Println("Invalid LeetCode cookie, it wasn't a valid JWT")
		return false
	}
	payload, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(splittedCookie[1])
	if err != nil {
		return false
	}
	return json.Unmarshal([]byte(payload), &json.RawMessage{}) == nil
}

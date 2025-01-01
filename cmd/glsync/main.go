package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/url"
	"strings"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/code"
	"github.com/ahmed-e-abdulaziz/gh-leet-sync/config"
	"github.com/ahmed-e-abdulaziz/gh-leet-sync/git"
	"github.com/ahmed-e-abdulaziz/gh-leet-sync/handler"
)

const (
	lcCookieCmd = "lc-cookie"
	repoUrlCmd  = "repo-url"
)

func main() {
	cfg := initConfig()
	lc := code.NewLeetCode(cfg, "https://leetcode.com/graphql/")
	gh := git.NewGithub(cfg)
	handler := handler.NewHandler(lc, gh)
	handler.Execute()
}

func initConfig() config.Config {
	cfg := config.Config{}
	flag.StringVar(&cfg.LcCookie, lcCookieCmd, "", "The cookie of your LeetCode session, refer to the README.md for more info")
	flag.StringVar(&cfg.RepoUrl, repoUrlCmd, "", "Github repo's url to push LC submissions to")
	flag.Parse()
	if !isValidCookie(cfg.LcCookie) {
		log.Fatalf("Invalid leet code session cookie provided, use -%v option to provide your leetcode cookie", lcCookieCmd)
	}
	if !isGitHubURL(cfg.RepoUrl) {
		log.Fatalf("Invalid github repo url provided provided, use -%v option to provide your github repo url ", repoUrlCmd)
	}
	log.Println("Input parsed successfully.")
	return cfg
}

func isValidCookie(cookie string) bool {
	payload, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(strings.Split(cookie, ".")[1])
	if err != nil {
		return false
	}
	return json.Unmarshal([]byte(payload), &json.RawMessage{}) == nil
}

func isGitHubURL(input string) bool {
	u, err := url.Parse(input)
	if err != nil {
		return false
	}
	host := u.Host
	if strings.Contains(host, ":") {
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			return false
		}
	}
	return host == "github.com"
}

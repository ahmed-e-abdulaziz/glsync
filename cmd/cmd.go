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
	lcCookieArg    = "lc-cookie"
	repoUrlArg     = "repo-url"
	bearerTokenArg = "bearer-token"
	siteArg           = "site"
	lcCsrfTokenArg    = "lc-csrf-token"
	lcCfClearanceArg  = "lc-cf-clearance"
)

var graphqlURLBySite = map[string]string{
	"com": "https://leetcode.com/graphql/",
	"cn":  "https://leetcode.cn/graphql/",
}

// Execute is the CLI entry point. urlOverride is used by tests to point at a
// mock server; pass an empty string in production and the URL will be derived
// from the --site flag.
func Execute(urlOverride string) {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	initUsageFunc()
	cfg := initConfig()

	graphqlURL := urlOverride
	if graphqlURL == "" {
		url, ok := graphqlURLBySite[cfg.LcSite]
		if !ok {
			log.Panicf("Unknown site %q, valid values are: com, cn", cfg.LcSite)
		}
		graphqlURL = url
	}

	lc := code.NewLeetCode(cfg, graphqlURL)
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
	flag.StringVar(&cfg.LcCookie, lcCookieArg, "", "The cookie of your LeetCode session, refer to the README.md for more info")
	flag.StringVar(&cfg.RepoUrl, repoUrlArg, "", "The git repo's url to push LC submissions to")
	flag.StringVar(&cfg.LcSite, siteArg, "com", "LeetCode site to sync from: \"com\" for leetcode.com (default) or \"cn\" for leetcode.cn")
	flag.StringVar(&cfg.LcCsrfToken, lcCsrfTokenArg, "", "CSRF token for leetcode.cn (value of the csrftoken cookie in your browser); required when -site=cn")
	flag.StringVar(&cfg.LcCfClearance, lcCfClearanceArg, "", "Cloudflare clearance token for leetcode.cn (value of the cf_clearance cookie in your browser); required when -site=cn")
	flag.Parse()
	if cfg.LcCookie == "" || !isValidCookie(cfg.LcCookie) {
		log.Panicf("Invalid leet code session cookie provided, use -%v option to provide your leetcode cookie", lcCookieArg)
	}
	if cfg.RepoUrl == "" {
		log.Panicf("No git repo url was provided, use -%v option to provide your git repo url ", repoUrlArg)
	}
	if cfg.LcSite == "cn" && cfg.LcCsrfToken == "" {
		log.Panicf("leetcode.cn requires a CSRF token, use -%v option to provide it", lcCsrfTokenArg)
	}
	if cfg.LcSite == "cn" && cfg.LcCfClearance == "" {
		log.Panicf("leetcode.cn requires a Cloudflare clearance token, use -%v option to provide it (get the cf_clearance cookie value from your browser after visiting leetcode.cn)", lcCfClearanceArg)
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

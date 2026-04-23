package config

type Config struct {
	LcCookie    string // LeetCode's cookie that you can get from Chrome Devtools->Application tab->Cookies->LEETCODE_SESSION
	RepoUrl     string // The repo to push the submitted code to
	BearerToken string // A user reported that LeetCode is now expecting a bearer token, this will be passed as Authorization: Bearer header to LeetCode. Check https://github.com/ahmed-e-abdulaziz/glsync/issues/5 for more info
	LcSite        string // Target LeetCode site: "com" for leetcode.com (default), "cn" for leetcode.cn
	LcCsrfToken   string // CSRF token required by leetcode.cn; get it from the csrftoken cookie in your browser
	LcCfClearance string // Cloudflare clearance cookie for leetcode.cn; get it from the cf_clearance cookie in your browser
}

// ratelimit-probe measures leetcode.cn's submissionDetail rate limit.
//
// Usage:
//
//	go run ./cmd/ratelimit-probe \
//	  -cookie=<LEETCODE_SESSION> \
//	  -csrf=<csrftoken> \
//	  -cf=<cf_clearance> \
//	  -id=719588908 \
//	  -interval=0   # seconds between requests (0 = burst)
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	graphqlURL = "https://leetcode.cn/graphql/"
	queryFmt   = `{"query":"\n    query submissionDetail($id: ID!) {\n  submissionDetail(submissionId: $id) {\n    code\n  }\n}\n    ","variables":{"id":"%s"},"operationName":"submissionDetail"}`
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
)

func main() {
	cookie := flag.String("cookie", "", "LEETCODE_SESSION cookie value")
	csrf := flag.String("csrf", "", "csrftoken cookie value")
	cf := flag.String("cf", "", "cf_clearance cookie value (optional)")
	id := flag.String("id", "719588908", "submission ID to probe")
	intervalSec := flag.Int("interval", 0, "seconds between requests (0 = burst)")
	maxReqs := flag.Int("max", 500, "max requests to fire before stopping")
	flag.Parse()

	if *cookie == "" || *csrf == "" {
		log.Fatal("must provide -cookie and -csrf flags")
	}

	interval := time.Duration(*intervalSec) * time.Second
	start := time.Now()
	mode := map[bool]string{true: "burst", false: "paced"}[*intervalSec == 0]

	fmt.Printf("mode=%s interval=%ds max=%d submission_id=%s\n",
		mode, *intervalSec, *maxReqs, *id)
	fmt.Println("req#   elapsed   status")
	fmt.Println("-----  --------  ------")

	for i := 1; i <= *maxReqs; i++ {
		elapsed := time.Since(start).Round(time.Millisecond)
		status, isRateLimit := probe(*id, *cookie, *csrf, *cf)
		fmt.Printf("%5d  %8s  %s\n", i, elapsed, status)

		if isRateLimit {
			fmt.Printf("\n==> RATE LIMIT triggered at request %d, elapsed %s\n", i, elapsed)
			fmt.Println("==> Polling every 10s to find exact recovery time...")
			pollRecovery(*id, *cookie, *csrf, *cf, start)
			return
		}

		if interval > 0 {
			time.Sleep(interval)
		}
	}

	fmt.Printf("\nNo rate limit after %d requests.\n", *maxReqs)
}

// probe fires one submissionDetail request and classifies the response.
func probe(id, cookie, csrf, cf string) (status string, isRateLimit bool) {
	body := fmt.Sprintf(queryFmt, id)
	req, err := http.NewRequest(http.MethodPost, graphqlURL, bytes.NewBufferString(body))
	if err != nil {
		return "request_build_error: " + err.Error(), false
	}

	req.AddCookie(&http.Cookie{Name: "LEETCODE_SESSION", Value: cookie, Domain: ".leetcode.cn", Secure: true})
	req.AddCookie(&http.Cookie{Name: "csrftoken", Value: csrf, Domain: ".leetcode.cn"})
	if cf != "" {
		req.AddCookie(&http.Cookie{Name: "cf_clearance", Value: cf, Domain: ".leetcode.cn", Secure: true})
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-csrftoken", csrf)
	req.Header.Set("Referer", "https://leetcode.cn/")
	req.Header.Set("Origin", "https://leetcode.cn")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "network_error: " + err.Error(), false
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	s := string(b)

	switch {
	// rate limit message: "超出访问限制，请稍后再试"
	case strings.Contains(s, `\u8d85\u51fa\u8bbf\u95ee\u9650\u5236`):
		return "RATE_LIMITED", true
	case strings.Contains(s, `"code":"`):
		return "ok", false
	case strings.Contains(s, `"submissionDetail":null`):
		return "null_no_error", false
	default:
		preview := s
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		return "unknown: " + preview, false
	}
}

// pollRecovery polls every 10s after a rate limit hit until the block clears.
func pollRecovery(id, cookie, csrf, cf string, start time.Time) {
	for {
		time.Sleep(10 * time.Second)
		elapsed := time.Since(start).Round(time.Second)
		status, isRateLimit := probe(id, cookie, csrf, cf)
		fmt.Printf("  poll %s: %s\n", elapsed, status)
		if !isRateLimit {
			fmt.Printf("\n==> UNBLOCKED at %s after rate limit hit\n", elapsed)
			return
		}
	}
}

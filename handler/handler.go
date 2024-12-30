package handler

import (
	"fmt"
	"log"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/config"
	"github.com/ahmed-e-abdulaziz/gh-leet-sync/leetcode"
)

type Handler struct {
	cfg      config.Config
	leetcode leetcode.LeetCode
}

func NewHandler(cfg config.Config, leetcode leetcode.LeetCode) Handler {
	return Handler{cfg, leetcode}
}

func (h Handler) Execute() {
	questions := h.leetcode.FetchQuestions()
	submissionsOverview := map[string]leetcode.SumbissionOverview{} // question title slug, the submission overview
	for _, q := range questions {
		submissionsOverview[q.TitleSlug] = h.leetcode.FetchSubmissionOverview(q.TitleSlug)
		s := submissionsOverview[q.TitleSlug]
		code := h.leetcode.FetchSubmissionCode(s.Id)
		fileName := h.buildFileName(q.FrontendId, q.Title, s.Lang)
		folderName := fmt.Sprintf("%v %v", q.FrontendId, q.Title)
		fmt.Printf(code, fileName, folderName, q.LastSubmittedAt)
	}

	log.Println("Not implemented yet")
}

func (Handler) buildFileName(id, titleSlug, lang string) string {
	langFileExtension := map[string]string{
		"cpp":        "cpp",
		"java":       "java",
		"python":     "py",
		"python3":    "py",
		"mysql":      "sql",
		"mssql":      "sql",
		"oraclesql":  "sql",
		"c":          "c",
		"csharp":     "cs",
		"javascript": "js",
		"typescript": "ts",
		"bash":       "sh",
		"php":        "php",
		"swift":      "swift",
		"kotlin":     "kt",
		"dart":       "dart",
		"golang":     "go",
		"ruby":       "rb",
		"scala":      "scala",
		"rust":       "rs",
		"racket":     "rkt",
		"erlang":     "erl",
		"elixir":     "ex",
		"postgresql": "sql",
	}
	return fmt.Sprintf("%s%s.%s", id, titleSlug, langFileExtension[lang])
}

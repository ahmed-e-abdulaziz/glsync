package handler

import (
	"fmt"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/github"
	"github.com/ahmed-e-abdulaziz/gh-leet-sync/leetcode"
)

type Handler struct {
	leetcode leetcode.LeetCode
	github   github.Github
}

func NewHandler(leetcode leetcode.LeetCode, github github.Github) Handler {
	return Handler{leetcode, github}
}

func (h Handler) Execute() {
	questions := h.leetcode.FetchQuestions()
	submissionsOverview := map[string]leetcode.SumbissionOverview{} // question title slug, the submission overview
	for _, q := range questions {
		submissionsOverview[q.TitleSlug] = h.leetcode.FetchSubmissionOverview(q.TitleSlug)
		s := submissionsOverview[q.TitleSlug]
		code := h.leetcode.FetchSubmissionCode(s.Id)
		fileName := h.buildFileName(q.FrontendId, q.TitleSlug, s.Lang)
		folderName := fmt.Sprintf("%v %v", q.FrontendId, q.Title)
		commitName := fmt.Sprintf("LeetCode submission for question: %v %v", q.FrontendId, q.Title)
		h.github.Commit(folderName, fileName, code, commitName, q.LastSubmittedAt)
	}
	h.github.Push()
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

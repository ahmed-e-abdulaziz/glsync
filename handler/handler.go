package handler

import (
	"fmt"

	"github.com/ahmed-e-abdulaziz/gh-leet-sync/code"
	"github.com/ahmed-e-abdulaziz/gh-leet-sync/git"
)

type Handler struct {
	codeClient code.CodeClient
	git        git.GitClient
}

func NewHandler(codeClient code.CodeClient, gitClient git.GitClient) Handler {
	return Handler{codeClient, gitClient}
}

func (h Handler) Execute() {
	submissions := h.codeClient.FetchSubmissions()
	for _, s := range submissions {
		fileName := h.buildFileName(s.Id, s.TitleSlug, s.Lang)
		folderName := fmt.Sprintf("%v %v", s.Id, s.Title)
		commitName := fmt.Sprintf("Code challenge submission for question: %v %v", s.Id, s.Title)
		h.git.Commit(folderName, fileName, s.Code, commitName, s.LastSubmittedAt)
	}
	h.git.Push()
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

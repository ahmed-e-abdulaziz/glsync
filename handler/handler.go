package handler

import (
	"fmt"
	"log"
	"strings"

	"github.com/ahmed-e-abdulaziz/glsync/code"
	"github.com/ahmed-e-abdulaziz/glsync/git"
)

type Handler struct {
	codeClient code.CodeClient
	git        git.GitClient
}

func NewHandler(codeClient code.CodeClient, gitClient git.GitClient) Handler {
	return Handler{codeClient, gitClient}
}

// It does three things:
//
//	1- Fetch submissions using codeClient
//	2- Loop through submissions and git commit each one
//	3- Use git to push to the repo set in the git client
func (h Handler) Execute() {
	submissions, err := h.codeClient.FetchSubmissions()
	if err != nil {
		panic("Error while fetching code submissions: " + err.Error())
	}
	log.Printf("Fetched %v submissions, will commit them next\n", len(submissions))
	for idx, s := range submissions {
		// ex. s.Id="10", s.TitleSlug="binary-tree", s.Lang="go" then fileName = "10binary-tree.go"
		fileName := h.buildFileName(s.Id, s.TitleSlug, s.Lang)
		// ex. s.Id="10", s.Title="Binary Tree", then folderName = "10 Binary Tree"
		folderName := fmt.Sprintf("%v %v", s.Id, s.Title)
		// ex. s.Id="10", s.Title="Binary Tree", then commitName = "Code challenge submission for question: 10 Binary Tree"
		commitName := fmt.Sprintf("Code challenge submission for question: %v %v", s.Id, s.Title)
		err := h.git.Commit(folderName, fileName, s.Code, commitName, s.LastSubmittedAt)
		if err != nil && !strings.Contains(err.Error(), "nothing to commit") {
			log.Println("\t" + err.Error())
			log.Printf("\tEncountered an error while commiting the code for question with ID: %v\n", s.Id)
		}
		log.Printf("\t%v%% questions committed. Committed question no. %v of total %v with ID: %v\n", int(float64(idx+1)/float64(len(submissions))*100), idx+1, len(submissions), s.Id)
	}
	err = h.git.Push()
	if err != nil {
		panic("Encountered an error while pushing to git, exiting...")
	}
}

// Takes the id, titleSlug and lang to return the fileName
//
// It will follow the format <id><titleSlug>.<langExtension>
//
// It figures out the lang extension using its internal langFileExtension map
// if there any code client support a new language then add it here to avoid future errors
// currently this map was only formed using LeetCode's lang name
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

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
	for _, q := range questions {
		fmt.Print(q.FrontendId)
		fmt.Print(" " + q.Title + "\n")
	}
	log.Println("Not implemented yet")
}

// This package is responsible for committing and pushing the code to a git repo
// It is currently implemented by [gitcli.go]
package git

import "time"

type GitClient interface {
	Commit(folderName, fileName, code, commitMessage string, timestamp time.Time) error
	Push() error
}

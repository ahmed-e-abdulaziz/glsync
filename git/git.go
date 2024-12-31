package git

import "time"

type GitClient interface {
	Commit(folderName, fileName, code, commitMessage string, timestamp time.Time)
	Push()
}

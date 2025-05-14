package rill

import (
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/status"
)

func main() {
	git.Status(status.Short, status.Porcelain("v2"), status.Branch, status.Null)
}

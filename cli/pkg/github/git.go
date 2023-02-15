package github

import (
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func PublicKey() (*ssh.PublicKeys, error) {
	var publicKey *ssh.PublicKeys
	sshPath := os.Getenv("HOME") + "/.ssh/id_ed25519"
	sshKey, _ := os.ReadFile(sshPath)
	publicKey, err := ssh.NewPublicKeys("git", sshKey, "")
	if err != nil {
		return nil, err
	}
	return publicKey, err
}

func CloneRepo(url string) (string, error) {
	repoName := strings.TrimSuffix(url[strings.LastIndex(url, "/")+1:], ".git")

	if strings.HasPrefix(url, "git@github.com") {
		auth, keyErr := PublicKey()
		if keyErr != nil {
			return repoName, keyErr
		}

		_, err := git.PlainClone(repoName, false, &git.CloneOptions{
			URL:      url,
			Progress: os.Stdout,
			Auth:     auth,
		})

		return repoName, err
	}

	_, err := git.PlainClone(repoName, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	return repoName, err
}

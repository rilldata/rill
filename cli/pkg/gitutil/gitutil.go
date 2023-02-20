package gitutil

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	exec "golang.org/x/sys/execabs"
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

func getAuth(url string) *http.BasicAuth {
	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("url=%s\n", url))
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("'git credential fill' failed: %v\n", err)
	}

	var username, password string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		frags := strings.SplitN(line, "=", 2)
		if len(frags) != 2 {
			continue // Ignore unrecognized response lines.
		}
		switch strings.TrimSpace(frags[0]) {
		case "username":
			username = frags[1]
		case "password":
			password = frags[1]
		}
	}

	authOpts := &http.BasicAuth{
		Username: username,
		Password: password,
	}

	return authOpts
}

func CloneRepo(url string) (string, error) {
	endpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}

	repoName := strings.TrimSuffix(endpoint.Path[strings.LastIndex(endpoint.Path, "/")+1:], ".git")
	fmt.Printf("Cloning into '%s'...\n", endpoint.String())

	if endpoint.Protocol == "ssh" {
		auth, keyErr := PublicKey()
		if keyErr != nil {
			return repoName, keyErr
		}

		_, err = git.PlainClone(repoName, false, &git.CloneOptions{
			URL:      endpoint.String(),
			Progress: os.Stdout,
			Auth:     auth,
		})

		return repoName, err
	}

	auth := getAuth(endpoint.String())
	_, err = git.PlainClone(repoName, false, &git.CloneOptions{
		URL:      endpoint.String(),
		Progress: os.Stdout,
		Auth:     auth,
	})

	return repoName, err
}

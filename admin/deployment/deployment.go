package deployment

import (
	"os/exec"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

type LocalDeployment struct {
	command *exec.Cmd
	Logger  *zap.Logger
}

func (l *LocalDeployment) DeployProject(project *database.Project) error {
	app := "rill"

	arg0 := "start"
	// remove username and pwd since no support for local runtime 
	ep, _ := transport.NewEndpoint(project.GitURL)
	ep.User = ""
	ep.Password = ""
	arg1 := ep.String()

	l.command = exec.Command(app, arg0, arg1)
	go func() {
		err := l.command.Run()
		if err != nil {
			l.Logger.Error("error in deploying ", zap.Error(err))
		}
	}()
	return nil
}

func (l *LocalDeployment) Close() error {
	return l.command.Process.Kill()
}

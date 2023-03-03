package deployment

import (
	"os/exec"

	"github.com/rilldata/rill/admin/database"
)

type LocalDeployment struct {
	command *exec.Cmd
}

func (l *LocalDeployment) DeployProject(project *database.Project) error {
	app := "rill"

	arg0 := "start"
	arg1 := project.GitURL

	l.command = exec.Command(app, arg0, arg1)
	go func() {
		_ = l.command.Run()
	}()
	return nil
}

func (l *LocalDeployment) Close() error {
	return l.command.Process.Kill()
}

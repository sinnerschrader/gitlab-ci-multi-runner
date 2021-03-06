package docker

import (
	"errors"

	"github.com/docker/docker/api/types"

	"gitlab.com/gitlab-org/gitlab-ci-multi-runner/common"
	"gitlab.com/gitlab-org/gitlab-ci-multi-runner/executors"
	"gitlab.com/gitlab-org/gitlab-ci-multi-runner/helpers/ssh"
)

type sshExecutor struct {
	executor
	sshCommand ssh.Client
}

func (s *sshExecutor) Prepare(options common.ExecutorPrepareOptions) error {
	err := s.executor.Prepare(options)
	if err != nil {
		return err
	}

	if s.Config.SSH == nil {
		return errors.New("Missing SSH configuration")
	}

	s.Debugln("Starting SSH command...")

	// Start build container which will run actual build
	container, err := s.createContainer("build", s.Build.Image, []string{}, []string{})
	if err != nil {
		return err
	}

	s.Debugln("Starting container", container.ID, "...")
	err = s.client.ContainerStart(s.Context, container.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	containerData, err := s.client.ContainerInspect(s.Context, container.ID)
	if err != nil {
		return err
	}

	// Create SSH command
	s.sshCommand = ssh.Client{
		Config: *s.Config.SSH,
		Stdout: s.Trace,
		Stderr: s.Trace,
	}
	s.sshCommand.Host = containerData.NetworkSettings.IPAddress

	s.Debugln("Connecting to SSH server...")
	err = s.sshCommand.Connect()
	if err != nil {
		return err
	}
	return nil
}

func (s *sshExecutor) Run(cmd common.ExecutorCommand) error {
	s.SetCurrentStage(DockerExecutorStageRun)

	err := s.sshCommand.Run(cmd.Context, ssh.Command{
		Environment: s.BuildShell.Environment,
		Command:     s.BuildShell.GetCommandWithArguments(),
		Stdin:       cmd.Script,
	})
	if _, ok := err.(*ssh.ExitError); ok {
		err = &common.BuildError{Inner: err}
	}
	return err
}

func (s *sshExecutor) Cleanup() {
	s.sshCommand.Cleanup()
	s.executor.Cleanup()
}

func init() {
	options := executors.ExecutorOptions{
		DefaultBuildsDir: "builds",
		SharedBuildsDir:  false,
		Shell: common.ShellScriptInfo{
			Shell:         "bash",
			Type:          common.LoginShell,
			RunnerCommand: "gitlab-runner",
		},
		ShowHostname: true,
	}

	creator := func() common.Executor {
		e := &sshExecutor{
			executor: executor{
				AbstractExecutor: executors.AbstractExecutor{
					ExecutorOptions: options,
				},
			},
		}
		e.SetCurrentStage(common.ExecutorStageCreated)
		return e
	}

	featuresUpdater := func(features *common.FeaturesInfo) {
		features.Variables = true
		features.Image = true
		features.Services = true
	}

	common.RegisterExecutor("docker-ssh", executors.DefaultExecutorProvider{
		Creator:         creator,
		FeaturesUpdater: featuresUpdater,
	})
}

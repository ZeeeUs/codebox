package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ZeeeUs/codebox/internal/config"
	"github.com/ZeeeUs/codebox/internal/docker"
	"github.com/ZeeeUs/codebox/internal/menu"
)

type App struct {
	args   []string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func New(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) App {
	return App{args: args, stdin: stdin, stdout: stdout, stderr: stderr}
}

func (a App) Run() error {
	fmt.Fprintln(a.stdout, "Codebox starting...")

	flags := flag.NewFlagSet("codebox", flag.ContinueOnError)
	flags.SetOutput(a.stderr)

	codexVersion := flags.String("codex-version", "", "override codebox-codex image version")
	if err := flags.Parse(a.args); err != nil {
		return err
	}

	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load(projectDir)
	if err != nil {
		return err
	}

	if *codexVersion != "" {
		cfg.Codex.Version = *codexVersion
	}

	if err := config.Validate(cfg); err != nil {
		return err
	}

	if err := docker.ValidateDocker(); err != nil {
		return errors.New("docker is not available")
	}

	runner := docker.NewRunner(os.Stdin, os.Stdout, os.Stderr)
	runtime := menu.Runtime{Reader: a.stdin, Writer: a.stdout}

	action, err := runtime.Run(projectDir, &cfg, runner)
	if err != nil {
		return err
	}

	if action == menu.ActionRestart {
		fmt.Fprintln(a.stdout, "Configuration updated. Restarting container is required.")
	}

	return nil
}

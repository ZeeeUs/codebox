package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

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

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "devel"
	}

	return info.Main.Version
}

func (a App) Run() error {
	flags := flag.NewFlagSet("codebox", flag.ContinueOnError)
	flags.SetOutput(a.stderr)

	showVersion := flags.Bool("version", false, "print codebox version")
	codexVersion := flags.String("codex-version", "", "override codebox-codex image version")
	if err := flags.Parse(a.args); err != nil {
		return err
	}

	if *showVersion {
		fmt.Fprintf(a.stdout, "codebox %s\n", buildVersion())
		return nil
	}

	fmt.Fprintln(a.stdout, "Codebox starting...")
	notifyCodeboxUpdate(a.stdout)

	projectDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, err := config.Load()
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

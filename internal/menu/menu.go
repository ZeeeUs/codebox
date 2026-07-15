package menu

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/ZeeeUs/codebox/internal/config"
	"github.com/ZeeeUs/codebox/internal/docker"
	"github.com/ZeeeUs/codebox/internal/paths"
	"github.com/ZeeeUs/codebox/internal/skills"
)

type Action int

const (
	ActionNone Action = iota
	ActionExit
	ActionRestart
)

type Runtime struct {
	Reader io.Reader
	Writer io.Writer
}

func (r Runtime) Run(projectDir string, cfg *config.Config, runner docker.Runner) (Action, error) {
	scanner := bufio.NewScanner(r.Reader)

	for {
		r.printMenu()
		fmt.Fprint(r.Writer, "> ")
		if !scanner.Scan() {
			return ActionExit, scanner.Err()
		}

		switch strings.TrimSpace(scanner.Text()) {
		case "1":
			if err := runner.Codex(projectDir, *cfg); err != nil {
				return ActionNone, err
			}
		case "2":
			return r.addMount(projectDir, cfg, scanner)
		case "3":
			if err := r.installSkill(scanner); err != nil {
				return ActionNone, err
			}
		case "4":
			if err := r.listSkills(); err != nil {
				return ActionNone, err
			}
		case "5":
			if err := runner.Shell(projectDir, *cfg); err != nil {
				return ActionNone, err
			}
		case "0":
			return ActionExit, nil
		default:
			fmt.Fprintln(r.Writer, "Unknown option")
		}
	}
}

func (r Runtime) printMenu() {
	fmt.Fprintln(r.Writer, "1. Run Codex")
	fmt.Fprintln(r.Writer, "2. Mount directory")
	fmt.Fprintln(r.Writer, "3. Install skill")
	fmt.Fprintln(r.Writer, "4. List skills")
	fmt.Fprintln(r.Writer, "5. Shell")
	fmt.Fprintln(r.Writer, "0. Exit")
}

func (r Runtime) addMount(projectDir string, cfg *config.Config, scanner *bufio.Scanner) (Action, error) {
	fmt.Fprint(r.Writer, "Path: ")
	if !scanner.Scan() {
		return ActionExit, scanner.Err()
	}

	hostPath, err := paths.ExpandPath(strings.TrimSpace(scanner.Text()))
	if err != nil {
		return ActionNone, err
	}

	target := docker.MountTargetForPath(hostPath)
	for _, mount := range cfg.Mounts {
		if mount.Target == target {
			return ActionNone, fmt.Errorf("mount target conflict: %s", target)
		}
	}

	cfg.Mounts = append(cfg.Mounts, config.Mount{
		Source: hostPath,
		Target: target,
		Mode:   "rw",
	})

	if err := config.Save(projectDir, *cfg); err != nil {
		return ActionNone, err
	}

	fmt.Fprintf(r.Writer, "Mount added: %s -> %s\n", hostPath, target)
	return ActionRestart, nil
}

func (r Runtime) installSkill(scanner *bufio.Scanner) error {
	fmt.Fprint(r.Writer, "Skill path: ")
	if !scanner.Scan() {
		return scanner.Err()
	}

	name, err := skills.Install(strings.TrimSpace(scanner.Text()))
	if err != nil {
		return err
	}

	fmt.Fprintf(r.Writer, "Installed skill: %s\n", name)
	return nil
}

func (r Runtime) listSkills() error {
	names, err := skills.List()
	if err != nil {
		return err
	}

	if len(names) == 0 {
		fmt.Fprintln(r.Writer, "No skills installed")
		return nil
	}

	for _, name := range names {
		fmt.Fprintln(r.Writer, name)
	}

	return nil
}

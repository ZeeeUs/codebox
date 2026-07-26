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

type menuItem struct {
	label    string
	selected bool
}

func (r Runtime) Run(projectDir string, cfg *config.Config, runner docker.Runner) (Action, error) {
	items := []menuItem{
		{label: "Run Codex"},
		{label: "Mount directory"},
		{label: "Install skill"},
		{label: "List skills"},
		{label: "Shell"},
		{label: "Configure languages"},
		{label: "Codex permissions"},
		{label: "Exit"},
	}

	for {
		choice, back, err := r.selectItem("Codebox", items, false, 0)
		if err != nil {
			return ActionExit, err
		}
		if back || choice == 7 {
			return ActionExit, nil
		}

		scanner := bufio.NewScanner(r.Reader)
		switch choice {
		case 0:
			if err := runner.Codex(projectDir, *cfg); err != nil {
				return ActionNone, err
			}
		case 1:
			return r.addMount(cfg, scanner)
		case 2:
			if err := r.installSkill(scanner); err != nil {
				return ActionNone, err
			}
			if err := r.waitForBack(); err != nil {
				return ActionNone, err
			}
		case 3:
			if err := r.listSkills(); err != nil {
				return ActionNone, err
			}
			if err := r.waitForBack(); err != nil {
				return ActionNone, err
			}
		case 4:
			if err := runner.Shell(projectDir, *cfg); err != nil {
				return ActionNone, err
			}
		case 5:
			if err := r.configureLanguages(cfg); err != nil {
				return ActionNone, err
			}
		case 6:
			if err := r.configureCodexPermissions(cfg); err != nil {
				return ActionNone, err
			}
		}
	}
}

func (r Runtime) addMount(cfg *config.Config, scanner *bufio.Scanner) (Action, error) {
	r.clear()
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

	cfg.Mounts = append(cfg.Mounts, config.Mount{Source: hostPath, Target: target, Mode: "rw"})
	if err := config.Save(*cfg); err != nil {
		return ActionNone, err
	}

	fmt.Fprintf(r.Writer, "Mount added: %s -> %s\n", hostPath, target)
	return ActionRestart, nil
}

func (r Runtime) installSkill(scanner *bufio.Scanner) error {
	r.clear()
	fmt.Fprint(r.Writer, "Skill path: ")
	if !scanner.Scan() {
		return scanner.Err()
	}

	name, err := skills.Install(strings.TrimSpace(scanner.Text()))
	if err != nil {
		return err
	}

	fmt.Fprintf(r.Writer, "Installed skill: %s\n\nPress Esc to go back", name)
	return nil
}

func (r Runtime) listSkills() error {
	r.clear()
	names, err := skills.List()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Fprintln(r.Writer, "No skills installed")
	} else {
		for _, name := range names {
			fmt.Fprintln(r.Writer, name)
		}
	}
	fmt.Fprintln(r.Writer, "\nPress Esc to go back")
	return nil
}

func (r Runtime) configureLanguages(cfg *config.Config) error {
	items := []menuItem{
		{label: "Go", selected: contains(cfg.Languages, "go")},
		{label: "Rust", selected: contains(cfg.Languages, "rust")},
		{label: "Save"},
	}

	cursor := 0
	for {
		choice, back, err := r.selectItem("Languages", items, true, cursor)
		if err != nil || back {
			return err
		}
		if choice < 2 {
			items[choice].selected = !items[choice].selected
			cursor = choice
			continue
		}

		cfg.Languages = cfg.Languages[:0]
		for i, language := range []string{"go", "rust"} {
			if items[i].selected {
				cfg.Languages = append(cfg.Languages, language)
			}
		}
		return config.Save(*cfg)
	}
}

func (r Runtime) configureCodexPermissions(cfg *config.Config) error {
	items := []menuItem{
		{label: "Ask when needed (workspace write)"},
		{label: "No prompts (workspace write)"},
		{label: "No prompts (full container access)"},
	}
	current := permissionProfileIndex(cfg.Codex)
	items[current].selected = true

	choice, back, err := r.selectItem("Codex permissions", items, true, current)
	if err != nil || back {
		return err
	}

	switch choice {
	case 0:
		cfg.Codex.ApprovalPolicy = "on-request"
		cfg.Codex.SandboxMode = "workspace-write"
	case 1:
		cfg.Codex.ApprovalPolicy = "never"
		cfg.Codex.SandboxMode = "workspace-write"
	case 2:
		cfg.Codex.ApprovalPolicy = "never"
		cfg.Codex.SandboxMode = "danger-full-access"
	}
	return config.Save(*cfg)
}

func (r Runtime) selectItem(title string, items []menuItem, showMarks bool, cursor int) (int, bool, error) {
	restore, err := makeRaw(r.Reader)
	if err != nil {
		return 0, false, err
	}
	defer restore()

	for {
		r.clear()
		fmt.Fprintln(r.Writer, title)
		fmt.Fprintln(r.Writer)
		for i, item := range items {
			pointer := "  "
			if i == cursor {
				pointer = "> "
			}
			mark := ""
			if showMarks && item.label != "Save" {
				mark = "[ ] "
				if item.selected {
					mark = "[x] "
				}
			}
			fmt.Fprintf(r.Writer, "%s%s%s\n", pointer, mark, item.label)
		}
		fmt.Fprintln(r.Writer, "Up/Down move  Space/Enter select  Esc back")

		key, err := readKey(r.Reader)
		if err != nil {
			return 0, false, err
		}
		switch key {
		case keyUp:
			cursor = (cursor - 1 + len(items)) % len(items)
		case keyDown:
			cursor = (cursor + 1) % len(items)
		case keySelect:
			return cursor, false, nil
		case keyEscape:
			return 0, true, nil
		}
	}
}

func (r Runtime) waitForBack() error {
	restore, err := makeRaw(r.Reader)
	if err != nil {
		return err
	}
	defer restore()

	for {
		key, err := readKey(r.Reader)
		if err != nil {
			return err
		}
		if key == keyEscape || key == keySelect {
			return nil
		}
	}
}

func (r Runtime) clear() {
	fmt.Fprint(r.Writer, "\x1b[2J\x1b[H")
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func permissionProfileIndex(cfg config.CodexConfig) int {
	if cfg.ApprovalPolicy == "never" && cfg.SandboxMode == "danger-full-access" {
		return 2
	}
	if cfg.ApprovalPolicy == "never" {
		return 1
	}
	return 0
}

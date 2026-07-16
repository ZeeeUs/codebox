package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const FileName = ".codebox.yaml"

type Config struct {
	Agent     string        `yaml:"agent"`
	Codex     CodexConfig   `yaml:"codex"`
	Project   ProjectConfig `yaml:"project"`
	Languages []string      `yaml:"languages"`
	Mounts    []Mount       `yaml:"mounts"`
	Skills    []Skill       `yaml:"skills"`
}

type CodexConfig struct {
	Version        string `yaml:"version"`
	ApprovalPolicy string `yaml:"approvalPolicy"`
	SandboxMode    string `yaml:"sandboxMode"`
}

type ProjectConfig struct {
	MountPath string `yaml:"mountPath"`
}

type Mount struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
	Mode   string `yaml:"mode"`
}

type Skill struct {
	Name string `yaml:"name"`
}

func Default() Config {
	return Config{
		Codex: CodexConfig{
			Version:        "latest",
			ApprovalPolicy: "on-request",
			SandboxMode:    "workspace-write",
		},
		Project:   ProjectConfig{MountPath: "/workspace"},
		Languages: []string{},
		Mounts:    []Mount{},
		Skills:    []Skill{},
	}
}

func Load(dir string) (Config, error) {
	path := filepath.Join(dir, FileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}

		return Config{}, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	cfg.applyDefaults()
	return cfg, nil
}

func Save(dir string, cfg Config) error {
	cfg.applyDefaults()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, FileName), data, 0o644)
}

func Validate(cfg Config) error {
	cfg.applyDefaults()

	if cfg.Codex.Version == "" {
		return errors.New("codex.version must not be empty")
	}

	if cfg.Project.MountPath == "" {
		return errors.New("project.mountPath must not be empty")
	}

	if cfg.Codex.ApprovalPolicy != "on-request" && cfg.Codex.ApprovalPolicy != "never" {
		return errors.New("codex.approvalPolicy must be on-request or never")
	}

	if cfg.Codex.SandboxMode != "workspace-write" && cfg.Codex.SandboxMode != "danger-full-access" {
		return errors.New("codex.sandboxMode must be workspace-write or danger-full-access")
	}

	for _, mount := range cfg.Mounts {
		if mount.Source == "" || mount.Target == "" {
			return errors.New("mount source and target must not be empty")
		}

		if mount.Mode != "ro" && mount.Mode != "rw" {
			return errors.New("mount mode must be ro or rw")
		}
	}

	seenLanguages := make(map[string]struct{}, len(cfg.Languages))
	for _, language := range cfg.Languages {
		if language != "go" && language != "rust" {
			return errors.New("language must be go or rust")
		}
		if _, ok := seenLanguages[language]; ok {
			return errors.New("languages must not contain duplicates")
		}
		seenLanguages[language] = struct{}{}
	}

	for _, skill := range cfg.Skills {
		if skill.Name == "" {
			return errors.New("skill name must not be empty")
		}
	}

	return nil
}

func (c *Config) applyDefaults() {
	if c.Codex.Version == "" {
		c.Codex.Version = "latest"
	}

	if c.Project.MountPath == "" {
		c.Project.MountPath = "/workspace"
	}

	if c.Codex.ApprovalPolicy == "" {
		c.Codex.ApprovalPolicy = "on-request"
	}

	if c.Codex.SandboxMode == "" {
		c.Codex.SandboxMode = "workspace-write"
	}

	if c.Mounts == nil {
		c.Mounts = []Mount{}
	}

	if c.Languages == nil {
		c.Languages = []string{}
	}

	if c.Skills == nil {
		c.Skills = []Skill{}
	}
}

package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const FileName = ".codebox.yaml"

type Config struct {
	Agent   string        `yaml:"agent"`
	Codex   CodexConfig   `yaml:"codex"`
	Project ProjectConfig `yaml:"project"`
	Mounts  []Mount       `yaml:"mounts"`
	Skills  []Skill       `yaml:"skills"`
}

type CodexConfig struct {
	Version string `yaml:"version"`
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
		Codex: CodexConfig{Version: "latest"},
		Project: ProjectConfig{MountPath: "/workspace"},
		Mounts:  []Mount{},
		Skills:  []Skill{},
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

	for _, mount := range cfg.Mounts {
		if mount.Source == "" || mount.Target == "" {
			return errors.New("mount source and target must not be empty")
		}

		if mount.Mode != "ro" && mount.Mode != "rw" {
			return errors.New("mount mode must be ro or rw")
		}
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

	if c.Mounts == nil {
		c.Mounts = []Mount{}
	}

	if c.Skills == nil {
		c.Skills = []Skill{}
	}
}

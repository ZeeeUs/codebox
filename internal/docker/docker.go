package docker

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ZeeeUs/codebox/internal/config"
	"github.com/ZeeeUs/codebox/internal/paths"
)

const runtimeDockerfile = `FROM golang:1.25-bookworm

ENV CGO_ENABLED=1
ENV PATH="/root/go/bin:${PATH}"

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    build-essential \
    ca-certificates \
    cargo \
    curl \
    git \
    gdb \
    less \
    nodejs \
    npm \
    procps \
    rustc \
    && rm -rf /var/lib/apt/lists/*

RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

RUN npm i -g @openai/codex@latest \
    && codex --version

WORKDIR /workspace

CMD ["bash"]
`

type Runner struct {
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

func NewRunner(stdin *os.File, stdout *os.File, stderr *os.File) Runner {
	return Runner{Stdin: stdin, Stdout: stdout, Stderr: stderr}
}

func BuildMounts(projectDir string, extra []config.Mount) ([]string, error) {
	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}

	mounts := []string{fmt.Sprintf("%s:/workspace:rw", projectDir)}
	for _, mount := range extra {
		source, err := paths.ExpandPath(mount.Source)
		if err != nil {
			return nil, err
		}

		if _, err := os.Stat(source); err != nil {
			return nil, fmt.Errorf("validate mount %q: %w", mount.Source, err)
		}

		mode := mount.Mode
		if mode == "" {
			mode = "rw"
		}

		if mode != "ro" && mode != "rw" {
			return nil, fmt.Errorf("invalid mount mode %q", mode)
		}

		target := mount.Target
		if target == "" {
			target = MountTargetForPath(source)
		}

		mounts = append(mounts, fmt.Sprintf("%s:%s:%s", source, target, mode))
	}

	return mounts, nil
}

func ValidateDocker() error {
	cmd := exec.Command("docker", "version")
	return cmd.Run()
}

func (r Runner) EnsureImage(cfg config.Config) error {
	image := imageName(cfg)
	if r.imageExists(image) {
		return nil
	}

	fmt.Fprintf(r.Stdout, "Docker image %s not found. Building...\n", image)
	buildDir, err := os.MkdirTemp("", "codebox-docker-build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(buildDir)

	if err := os.WriteFile(filepath.Join(buildDir, "Dockerfile"), []byte(runtimeDockerfile), 0o644); err != nil {
		return err
	}

	cmd := exec.Command("docker", "build", "-t", image, buildDir)
	cmd.Stdin = r.Stdin
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	return cmd.Run()
}

func (r Runner) RunContainer(projectDir string, cfg config.Config, command ...string) error {
	if err := r.EnsureImage(cfg); err != nil {
		return err
	}

	mounts, err := BuildMounts(projectDir, cfg.Mounts)
	if err != nil {
		return err
	}

	args := []string{"run", "--rm", "-it"}
	for _, mount := range mounts {
		args = append(args, "-v", mount)
	}

	args = append(args, "-w", cfg.Project.MountPath, imageName(cfg))
	args = append(args, command...)

	cmd := exec.Command("docker", args...)
	cmd.Stdin = r.Stdin
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	return cmd.Run()
}

func (r Runner) Shell(projectDir string, cfg config.Config) error {
	return r.RunContainer(projectDir, cfg, "bash")
}

func (r Runner) Codex(projectDir string, cfg config.Config) error {
	return r.RunContainer(projectDir, cfg, "codex")
}

func MountTargetForPath(hostPath string) string {
	return filepath.ToSlash(filepath.Join("/mnt", paths.Basename(strings.TrimSpace(hostPath))))
}

func imageName(cfg config.Config) string {
	return fmt.Sprintf("codebox-codex:%s", cfg.Codex.Version)
}

func (r Runner) imageExists(image string) bool {
	cmd := exec.Command("docker", "image", "inspect", image)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

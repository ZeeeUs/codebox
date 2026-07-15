package docker

import (
	_ "embed"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ZeeeUs/codebox/internal/config"
	"github.com/ZeeeUs/codebox/internal/paths"
)

const runtimeVersion = "2"

//go:embed runtime.Dockerfile
var runtimeDockerfile string

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
	if !r.imageExists(image) {
		fmt.Fprintf(r.Stdout, "Docker image %s not found. Building...\n", image)
		return r.buildImage(image, cfg.Codex.Version, cfg.Codex.Version == "latest")
	}

	if !r.runtimeImageCurrent(image) {
		fmt.Fprintf(r.Stdout, "Docker runtime image %s is outdated. Rebuilding...\n", image)
		return r.buildImage(image, cfg.Codex.Version, false)
	}

	if cfg.Codex.Version != "latest" {
		return nil
	}

	installedVersion, err := r.imageCodexVersion(image)
	if err != nil {
		fmt.Fprintf(r.Stderr, "Unable to check installed Codex version: %v\n", err)
		return nil
	}

	latestVersion, err := r.latestCodexVersion(image)
	if err != nil {
		fmt.Fprintf(r.Stderr, "Unable to check latest Codex version: %v\n", err)
		return nil
	}

	if installedVersion == latestVersion {
		fmt.Fprintf(r.Stdout, "Codex is up to date (%s).\n", installedVersion)
		return nil
	}

	fmt.Fprintf(r.Stdout, "Codex update available: %s -> %s. Update? [y/N] ", installedVersion, latestVersion)
	answer, err := bufio.NewReader(r.Stdin).ReadString('\n')
	if err != nil && len(answer) == 0 {
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		fmt.Fprintf(r.Stdout, "Updating Docker image %s...\n", image)
		return r.buildImage(image, latestVersion, false)
	default:
		fmt.Fprintln(r.Stdout, "Codex update skipped.")
		return nil
	}
}

func (r Runner) buildImage(image, version string, noCache bool) error {
	buildDir, err := os.MkdirTemp("", "codebox-docker-build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(buildDir)

	if err := os.WriteFile(filepath.Join(buildDir, "Dockerfile"), []byte(runtimeDockerfile), 0o644); err != nil {
		return err
	}

	args := []string{"build", "--pull"}
	if noCache {
		args = append(args, "--no-cache")
	}

	args = append(
		args,
		"--build-arg", "CODEX_VERSION="+version,
		"-t", image,
		buildDir,
	)

	cmd := exec.Command("docker", args...)
	cmd.Stdin = r.Stdin
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	return cmd.Run()
}

func (r Runner) imageCodexVersion(image string) (string, error) {
	return commandVersion("docker", "run", "--rm", image, "codex", "--version")
}

func (r Runner) latestCodexVersion(image string) (string, error) {
	return commandVersion("docker", "run", "--rm", image, "npm", "view", "@openai/codex", "version")
}

func commandVersion(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s: %w", strings.TrimSpace(string(exitErr.Stderr)), err)
		}

		return "", err
	}

	fields := strings.Fields(string(output))
	if len(fields) == 0 {
		return "", fmt.Errorf("empty version output")
	}

	return strings.TrimPrefix(fields[len(fields)-1], "v"), nil
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

func (r Runner) runtimeImageCurrent(image string) bool {
	cmd := exec.Command(
		"docker", "image", "inspect",
		"--format", "{{ index .Config.Labels \"io.codebox.runtime-version\" }}",
		image,
	)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == runtimeVersion
}

func (r Runner) imageExists(image string) bool {
	cmd := exec.Command("docker", "image", "inspect", image)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

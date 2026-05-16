package skills

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func Root() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".codebox", "skills"), nil
}

func Install(sourceDir string) (string, error) {
	info, err := os.Stat(sourceDir)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", sourceDir)
	}

	root, err := Root()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}

	name := filepath.Base(filepath.Clean(sourceDir))
	targetDir := filepath.Join(root, name)
	if err := os.RemoveAll(targetDir); err != nil {
		return "", err
	}

	if err := copyDir(sourceDir, targetDir); err != nil {
		return "", err
	}

	return name, nil
}

func List() ([]string, error) {
	root, err := Root()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	sort.Strings(names)
	return names, nil
}

func copyDir(sourceDir string, targetDir string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath, info.Mode())
	})
}

func copyFile(sourcePath string, targetPath string, mode os.FileMode) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	return err
}

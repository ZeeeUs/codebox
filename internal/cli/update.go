package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const latestReleaseURL = "https://api.github.com/repos/ZeeeUs/codebox/releases/latest"

var updateClient = &http.Client{Timeout: 2 * time.Second}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func notifyCodeboxUpdate(writer io.Writer) {
	current := buildVersion()
	if current == "devel" {
		return
	}

	latest, err := fetchLatestCodeboxVersion()
	if err != nil || !newerVersion(latest, current) {
		return
	}

	fmt.Fprintf(
		writer,
		"A new Codebox version is available: %s -> %s. Update with: go install github.com/ZeeeUs/codebox@%s\n",
		current,
		latest,
		latest,
	)
}

func fetchLatestCodeboxVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "codebox")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := updateClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub releases API returned %s", resp.Status)
	}

	var release githubRelease
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 1<<20))
	if err := decoder.Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("latest release has no tag")
	}

	return release.TagName, nil
}

func newerVersion(candidate, current string) bool {
	candidateParts, ok := parseVersion(candidate)
	if !ok {
		return false
	}
	currentParts, ok := parseVersion(current)
	if !ok {
		return false
	}

	for i := range candidateParts {
		if candidateParts[i] != currentParts[i] {
			return candidateParts[i] > currentParts[i]
		}
	}

	return false
}

func parseVersion(version string) ([3]int, bool) {
	var result [3]int
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	version = strings.SplitN(version, "-", 2)[0]
	parts := strings.Split(version, ".")
	if len(parts) != len(result) {
		return result, false
	}

	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return result, false
		}
		result[i] = value
	}

	return result, true
}

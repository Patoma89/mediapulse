package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// LatestReleaseAPIURL is queried only when the user explicitly clicks "Check for updates".
	LatestReleaseAPIURL = "https://api.github.com/repos/Patoma89/mediapulse/releases/latest"
	// ReleasesPageURL is the fallback browser link to MediaPulse releases.
	ReleasesPageURL = "https://github.com/Patoma89/mediapulse/releases"
)

// UpdateCheckResult describes the result of a manual opt-in update check.
type UpdateCheckResult struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleaseURL      string `json:"releaseUrl"`
	ReleaseName     string `json:"releaseName,omitempty"`
	PublishedAt     string `json:"publishedAt,omitempty"`
	ReleaseNotes    string `json:"releaseNotes,omitempty"`
	CheckedAt       string `json:"checkedAt"`
	Error           string `json:"error,omitempty"`
}

type githubReleasePayload struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	HTMLURL     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
	Body        string `json:"body"`
}

// CheckLatest queries GitHub's latest release endpoint for MediaPulse.
func CheckLatest(ctx context.Context, currentVersion string) UpdateCheckResult {
	return CheckURL(ctx, currentVersion, LatestReleaseAPIURL, nil)
}

// CheckURL queries a release JSON endpoint (useful for testing with httptest).
func CheckURL(ctx context.Context, currentVersion, endpointURL string, client *http.Client) UpdateCheckResult {
	curClean := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(currentVersion), "v"), "V")
	result := UpdateCheckResult{
		CurrentVersion: curClean,
		ReleaseURL:     ReleasesPageURL,
		CheckedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 12*time.Second)
		defer cancel()
	}

	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "MediaPulse/"+curClean)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("no se pudo conectar con GitHub: %v", err)
		return result
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		result.Error = "no se encontraron releases publicadas (HTTP 404)"
		return result
	case http.StatusForbidden, http.StatusTooManyRequests:
		result.Error = fmt.Sprintf("limite de consultas de GitHub alcanzado (HTTP %d)", resp.StatusCode)
		return result
	default:
		result.Error = fmt.Sprintf("GitHub devolvio HTTP %d", resp.StatusCode)
		return result
	}

	var payload githubReleasePayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		result.Error = fmt.Sprintf("respuesta invalida de GitHub: %v", err)
		return result
	}

	latest := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(payload.TagName), "v"), "V")
	if latest == "" {
		result.Error = "la release no incluye numero de version"
		return result
	}

	result.LatestVersion = latest
	result.UpdateAvailable = CompareVersions(curClean, latest) < 0
	if strings.TrimSpace(payload.HTMLURL) != "" {
		result.ReleaseURL = strings.TrimSpace(payload.HTMLURL)
	}
	result.ReleaseName = strings.TrimSpace(payload.Name)
	result.PublishedAt = strings.TrimSpace(payload.PublishedAt)
	result.ReleaseNotes = truncateRunes(strings.TrimSpace(payload.Body), 500)
	return result
}

// CompareVersions compares two semver strings, returning -1 if a < b, 0 if a == b, 1 if a > b.
func CompareVersions(a, b string) int {
	aParts, aPre := parseVersion(a)
	bParts, bPre := parseVersion(b)

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}
	for i := 0; i < maxLen; i++ {
		av := 0
		if i < len(aParts) {
			av = aParts[i]
		}
		bv := 0
		if i < len(bParts) {
			bv = bParts[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}

	if aPre == bPre {
		return 0
	}
	if aPre == "" {
		return 1
	}
	if bPre == "" {
		return -1
	}
	if aPre < bPre {
		return -1
	}
	return 1
}

func parseVersion(raw string) ([]int, string) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	if idx := strings.IndexByte(s, '+'); idx >= 0 {
		s = s[:idx]
	}

	pre := ""
	if idx := strings.IndexByte(s, '-'); idx >= 0 {
		pre = strings.TrimSpace(s[idx+1:])
		s = s[:idx]
	}

	rawParts := strings.Split(s, ".")
	nums := make([]int, 3)
	for i := 0; i < len(rawParts); i++ {
		n, err := strconv.Atoi(strings.TrimSpace(rawParts[i]))
		if err != nil {
			n = 0
		}
		if i < len(nums) {
			nums[i] = n
		} else {
			nums = append(nums, n)
		}
	}
	return nums, pre
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return strings.TrimSpace(string(runes[:max])) + "..."
}

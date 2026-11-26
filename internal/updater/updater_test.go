package updater

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func mockClient(status int, body string, assertReq func(*http.Request)) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if assertReq != nil {
				assertReq(r)
			}
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}),
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.1.0", -1},
		{"1.2.0", "1.10.0", -1},
		{"v1.1.0", "1.1.0", 0},
		{"1.1.0-rc1", "1.1.0", -1},
		{"1.2.0", "1.1.0", 1},
		{"1.1.0+build5", "1.1.0", 0},
		{"1.1.0-alpha", "1.1.0-beta", -1},
	}

	for _, tc := range cases {
		got := CompareVersions(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCheckURLUpdateAvailable(t *testing.T) {
	longBody := strings.Repeat("a", 550)
	client := mockClient(http.StatusOK, `{
		"tag_name": "v1.2.0",
		"name": "MediaPulse v1.2.0",
		"html_url": "https://github.com/Patoma89/mediapulse/releases/tag/v1.2.0",
		"published_at": "2026-12-01T10:00:00Z",
		"body": "`+longBody+`"
	}`, func(r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("unexpected Accept header: %q", r.Header.Get("Accept"))
		}
	})

	res := CheckURL(context.Background(), "1.1.0", "https://api.github.com/repos/Patoma89/mediapulse/releases/latest", client)
	if res.Error != "" {
		t.Fatalf("unexpected error: %s", res.Error)
	}
	if !res.UpdateAvailable {
		t.Fatalf("expected UpdateAvailable=true")
	}
	if res.LatestVersion != "1.2.0" {
		t.Fatalf("LatestVersion = %q, want 1.2.0", res.LatestVersion)
	}
	if res.ReleaseURL != "https://github.com/Patoma89/mediapulse/releases/tag/v1.2.0" {
		t.Fatalf("unexpected ReleaseURL: %q", res.ReleaseURL)
	}
	if len([]rune(res.ReleaseNotes)) != 503 {
		t.Fatalf("expected truncated notes of 503 runes, got %d", len([]rune(res.ReleaseNotes)))
	}
}

func TestCheckURLUpToDateAndAhead(t *testing.T) {
	client := mockClient(http.StatusOK, `{"tag_name":"v1.1.0","html_url":"https://github.com/Patoma89/mediapulse/releases/tag/v1.1.0"}`, nil)

	same := CheckURL(context.Background(), "1.1.0", "https://api.github.com/repos/Patoma89/mediapulse/releases/latest", client)
	if same.Error != "" || same.UpdateAvailable {
		t.Fatalf("expected up-to-date result, got %+v", same)
	}

	ahead := CheckURL(context.Background(), "1.2.0", "https://api.github.com/repos/Patoma89/mediapulse/releases/latest", client)
	if ahead.Error != "" || ahead.UpdateAvailable {
		t.Fatalf("expected local ahead result without update, got %+v", ahead)
	}
}

func TestCheckURLErrors(t *testing.T) {
	for _, tc := range []struct {
		name     string
		status   int
		body     string
		wantSub  string
	}{
		{"not found", http.StatusNotFound, `{}`, "HTTP 404"},
		{"rate limit", http.StatusForbidden, `{}`, "HTTP 403"},
		{"invalid json", http.StatusOK, `{not-json`, "respuesta invalida"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := mockClient(tc.status, tc.body, nil)
			res := CheckURL(context.Background(), "1.1.0", "https://api.github.com/repos/Patoma89/mediapulse/releases/latest", client)
			if !strings.Contains(res.Error, tc.wantSub) {
				t.Fatalf("expected error containing %q for %s, got %q", tc.wantSub, tc.name, res.Error)
			}
			if res.ReleaseURL != ReleasesPageURL {
				t.Fatalf("expected fallback ReleaseURL %q, got %q", ReleasesPageURL, res.ReleaseURL)
			}
		})
	}
}

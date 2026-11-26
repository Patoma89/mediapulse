package main

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestExtractMediaArgPaths(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "clips")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	vid1 := filepath.Join(tmpDir, "video1.mp4")
	vid2 := filepath.Join(subDir, "clip two.mkv")
	selfExe := filepath.Join(tmpDir, "MediaPulse.exe")
	dashFile := filepath.Join(tmpDir, "-leading-dash.mp4")

	for _, f := range []string{vid1, vid2, selfExe, dashFile} {
		if err := os.WriteFile(f, []byte("test"), 0o644); err != nil {
			t.Fatalf("write %s failed: %v", f, err)
		}
	}

	selfArg := selfExe
	dupVid1 := vid1
	if runtime.GOOS == "windows" {
		selfArg = strings.ToUpper(selfExe)
		dupVid1 = strings.ToUpper(vid1)
	}

	args := []string{
		"",
		"   ",
		selfArg,
		"--dev",
		"-v",
		`"` + vid1 + `"`,
		`'clips` + string(filepath.Separator) + `clip two.mkv'`,
		subDir, // directory should be skipped
		filepath.Join(tmpDir, "nonexistent.mp4"),
		dupVid1, // duplicate should be skipped
		"-leading-dash.mp4",
	}

	got := extractMediaArgPaths(args, tmpDir, selfExe)
	want := []string{vid1, vid2, dashFile}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractMediaArgPaths mismatch:\n got: %v\nwant: %v", got, want)
	}
}

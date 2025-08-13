package version

import (
	"testing"
)

func TestGetVersion(t *testing.T) {
	t.Parallel()
	version := GetVersion()
	if version == "" {
		t.Error("Version should not be empty")
	}
}

func TestGetBuildInfo(t *testing.T) {
	t.Parallel()
	version, gitCommit, buildDate := GetBuildInfo()

	if version == "" {
		t.Error("Version should not be empty")
	}
	if gitCommit == "" {
		t.Error("GitCommit should not be empty")
	}
	if buildDate == "" {
		t.Error("BuildDate should not be empty")
	}
}

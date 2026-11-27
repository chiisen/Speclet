package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestVersionDoesNotRequireWorkingDirectory(t *testing.T) {
	var stdout, stderr bytes.Buffer
	getwdCalled := false
	code := run([]string{"--version"}, &stdout, &stderr, func() (string, error) {
		getwdCalled = true
		return "", errors.New("working directory unavailable")
	})
	if getwdCalled {
		t.Fatal("version command requested the current working directory")
	}
	if code != 0 || !strings.Contains(stdout.String(), "Speclet version: ") && !strings.Contains(stdout.String(), "Speclet 版本：") || stderr.Len() != 0 {
		t.Fatalf("run() = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
}

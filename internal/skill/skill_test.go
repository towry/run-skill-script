package skill

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFindPrefersExactLocalSkill(t *testing.T) {
	home := t.TempDir()
	cwd := t.TempDir()
	local := filepath.Join(cwd, ".agents", "skills", "git-jj")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(home, ".cache", "amp", "global-skills", "ampcode.com", "user", "git-jj@deadbeef")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Lookup{Home: home, Cwd: cwd}.Find("git-jj")
	if err != nil {
		t.Fatal(err)
	}
	if got != local {
		t.Fatalf("Find = %s, want %s", got, local)
	}
}

func TestFindCacheHashedSkill(t *testing.T) {
	home := t.TempDir()
	cache := filepath.Join(home, ".cache", "amp", "global-skills", "ampcode.com", "user", "git-jj@c4b90d11")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Lookup{Home: home, Cwd: t.TempDir()}.Find("git-jj")
	if err != nil {
		t.Fatal(err)
	}
	if got != cache {
		t.Fatalf("Find = %s, want %s", got, cache)
	}
}

func TestFindNewestHashedSkill(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(home, ".cache", "amp", "global-skills", "ampcode.com", "user")
	oldDir := filepath.Join(root, "git-jj@aaa")
	newDir := filepath.Join(root, "git-jj@bbb")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(oldDir)
	if err != nil {
		t.Fatal(err)
	}
	newer := info.ModTime().Add(2 * time.Second)
	if err := os.Chtimes(newDir, newer, newer); err != nil {
		t.Fatal(err)
	}
	got, err := Lookup{Home: home, Cwd: t.TempDir()}.Find("git-jj")
	if err != nil {
		t.Fatal(err)
	}
	if got != newDir {
		t.Fatalf("Find = %s, want %s", got, newDir)
	}
}

func TestLoadMetaMissing(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadMeta(dir)
	var noMeta NoMetaError
	if !errors.As(err, &noMeta) {
		t.Fatalf("LoadMeta error %v, want NoMetaError", err)
	}
	if !strings.Contains(err.Error(), "refusing to guess") {
		t.Fatalf("error %q should refuse to guess", err)
	}
}

func TestRunDeclaredBashScript(t *testing.T) {
	dir := t.TempDir()
	scriptDir := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "#!/usr/bin/env bash\nprintf '%s\\n' \"$(pwd)\" \"$1\"\n"
	if err := os.WriteFile(filepath.Join(scriptDir, "echo_cwd.sh"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	writeMeta(t, dir, Meta{Scripts: map[string]Script{
		"echo-cwd": {File: "scripts/echo_cwd.sh", Runtime: "bash"},
	}})
	meta, err := LoadMeta(dir)
	if err != nil {
		t.Fatal(err)
	}

	cwd := t.TempDir()
	t.Chdir(cwd)
	got := withStdout(t, func() {
		if err := Run(dir, meta, "echo-cwd", []string{"from-arg"}); err != nil {
			t.Fatal(err)
		}
	})
	want := cwd + "\nfrom-arg\n"
	if got != want {
		t.Fatalf("output %q, want %q", got, want)
	}
}

func TestRunUnknownScriptListsDeclared(t *testing.T) {
	dir := t.TempDir()
	writeMeta(t, dir, Meta{Scripts: map[string]Script{
		"repo-check": {File: "scripts/repo_check.sh", Runtime: "bash", Description: "Detect repo type"},
	}})
	meta, err := LoadMeta(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = Run(dir, meta, "missing", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "not declared in meta.json") || !strings.Contains(msg, "repo-check") {
		t.Fatalf("error %q should list declared scripts", msg)
	}
}

func TestRejectsPathEscapeInMetaFile(t *testing.T) {
	dir := t.TempDir()
	writeMeta(t, dir, Meta{Scripts: map[string]Script{
		"bad": {File: "../secret.sh", Runtime: "bash"},
	}})
	meta, err := LoadMeta(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := Run(dir, meta, "bad", nil); err == nil || !strings.Contains(err.Error(), "relative") {
		t.Fatalf("got %v, want relative-path error", err)
	}
}

func writeMeta(t *testing.T, dir string, meta Meta) {
	t.Helper()
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func withStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()
	fn()
	w.Close()
	os.Stdout = old
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

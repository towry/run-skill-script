package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/towry/run-skill-script/internal/skill"
)

func TestSkillHelpListsScripts(t *testing.T) {
	home := t.TempDir()
	skillDir := filepath.Join(home, ".config", "agents", "skills", "git-jj")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	meta := skill.Meta{Scripts: map[string]skill.Script{
		"repo-check": {File: "scripts/repo_check.sh", Runtime: "bash", Description: "Detect repo type"},
	}}
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "meta.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("AMP_WORKING_DIRECTORY", t.TempDir())
	t.Setenv("RUN_SKILL_SCRIPT_SKILLS", "")
	t.Chdir(t.TempDir())

	for _, args := range [][]string{
		{"git-jj"},
		{"git-jj", "--help"},
		{"git-jj", "-h"},
	} {
		out, code := captureMain(t, args)
		if code != 0 {
			t.Fatalf("%v exit %d, output %q", args, code, out)
		}
		if !strings.Contains(out, "Usage: run-skill-script git-jj <script> [args...]") {
			t.Fatalf("%v missing usage: %q", args, out)
		}
		if !strings.Contains(out, "repo-check") || !strings.Contains(out, "Detect repo type") {
			t.Fatalf("%v missing script list: %q", args, out)
		}
	}
}

func TestUnknownScriptIsNotHelp(t *testing.T) {
	home := t.TempDir()
	skillDir := filepath.Join(home, ".config", "agents", "skills", "git-jj")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	meta := skill.Meta{Scripts: map[string]skill.Script{
		"repo-check": {File: "scripts/repo_check.sh", Runtime: "bash"},
	}}
	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "meta.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("AMP_WORKING_DIRECTORY", t.TempDir())
	t.Setenv("RUN_SKILL_SCRIPT_SKILLS", "")
	t.Chdir(t.TempDir())

	out, code := captureMain(t, []string{"git-jj", "missing"})
	if code == 0 {
		t.Fatalf("expected failure, got %q", out)
	}
	if !strings.Contains(out, "not declared in meta.json") {
		t.Fatalf("output %q", out)
	}
}

func captureMain(t *testing.T, args []string) (string, int) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = w, w
	code := run(args)
	w.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(b), code
}

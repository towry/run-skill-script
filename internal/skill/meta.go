package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Meta struct {
	Scripts map[string]Script `json:"scripts"`
}

type Script struct {
	File        string `json:"file"`
	Runtime     string `json:"runtime"`
	Description string `json:"description"`
}

type NoMetaError struct {
	Skill string
	Dir   string
}

func (e NoMetaError) Error() string {
	return fmt.Sprintf("skill %q has no meta.json at %s; refusing to guess runnable scripts", e.Skill, e.Dir)
}

func LoadMeta(dir string) (*Meta, error) {
	path := filepath.Join(dir, "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NoMetaError{Skill: skillNameFromDir(filepath.Base(dir)), Dir: dir}
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if meta.Scripts == nil {
		meta.Scripts = map[string]Script{}
	}
	return &meta, nil
}

func (m *Meta) Script(name string) (Script, error) {
	s, ok := m.Scripts[name]
	if !ok {
		return Script{}, fmt.Errorf("script %q is not declared in meta.json\navailable scripts:\n%s", name, formatScripts(m))
	}
	if strings.TrimSpace(s.File) == "" {
		return Script{}, fmt.Errorf("script %q in meta.json has no file", name)
	}
	if filepath.IsAbs(s.File) || !isRelFile(s.File) {
		return Script{}, fmt.Errorf("script %q file must be a path relative to the skill directory", name)
	}
	return s, nil
}

func isRelFile(file string) bool {
	clean := filepath.ToSlash(filepath.Clean(file))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return false
	}
	return !strings.HasPrefix(clean, "/")
}

func formatScripts(m *Meta) string {
	names := make([]string, 0, len(m.Scripts))
	for name := range m.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "  (none)"
	}
	var b strings.Builder
	for _, name := range names {
		s := m.Scripts[name]
		fmt.Fprintf(&b, "  %s\t%s", name, s.File)
		if s.Runtime != "" {
			fmt.Fprintf(&b, " (%s)", s.Runtime)
		}
		if s.Description != "" {
			fmt.Fprintf(&b, " — %s", s.Description)
		}
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

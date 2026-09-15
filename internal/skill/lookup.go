package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Lookup locates skill directories the same way Amp discovers them.
type Lookup struct {
	Home    string
	Cwd     string
	WorkDir string
	Extra   []string
}

func DefaultLookup() Lookup {
	cwd, _ := os.Getwd()
	return Lookup{
		Home:    os.Getenv("HOME"),
		Cwd:     cwd,
		WorkDir: os.Getenv("AMP_WORKING_DIRECTORY"),
		Extra:   splitPath(os.Getenv("RUN_SKILL_SCRIPT_SKILLS")),
	}
}

func splitPath(v string) []string {
	if v == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(v, string(os.PathListSeparator)) {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (l Lookup) Find(name string) (string, error) {
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid skill name %q", name)
	}
	var searched []string
	for _, root := range l.roots() {
		searched = append(searched, root)
		if dir, ok := findNamed(root, name); ok {
			return dir, nil
		}
	}
	if len(searched) == 0 {
		return "", fmt.Errorf("skill %q not found (no skill directories to search)", name)
	}
	return "", fmt.Errorf("skill %q not found\nsearched:\n  %s", name, strings.Join(searched, "\n  "))
}

type listedSkill struct {
	Name string
	Path string
}

func (l Lookup) List() ([]listedSkill, error) {
	seen := map[string]listedSkill{}
	var order []string
	for _, root := range l.roots() {
		entries, err := os.ReadDir(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := skillNameFromDir(e.Name())
			if name == "" {
				continue
			}
			path := filepath.Join(root, e.Name())
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = listedSkill{Name: name, Path: path}
			order = append(order, name)
		}
	}
	out := make([]listedSkill, 0, len(order))
	for _, name := range order {
		out = append(out, seen[name])
	}
	return out, nil
}

func (l Lookup) roots() []string {
	var roots []string
	add := func(p string) {
		if p == "" {
			return
		}
		p = filepath.Clean(p)
		for _, existing := range roots {
			if existing == p {
				return
			}
		}
		roots = append(roots, p)
	}

	for _, extra := range l.Extra {
		add(extra)
	}
	if l.Home != "" {
		add(filepath.Join(l.Home, ".config", "agents", "skills"))
		add(filepath.Join(l.Home, ".agents", "skills"))
		add(filepath.Join(l.Home, ".config", "amp", "skills"))
	}
	for _, start := range []string{l.Cwd, l.WorkDir} {
		for _, dir := range walkUp(start) {
			add(filepath.Join(dir, ".agents", "skills"))
			add(filepath.Join(dir, ".claude", "skills"))
		}
	}
	if l.Home != "" {
		add(filepath.Join(l.Home, ".claude", "skills"))
		addCacheRoots(l.Home, add)
	}
	return roots
}

func addCacheRoots(home string, add func(string)) {
	walkHostScope(filepath.Join(home, ".cache", "amp", "global-plugins"), func(scopePath string) {
		addPluginSkillRoots(scopePath, add)
	})
	add(filepath.Join(home, ".cache", "find-hot-skill", "skills"))
	walkHostScope(filepath.Join(home, ".cache", "amp", "global-skills"), func(scopePath string) {
		add(scopePath)
	})
}

func walkHostScope(base string, fn func(scopePath string)) {
	hosts, err := os.ReadDir(base)
	if err != nil {
		return
	}
	for _, host := range hosts {
		if !host.IsDir() {
			continue
		}
		hostPath := filepath.Join(base, host.Name())
		scopes, err := os.ReadDir(hostPath)
		if err != nil {
			continue
		}
		for _, scope := range scopes {
			if scope.IsDir() {
				fn(filepath.Join(hostPath, scope.Name()))
			}
		}
	}
}

func addPluginSkillRoots(scopePath string, add func(string)) {
	entries, err := os.ReadDir(scopePath)
	if err != nil {
		return
	}
	var plugins []os.DirEntry
	for _, e := range entries {
		if e.IsDir() && isDir(filepath.Join(scopePath, e.Name(), "skills")) {
			plugins = append(plugins, e)
		}
	}
	sortNewestFirst(scopePath, plugins)
	for _, p := range plugins {
		add(filepath.Join(scopePath, p.Name(), "skills"))
	}
}

func walkUp(start string) []string {
	if start == "" {
		return nil
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil
	}
	var dirs []string
	for {
		dirs = append(dirs, dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return dirs
}

func findNamed(root, name string) (string, bool) {
	exact := filepath.Join(root, name)
	if isDir(exact) && skillNameFromDir(name) != "" {
		return exact, true
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", false
	}
	var matches []os.DirEntry
	prefix := name + "@"
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			matches = append(matches, e)
		}
	}
	if len(matches) == 0 {
		return "", false
	}
	sortNewestFirst(root, matches)
	return filepath.Join(root, matches[0].Name()), true
}

func sortNewestFirst(root string, entries []os.DirEntry) {
	sort.Slice(entries, func(i, j int) bool {
		pi := filepath.Join(root, entries[i].Name())
		pj := filepath.Join(root, entries[j].Name())
		si, _ := os.Stat(pi)
		sj, _ := os.Stat(pj)
		if si != nil && sj != nil && !si.ModTime().Equal(sj.ModTime()) {
			return si.ModTime().After(sj.ModTime())
		}
		return entries[i].Name() > entries[j].Name()
	})
}

func skillNameFromDir(dir string) string {
	if dir == "" || strings.HasPrefix(dir, ".") {
		return ""
	}
	if strings.HasSuffix(dir, ".tmp") {
		return ""
	}
	if i := strings.IndexByte(dir, '@'); i > 0 {
		return dir[:i]
	}
	return dir
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

package skill

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ExitError struct {
	Code int
	Err  error
}

func (e ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit status %d", e.Code)
}

func (e ExitError) Unwrap() error { return e.Err }

func ExitCode(err error) (int, bool) {
	var ee ExitError
	if errors.As(err, &ee) {
		return ee.Code, true
	}
	return 0, false
}

func Run(dir string, meta *Meta, scriptName string, args []string) error {
	script, err := meta.Script(scriptName)
	if err != nil {
		return err
	}
	file := filepath.Join(dir, filepath.FromSlash(script.File))
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("script %q file %s does not exist", scriptName, file)
		}
		return err
	}

	argv, err := commandFor(script.Runtime, script.File, file)
	if err != nil {
		return fmt.Errorf("script %q: %w", scriptName, err)
	}
	cmd := exec.Command(argv[0], append(argv[1:], args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir, err = os.Getwd()
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		var xerr *exec.ExitError
		if errors.As(err, &xerr) {
			return ExitError{Code: xerr.ExitCode(), Err: err}
		}
		return err
	}
	return nil
}

func commandFor(runtime, declaredFile, absFile string) ([]string, error) {
	rt := strings.ToLower(strings.TrimSpace(runtime))
	if rt == "" {
		rt = runtimeFromFile(declaredFile)
	}
	switch rt {
	case "bash":
		return []string{"bash", absFile}, nil
	case "sh":
		return []string{"sh", absFile}, nil
	case "bun":
		return []string{"bun", "run", absFile}, nil
	case "node", "js":
		return []string{"node", absFile}, nil
	case "python", "python3", "py":
		return []string{"python3", absFile}, nil
	case "":
		return nil, fmt.Errorf("meta.json does not set runtime, and file %q has no known runtime", declaredFile)
	default:
		return nil, fmt.Errorf("unknown runtime %q", runtime)
	}
}

func runtimeFromFile(file string) string {
	switch strings.ToLower(filepath.Ext(file)) {
	case ".sh", ".bash":
		return "bash"
	case ".ts", ".js", ".mjs":
		return "bun"
	case ".py":
		return "python3"
	default:
		return ""
	}
}

func PrintScripts(w io.Writer, name, dir string, meta *Meta) error {
	fmt.Fprintf(w, "Usage: run-skill-script %s <script> [args...]\n", name)
	fmt.Fprintf(w, "Skill: %s\n", name)
	fmt.Fprintf(w, "Path:  %s\n", dir)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Scripts:")
	fmt.Fprintln(w, formatScripts(meta))
	return nil
}

func PrintSkills(w io.Writer, lookup Lookup) error {
	skills, err := lookup.List()
	if err != nil {
		return err
	}
	if len(skills) == 0 {
		fmt.Fprintln(w, "(no skills found)")
		return nil
	}
	for _, s := range skills {
		fmt.Fprintf(w, "%s\t%s\n", s.Name, s.Path)
	}
	return nil
}

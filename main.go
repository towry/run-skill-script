package main

import (
	"fmt"
	"os"

	"github.com/towry/run-skill-script/internal/skill"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprint(os.Stderr, usage())
		if len(args) == 0 {
			return 2
		}
		return 0
	}
	if args[0] == "-v" || args[0] == "--version" {
		fmt.Println(version)
		return 0
	}

	lookup := skill.DefaultLookup()
	if args[0] == "--list-skills" {
		if err := skill.PrintSkills(os.Stdout, lookup); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}
	if args[0] != "" && args[0][0] == '-' {
		fmt.Fprintf(os.Stderr, "unknown flag %s\n\n%s", args[0], usage())
		return 2
	}

	name := args[0]
	dir, err := lookup.Find(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	meta, err := skill.LoadMeta(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if len(args) == 1 {
		if err := skill.PrintScripts(os.Stdout, name, dir, meta); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}

	if err := skill.Run(dir, meta, args[1], args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if code, ok := skill.ExitCode(err); ok {
			return code
		}
		return 1
	}
	return 0
}

func usage() string {
	return `Usage:
  run-skill-script <skill> <script> [args...]
  run-skill-script <skill>
  run-skill-script --list-skills
  run-skill-script --version

Runs a script declared in the skill's meta.json.
If meta.json is missing, the command exits without guessing a script path.
`
}

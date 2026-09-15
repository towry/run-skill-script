# run-skill-script

CLI that runs a script declared in an Amp skill's `meta.json`.

```
run-skill-script git-jj repo-check
```

If the skill has no `meta.json`, the command exits and does not guess a script path.

## Install

```
curl -fsSL https://raw.githubusercontent.com/towry/run-skill-script/main/install.sh | bash
```

This installs `run-skill-script` to `~/.local/bin`. Override with `PREFIX` or `RUN_SKILL_SCRIPT_BIN_DIR`:

```
curl -fsSL https://raw.githubusercontent.com/towry/run-skill-script/main/install.sh | PREFIX=/usr/local bash
curl -fsSL https://raw.githubusercontent.com/towry/run-skill-script/main/install.sh | RUN_SKILL_SCRIPT_VERSION=v0.1.0 bash
```

## Usage

```
run-skill-script <skill> <script> [args...]
run-skill-script <skill>
run-skill-script <skill> --help
run-skill-script --list-skills
```

`<skill>` is a directory name such as `git-jj` or `task-notes`, not a file path. See [docs/skill-meta.md](docs/skill-meta.md) for `meta.json`.

## Skill search paths

The first match wins. Duplicate names later in the list are ignored.

Hashed directories such as `git-jj@c4b90d11` match the skill name `git-jj`. If several hashed copies exist in the same root, the newest directory wins.

| Order | Path |
|---|---|
| 1 | Directories in `RUN_SKILL_SCRIPT_SKILLS` (the OS path list, `:` on Unix) |
| 2 | `~/.config/agents/skills` |
| 3 | `~/.agents/skills` |
| 4 | `~/.config/amp/skills` |
| 5 | `<dir>/.agents/skills` and `<dir>/.claude/skills`, walking from the current directory toward `/`, then the same walk from `AMP_WORKING_DIRECTORY` |
| 6 | `~/.claude/skills` |
| 7 | `~/.cache/amp/global-plugins/<host>/<scope>/<plugin@hash>/skills` (plugin dirs newest first) |
| 8 | `~/.cache/find-hot-skill/skills` |
| 9 | `~/.cache/amp/global-skills/<host>/<scope>` |

The plugin cache root is the plugin's `skills/` directory, not the plugin directory. `find-hot-skill@09a84769.../skills/task-notes` matches `task-notes`. `find-hot-skill@09a84769...` itself is not a skill. Directories ending in `.tmp` are ignored.

## Release

1. Merge the commit you want onto `main`.
2. Tag it and push the tag:

```
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions (`.github/workflows/release.yml`) then builds `run-skill-script` for Linux and macOS (`amd64` and `arm64`) and publishes a GitHub Release. `install.sh` downloads that release.

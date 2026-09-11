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

Skill lookup follows Amp's skill directories, including hashed cache dirs such as `git-jj@c4b90d11`.
It also looks in plugin caches at `~/.cache/amp/global-plugins/<host>/<scope>/<plugin@hash>/skills/`.

## Release

1. Merge the commit you want onto `main`.
2. Tag it and push the tag:

```
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions (`.github/workflows/release.yml`) then builds `run-skill-script` for Linux and macOS (`amd64` and `arm64`) and publishes a GitHub Release. `install.sh` downloads that release.

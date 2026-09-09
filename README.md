# run-skill-script

CLI that runs a script declared in an Amp skill's `meta.json`.

```
run-skill-script git-jj repo-check
```

If the skill has no `meta.json`, the command exits and does not guess a script path.

## Usage

```
run-skill-script <skill> <script> [args...]
run-skill-script <skill>
run-skill-script --list-skills
```

Skill lookup follows Amp's skill directories, including hashed cache dirs such as `git-jj@c4b90d11`.

## Release

Push a tag `vX.Y.Z`. GitHub Actions builds `run-skill-script` for Linux and macOS (`amd64` and `arm64`) and publishes a GitHub Release.

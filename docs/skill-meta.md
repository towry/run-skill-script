# Skill `meta.json`

`run-skill-script` runs only scripts declared in a skill's `meta.json`. If that file is missing, the command exits and does not guess a path.

Place `meta.json` in the skill directory, next to `SKILL.md`.

## Command

```
run-skill-script <skill> <script> [args...]
```

`<skill>` is the skill directory name, not a file path. `<script>` is a key under `scripts` in `meta.json`. Extra arguments are passed to the script unchanged.

```
run-skill-script git-jj repo-check
run-skill-script task-notes notes latest --slug example --limit 8
```

List declared scripts:

```
run-skill-script <skill>
run-skill-script <skill> --help
```

The process working directory is the caller's current directory, not the skill directory. The script file itself is resolved from the skill directory.

## Schema

```json
{
  "scripts": {
    "repo-check": {
      "file": "scripts/repo_check.sh",
      "runtime": "bash",
      "description": "Detect repository type"
    }
  }
}
```

| Field | Required | Meaning |
|---|---|---|
| `scripts` | yes | Map of script name to declaration. Missing `scripts` is treated as empty. |
| `scripts.<name>` | yes | Name used as `<script>` on the command line. |
| `scripts.<name>.file` | yes | Path to the script, relative to the skill directory. |
| `scripts.<name>.runtime` | no | Interpreter. If omitted, inferred from the file extension. |
| `scripts.<name>.description` | no | Shown in `run-skill-script <skill>` output. |

Unknown JSON fields are ignored.

## `file`

`file` must be a relative path inside the skill directory.

Rejected values:

- empty
- absolute paths (`/usr/bin/env`, `C:\...`)
- paths that escape the skill directory (`../secret.sh`)

The file must exist at run time. `run-skill-script` joins the skill directory with `file` and fails if that path is missing.

## `runtime`

| `runtime` | Command |
|---|---|
| `bash` | `bash <file>` |
| `sh` | `sh <file>` |
| `bun` | `bun run <file>` |
| `node`, `js` | `node <file>` |
| `python`, `python3`, `py` | `python3 <file>` |

Matching is case-insensitive. Surrounding whitespace is trimmed.

If `runtime` is omitted, the extension of `file` selects one:

| Extension | Runtime |
|---|---|
| `.sh`, `.bash` | `bash` |
| `.ts`, `.js`, `.mjs` | `bun` |
| `.py` | `python3` |

Any other extension with no `runtime` is an error. An unknown `runtime` value is an error.

## Errors

| Situation | Result |
|---|---|
| No `meta.json` | Exit without guessing a script path |
| `<script>` not in `scripts` | Error listing the declared names |
| `file` empty, absolute, or outside the skill directory | Error |
| Declared `file` does not exist | Error |
| Unknown `runtime`, or none that can be inferred | Error |
| Script process exits non-zero | That exit code is returned |

## Example

`git-jj/meta.json`:

```json
{
  "scripts": {
    "repo-check": {
      "file": "scripts/repo_check.sh",
      "runtime": "bash",
      "description": "Detect repository type. Prints jj, git, or no-repo."
    }
  }
}
```

`run-skill-script git-jj repo-check` runs `scripts/repo_check.sh` from the resolved `git-jj` skill directory.

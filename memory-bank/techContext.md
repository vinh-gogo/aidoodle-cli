# Tech Context

## Stack
- **Go** (`go.mod`: 1.25.5; local `go version` = 1.27.0 windows/amd64)
- **agentcore** `github.com/voocel/agentcore v1.8.3` — minimal agent kernel (tool-calling, streaming, `context` compression helpers)
- **litellm** `github.com/voocel/litellm v1.8.10` — unified LLM provider adapter
- **Bubble Tea** `bubbletea v1.3.10`, `bubbles v1.0.0`, `lipgloss v1.1.0`, `charmbracelet/x/ansi`, `termenv` — TUI
- `gofrs/flock v0.13.0` — per-book directory lease (`.ainovel.lock`)
- `golang.org/x/{mod,text,image,sys}` — semver / encoding (GB18030 for import) / EPUB images / syscalls

## Layout
```
cmd/ainovel-cli/main.go      CLI entry (flags, setup, update, eval subcommand)
assets/                      embedded prompts/, references/ (+genres/<style>/), styles/, voice.md, load.go, testdata/
internal/                    25 packages (see systemPatterns.md)
docs/                        design docs (Chinese)
evals/cases/smoke/           architect_long/short.json, writer_first_chapter.json
scripts/                     install.sh, check_chapter_wordcount.py, sample.gif, novel.png
.github/workflows/           ci.yml, docker.yml, release.yml (+ scripts/gen-changelog.sh)
output/                      generated books (gitignored: `output*`)
run.sh                       Vietnamese Ollama setup script (Git Bash/WSL/macOS/Linux)
Modelfile                    FROM qwen3.5:4b ; num_ctx 32768
OllamaSetup.exe, ainovel-cli.exe   local binaries (*.exe gitignored)
```

## CLI surface (`cmd/ainovel-cli/main.go`)
`ainovel-cli` (TUI) · `--headless [--prompt S | --prompt-file F|-]` · `--style/-s NAME` · `--dir/-d OUTDIR` · `--version/-v` / `version` · `update [version]` · `eval …` (offline harness, separate flags). Positional args are rejected (novel request must be typed in TUI). `--prompt*` only valid with `--headless`. Headless refuses first-run setup. Unknown style → error listing available styles from `bundle.Styles`.

## Dev commands (from CI)
```powershell
gofmt -l .                      # must be empty (CI: test -z "$(gofmt -l .)")
go vet ./...                    # CI sets GOWORK=off
go test -buildvcs=false -count=1 ./...
go test -race -buildvcs=false -count=1 ./internal/host ./internal/store ./internal/tools   # CI: Linux only
go build -o ainovel-cli.exe ./cmd/ainovel-cli
sh -n scripts/install.sh        # Linux CI only
```
CI matrix: ubuntu-latest + windows-latest, Go from `go.mod`.
Release: GoReleaser (`.goreleaser.yml`), Docker image `ghcr.io/voocel/ainovel-cli`; `docker-compose.yml` mounts `./config:/root/.ainovel` and `./workspace:/workspace`.

> I have NOT run build/tests in this workspace yet — state unknown. Do this at the start of any code-changing task.

## Config
- Global `~/.ainovel/config.json`; project `./.ainovel/config.json` (gitignored — contains secrets); example `config.example.jsonc` (Vietnamese comments).
- Keys: `provider`, `model`, `reasoning_effort` (off|low|medium|high|xhigh|max), `providers{type, api_key, base_url, models[{name,context_window,json_schema}], api(chat|responses), extra{user_agent,headers,anthropic_beta}, extra_body, stream_idle_timeout}`, `roles{architect|writer|editor|import_segment|import_analyze|import_synthesize → provider/model/reasoning_effort/fallbacks}`, `style`, `context_window` (legacy), `budget{book_usd,warn_ratio,hard_stop}`, `notify{enabled,command,events}`, `disable_update_check`.
- Ollama: `base_url` must include `/v1`; raise `stream_idle_timeout` (e.g. `"15m"`) for slow local inference.
- Rules dirs: `~/.ainovel/rules/*.md`, `./.ainovel/rules/*.md` (natural language, normalized by model into `meta/user_rules.json`).
- Startup errors persisted to `~/.ainovel/last-error.log`.

## Styles (config `style`, file name = value)
Built-in: `default`, `suspense`, `fantasy`, `romance`. Added in this fork: `psychological` (= `psychology` alias file), `stone-age-doodle` (= `stone-age` alias, identical size), `stone-age-doodle-explain`, `vietnamese-history` (= `viet-history` alias). Genre reference dirs exist for: fantasy, psychological, romance, stone-age-doodle, stone-age-doodle-explain, suspense, vietnamese-history (the alias style names have **no** genre dir of their own — unverified whether loader maps aliases).

## Constraints & gotchas
- **Windows terminal encoding**: PowerShell output of UTF-8 Vietnamese/Chinese files renders as mojibake. Don't treat garbled output as corrupted files; use `view_file` or force UTF-8 (`[Console]::OutputEncoding = [Text.Encoding]::UTF8`).
- File-lock file permission was changed `0o600 → 0o666` in `host/book_lock.go` to avoid "Access is denied" on Windows (with explicit Vietnamese error messages and `isAccessDenied`).
- Import supports `txt`/`md`, UTF-8/GB18030 only.
- Small local models (4B, 16K ctx) are weak at the structured tool calls the system needs — expect planning failures (see progress.md).
- `.gitignore` ignores `*.exe`, `.ainovel/`, `output*`, `workspace/`, `refer/`, `tasks`, `ainovel.db`, `dist/`. `memory-bank/` is NOT ignored (will show as untracked until committed).
- Cost accounting needs registry pricing (`internal/models`, generated by `gen_models.go`); custom models without price won't trigger the budget fuse.

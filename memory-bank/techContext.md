# Tech Context

## Stack
- **Go** (`go.mod`: 1.25.5; local `go version` = 1.27.0 windows/amd64)
- **agentcore** `github.com/voocel/agentcore v1.8.3` — minimal agent kernel (tool-calling, streaming, `context` compression helpers)
- **litellm** `github.com/voocel/litellm v1.8.10` — unified LLM provider adapter
- **Bubble Tea** `bubbletea v1.3.10`, `bubbles v1.0.0`, `lipgloss v1.1.0`, `charmbracelet/x/ansi`, `termenv` — TUI
- **Web UI & Server**: Standard library `net/http` + SSE (Server-Sent Events) + embedded HTML5/CSS/JS (`go:embed`)
- **Word OpenXML (.docx) Exporter**: Native Go implementation via `archive/zip` (zero external dependencies)
- `gofrs/flock v0.13.0` — per-book directory lease (`.ainovel.lock`)
- `golang.org/x/{mod,text,image,sys}` — semver / encoding (GB18030 for import) / EPUB images / syscalls

## Layout
```
cmd/ainovel-cli/main.go      CLI entry (flags, setup, update, eval, web subcommand)
assets/                      embedded prompts/, references/ (+genres/<style>/), styles/, voice.md, load.go, testdata/
internal/entry/web/          Web server, REST handlers, SSE broadcaster, embedded web/ui/index.html
internal/host/exp/           multi-format export engine (docx, epub, txt, video scripts)
internal/                    26 packages (see systemPatterns.md)
docs/                        design docs
evals/cases/smoke/           architect_long/short.json, writer_first_chapter.json
scripts/                     install.sh, check_chapter_wordcount.py, sample.gif, novel.png
.github/workflows/           ci.yml, docker.yml, release.yml, build-windows.yml
output/                      generated books (gitignored: `output*`)
run.sh                       Vietnamese Ollama setup script & runner (run.sh web)
Modelfile                    FROM qwen3.5:4b ; num_ctx 32768
OllamaSetup.exe, ainovel-cli.exe   local binaries (*.exe gitignored)
```

## CLI surface (`cmd/ainovel-cli/main.go`)
`ainovel-cli` (TUI) · `ainovel-cli web [--port P]` (Web Dashboard) · `--headless [--prompt S | --prompt-file F|-]` · `--style/-s NAME` · `--dir/-d OUTDIR` · `--version/-v` / `version` · `update [version]` · `eval …` (offline harness, separate flags).

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

> Verified build/test status: 100% PASS across all 26 packages (unit, integration, e2e, evals: 315+ tests). Verified locally on Windows AMD64.

## Config
- Global `~/.ainovel/config.json`; project `./.ainovel/config.json` (gitignored — contains secrets); example `config.example.jsonc` (Vietnamese comments).
- Keys: `provider`, `model`, `reasoning_effort` (off|low|medium|high|xhigh|max), `providers{type, api_key, base_url, models[{name,context_window,json_schema}], api(chat|responses), extra{user_agent,headers,anthropic_beta}, extra_body, stream_idle_timeout}`, `roles{architect|writer|editor|import_segment|import_analyze|import_synthesize → provider/model/reasoning_effort/fallbacks}`, `style`, `context_window` (legacy), `budget{book_usd,warn_ratio,hard_stop}`, `notify{enabled,command,events}`, `disable_update_check`.
- **Colab / Local Server Inference via Cloudflare Tunnel (`kaggle/colab_llama_cpp.ipynb`)**:
  - Unified FastAPI Gateway (Port 8000) published via Cloudflare Tunnel at fixed URL: `https://api.gloryvocab.com/v1`
  - **Main LLM (Port 8001)**: `Gemma 4 26B A4B` (`model: "main-llm"` or `"gemma-4-26B-A4B-it-Q4_K_M.gguf"`, `--ctx-size 65536`, `context_window: 32768`).
  - **STT Backend (Port 8003)**: `Gemma 4 E2B + Audio mmproj` (`model: "gemma-4-stt"` or `"google/gemma-4-E2B-it"`, `/v1/audio/transcriptions`).
  - `kaggle-llama` / `colab-llama` provider config: Uses `type: "openai"`, `base_url: "https://api.gloryvocab.com/v1"`, `model: "main-llm"`, `context_window: 32768`, `stream_idle_timeout: "15m"` (không còn cần `ngrok-skip-browser-warning`).
  - **Repetition Trap Mitigation**: LLMs on open-source weights (Qwen, Llama) can enter infinite loops (e.g., `- true\n- true...`). Mitigated by adding `extra_body` to provider config:
    ```json
    "extra_body": {
      "temperature": 0.7,
      "top_p": 0.9,
      "frequency_penalty": 0.3,
      "presence_penalty": 0.2,
      "repetition_penalty": 1.1
    }
    ```
- Rules dirs: `~/.ainovel/rules/*.md`, `./.ainovel/rules/*.md` (natural language, normalized by model into `meta/user_rules.json`).
- Startup errors persisted to `~/.ainovel/last-error.log`.

## Styles & Script Limits
- **7 Primary AI Modes**:
  1. `novel-manga` (Tiểu thuyết / Manga Đa năng, văn xuôi 2.000 – 4.000 từ/chương)
  2. `suspense` (Tiểu thuyết Trinh thám – Tâm lý Tội phạm & Kinh dị Tâm linh, 2.500 – 4.500 từ/chương; aliases: `trinh-tham`, `mystery`, `thriller`, `horror`)
  3. `romance` (Tiểu thuyết Ngôn tình – Chữa lành & Tâm lý Tình cảm Hiện đại, 2.500 – 4.000 từ/chương; aliases: `ngon-tinh`, `healing`)
  4. `fantasy` (Tiểu thuyết Kỳ ảo – Tiên hiệp / Thăng cấp & Thế giới Đa tầng, 2.500 – 4.500 từ/chương; aliases: `ky-ao`, `tien-hiep`, `xianxia`)
  5. `vietnamese-history` (Tiểu thuyết nhân vật lịch sử Việt Nam, 2.000 – 4.000 từ/chương; aliases: `history`, `lich-su`)
  6. `doodle-explainer` (Doodle Explainer video script 5+ phút)
  7. `behavioral-psychology` (Kịch bản video tâm lý học hành vi 5+ phút; alias: `psychology`)
- Legacy styles: `stone-age-doodle`, `stone-age-doodle-explain`, `psychological`, `default`.
- **Interactive Commands**:
  - `/mode` (alias: `/topic`, `/topics`): Chuyển đổi giữa 7 chế độ làm việc (phím tắt `1`–`7`) và lưu vào `.ainovel/config.json`.
  - `/dash` (alias: `/dashboard`, `/projects`, `/history`): Mở modal quản lý toàn bộ outputs đã làm, hiển thị Tag Mode (`[📖 Tiểu thuyết / Manga]`, `[🕵️ Trinh thám & Kinh dị]`, `[🌸 Ngôn tình & Chữa lành]`, `[🐉 Kỳ ảo & Tiên hiệp]`, `[⚔️ Lịch sử Việt Nam]`, `[🦴 Video TikTok]`, `[🧠 Tâm lý học hành vi]`), tiến độ, đường dẫn thư mục, cho phép chọn bằng phím mũi tên và ấn Enter để chuyển đổi dự án.
  - `/new`: Khởi tạo phiên làm việc mới trong thư mục output timestamp `output/novel-YYYYMMDD-HHMM` mà không cần nhập tham số.
- **Script validator bounds (`internal/tools/script_format.go`)**:
  - Chỉ áp dụng cho chế độ video (`doodle-explainer`, `behavioral-psychology`). Hoàn toàn không áp dụng cho 5 chế độ tiểu thuyết (`novel-manga`, `suspense`, `romance`, `fantasy`, `vietnamese-history`).
  - `scriptMinWords = 700`, `scriptMaxWords = 1500` (words of `LỜI:`).
  - `scriptMinSeconds = 300`, `scriptMaxSeconds = 600` (5 đến 10 phút thời lượng đọc).
  - Yêu cầu: dòng tiêu đề, HOOK đầu tiên, CHỐT cuối cùng, mỗi block có ít nhất một cặp LỜI và HÌNH, và phần chân kịch bản (CAPTION, HASHTAG, NGUỒN, CẦN KIỂM CHỨNG).
- **Tavily Integration & Graceful Degradation**:
  - Tự động nạp file `.env` tại thư mục làm việc khi khởi động (`main.go` -> `godotenv.Load()`).
  - Sử dụng biến môi trường `TAVILY_API_KEY` cho các công cụ `tavily_search` và `tavily_crawl`.
  - Cơ chế Graceful Fallback: Khi không có key hoặc API Tavily trả về lỗi HTTP 40x/50x/mạng, công cụ tự động trả về phản hồi mềm hướng dẫn LLM sử dụng tri thức sẵn có, không throw error, tránh deadlock và vòng lặp bế tắc.

## Constraints & gotchas
- **Windows terminal encoding**: PowerShell output of UTF-8 Vietnamese/Chinese files renders as mojibake. Don't treat garbled output as corrupted files; use `view_file` or force UTF-8 (`[Console]::OutputEncoding = [Text.Encoding]::UTF8`).
- File-lock file permission was changed `0o600 → 0o666` in `host/book_lock.go` to avoid "Access is denied" on Windows (with explicit Vietnamese error messages and `isAccessDenied`).
- Import supports `txt`/`md`, UTF-8/GB18030 only.
- Small local models (4B, 16K ctx) are weak at the structured tool calls the system needs — expect planning failures; recommended model is 14B–27B+ (e.g. `qwen-27b`).
- `.gitignore` ignores `*.exe`, `.ainovel/`, `output*`, `workspace/`, `refer/`, `tasks`, `ainovel.db`, `dist/`. `memory-bank/` is NOT ignored (will show as untracked until committed).
- Cost accounting needs registry pricing (`internal/models`, generated by `gen_models.go`); custom models without price won't trigger the budget fuse.

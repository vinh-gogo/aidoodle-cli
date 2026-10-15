# Project Brief — ainovel-cli (aidoodle-cli)

> Source of truth for scope. Updated 2026-10-15 (Expanded to 7 Specialized AI Modes with 3 Standout Novel Genres).

## What it is
`ainovel-cli` (aidoodle-cli) is an **autonomous, deterministic engine for generating Vietnamese content across 7 specialized AI working modes** (selectable via the interactive `/mode` command in both TUI and Web UI):
1. **Tiểu thuyết / Manga Đa năng (`novel-manga`)**: Sáng tác tiểu thuyết dài kỳ, Light Novel, Manga/Webtoon đa thể loại (2.000 – 4.000 từ/chương).
2. **Tiểu thuyết Trinh thám & Kinh dị (`suspense`)**: Sáng tác tiểu thuyết trinh thám suy luận công bằng (Fair-Play Whodunit), tâm lý tội phạm đen tối và kinh dị tâm linh dân gian Việt Nam (2.500 – 4.500 từ/chương).
3. **Tiểu thuyết Ngôn tình & Chữa lành (`romance`)**: Sáng tác tiểu thuyết ngôn tình hiện đại, tâm lý tình cảm sâu sắc (Slow-burn) và chữa lành vết thương nội tâm (2.500 – 4.000 từ/chương).
4. **Tiểu thuyết Kỳ ảo & Tiên hiệp (`fantasy`)**: Sáng tác tiểu thuyết kỳ ảo sử thi, tiên hiệp tu chân và thăng cấp sức mạnh (Progression) với hệ thống quy tắc cứng và thế giới đa tầng (2.500 – 4.500 từ/chương).
5. **Tiểu thuyết Lịch sử Việt Nam (`vietnamese-history`)**: Sáng tác tiểu thuyết và dã sử hào sảng về các nhân vật lịch sử Việt Nam (2.000 – 4.000 từ/chương), tích hợp tra cứu Tavily Search.
6. **Doodle Explainer (`doodle-explainer`)**: Biên kịch video người que đồ đá giải thích các chủ đề xu hướng & kiến thức hiện đại cho TikTok / YouTube Shorts (5+ phút, nhịp 1:1 Thoại - Hình).
7. **Tâm lý học hành vi (`behavioral-psychology`)**: Biên kịch video giải mã bẫy nhận thức, cơ chế Não Bò Sát vs Não Lý Trí, thí nghiệm khoa học chuẩn xác và Cú hích hành vi (Nudge) thực chiến.

- Module: `github.com/voocel/ainovel-cli` (Fork repo: `aidoodle-cli`, branches `doodle-explainer` & `main`)
- Go `1.25.5` in `go.mod` (local toolchain: go 1.27.0 windows/amd64)
- Entrypoint: `cmd/ainovel-cli/main.go`

## Evolution of the Fork
- Originally forked from `ainovel-cli` (novel writing engine).
- Successfully expanded into an extensible multi-mode engine with 7 dedicated modes:
  - **5 Novel modes** (`novel-manga`, `suspense`, `romance`, `fantasy`, `vietnamese-history`): 1 chapter = 2.000 – 4.500 từ văn xuôi tiểu thuyết, kết cấu chương hồi chuyên sâu theo từng thể loại, không áp dụng ràng buộc kịch bản video.
  - **2 Video Script modes** (`doodle-explainer`, `behavioral-psychology`): 1 video = 1 "chapter" (thời lượng 5+ phút: 300–600s, ~700–1500 từ lời đọc, nhịp LỜI-HÌNH xen kẽ 1:1, chạy bộ kiểm tra kịch bản `lintScript`).
  - **Tích hợp Tavily Search & Crawl (Graceful Fallback)**: Nạp `source_pack` tự động, kiểm chứng dữ liệu thực tế và tài liệu lịch sử, tự động nạp file `.env` cục bộ; cơ chế suy giảm chức năng an toàn tự động vận dụng tri thức sẵn có khi API bên ngoài gặp sự cố.
  - **TUI Dashboard (`/dash`) & Khởi tạo phiên mới (`/new`)**: Xem lại toàn bộ outputs, quản lý theo Tag Mode trực quan và chuyển đổi dự án không ngắt quãng.

## Core requirements (from upstream design, still binding)
1. **Stability first** — one sentence in → whole book out, no architectural self-interruption.
2. **Quality is iterable** — prompts / references / review dimensions / context strategy adjustable without touching architecture.
3. **Recoverable** — resume from latest checkpoint after crash / network loss / Ctrl+C (step-level).
4. **Observable** — per-chapter, per-step progress, artifacts, timing.

## Scope (in)
- Deterministic Engine + 3 autonomous workers (Architect / Writer / Editor) + on-demand Arbiter (LLM function)
- Rolling volume/arc planning for long-form; layered summaries; context compression pipeline
- Seven-dimension Editor review; live user intervention (steer); optional per-chapter acceptance gate (`/review on`, `/next`)
- TUI + `--headless`; multi-provider LLM (OpenRouter/Anthropic/Gemini/OpenAI/DeepSeek/Qwen/GLM/Grok/Ollama/Bedrock/custom proxy)
- Import (`/import`), export (`/export` TXT/EPUB), sync of manual edits (`/sync`), simulation profile (`/simulate`), diagnostics (`/diag`)
- Customizable voice/style/rules layers without recompiling

## Scope (out / non-goals)
- No task queue, no policy engine, no WorkflowInstance abstractions ("reject complex orchestration")
- No keyword/score-threshold rules pretending to be understanding (iron rule #4 in docs/architecture.md)
- Observability layer (`internal/diag`) is observe-only — never auto-fixes or resumes

## Key reference docs (in-repo)
`README.md` (Chinese, features + usage), `docs/architecture.md` (runtime architecture, authoritative), `docs/engine-arbiter.md`, `docs/engine-rfc.md`, `docs/context-management.md`, `docs/import-pipeline.md`, `docs/chapter-advance-gate.md`, `docs/user-rules-runtime.md`, `docs/voice-layer.md`, `docs/observability.md`, `docs/prompt-cache-design.md`, `docs/evaluation-system.md`, `docs/refactor-flow-driven.md`, `assets/README.md` (content-ownership map).

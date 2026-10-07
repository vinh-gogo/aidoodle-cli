# Project Brief — ainovel-cli (aidoodle-cli)

> Source of truth for scope. Updated 2026-10-07 (Doodle Explainer conversion completed).

## What it is
`ainovel-cli` (aidoodle-cli) is an **autonomous, deterministic engine for generating Vietnamese content across 4 specialized AI working modes** (selectable via the interactive `/mode` command):
1. **Tiểu thuyết / Manga (`novel-manga`)**: Sáng tác tiểu thuyết dài kỳ, manga, thế giới quan đa tầng, chiều sâu tâm lý nhân vật (2.000 – 4.000 từ/chương).
2. **Doodle Explainer (`doodle-explainer`)**: Biên kịch video người que đồ đá giải thích các chủ đề xu hướng & kiến thức hiện đại cho TikTok / YouTube Shorts (5+ phút, nhịp 1:1 Thoại - Hình).
3. **Tâm lý học hành vi (`behavioral-psychology`)**: Biên kịch video giải mã bẫy nhận thức, cơ chế Não Bò Sát vs Não Lý Trí, thí nghiệm khoa học chuẩn xác và Cú hích hành vi (Nudge) thực chiến.
4. **Tiểu thuyết Lịch sử Việt Nam (`vietnamese-history`)**: Sáng tác tiểu thuyết và dã sử hào sảng về các nhân vật lịch sử Việt Nam (2.000 – 4.000 từ/chương), tích hợp tra cứu Tavily Search, khắc họa chân dung con người thật đa chiều, bối cảnh lịch sử đa tầng và những nghịch cảnh sinh tử bi tráng.

- Module: `github.com/voocel/ainovel-cli` (Fork repo: `aidoodle-cli`, branch `route` / `doodle-explainer`)
- Go `1.25.5` in `go.mod` (local toolchain: go 1.27.0 windows/amd64)
- Entrypoint: `cmd/ainovel-cli/main.go`

## Evolution of the Fork
- Originally forked from `ainovel-cli` (novel writing engine).
- Successfully expanded into an extensible multi-mode engine with 4 dedicated modes:
  - **Video Script modes** (`doodle-explainer`, `behavioral-psychology`): 1 video = 1 "chapter" (thời lượng 5+ phút: 300–600s, ~700–1500 từ lời đọc, nhịp LỜI-HÌNH xen kẽ 1:1, chạy bộ kiểm tra kịch bản `lintScript`).
  - **Novel modes** (`novel-manga`, `vietnamese-history`): 1 chapter = 2.000 – 4.000 từ văn xuôi tiểu thuyết, kết cấu chương hồi, không áp dụng ràng buộc kịch bản video.
  - **Tích hợp Tavily Search & Crawl**: Nạp `source_pack` tự động, kiểm chứng dữ liệu thực tế và tài liệu lịch sử, tự động nạp file `.env` cục bộ.

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

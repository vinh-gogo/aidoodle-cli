# Project Brief — ainovel-cli (aidoodle-cli)

> Source of truth for scope. Updated 2026-10-07 (Doodle Explainer conversion completed).

## What it is
`ainovel-cli` (aidoodle-cli) is an **autonomous, deterministic engine for generating Vietnamese TikTok Doodle Explainer scripts** (nhân vật que đồ đá giải thích các chủ đề xu hướng & kiến thức hiện đại). From a single topic prompt or auto-fetched trends, it designs the series bible, creates characters, outlines deep-dive episodes, drafts scripts, conducts editorial reviews, and exports production packages with zero or minimal human intervention.

- Module: `github.com/voocel/ainovel-cli` (Fork repo: `aidoodle-cli`, branch `doodle-explainer`)
- Go `1.25.5` in `go.mod` (local toolchain: go 1.27.0 windows/amd64)
- Entrypoint: `cmd/ainovel-cli/main.go`

## Evolution of the Fork
- Originally forked from `ainovel-cli` (novel writing engine).
- Successfully converted into a specialized **TikTok Doodle Explainer Video Scriptwriter**:
  - **1 video = 1 "chapter"** (thời lượng từ 5 phút trở lên: 300–600s, ~750–1200 từ lời đọc).
  - **1 series = 1 "book"** (mặc định quy hoạch series 3 tập chuyên sâu đào sâu các góc độ của cùng một chủ đề mà không bị phân tán hay lan man).
  - **Doodle Visuals**: Phân chia cặp `LỜI:` và `HÌNH:` xen kẽ 1:1 theo từng câu thoại (mỗi 3–6s đổi hình một lần, loại bỏ hoàn toàn thẻ `CHỮ:`).
  - **Tone & Voice**: Nhân vật que thời đồ đá (host + phụ + linh vật), ẩn dụ đồ đá giải thích thế giới hiện đại, đệm nhẹ tiếng cười đồng cảm nhân sinh.

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

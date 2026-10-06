# Project Brief — ainovel-cli

> Source of truth for scope. Created 2026-10-06 (first Memory Bank bootstrap, derived from README, docs/, source tree and git history).

## What it is
`ainovel-cli` is a **fully automatic AI long-form novel writing engine** (Go CLI, TUI + headless). From a one-sentence request it plans, writes, reviews and rewrites an entire novel (target 200–500+ chapters) with no human intervention. Human steering is optional and live.

- Module: `github.com/voocel/ainovel-cli` (upstream author: `voocel`; license MIT)
- Go `1.25.5` in `go.mod` (local toolchain observed: go 1.27.0 windows/amd64)
- Entrypoint: `cmd/ainovel-cli/main.go`

## This repository is a FORK
- Local repo `D:\ainovel-cli` has only 2 commits:
  - `b67734c` (2026-10-03) "init vietnamese from ainovel-cli" — Vietnamese localization of the upstream project
  - `13f98eb` (2026-10-04) "add style" — new genre styles + Windows lock fix + Vietnamese-language summary prompts (author `vinh-gogo`)
- Goal of the fork: **write novels in Vietnamese** (and run locally, e.g. via Ollama), keeping upstream architecture intact.

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

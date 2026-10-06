# Product Context

## Why this project exists
Writing a 300–500 chapter web novel by hand takes months; naive "one big LLM loop" approaches drift, forget foreshadowing, repeat structures, and die on crashes. ainovel-cli replaces a fragile agent-orchestration loop with a **deterministic engine** that calls LLMs only where judgment is needed, so a book can run for hours/days unattended.

## Problems it solves
| Problem | Solution |
|---|---|
| Outline becomes hollow at 300+ chapters | **Rolling compass + horizon planning**: plan 2 volumes' arc skeletons + arc 1 in detail; expand next arc/volume only when reached, using prior summaries + character snapshots |
| Model forgets earlier chapters | 3-level summaries (chapter→arc→volume), related-chapter recommendation (foreshadow / appearance / state change / relationship), next-chapter preview |
| Context window overflow | 4-stage compression: ToolResultMicrocompact → LightTrim → StoreSummaryCompact (zero-LLM) → FullSummary; CJK token estimate runes×1.5 |
| Crash mid-chapter | Step-level checkpoints; resume = read store and re-route (no session to restore) |
| "AI flavor" prose | Built-in anti-AI-tone baseline (mechanical blacklist + semantic criteria), user rules in natural language |
| Loss of user control | Live steer input, optional per-chapter gate, budget fuse, notifications |

## How it should work (user flow)
1. First run → bootstrap wizard writes `~/.ainovel/config.json` (provider → API key → base URL → model).
2. Start in a directory; novel artifacts go to `{cwd}/output/novel/` (override with `--dir`). One directory = one book; re-running there auto-resumes.
3. TUI start modes: **quick start** (one sentence) or **co-create** (multi-turn clarification; right pane shows live draft of the creative brief, `Ctrl+S` to start); `/start ./outline.md` to seed from a file.
4. Engine loops: Arbiter picks planner → Architect plans → Writer writes chapter by chapter (novel_context → read_chapter → plan_chapter → draft_chapter → check_consistency → commit_chapter) → Editor reviews at arc/volume ends → rewrite/polish or expand next arc.
5. User may inject steer text anytime; Arbiter triages (settings change → Architect; rewrite → Editor queue; rules → immediate).
6. `--headless --prompt "…"` / `--prompt-file` for servers, NAS, CI, Docker.

## UX goals
- Zero-intervention default; precise control available when wanted.
- Every semantic decision audit-logged (`meta/decisions.jsonl`) and replayable.
- TUI shows context health gradient (green <70%, yellow 70–85%, red >85%).
- Errors should not vanish: fatal startup errors are written to `~/.ainovel/last-error.log` and the console pauses for Enter (non-headless).

## Fork-specific product intent (Vietnamese edition)
- Output language must be **100% Vietnamese** (prompts instruct "tuyệt đối KHÔNG dùng tiếng Trung" — never use Chinese).
- Runtime messages in `main.go`/`book_lock.go` and worker summary prompts translated to Vietnamese; `config.example.jsonc` comments Vietnamese; **README/docs remain Chinese**.
- Target hardware: local models via **Ollama** (`run.sh` creates a `qwen3.5-4b-16k` variant, num_ctx 16384, ~8GB-class VRAM tuning via `OLLAMA_FLASH_ATTENTION`, `OLLAMA_KV_CACHE_TYPE=q8_0`), but any provider works.
- Extra genre styles for Vietnamese content: psychological, stone-age-doodle(+explain), vietnamese-history.

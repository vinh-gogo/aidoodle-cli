# System Patterns

> Authoritative detail lives in `docs/architecture.md` (522 lines; I only read §1–4 so far). This file is the condensed map.

## Guiding principle
**Fact layer deterministic, semantic layer autonomous.** Models are free where verification is impossible (what/how to write); constrained where verification is possible (order, idempotency, phase).

## Three-way split of decisions
1. **Enumerable state transitions → code**: `flow.LoadState → flow.Route → Instruction` (pure function, ten-thousand-combination exhaustive spec test; zero LLM cost).
2. **Bounded semantic judgments → Arbiter (LLM function)**: `arbiter.Collect* → arbiter.Decide* → XxxDecision`. Scenarios: pick planner at start, user-steer triage, failure/deadlock exit. Facts in, structured decision out, mechanical validation, persisted to `meta/decisions.jsonl` (replayable / eval regression). Arbiter uses the *default* model (no role override).
3. **Open-ended creation → Workers (LLM loops)**: Architect (`architect_short` / `architect_long`), Writer, Editor — each own run, context, and model.

New decision points must follow this two-plane symmetry; don't invent new patterns.

## Layers & dependency direction
`entry → host → agents/arbiter → tools → store → domain`; `flow` = pure policy package (above store, below host); `errs/` usable anywhere; `diag/` subscribes to host events + reads store read-only.

| Layer | Package(s) | Does | Doesn't |
|---|---|---|---|
| Entry | `internal/entry/{tui,headless,startup}` | display, input | business decisions |
| Host/Engine | `internal/host` (73 go files, largest) | lifecycle, Route execution, run Workers (`subagent.Runner.Run`), sentinel boundaries, steer orchestration, usage/budget, book lock | literary judgment |
| Arbiter | `internal/arbiter` | structured semantic decisions | create, execute |
| Workers | `internal/agents` | think/write/review | touch Store directly |
| Tools | `internal/tools` | atomic single-file IO, explicit errors, idempotent; commit uses Saga | cross-agent dispatch instructions |
| Store | `internal/store` | filesystem persistence | business logic |
| Domain | `internal/domain` | types (Progress, Checkpoint, …) | — |

Other packages: `bootstrap` (config + first-run setup), `assets` (embedded prompts/refs/styles + loader), `rules` & `userrules` (mechanical baseline + NL rule normalization), `revision` (`/sync` manual edit acceptance), `stylestat`, `models` (registry/pricing, generated), `llmcontract`/`llmretry`, `notify`, `version` (self-update), `eval` (offline harness, `ainovel-cli eval`), `diag`, `chapterfacts`, `logger`, `utils`.

## The four iron laws
1. **Tools return facts, not cross-dispatch instructions** (e.g. `commit_chapter` returns `arc_end`, `needs_expansion`, `final_verdict`, `pending_rewrites` — no `[系统]` strings).
2. **Routing = Flow Router; execution = Engine.** Engine loop: read store → Route → pre-checks → run Worker → sentinel boundary. `nil` = semantic scenario or natural stop. **Deadlock bounds**: same `Agent+Task` after a round → 3 times consult Arbiter, 5 times hard-fuse pause.
3. **Semantic rulings go through Arbiter, every ruling persisted.** Workers keep `CheckpointDeltaGuard` (can't finish if artifact not persisted).
4. **Hard-code boundaries, not unenumerable judgments.** No keyword/score-threshold/rule tables substituting for understanding. Prove decision space is closed and mechanically verifiable before adding a code rule; otherwise improve context/tool expressiveness.

## Observation layer rules
- UI/diag/logs are passive projections of events and read-only artifacts. **`internal/diag` never self-heals, resumes, or alters flow.**
- Three tiers: `agentcore.ProgressPayload` (transport; full error text, no truncation) → `host.Event.Summary` (short) / `Detail` (full) → file logs prefer `Detail`; TUI truncates at render time.
- Host takes the novel-directory lease (flock `.ainovel.lock`) **first**, then logging session, then Store/models/Engine.

## Flat fact layer (only 3 kinds)
- **Progress** (`meta/progress.json`): Phase (`init→premise→outline→writing→complete`, forward-only) + Flow (`writing/reviewing/rewriting/polishing/steering`, validated transitions), CurrentChapter, CompletedChapters, PendingRewrites, Strand/Hook history, CurrentVolume/Arc, Layered.
- **Checkpoint** (`meta/checkpoints.jsonl`): monotonic `Seq`, `Scope{chapter|arc|volume|global}`, appended after each successful tool (plan/draft/commit/review/arc_summary).
- **Artifact**: chapters, outline, characters, world rules, summaries, etc.
- Plus `RunMeta` (`meta/run.json`): user intent — PlanningTier, PlanStart (planner decision), PendingSteer (single in-flight slot), AdvanceMode/AdvancePermitChapter (chapter gate), AdvanceHold.
- `BookMetadata` (`meta/book.json`) is the single source of title/synopsis; `book.md` is a readable projection.

## Persistence patterns
- Single file: `temp + fsync + rename` atomic replace.
- Cross-file: **durable `PendingCommit` Saga** for chapter commit; structural writes use deterministic idempotent replay; failures surfaced explicitly. Don't pretend to have DB transactions.
- Append logs (`*.jsonl`) have unique producers/consumers; derivable views aren't persisted twice.

## Writer fixed tool order (per chapter)
`novel_context → read_chapter → plan_chapter → draft_chapter → check_consistency (must be after draft) → commit_chapter`. `edit_chapter` also exists. Editor tools: `novel_context read_chapter save_review save_arc_summary save_volume_summary`. Architect: `novel_context save_book save_foundation` (+ `expand_next_arc`, `revise_outline`, `reopen_book`, `resolve_outline_feedback` in tools/).

## Content ownership rules (assets/README.md "five questions")
1. Must be *guaranteed*? → code (StopAfterTools / tool guard / Flow Router), not prompt.
2. Adjudication criterion? → table-type: `internal/flow/router.go`; semantic: `assets/prompts/arbiter-*.md`.
3. Role's aesthetic/execution standard? → `assets/prompts/<role>.md`.
4. Mechanically enumerable default rule? → `internal/rules/snapshot.go` `SystemDefaults()`; user rules in `.ainovel/rules/*.md`.
5. Writing knowledge material? → `assets/references/` — **needs 3 wire-ups**: `tools.References` field + `assets/load.go` loadReferences + `novel_context.go` writerReferences/architectReferences. Dropping a file in doesn't auto-load.
Prompts must stay consistent with `novel_context` envelope paths (`working_memory.*`); tool arg shapes only in Tool Schema.

## Context management
Summaries: Volume → Arc → Chapter (sliding window of last 3 chapter summaries). Strategy auto-switches full / sliding / layered by total chapter count. Compression circuit breaker (half-open, retries next round). After FullSummary a recovery pack (chapter plan, outline, character snapshot) is injected.

## Config override layering
`~/.ainovel/config.json` (global) < `./.ainovel/config.json` (project). Scalars override; `providers`/`roles` merge by key. `provider` is a **key name pointer** into `providers`, not a protocol name. Effective reasoning_effort = role intent clamped by model capability at send time (stored intent never mutated). Context window resolution: model-specific → legacy top-level → model registry → 200K fallback.

## Voice / rules customization
Two override dirs: `<outputDir>/style/` (per book) > `~/.ainovel/style/` (global). **Guidance text (voice.md, anti-ai-tone.md) is appended; style presets (styles/, genres/) are replaced whole-file.** Execution-protocol prompts are not overridable. Mechanical constraints belong in `rules/`.

## Doodle Explainer Architecture Patterns (Fork-specific)
1. **Doodle Concept Mapping**:
   - 1 Book = 1 Series/Season (default 3 deep-dive episodes for a single topic).
   - 1 Chapter = 1 Video Script (> 5 minutes, 300–600s, ~750–1200 spoken words).
   - Premise = Series Bible (11 standardized Vietnamese H2 headings, `docs/script-format.md`).
   - Characters = Recurring stick figures (Host + Companion/Skeptic + Mascot).
   - World Rules = Stick-figure universe rules + purposeful anachronism.
2. **5-Stage Script Architecture**:
   - `HOOK 0:00-0:03` -> `CẢNH 1 0:03-0:45` (đặt vấn đề đời thực) -> `CẢNH 2-4 0:45-3:45` (thân bài 3 chặng, Tái Hook mỗi 60-90s) -> `CẢNH 5 3:45-4:30` (Reframe & hành động nhỏ) -> `CHỐT 4:30-5:15+` (Loop hook & tương tác).
3. **1:1 Alternating Voice-Visual Beats Pattern**:
   - Within each scene, strictly avoid monolithic paragraphs.
   - Alternate pairs of `LỜI:` (1-2 sentences, 10-20 words / 3-6s) and immediate `HÌNH:` (matching stick-figure action/diagram).
   - Eliminate `CHỮ:` tag completely; visuals change dynamically every 3-6 seconds.
4. **Dynamic Output Isolation**:
   - Every `/new` command automatically spins up a timestamped folder: `output/novel-YYYYMMDD-HHMM`.

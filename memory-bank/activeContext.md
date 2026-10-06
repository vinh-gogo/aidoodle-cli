# Active Context

_Last updated: 2026-10-06 (initial bootstrap + direction decision)_

## DIRECTION DECISION (2026-10-06) — repo is being converted
User wants this repo to become a **Vietnamese doodle-explainer script engine for TikTok on trending topics**. Confirmed choices:
- Topics: **auto-fetched trends (VN)**. Sources verified: Google Trends RSS `https://trends.google.com/trending/rss?geo=VN` (works; snippets empty, contains junk like lottery `xsmb`), VnExpress RSS works. **TikTok Creative Center has no public API; scraping violates ToS → manual import only.**
- Unit: **1 video = 1 "chapter"**, 1 "book" = 1 series/season. Reuse the whole Engine/Route/Store/checkpoint pipeline.
- Style: **stone-age stick figures explaining modern topics** (successor to `stone-age-doodle-explain`).
- Length: **60–180s (~150–450 spoken words)**.
- Full plan (P0–P6, risks, open questions): artifact `doodle_explainer_plan.md` in `C:\Users\lea26\.gemini\antigravity-cli\brain\ddc11f54-f02e-4ccc-9426-12497c9b6842\`. Plan is NOT yet approved/started.
- Before starting: `git tag novel-baseline`, branch `doodle-explainer`.

### Code-verified findings that block Vietnamese output (fix first = P1)
1. `internal/rules/snapshot.go` `SystemDefaults()` blacklist/fatigue words are **Chinese only** → mechanical check useless for Vietnamese.
2. `internal/rules/lint.go` `non_cjk_fragments` regex `[A-Za-z]{2,}` matches ordinary Vietnamese words (false warnings) and does NOT catch Han leakage → replace with `han_residue`.
3. `internal/domain/chapter.go:31` `WordCount` = rune count → wrong for Vietnamese; switch to whitespace tokens + speech-seconds.
4. `internal/tools/premise_structure.go` heading aliases are Chinese; Vietnamese architect prompt emits Vietnamese headings → `template_ready` always false (informational only, used in `novel_context`).
5. `lint.go` `markdown_residue` warns on `**` and any `#` line after the first → script format must use plain tag lines (`HOOK/CẢNH/LỜI:/HÌNH:/CHỮ:`), not Markdown bold/headings.
6. Many Chinese runtime strings still reach the model from Go (`novel_context.go`, `commit_chapter.go`, `save_review.go`, `save_foundation.go`, `chapterfacts/facts.go`, `llmcontract/validate.go`, `ctxpack/builder.go`, `domain/writing.go`). `assets/prompts/*.md` are ALREADY Vietnamese (correction to earlier note).
7. `assets/styles/stone-age*.md` are 3 same-size copies; fantasy/vietnamese-history genre refs forbid "xuyên không ngôn ngữ" (anachronism) — new style must explicitly allow it.
8. Not yet read: `save_review.go`/`domain/review.go` (are the 7 review dimension keys hard-coded?), `stylestat` Chinese bias, `agentcore.EstimateTokens` on Vietnamese.

## Current focus (updated 2026-10-06, later session)
Branch `doodle-explainer` (tag `novel-baseline` = pre-change state; recover anything with `git checkout novel-baseline -- <path>`).
- **P0 done.**
- **P1 committed** (`d649d58 P1: harden engine for Vietnamese output (word count, rules, lint, premise headings, model-facing strings)`).
- **P2 committed** (`45b228e P2: script format, validator, series bible headings, doodle-explainer style and prompts`).
- **P3 done** (Editor rubric, stylestat Vietnamese port, eval smoke cases):
  - `internal/stylestat/stylestat.go`: Đổi 8 pattern AI cliché sang tiếng Việt (hỗ trợ song ngữ), regex câu thời gian, n-gram tiếng Việt 2-4 từ với stopwords cạnh `vnEdgeStops`, nhận diện tiền tố tiêu đề `Tập N`.
  - `internal/stylestat/stylestat_test.go`: Cập nhật `want` map sang tên tiếng Việt, thêm test suite tiếng Việt đầy đủ (`TestComputePatterns_Vietnamese`, `TestComputeTopPhrases_Vietnamese`, `TestComputeRepeatedSentences_Vietnamese`, `TestComputeTitleFormats_Vietnamese`).
  - `internal/tools/commit_chapter_test.go`: Cập nhật test pattern `Câu phủ định`.
  - `internal/diag/rules_quality.go`: Đổi thông báo `HookWeakChain` sang "liên tiếp %d tập".
  - `evals/cases/smoke/`: Cập nhật 3 case smoke (`architect_short.json`, `writer_first_chapter.json`, `architect_long.json`) sang kịch bản doodle explainer tiếng Việt (tài chính, lạm phát, công nghệ).
  - Toàn bộ test suite `go test -buildvcs=false -count=1 ./...` đều PASS.

**Decision: user deleted `.github/` (CI/docker/release workflows). Keep deleted; do NOT restore unless asked.**

**Next: Commit P3** sau đó chuyển sang **P4** (Trend Intake tự động: Google Trends RSS VN, VnExpress RSS, article fetcher, Arbiter topic selection).


## Recent changes (from git history)
- **2026-10-03 `b67734c` "init vietnamese from ainovel-cli"** — fork baseline; Vietnamese localization of user-facing/runtime strings.
- **2026-10-04 `13f98eb` "add style"** (31 files, +446/−68):
  - New styles: `psychological` / `psychology`, `stone-age-doodle` / `stone-age-doodle-explain` / `stone-age`, `vietnamese-history` / `viet-history`, plus matching `assets/references/genres/*/{arc-templates,style-references}.md`.
  - `assets/prompts/{architect-long,architect-short,editor}.md`, `assets/voice.md`, `assets/testdata/writer-golden.md` edited (a few lines each) — likely Vietnamese-output enforcement.
  - `assets/load_test.go` +12 lines (style loading test).
  - `internal/agents/{architect,editor}_context.go`: summary system/user prompts rewritten Chinese → Vietnamese ("BẮT BUỘC: … 100% tiếng Việt").
  - `internal/host/book_lock.go`: `isAccessDenied()`, friendlier errors, lock perms `0o666`.
  - Test strings adjusted in `guard_test.go`, `diag/export_test.go`, `diag/rules_quality_test.go`, `host/{budget,cocreate_stage,engine}_test.go`.
  - `cmd/ainovel-cli/main.go`: messages now Vietnamese.
- Git working tree was clean when inspected (before creating `memory-bank/`).

## Local environment state observed
- `~/.ainovel` config targets **Ollama**, model `qwen3.5-4b-16k` (from `output/novel/meta/run.json`; `Modelfile` in repo says num_ctx 32768 while `run.sh` default is 16384 — inconsistent).
- Two books under `output/`:
  - `output/novel/` — phase `outline`, 0 chapters, `layered: true`, planner `architect_long`, start prompt was literally `"hi"` (run started 2026-10-03T22:04); `book.json` has a Vietnamese title ("Minh Va Loi Logic Cuu Vu Tu") — a cyberpunk/ma-pháp concept the model invented.
  - `output/stone-age/` — phase `outline`, `total_chapters: 94`, volume 1 / arc 1, 0 chapters written; has `.ainovel.lock`; Vietnamese title/synopsis present in `meta/book.json` (exact text not decoded — PowerShell mojibake). Style used is unknown (likely a stone-age one).
- Neither book has written a chapter yet.

## Active decisions / patterns to follow
- Keep upstream architecture untouched; fork changes are **language + styles + Windows robustness**. Prefer minimal diffs against upstream so merges stay possible.
- Any new Vietnamese-output prompt text must forbid Chinese output explicitly (existing convention).
- When adding a style: add `assets/styles/<name>.md` (filename = style name) **and** `assets/references/genres/<name>/{arc-templates,style-references}.md`; check `assets/load.go` and `load_test.go`.
- When adding any reference file: remember the 3 wire-ups (see systemPatterns.md).
- Decision-table (`flow.Route`) changes: update the exhaustive spec **before** the implementation.

## Next steps (suggested, pending user direction)
1. Ask user what to work on (candidates below).
2. Run `go build ./...`, `go vet ./...`, `go test ./...` to establish baseline on this machine (Windows, Go 1.27) — not yet done.
3. Candidates the repo state hints at:
   - Make a full run succeed with the local 4B model (planning stalled at `outline` for both books).
   - Finish Vietnamese localization (remaining Chinese in prompts `assets/prompts/*.md`, `docs/`, README, TUI strings; check `rg "[\u4e00-\u9fff]"` in `internal/`).
   - Reconcile alias styles (`psychology`, `stone-age`, `viet-history`) — duplicates of the canonical files; verify genre reference loading for aliases.
   - Reconcile `Modelfile` (32768) vs `run.sh` (16384).

## Learnings / insights
- Memory bank was created by reading: README, go.mod, main.go, config.example.jsonc, run.sh, ci.yml, docs/architecture.md §1–4 (rest of docs unread), assets/README.md, last commit diff. **Source code of host/tools/store/flow/arbiter has not been read** — descriptions of them come from docs, not code. Verify before relying on specifics.
- The repo root contains stray binaries (`OllamaSetup.exe`, `ainovel-cli.exe`) — ignored by git.

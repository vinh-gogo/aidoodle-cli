# Progress

_Last updated: 2026-10-06 (initial bootstrap)_

## What works (per upstream docs; not re-verified locally)
- Deterministic Engine + Route table, Arbiter decisions, 3 Workers, Store with atomic IO, checkpoints, commit Saga.
- Rolling volume/arc planning, layered summaries, related-chapter recommendation, compression pipeline.
- TUI (+ co-create, `/config`, `/model`, `/review`, `/next`, `/diag`, `/simulate`, `/importsim`, `/sync`, `/import`, `/export`, `/start`) and headless mode; Docker image; self-update; notify; budget fuse; eval harness (`evals/cases/smoke`).
- Many test files exist in every package (e.g. `internal/tools/commit_chapter_test.go` 60KB, `internal/host` 73 files).

## Fork-specific work done
- [x] Vietnamese localization of CLI messages (`main.go`), lock errors (`book_lock.go`), architect/editor summary prompts, `config.example.jsonc`, `run.sh`.
- [x] New genre styles: psychological, stone-age-doodle(+explain), vietnamese-history (+ alias files).
- [x] Windows "Access is denied" handling for the book lock file.
- [x] Memory Bank bootstrapped (this folder).

## Doodle-explainer conversion progress (see activeContext.md for details)
- [x] P0: tag `novel-baseline`, branch `doodle-explainer`; baseline build/vet/test were green (Go 1.27.0, go.mod 1.25.5).
- [x] P1: WordCount/SpeechSeconds, Vietnamese rule defaults + lint (`han_residue`), premise headings, tools/*.go strings translated; full `go test` green after test-needle updates.
- [x] P1.5 second pass: model-facing Chinese literals translated in tools, ctxpack, chapterfacts, llmcontract, store, flow, revision (test needles updated; `go build/vet/test ./...` and gofmt all green on 2026-10-06). Left: slog, comments, `guard/subagent_guards.go` match literals "卷摘要"/"弧摘要", `store/session.go` chapterRe, `host/imp`, tui, eval, diag, host/sim. `internal/bootstrap/config.example.jsonc` MUST stay byte-identical to root `config.example.jsonc` (test enforces).
- [x] Doc note on token under-estimation in `config.example.jsonc`.
- [x] P2 script format + prompts + style:
  - Spec [docs/script-format.md](file:///D:/ainovel-cli/docs/script-format.md) chốt chuẩn kịch bản (HOOK/CẢNH/CHỐT, LỜI/HÌNH/CHỮ/ÂM, CAPTION/HASHTAG/NGUỒN/CẦN KIỂM CHỨNG) và series bible headings.
  - Validator [`internal/tools/script_format.go`](file:///D:/ainovel-cli/internal/tools/script_format.go) (chỉ cảnh báo sự thật, không chặn commit) tích hợp vào `commit_chapter`.
  - Tiêu đề series bible trong [`internal/tools/premise_structure.go`](file:///D:/ainovel-cli/internal/tools/premise_structure.go) đồng bộ 100% với spec.
  - Tạo style mới [`assets/styles/doodle-explainer.md`](file:///D:/ainovel-cli/assets/styles/doodle-explainer.md), đặt làm style mặc định (`doodle-explainer`).
  - Tạo references thể loại [`assets/references/genres/doodle-explainer/{style-references,arc-templates}.md`](file:///D:/ainovel-cli/assets/references/genres/doodle-explainer/) và 2 tài liệu tham khảo mới [`doodle-visual-language.md`](file:///D:/ainovel-cli/assets/references/doodle-visual-language.md), [`fact-grounding.md`](file:///D:/ainovel-cli/assets/references/fact-grounding.md) (nối dây qua `novel_context.go` và `assets/load.go`).
  - Viết lại toàn bộ 13 references sang lĩnh vực kịch bản video ngắn tiếng Việt, sạch chữ Hán.
  - Viết lại [`assets/prompts/writer.md`](file:///D:/ainovel-cli/assets/prompts/writer.md), [`assets/voice.md`](file:///D:/ainovel-cli/assets/voice.md), tái tạo [`assets/testdata/writer-golden.md`](file:///D:/ainovel-cli/assets/testdata/writer-golden.md).
  - Viết lại [`assets/prompts/architect-short.md`](file:///D:/ainovel-cli/assets/prompts/architect-short.md), [`assets/prompts/architect-long.md`](file:///D:/ainovel-cli/assets/prompts/architect-long.md), [`assets/prompts/editor.md`](file:///D:/ainovel-cli/assets/prompts/editor.md) và tinh chỉnh các prompt Arbiter.
- [x] P3 editor rubric, stylestat Vietnamese port, smoke eval cases:
  - Cập nhật `internal/stylestat/stylestat.go` cho tiếng Việt: 8 nhóm AI cliché, n-gram tiếng Việt 2-4 từ, `vnEdgeStops`, `openingTimeRe` tiếng Việt, nhận diện tiền tố `Tập N`.
  - Cập nhật và bổ sung test tiếng Việt trong `internal/stylestat/stylestat_test.go` (`TestComputePatterns_Vietnamese`, `TestComputeTopPhrases_Vietnamese`, `TestComputeRepeatedSentences_Vietnamese`, `TestComputeTitleFormats_Vietnamese`).
  - Cập nhật test `internal/tools/commit_chapter_test.go` chấp nhận pattern "Câu phủ định".
  - Chuẩn hóa thông báo `HookWeakChain` trong `internal/diag/rules_quality.go` thành "liên tiếp %d tập".
  - Chuyển 3 case smoke trong `evals/cases/smoke/` (`architect_short.json`, `writer_first_chapter.json`, `architect_long.json`) sang nội dung doodle explainer tiếng Việt.
  - Nghiệm thu: `go test -buildvcs=false -count=1 ./...` toàn bộ repo PASS 100%.
- [ ] P4 trend intake (Google Trends RSS / VnExpress RSS fetcher, source pack, Arbiter topic selection).
- [ ] P5 safety & human gate; P6 export & docs & TUI.
- `.github/` deleted by user's choice (recoverable from `novel-baseline`).

## Verified on this machine
- [x] `go build ./...` / `go vet ./...` / `go test -buildvcs=false -count=1 ./...` green at baseline (needs `-buildvcs=false` in sandbox).
- [ ] Whether `ainovel-cli.exe` in repo root corresponds to current source.
- [ ] Whether alias styles load genre references.


## What's left / candidate work
- [ ] Get an end-to-end novel run past `outline` phase with the local Ollama model (both `output/novel` and `output/stone-age` have 0 chapters).
- [ ] Complete Vietnamese localization (README, docs, remaining Chinese prompts in `assets/prompts/*.md`, writer/arbiter/import prompts, TUI text) — scope unknown until grepped.
- [ ] Clean up duplicate style aliases or document them.
- [ ] Align `Modelfile` num_ctx (32768) with `run.sh` default (16384).
- [ ] Decide whether `memory-bank/` gets committed.

## Known issues / risks
- 4B local model (`qwen3.5-4b-16k`) likely too weak/slow for structured tool-calling long-form planning; first book (`output/novel`) was started with prompt `"hi"` and stalled at `outline` with 0 total chapters.
- `run.json` for `output/novel` shows plan_start text possibly in a different language than requested — model drift on weak models.
- Mojibake when reading UTF-8 files through PowerShell (display only).
- Lock file perms `0o666` is more permissive than upstream `0o600` (intentional Windows workaround; revisit for security on multi-user hosts).
- `.gitignore` contains a garbled comment line (encoding damage in a Chinese comment about `.ainovel/`); harmless.

## Evolution of project decisions
- 2026-07-12 (upstream): Coordinator LLM long loop retired → deterministic Engine + Arbiter functions (see `docs/engine-arbiter.md`, `docs/engine-rfc.md`).
- Upstream moved built-in rules from `assets/rules/` (deprecated) to code `rules.SystemDefaults()` + natural-language user rules normalized into a snapshot.
- Upstream added Voice Layer (`docs/voice-layer.md`): append-only guidance, whole-file replace for style presets.
- 2026-10-03/04 (fork): forked for Vietnamese output + local Ollama usage + extra genre styles.

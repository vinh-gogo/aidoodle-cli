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
- [x] P4 trend intake & fact-grounding:
  - Gói mới `internal/host/trend/`: `types.go`, `fetcher.go`, `article.go`, `store.go`, `runner.go` và toàn bộ unit test với httptest fixtures.
  - Arbiter topic selection: `internal/arbiter/topics.go`, `topics_test.go`, prompt `assets/prompts/arbiter-topics.md` lọc rác/xổ số/tai nạn và ưu tiên góc nhìn đồ đá viral. Nối vào `assets/load.go`.
  - Cấu hình: `TrendsConfig` trong `internal/bootstrap/config.go`, `configfile.go`, cập nhật `config.example.jsonc` (giữ byte-identical).
  - Tích hợp: `novel_context.go` và `novel_context_builders.go` nạp `source_pack` cho Writer và `trend_brief` cho Architect.
  - CLI: `cmd/ainovel-cli/main.go` hỗ trợ cờ `--trends`.
  - Nghiệm thu: `go test -buildvcs=false -count=1 ./...` toàn bộ 35 package pass 100%.
- [x] P5 safety & human gate:
  - Tài liệu chuẩn [`assets/references/tiktok-content-safety.md`](file:///D:/ainovel-cli/assets/references/tiktok-content-safety.md) kết nối Luật An ninh mạng 2018, Nghị định 15/2020/NĐ-CP Điều 101 và Tiêu chuẩn cộng đồng TikTok (Vùng đỏ cấm tuyệt đối, Vùng vàng thận trọng, Vùng xanh, cơ chế bám nguồn và cổng duyệt).
  - Nạp vùng cấm kỵ mặc định vào `rules.SystemDefaults().Preferences` tự động bảo vệ mọi phiên sáng tác.
  - Validator `internal/tools/script_format.go` bắt buộc kiểm tra thẻ chân `CẦN KIỂM CHỨNG:` (`script_missing_unverified` warning), cập nhật test và `docs/script-format.md`.
  - Cấu hình `advance_mode` ("auto" / "review") trong `bootstrap.Config`, `config.example.jsonc` và mặc định `review` cho series xu hướng mới trong `host.New`. Cung cấp cờ `--review` và `--next` (cho headless) trong `main.go`.
  - Nghiệm thu: `go test -buildvcs=false -count=1 ./...` toàn bộ 35 package pass 100%.
- [x] P6 export video formats, TUI polish & Vietnamese README:
  - Gói xuất video TikTok `internal/host/exp/video.go`: `slugify` tiếng Việt chuẩn (bỏ dấu NFD, chuyển đ/Đ thành d, sinh slug URL an toàn), `renderVideoPackage` xuất `scripts/NN-slug.md`, `voiceover/NN-slug.txt`, `shotlist.csv` (UTF-8 BOM), `publish.csv` (UTF-8 BOM).
  - Tích hợp `FormatVideo` ("video") trong `internal/host/exp/types.go` và `exporter.go`. Unit test `TestSlugify` và `TestRun_VideoFormat` pass 100%.
  - TUI & CLI: `/export --video` hoặc `format=video` trong `internal/entry/tui/export.go`, `commands.go`.
  - Cập nhật TUI polish: màn hình chào mừng (`panels.go`), gợi ý chủ đề doodle đồ đá, placeholder cocreate và model update (`cocreate.go`, `model_update.go`).
  - Viết lại toàn bộ `README.md` sang tiếng Việt, định vị đúng sản phẩm công cụ tạo kịch bản video TikTok Doodle Explainer.
  - Nghiệm thu: `gofmt`, `go vet ./...`, `go test -buildvcs=false -count=1 ./...` toàn bộ 35 package pass 100%.
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

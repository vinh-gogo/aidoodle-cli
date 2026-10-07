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

## Post-P6 Enhancements & Fixes (2026-10-07)
- [x] **Branch `route`: AI Working Mode Selector (`/mode`) & Specific Style Keys**:
  - Tạo nhánh mới `route`.
  - Lệnh `/mode` (alias: `/topic`, `/topics`) mở modal overlay trực quan chọn chế độ AI làm việc:
    1. `[1] Tiểu thuyết / Manga` (Style key: `"novel-manga"`): Sáng tác tiểu thuyết dài kỳ, manga, thế giới quan và diễn biến tâm lý sâu sắc.
    2. `[2] Doodle Explainer` (Style key: `"doodle-explainer"`): Biên kịch video người que đồ đá giải thích kiến thức TikTok / YouTube Shorts.
    3. `[3] Chế độ mở rộng` (Chưa có / Đang suy nghĩ ở bước tiếp theo): Placeholder cho các chế độ tiếp theo.
  - Đặt tên cụ thể cho 2 style (`"novel-manga"` và `"doodle-explainer"`), loại bỏ phụ thuộc vào `style == "default"`.
  - Phân tách độc lập kho prompt, voice và reference cho từng mode trong `assets/`:
    - `assets/prompts/modes/novel-manga/` vs `assets/prompts/modes/doodle-explainer/` (Writer, Architect-Short, Architect-Long, Editor).
    - `assets/voices/novel-manga.md` vs `assets/voices/doodle-explainer.md`.
    - `assets/references/modes/novel-manga/` vs `assets/references/modes/doodle-explainer/`.
    - Style: `assets/styles/novel-manga.md` vs `assets/styles/doodle-explainer.md`.
  - Cơ chế nạp động theo style trong `assets/load.go`.
  - Cô lập kiểm tra `lintScript` trong `internal/tools/commit_chapter.go` chỉ dành riêng cho `doodle-explainer`.
  - Tự động kích hoạt, cập nhật runtime và ghi nhận vào `.ainovel/config.json`.
  - Hỗ trợ auto-loading file `.env` cục bộ cho các khóa API (Tavily...).
  - Bộ kiểm thử unit test và toàn bộ repo 36 packages pass 100%.
- [x] **Timestamped Output Directory Isolation** (commit `e81b404`):
  - Khi người dùng gõ `/new`, hệ thống tự động sinh thư mục output có timestamp `output/novel-YYYYMMDD-HHMM` thay vì đè `output/novel`.
- [x] **5-Minute+ Doodle Explainer Standard** (commit `e21740a`):
  - Chuẩn hóa thời lượng video: từ 5 phút trở lên (300–600 giây, 700–1500 từ LỜI).
  - Cấu trúc 5 giai đoạn: Hook 3s -> Phần đầu (0:03-0:45) -> Thân 3 chặng có Tái Hook mỗi 60-90s (0:45-3:45) -> Reframe & Hành động nhỏ (3:45-4:30) -> Chốt loop (4:30-5:15+).
  - Cập nhật validator `internal/tools/script_format.go` (`scriptMinWords=700`, `scriptMaxWords=1500`, `scriptMinSeconds=300`, `scriptMaxSeconds=600`).
- [x] **Loại bỏ hoàn toàn thẻ `CHỮ:`**:
  - Script format chỉ giữ lại các thẻ `LỜI:` (voiceover), `HÌNH:` (mô tả hình vẽ que), `ÂM:` (sfx/nhạc) và chân kịch bản. Mọi thông tin chữ được lồng vào HÌNH.
- [x] **Humor Specialist & Archetypes** (commit `f022efa`):
  - Tạo `assets/references/humor-relatability.md`, nối dây qua `novel_context.go`, `load.go`.
  - Hướng dẫn sáng tạo nhân vật que theo chủ đề (Que Lanh, Que Bự, Cục Đá Im Lặng, Que Mồ, Que Chạy...), đệm nhẹ tiếng cười đồng cảm từ nghịch lý đời thường (Tây Du Ký, Thủy Hử).
- [x] **Anti-Drift & Topic Fidelity** (commit `9e1c1d6`):
  - Ràng buộc 100% bám sát chủ đề được giao, nghiêm cấm phân nhánh lan man (ví dụ hỏi mất lông thì không tự ý nhảy sang não to, đứng thẳng, phát minh lửa). Kết luận phải giải thích trọn vẹn chủ đề.
- [x] **Sửa lỗi `save_foundation` InputValidationError** (commit `dd1b500`):
  - Lỗi `The required parameter 'type' is missing`: nới lỏng schema bắt buộc, bổ sung suy luận tham số từ cú pháp Markdown/JSON, nhận diện trường chuyên dụng, chuẩn hóa alias tiếng Việt ("dàn ý", "nhân vật"...). Bổ sung 7 unit tests (29/29 tests pass).
- [x] **Mặc định Series 3 tập chuyên sâu** (commit `bc2230c`):
  - Khi người dùng đưa vào một chủ đề đơn lẻ, hệ thống mặc định lập dàn ý gồm **đúng 3 tập chuyên sâu** (mỗi tập > 5 phút) mổ xẻ các góc cạnh khác nhau của cùng một chủ đề (Tập 1: Nghịch lý; Tập 2: So sánh & Tình huống thực chiến; Tập 3: Tranh luận khoa học & Dấu ấn hiện đại). Cho phép tùy biến số tập nếu người dùng chỉ định.
- [x] **1:1 Alternating Voice-Visual Beats (Khớp nhịp LỜI - HÌNH)** (commit `93d868a`):
  - Giải quyết triệt để vấn đề "1 đoạn văn dài 40-60s nhưng chỉ có 1 hình ảnh gây chết hình": Bắt buộc phân tách cảnh thành các cặp `LỜI:` - `HÌNH:` xen kẽ liên tục (mỗi câu thoại 10-20 từ / 3-6s đi liền ngay một thẻ HÌNH tương ứng). Cập nhật `docs/script-format.md`, `writer.md`, `editor.md`, `chapter-guide.md`, `chapter-template.md`, `doodle-visual-language.md`, `writer-golden.md`.
- [x] **Khắc phục lỗi LLM Repetition Loop (Degeneration Trap)**:
  - Phân tích hiện tượng model lặp vô tận (ví dụ `- true\n- true...`). Hướng dẫn cấu hình `extra_body` với `frequency_penalty: 0.3`, `presence_penalty: 0.2`, `temperature: 0.7`, `repetition_penalty: 1.1`.
- [x] **Tích hợp Tavily Search & Crawl cho cơ sở khoa học & kiểm chứng** (commit `942b895`):
  - Tạo gói `internal/tavily/` (`Client`, `Search`, `Crawl`, `Extract`, `SearchAndBuildSourcePack`).
  - Tạo 2 LLM tools: `tavily_search` và `tavily_crawl` trong `internal/tools/` cho Architect và Writer.
  - Tự động tra cứu Tavily Search khi khởi tạo chủ đề mới (`StartPrepared`), tự động lưu `SourcePack` vào `outputDir/meta/trends/sources/`.
  - Nạp `source_pack` tự động vào `novel_context` của cả Architect và Writer.
  - Cập nhật prompt: bắt buộc trích dẫn bài báo/tạp chí uy tín và URL vào thẻ `NGUỒN:`, nêu các điểm khoa học tranh luận/giả thuyết đối trọng vào `CẦN KIỂM CHỨNG:`.
  - Tái tạo `assets/testdata/writer-golden.md` chính xác từng byte.
  - Toàn bộ test của 36 package pass 100%.

## Verified on this machine
- [x] `go build ./...` / `go vet ./...` / `go test -buildvcs=false -count=1 ./assets/... ./internal/tools/...` green.
- [x] `ainovel-cli.exe` biên dịch thành công từ commit mới nhất.
- [x] Tạo kịch bản hoàn chỉnh thực tế trên Kaggle ngrok (`qwen-27b`) ra kết quả 5 phút 15 giây, đúng format, đầy đủ 3 tập.

## Evolution of project decisions
- 2026-07-12 (upstream): Coordinator LLM long loop retired → deterministic Engine + Arbiter functions.
- 2026-10-03/04 (fork): forked for Vietnamese output + local Ollama usage + extra genre styles.
- 2026-10-06 (fork): Chuyển hướng toàn diện sang TikTok Doodle Explainer Video Scriptwriting Engine.
- 2026-10-07 (fork): Chuẩn hóa video 5 phút, 5 giai đoạn, cặp LỜI-HÌNH 1:1, series 3 tập chuyên sâu, thư mục timestamp, sửa lỗi save_foundation.

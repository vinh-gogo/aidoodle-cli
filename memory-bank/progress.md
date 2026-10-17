# Progress

_Last updated: 2026-10-17 (Production Readiness Upgrade v1.0.1: Security Hardening M1, Context & Mechanical Validators M2, W3C EPUB 3.3 Exporter M3, 7-Mode Evals, CI Release Hygiene & Documentation Alignment M4)_

## What works (per upstream docs; verified locally)
- Deterministic Engine + Route table, Arbiter decisions, 3 Workers, Store with atomic IO, checkpoints, commit Saga.
- Rolling volume/arc planning, layered summaries, related-chapter recommendation, compression pipeline.
- TUI (+ co-create, `/config`, `/model`, `/review`, `/next`, `/diag`, `/simulate`, `/importsim`, `/sync`, `/import`, `/export`, `/start`, `/mode`, `/dash`, `/new`) and headless mode.
- **Web UI Dashboard (`ainovel-cli web`)**: Retro terminal layout, 3-panel UI, realtime SSE stream, loopback 127.0.0.1 default, Origin/CSRF defense, SafeProjectDir traversal defense, XSS sanitized, non-blocking toast notifications.
- **Multi-format Exporter**: TXT, Video Scripts Package, native Word (.docx) OpenXML, and **W3C EPUB 3.3 compliant exporter** with OCF uncompressed mimetype, OPF 3.0, Accessibility 1.1, hierarchical nav.xhtml, XHTML5 Unicode NFC normalization and XML control char stripping.
- **Multi-mode support (7 specialized AI modes)**: 5 Novel modes (`novel-manga`, `suspense`, `romance`, `fantasy`, `vietnamese-history`) + 2 Video Script modes (`doodle-explainer`, `behavioral-psychology`).
- **6 Mechanical Output Validators**: CJK character rejection, word bounds validation, repetition loop detection, truncated mid-sentence detection, banned convert terms, leaked video tags guard.
- **7-Mode Evaluation Suite & Mock LLM Smoke Suite**: Genre-specific eval cases and rubrics for all 7 modes in `evals/cases/` and `evals/rubrics/`, deterministic offline smoke test in `evals/smoke_test.go` generating short book end-to-end without network/API keys.
- **CI/CD & Release Hygiene**: Matrix CI on Ubuntu + Windows (`.github/workflows/ci.yml`), multi-platform release workflow (`.github/workflows/release.yml`) generating SHA-256 `checksums.txt` compatible with `internal/version/update.go`, operational runbook `docs/runbook.md`, comprehensive Vietnamese `README.md`, and `CHANGELOG.md`.
- Automated CI/CD: GitHub Actions building Windows AMD64 & ARM64 executables, Release v1.0.1.
- Full test suite passing 100% across all 26 packages (unit, integration, e2e, evals).

---

## Recent Milestones (2026-10-17)
- [x] **Milestone M1 (Security Hardening — SEC-01 to SEC-07)**: Web UI host binding & opt-in auth, CSRF & Origin defense, SafeProjectDir path traversal defense, frontend XSS sanitization & toast notifications, Colab gateway API key auth & rate limiting, diagnostic secret redaction, Tavily prompt injection isolation.
- [x] **Milestone M2 (Context Budget & Mechanical Validators — CTX-01 to VAL-06)**: Context window 32k/64k unification, token budget enforcement $\le$ 60% across 7 modes, non-retryable 4xx handling, proactive gateway health checks, $0 model circuit breakers, 6 mechanical output validators.
- [x] **Milestone M3 (W3C EPUB 3.3 Exporter — EPUB-01 to EPUB-07)**: OCF container zip.Store, OPF 3.0 metadata, EPUB Accessibility 1.1, nested Volume->Arc->Chapter nav.xhtml with landmarks, XHTML5 NFC normalization & XML control character stripping, atomic file writes, fallback chapter pipeline, RFC 5987 headers, TUI `--epub` flag.
- [x] **Milestone M4 (7-Mode Evals, CI Hygiene & Documentation — EVL-01 to DOC-01)**: 7-mode eval cases & rubrics, offline mock-LLM smoke test for all 7 modes, GitHub Actions CI & release workflows with SHA-256 checksums, CHANGELOG.md, operational runbook (`docs/runbook.md`), comprehensive Vietnamese README.md, memory-bank alignment, production-readiness checkoff.

---

## Historical Milestones (2026-10-15)
- [x] **Mở Rộng 3 Thể Loại Tiểu Thuyết Nổi Bật Nhất (`suspense`, `romance`, `fantasy`) & Hệ Thống 7 Modes (`/mode` 1–7)**:
  - Xuất bản báo cáo nghiên cứu thị trường `docs/novel-genres-report.md` phân tích 8 nhóm thể loại và chọn 3 trụ cột tiểu thuyết ăn khách nhất.
  - Xây dựng trọn bộ 24 file tài nguyên chuyên sâu (Prompts `architect-short`, `architect-long`, `writer`, `editor`; `voices/`; `styles/`; `references/genres/{suspense,romance,fantasy}/{style-references,arc-templates}.md`) cho:
    1. **Trinh thám – Tâm lý Tội phạm & Kinh dị Tâm linh (`suspense`)**: Fair-Play Whodunit, Sổ cái Manh mối 3:2, Ma trận Chân tướng Kép, Đồng hồ Đếm ngược & Kinh dị Dân gian Việt Nam.
    2. **Ngôn tình – Chữa lành & Tâm lý Tình cảm Hiện đại (`romance`)**: Hồ sơ Vết thương Tâm lý (Lie/Ghost), Cân bằng Sức hút Kép, Thang đo Cảm xúc 6 Cấp độ, Đối thoại Subtext & Chống sáo ngữ tổng tài.
    3. **Kỳ ảo – Tiên hiệp / Thăng cấp & Thế giới Đa tầng (`fantasy`)**: Hệ thống Sức mạnh Quy tắc Cứng (Định luật Sanderson), Bản đồ Đa tầng, Chiến đấu 3 Pha & Bộ lọc Chống văn phong Convert Hán-Việt.
  - Cập nhật `assets/load.go` và `internal/tools/novel_context.go` định tuyến chuẩn xác 3 thể loại mới, loại bỏ rò rỉ tài liệu doodle và giữ `style_reference` xuyên suốt các chương.
  - Đồng bộ hóa TUI (`internal/entry/tui/command_mode.go`, `command_dash.go`) và Web UI (`internal/entry/web/project.go`, `ui/index.html`) lên 7 chế độ với phím tắt `1`–`7` và nhãn `/dash` trực quan.
- [x] **Chuyển Đổi Từ Ngrok Sang Cloudflare Tunnel Cố Định (`https://api.gloryvocab.com/v1`)**:
  - Cập nhật `kaggle/colab_llama_cpp.ipynb` chạy Unified FastAPI Gateway (Port 8000) cho cả LLM 26B (`main-llm` / `gemma-4-26B-A4B-it-Q4_K_M.gguf` / `gemma-4-26b`) và STT (`gemma-4-stt`) qua Cloudflare Tunnel `https://api.gloryvocab.com/v1`.
  - Dọn dẹp các file notebook/script Ngrok cũ trong `kaggle/`.
  - Cập nhật `config/config.json`, `.ainovel/config.json`, `~/.ainovel/config.json`, `config.example.jsonc` và `internal/bootstrap/configfile.go` (bổ sung gộp `StreamIdleTimeout` trong `mergeConfig`), đồng bộ `context_window: 32768`.
- [x] **Sửa Triệt Để Lỗi `openai: stream ended before [DONE]` & Tràn Ngữ Cảnh 32K**:
  - Thu gọn `reference_pack.references` trong `internal/tools/novel_context.go` (`compactReference`, lọc bỏ tài liệu doodle trong mode `novel-manga` và `vietnamese-history`) và `compactSourcePack` trong `internal/tools/novel_context_builders.go`, giảm dung lượng `novel_context` từ 70–122 KB xuống ~20–25 KB.
  - Tinh chỉnh `ToolResultMicrocompactConfig` cho `writer` (`KeepRecent: 2`) và `architect` (`KeepRecentReads: 1`, bao gồm cả `tavily_search` & `tavily_crawl`) trong `internal/agents/build.go`, `internal/agents/architect_context.go`, `internal/agents/context_manager.go`.
  - Cập nhật `kaggle/colab_llama_cpp.ipynb`: Nâng `--ctx-size` lên `65536` và sửa proxy streaming `/v1/chat/completions` trong `gateway.py` kiểm tra mã lỗi HTTP `>= 400` trước khi trả `StreamingResponse`.
- [x] **Web UI Dashboard Toàn Diện (`internal/entry/web`)**:
  - Máy chủ HTTP độc lập hỗ trợ RESTful API và Server-Sent Events (SSE).
  - Giao diện Web SPA phong cách Retro Terminal TUI, hỗ trợ theo dõi bản thảo trực tiếp, quản lý dự án (`/dash`), tạo mới (`/new`), đổi mode (`/mode`), duyệt tiếp (`/next`).
- [x] **Trình Xuất File Word (.docx) Native (`internal/host/exp/docx.go`)**:
  - Tạo tài liệu Microsoft Word chuẩn OpenXML (.docx) không phụ thuộc thư viện thứ 3.
  - Chuẩn định dạng sách xuất bản: Trang tiêu đề, ngắt trang giữa các chương, thụt đầu dòng 0.5 inch, Times New Roman 13pt.
  - Tích hợp cờ `--word` / format `word` cho CLI, TUI (`/export --word`) và Web UI.
- [x] **Khắc Phục Toàn Diện Lỗi Treo / Phản Hồi Khi Xuất Word Trên Web**:
  - Sửa lỗi crash Nil Pointer Panic trong `SwitchProject` khi chuyển sang dự án đã hoàn thành (`h == nil`).
  - Phản hồi xác nhận thành công rõ ràng: Web UI nhận JSON kết quả (`ok: true`, `filename`, `path`, `chapters`, `bytes`) và bật `alert(...)` thông báo chi tiết tức thì.
  - Tự động kích hoạt tải tệp trực tiếp qua iframe ẩn + link dự phòng, chống 100% việc trình duyệt Chrome/Edge âm thầm chặn download tự động sau tác vụ bất đồng bộ.
  - Loại bỏ hoàn toàn tình trạng tab trình duyệt bị xoay/load mãi không dừng và hiển thị trạng thái nút `⏳ Đang xuất Word...`.
  - Chuẩn hóa header `Content-Disposition` theo RFC 6266 / RFC 5987 hỗ trợ tiếng Việt có dấu.
  - Bổ sung fallback tự động tìm kịch bản chương trong `*-video/scripts/` và `drafts/`.
  - Tự động nạp snapshot dự phòng từ `meta/` khi host rỗng để dashboard luôn hiển thị đúng thông tin sách đã hoàn thành.
- [x] **GitHub Actions CI/CD Tự Động Build Bản Cài Đặt Windows**:
  - Workflow `.github/workflows/build-windows.yml` tự động build và upload bản thực thi `amd64` và `arm64`.
  - Cập nhật GitHub Release `v1.0.0`.
- [x] **Đồng Bộ Nhánh & Đặt `doodle-explainer` Làm Mặc Định**:
  - Đồng bộ commit giữa `doodle-explainer` và `main`.
- [x] **Tạo nhánh mới `history-vietnam`**:
  - Tách nhánh từ `tam-ly-hoc`, kế thừa toàn bộ tính năng `/dash`, `/new`, Tavily Graceful Fallback và hệ thống 4 modes.
- [x] **Đại tu toàn diện chế độ Tiểu thuyết Lịch sử Việt Nam (`vietnamese-history`)**:
  - **Tôn chỉ tối thượng**: *"Tiểu thuyết lịch sử hay phải đúng đủ để người đọc tin và sống đủ để người đọc quan tâm."*
  - **Cập nhật hệ thống Prompt cốt lõi**:
    - `assets/prompts/modes/vietnamese-history/architect-short.md`: Quy hoạch đoản thiên/trung thiên với Bảng bất biến, Bảng vùng tự do, Xung đột kép, 11 tiêu đề `# Series bible`.
    - `assets/prompts/modes/vietnamese-history/architect-long.md`: Quy hoạch trường thiên theo Bảng bất biến, Bảng vùng tự do, 2 dòng chảy song hành, 13 tiêu đề Premise.
    - `assets/prompts/modes/vietnamese-history/writer.md`: Cấu trúc 7 phần của một chương (mở đầu khoảnh khắc sống & chi tiết giác quan, hai dòng chảy, xung đột kép, cái giá cao trào, kết thúc biến chuyển số phận, mục `# Hậu ký và Ghi chú lịch sử của tác giả` ở chương cuối).
    - `assets/prompts/modes/vietnamese-history/editor.md`: Thẩm định 7 chiều tích hợp **Bộ kiểm tra nhanh (5 câu hỏi vàng)**, kiểm soát nghiêm ngặt ranh giới sử/hư cấu và chi tiết đúng thời.
    - `assets/voices/vietnamese-history.md`: Tiêu chuẩn giọng văn của nhà chép sử đầy tâm huyết kết hợp bút lực văn chương trầm hùng, đĩnh đạc; chi tiết đời sống chuẩn thời, xưng hô Đại Việt, cấm từ convert kiếm hiệp, 100% tiếng Việt có dấu, sạch chữ Hán.
    - `assets/styles/vietnamese-history.md`: Tôn chỉ sáng tác, bảng bất biến, bảng vùng tự do, xung đột kép, cái giá đánh đổi và hậu ký tác giả.
  - **Cập nhật đồng bộ hệ thống tài liệu Reference chuyên sâu**:
    - `chapter-guide.md`: Hướng dẫn toàn diện 7 bước triển khai chương văn xuôi lịch sử (2.000 - 4.000 từ).
    - `chapter-template.md`: Mẫu chương chuẩn mực *Khói Lò Rèn Bên Bến Vạn Kiếp* đầy đủ hành động, giác quan, xung đột kép, mất mát thực tế và mục Hậu ký tác giả mẫu.
    - `quality-checklist.md`: Bộ kiểm tra nhanh 5 câu hỏi vàng & Checklist thẩm định 7 tiêu chuẩn chi tiết.
    - `hook-techniques.md`: Cẩm nang 12 kỹ thuật hook chuyên biệt cho đề tài lịch sử, 7 lỗi cần tránh, chọn theo mục đích, chuyển sang video/doodle explainer và quy trình 4 bước luyện hook nhanh.
    - `character-building.md`: Xây dựng nhân vật đa chiều, cửa vào độc giả, tránh thần thánh hóa một chiều, nhân vật phụ soi chiếu góc khuất sử sách.
    - `fact-grounding.md`: Nền móng trước khi viết, lập Bảng bất biến & Bảng vùng tự do, phân cấp 4 tầng nguồn tư liệu.
    - `adversity-conflict.md`: Mô hình xung đột kép (đời tư va đập thời cuộc), hai dòng chảy song hành và khoảnh khắc lựa chọn/đánh đổi ở cao trào.
    - `dialogue-writing.md`: Nghệ thuật đối thoại và chuẩn mực xưng hô Đại Việt, danh mục cấm từ ngữ convert kiếm hiệp lai căng.
    - `plot-structures.md`: Cấu trúc 8 giai đoạn chuẩn mực tiểu thuyết lịch sử Việt Nam từ kinh nghiệm các tác phẩm lớn (*Hồ Quý Ly*, *Hội thề*, *Bão táp triều Trần*, *Sống mãi với thủ đô*).
    - `character-template.md` & `outline-template.md`: Cấu trúc dữ liệu JSON mẫu tích hợp trường Bảng bất biến, Vùng tự do và Xung đột kép.
- [x] **Cơ chế Bắt buộc Tra cứu Realtime Tavily Search (`tavily_search` & `tavily_crawl`)**:
  - Bổ sung `tavilyCrawl` vào `architectTools` trong `internal/agents/build.go`.
  - Thiết lập **BƯỚC 0 BẮT BUỘC** trong `architect-short.md` và `architect-long.md`: Ép LLM thực hiện 1-3 truy vấn `tavily_search` tra cứu thời gian thực về niên đại, địa danh cổ, nhân vật, diễn biến trận đánh, giai thoại dã sử TRƯỚC KHI gọi `novel_context`, `save_book` hay `save_foundation`.
  - Bổ sung chỉ dẫn và danh sách công cụ trong `writer.md` cho phép Writer tra cứu sử liệu chi tiết (vũ khí, trang phục, phong tục, địa hình) trước khi viết bản thảo.
  - Đồng bộ cơ chế cho cả mode `vietnamese-history` và `behavioral-psychology`.
- [x] **Kiểm thử & Biên dịch**:
  - 100% test suite `go test ./...` pass green.
  - `ainovel-cli.exe` biên dịch thành công.

---

## Previous Milestones on branch `tam-ly-hoc` (2026-10-08)
- [x] **TUI Dashboard (`/dash`)**: Quét outputs, gắn tag mode trực quan, hỗ trợ chuyển đổi trực tiếp.
- [x] **Khởi tạo phiên mới sạch (`/new`)**: Bắt đầu phiên mới nhanh chóng.
- [x] **Tavily Graceful Fallback**: Chống lỗi đứt gãy luồng ReAct khi thiếu key hoặc gặp lỗi HTTP từ Tavily.
- [x] **Behavioral Psychology 8-Step Blueprint**: Chuẩn hóa chế độ tâm lý học hành vi theo cấu trúc 8 bước khoa học & nghệ thuật kể chuyện.
- [x] **Hoàn thành trọn vẹn bộ kịch bản *Vùng Xám: Giải Mã Ranh Giới Thiện - Ác*** (3/3 tập).

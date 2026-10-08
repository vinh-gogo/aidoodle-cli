# Progress

_Last updated: 2026-10-08 (Branches `doodle-explainer` & `main`: Web Dashboard, Native Word Exporter .docx, Fix Export Hang, GitHub Actions Release v1.0.0)_

## What works (per upstream docs; verified locally)
- Deterministic Engine + Route table, Arbiter decisions, 3 Workers, Store with atomic IO, checkpoints, commit Saga.
- Rolling volume/arc planning, layered summaries, related-chapter recommendation, compression pipeline.
- TUI (+ co-create, `/config`, `/model`, `/review`, `/next`, `/diag`, `/simulate`, `/importsim`, `/sync`, `/import`, `/export`, `/start`, `/mode`, `/dash`, `/new`) and headless mode.
- **Web UI Dashboard (`ainovel-cli web`)**: Retro terminal layout, 3-panel UI, realtime SSE stream, command executor, chapter reader, visual project manager.
- **Multi-format Exporter**: TXT, EPUB, Video Scripts Package, and **Microsoft Word (.docx)** native OpenXML zero-dependency.
- Multi-mode support: 4 specialized AI modes (`novel-manga`, `doodle-explainer`, `behavioral-psychology`, `vietnamese-history`).
- Automated CI/CD: GitHub Actions building Windows AMD64 & ARM64 executables, Release v1.0.0.
- Full unit test suite and integration tests passing 100% across all packages.

---

## Recent Milestones on branches `doodle-explainer` & `main` (2026-10-08)
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

# Progress

_Last updated: 2026-10-08 (Branch tam-ly-hoc, /dash Dashboard, Clean /new, Tavily Graceful Fallback)_

## What works (per upstream docs; verified locally)
- Deterministic Engine + Route table, Arbiter decisions, 3 Workers, Store with atomic IO, checkpoints, commit Saga.
- Rolling volume/arc planning, layered summaries, related-chapter recommendation, compression pipeline.
- TUI (+ co-create, `/config`, `/model`, `/review`, `/next`, `/diag`, `/simulate`, `/importsim`, `/sync`, `/import`, `/export`, `/start`, `/mode`, `/dash`, `/new`) and headless mode.
- Multi-mode support: 4 specialized AI modes (`novel-manga`, `doodle-explainer`, `behavioral-psychology`, `vietnamese-history`).
- Full unit test suite and integration tests passing 100% across all 36 packages.

---

## Recent Milestones on branch `tam-ly-hoc` (2026-10-08)
- [x] **TUI Dashboard (`/dash`)**:
  - Triển khai lệnh `/dash` (aliases: `/dashboard`, `/projects`, `/history`) với `AutoExecute: true`.
  - Quét toàn bộ các outputs trong `output/`, trích xuất Title, Style, Progress, ModTime.
  - Tự động gắn tag mode trực quan: `[🧠 Tâm lý học hành vi]`, `[🦴 Video TikTok]`, `[⚔️ Lịch sử Việt Nam]`, `[📖 Tiểu thuyết / Manga]`.
  - Hiển thị tiến độ: `✓ Đã xong (3/3)`, `⏳ Đang viết`, `🌱 Khởi tạo`, và đánh dấu `⭐ [Đang mở]`.
  - Bố cục 2 dòng/dự án với hàm `dashModalSize` chống tràn chữ trên mọi kích thước terminal.
  - Điều hướng bằng phím mũi tên `↑`/`↓` hoặc `j`/`k`, nhấn `Enter` để lập tức chuyển sang dự án đã chọn, nạp lại đúng bundle assets và tiếp tục phiên làm việc.
  - Bổ sung unit test bao phủ toàn diện cho `/dash` (`command_dash_test.go`).
- [x] **Khởi tạo phiên mới sạch (`/new`)**:
  - Gõ `/new` -> Enter lập tức bắt đầu session mới với thư mục output timestamp riêng biệt `output/novel-YYYYMMDD-HHMM` và quay về màn hình chào mừng sạch mà không yêu cầu thêm tham số.
- [x] **Tavily Graceful Fallback & Dọn dẹp Key ảo**:
  - Gỡ bỏ fallback API key dev ảo hardcoded khỏi `internal/bootstrap/config.go` và `.env`.
  - Nâng cấp `TavilySearchTool` và `TavilyCrawlTool`: khi không có API key hoặc gặp lỗi HTTP 40x/50x/mạng, công cụ không báo lỗi cứng làm đứt luồng ReAct mà trả về dữ liệu fallback kèm lời khuyên để LLM tiếp tục sáng tác dựa trên tri thức học thuật sẵn có.
  - Triệt tiêu hoàn toàn sự cố kích hoạt bộ ngắt mạch bế tắc: *"Chỉ lệnh liên tiếp 5 lần không có tiến triển"*.
  - Bổ sung unit test kiểm thử các kịch bản fallback.
- [x] **Behavioral Psychology 8-Step Blueprint & Prompt Overhaul**:
  - Tinh chỉnh toàn bộ hệ thống prompt của mode `behavioral-psychology` theo chuẩn mực 8 phần:
    1. Mở bài từ trải nghiệm người đọc (3-5 câu).
    2. Nêu vấn đề và lời hứa ngắn gọn.
    3. Đặt tên hiện tượng và định nghĩa bằng lời thường (1 ví dụ).
    4. Cơ chế và bằng chứng (phân biệt tương quan vs nhân quả, thí nghiệm vs khảo sát, khủng hoảng tái lập).
    5. Giới hạn, ngoại lệ, phản biện (mẫu WEIRD).
    6. Ứng dụng: 1-3 việc nhỏ làm thử trong vài ngày.
    7. Kết bài: nhìn lại tình huống ban đầu bằng góc nhìn mới.
    8. Nguồn và lưu ý: không chẩn đoán, không phán xét, một bài một ý lớn.
- [x] **Hoàn thành trọn vẹn bộ kịch bản *Vùng Xám: Giải Mã Ranh Giới Thiện - Ác***:
  - Cứu vãn và hoàn thành Tập 3 (`output/novel-20261008-1006`), bản thảo 1.901 từ nhịp 1:1 Thoại - Hình.
  - Checkpoint, summaries và chuyển trạng thái sang `phase: complete` (3/3 tập).

---

## Previous Milestones on branch `route` & `doodle-explainer` (2026-10-07)
- [x] **Branch `route`: 4 AI Working Modes (`/mode`)**:
  - Lệnh `/mode` (alias: `/topic`, `/topics`) mở modal overlay trực quan chọn giữa 4 chế độ.
  - Phân tách và nạp tài nguyên độc lập 100% cho cả 4 chế độ trong `assets/load.go`.
  - Gated validation: `lintScript` chỉ áp dụng cho 2 mode video (`doodle-explainer`, `behavioral-psychology`), giải phóng các mode tiểu thuyết (`novel-manga`, `vietnamese-history`).
- [x] **Tích hợp Tavily Search & Crawl cho cơ sở khoa học & kiểm chứng**:
  - Gói `internal/tavily/` và 2 công cụ LLM `tavily_search`, `tavily_crawl`.
  - Tự động nạp `source_pack` vào ngữ cảnh `novel_context`.
- [x] **1:1 Alternating Voice-Visual Beats**:
  - Bắt buộc phân tách cảnh thành các cặp `LỜI:` - `HÌNH:` xen kẽ liên tục (mỗi câu thoại 10-20 từ / 3-6s đi liền ngay một thẻ HÌNH).
  - Loại bỏ hoàn toàn thẻ `CHỮ:`.
- [x] **Series 3 tập chuyên sâu mặc định**:
  - Mặc định lập kế hoạch 3 tập chuyên sâu cho một chủ đề đơn lẻ.
- [x] **5-Minute+ Doodle Explainer Standard**:
  - Chuẩn hóa thời lượng video 5+ phút (300–600s, 700–1500 từ LỜI), 5 giai đoạn có Tái Hook.
- [x] **Thư mục output phân lập theo timestamp (`output/novel-YYYYMMDD-HHMM`)**.
- [x] **Humor Specialist & Archetypes** (`assets/references/humor-relatability.md`).
- [x] **Sửa lỗi `save_foundation` InputValidationError**.

---

## Verified on this machine
- [x] `go build -o ainovel-cli.exe ./cmd/ainovel-cli` thành công.
- [x] `go test ./...` 36 packages pass 100%.
- [x] Đã push toàn bộ commit lên nhánh `tam-ly-hoc` trên GitHub.

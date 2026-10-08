# Active Context

_Last updated: 2026-10-08 (Branches: `doodle-explainer` & `main` - Đồng bộ Web Dashboard, Native Word Exporter .docx, Fix Lỗi Treo Export & GitHub Actions CI/CD Windows)_

## Current Status & Direction
Repo đang hoạt động đồng bộ trên hai nhánh chính: **`doodle-explainer`** (mặc định) và **`main`**:
- **Phiên bản phát hành chính thức v1.0.0**: Đã thiết lập GitHub Actions tự động build và xuất bản các bản thực thi Windows Portable (AMD64 & ARM64).
- **Hệ thống 4 Chế độ Sáng tác Chuyên biệt (`/mode`)**:
  1. `[1] Tiểu thuyết / Manga` (`novel-manga`): Tiểu thuyết văn xuôi chương hồi (2.000 – 4.000 từ/chương).
  2. `[2] Doodle Explainer` (`doodle-explainer`): Video người que giải thích kiến thức TikTok / YouTube Shorts (5+ phút, nhịp 1:1 Thoại - Hình).
  3. `[3] Tâm lý học hành vi` (`behavioral-psychology`): Video giải mã bẫy nhận thức, cơ chế não bộ, chuẩn mực 8 phần khoa học kết hợp nghệ thuật kể chuyện cuốn hút.
  4. `[4] Tiểu thuyết Lịch sử Việt Nam` (`vietnamese-history`): Tiểu thuyết văn xuôi chuẩn mực "Đúng đủ để người đọc tin, sống đủ để người đọc quan tâm" (Bảng bất biến, Vùng tự do, Xung đột kép, 12 kỹ thuật hook, 4 tầng sử liệu, Tavily Search realtime).

---

## Recent Major Features & Enhancements

### 1. Web UI Dashboard Phong cách Retro Terminal (`ainovel-cli web`)
- **Kiến trúc Server (`internal/entry/web`)**:
  - Giao diện Web SPA phong cách Retro Terminal TUI tông màu amber/neon sắc sảo, chia bố cục 3 panel:
    - **Panel 1 (Bên trái)**: Thông tin tác phẩm, thanh tiến độ, chi phí API thực tế và bảng danh sách chương có thể click đọc tức thì.
    - **Panel 2 (Ở giữa)**: Khung hiển thị trực tiếp (Live Stream Markdown) bản thảo AI đang viết theo thời gian thực hoặc xem lại các chương đã chọn.
    - **Panel 3 (Bên phải)**: Terminal console tương tác với thanh gợi ý lệnh nhanh (`/mode`, `/dash`, `/next`, `/new`, `/export`, `/pause`, `/resume`) và bảng Event Logs theo dõi chi tiết hoạt động của các Agents.
  - Tích hợp SSE (`/api/stream`) truyền phát real-time stream token, delta và system lifecycle events.
  - Tích hợp Modal chọn chế độ sáng tác trực quan (`/mode`) và Thư viện quản lý outputs (`/dash`).

### 2. Bộ Xuất File Word (.docx) Native Không Phụ Thuộc Thư Viện Ngoài
- **Engine Word OpenXML (`internal/host/exp/docx.go`)**:
  - Tự tay xây dựng cấu trúc file Microsoft Word chuẩn `.docx` (Office OpenXML standard) bằng Go thuần với `archive/zip` (zero external dependencies).
  - Định dạng chuẩn xuất bản: Trang bìa tiêu đề sách, ngắt trang trang trọng (`w:br w:type="page"`), Heading 1 phân tách chương, căn lề văn bản thụt đầu dòng (indent 0.5 inch / 720 dxa), phông chữ Times New Roman 13pt, khoảng cách dòng thoáng (1.25x line spacing).
  - Tích hợp cờ `--word` / format `word` trong CLI, lệnh TUI (`/export --word`) và các nút xuất nhanh trên Web UI (Action Toolbar, Panel Header, Celebration Banner).

### 3. Sửa Triệt Để Lỗi Treo / Phản Hồi Khi Xuất File Word & Chuyển Dự Án
- **Khắc phục Crash Nil Pointer Panic**:
  - Trong `server.SwitchProject`: Khắc phục lỗi `panic: nil pointer dereference` khi gọi `h.ReplayQueue()` trên dự án đã hoàn thành (không có active LLM host). Bọc điều kiện an toàn `h != nil`.
- **Nạp Snapshot Dự Phòng từ Store**:
  - Khi chuyển sang một dự án đã viết xong trong quá khứ (`s.host == nil`), `/api/status` tự động đọc dữ liệu từ thư mục `meta/` (tên sách, số chương hoàn thành, trạng thái `COMPLETE`) để giao diện web lập tức hiển thị đầy đủ thông tin sách và nút xuất file.
- **Phản Hồi Xác Nhận Thành Công Rõ Ràng & Tự Động Tải Tệp Trực Tiếp**:
  - Web UI gọi endpoint `/api/export` nhận JSON kết quả (`ok: true`, `filename`, `path`, `chapters`, `bytes`, `download_url`).
  - Hiển thị hộp thoại `alert(...)` xác nhận thành công với đầy đủ chi tiết: tên file, số chương, dung lượng và đường dẫn lưu trên máy tính.
  - Tự động kích hoạt tải tệp trực tiếp về thư mục Downloads qua iframe ẩn kết hợp link dự phòng (tránh 100% việc trình duyệt Chrome/Edge âm thầm chặn popup download tự động sau tác vụ bất đồng bộ).
  - Ghi log nổi bật màu xanh vào bảng Event Log của giao diện Retro Terminal.
- **Chuẩn Hóa Mã Hóa Header Tên File Tiếng Việt (RFC 6266 / RFC 5987)**:
  - Header `Content-Disposition` sử dụng cú pháp `filename*=UTF-8''...` kết hợp fallback ASCII an toàn, tải đúng tên tệp tiếng Việt có dấu trên mọi trình duyệt.
- **Bổ Sung Fallback Đọc Nội Dung Chương Trong Exporter**:
  - Bổ sung cơ chế fallback tự động tìm kịch bản chương trong `*-video/scripts/%02d-*.md` và `drafts/` nếu file `chapters/%02d.md` không nằm ở vị trí mặc định.

### 4. GitHub Actions CI/CD Tự Động Build Bản Cài Đặt Windows
- File cấu hình `.github/workflows/build-windows.yml` tự động build và phát hành các bản:
  - `ainovel-cli-windows-amd64.exe` (Intel / AMD 64-bit)
  - `ainovel-cli-windows-arm64.exe` (Windows on ARM)
- Đính kèm tự động vào GitHub Release `v1.0.0`.

---

## Active Workspace State
- **Branches**: Đồng bộ giữa `doodle-explainer` (mặc định) và `main`.
- **Tests**: 100% test suite `go test ./...` pass green.
- **Binary**: `ainovel-cli.exe` mới nhất biên dịch thành công.

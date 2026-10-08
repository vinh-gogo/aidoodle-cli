# Active Context

_Last updated: 2026-10-08 (Dashboard /dash, Clean /new, Tavily Graceful Fallback & Behavioral Psychology 8-Step Blueprint)_

## Current Status & Direction
Repo đang hoạt động trên branch **`tam-ly-hoc`** (forked từ `route` / `aidoodle-cli`):
- **4 Chế độ làm việc chuyên biệt (`/mode`)**:
  1. `[1] Tiểu thuyết / Manga` (`novel-manga`): Tiểu thuyết văn xuôi chương hồi (2.000 – 4.000 từ/chương).
  2. `[2] Doodle Explainer` (`doodle-explainer`): Video người que giải thích kiến thức TikTok / YouTube Shorts (5+ phút, nhịp 1:1 Thoại - Hình).
  3. `[3] Tâm lý học hành vi` (`behavioral-psychology`): Video giải mã bẫy nhận thức, cơ chế não bộ, Cú hích hành vi (Nudge), chuẩn mực 8 phần khoa học kết hợp nghệ thuật kể chuyện cuốn hút và tự trắc ẩn vỗ về đứa trẻ bên trong.
  4. `[4] Tiểu thuyết Lịch sử Việt Nam` (`vietnamese-history`): Tiểu thuyết văn xuôi hào sảng tái hiện danh nhân lịch sử Đại Việt (2.000 – 4.000 từ/chương), tra cứu Tavily Search.
- **Dashboard Quản lý Outputs (`/dash`)**:
  - Gõ `/dash` -> Enter (hoặc `/dashboard`, `/projects`, `/history`) mở modal trực quan liệt kê tất cả các dự án trong `output/` (mới nhất lên đầu).
  - Tự động gắn **Tag Mode** trực quan: `[🧠 Tâm lý học hành vi]`, `[🦴 Video TikTok]`, `[⚔️ Lịch sử Việt Nam]`, `[📖 Tiểu thuyết / Manga]`.
  - Hiển thị tiến độ: `✓ Đã xong (3/3)`, `⏳ Đang viết (1/3)`, `🌱 Khởi tạo`, đường dẫn `output\novel-...`, timestamp và đánh dấu `⭐ [Đang mở]`.
  - Hỗ trợ di chuyển `↑`/`↓`/`j`/`k`, nhấn `Enter` để lập tức mở lại và chuyển đổi dự án (tự động nạp đúng mode và khôi phục tiến trình).
- **Khởi tạo phiên mới sạch (`/new`)**:
  - Gõ `/new` -> Enter lập tức bắt đầu session mới với thư mục output timestamp riêng biệt `output/novel-YYYYMMDD-HHMM` và quay về màn hình chào mừng sạch mà không yêu cầu thêm tham số.
- **Tavily Graceful Fallback (Suy giảm chức năng an toàn)**:
  - Loại bỏ hoàn toàn API key dev ảo hardcoded khỏi hệ thống.
  - Công cụ `tavily_search` và `tavily_crawl` tự động chuyển đổi sang phản hồi mềm khi chưa cấu hình key hoặc máy chủ Tavily trả về lỗi HTTP 40x/50x/mạng, hướng dẫn LLM vận dụng tri thức học thuật sẵn có, triệt tiêu hoàn toàn sự cố đứt gãy luồng ReAct và bế tắc ngắt mạch.
- **Tác phẩm mẫu hoàn thành**:
  - Bộ *Vùng Xám: Giải Mã Ranh Giới Thiện - Ác* (`output/novel-20261008-1006`) đã hoàn thành trọn vẹn cả 3 tập (phase: complete), bản thảo 1.901 từ nhịp 1:1 Thoại - Hình.

---

## Recent Milestones & Commits (2026-10-08)

1. **`34a3302` — TUI Dashboard (`/dash`) & Tavily Graceful Fallback**:
   - **Lệnh `/dash` (Dashboard)**:
     - Tạo `internal/entry/tui/command_dash.go` và `command_dash_test.go`.
     - Quét động toàn bộ thư mục `output/`, trích xuất Title, Style, Progress, ModTime.
     - Hàm `detectStyleBadge`: gán tag mode trực quan cho từng thể loại.
     - Hàm `dashModalSize`: căn chỉnh kích thước modal tối ưu, bố cục 2 dòng/dự án chống tràn chữ trên mọi kích thước terminal.
     - Hỗ trợ di chuyển phím `↑`/`↓`/`j`/`k`, phím `Enter` kích hoạt `m.switchOutputDir` và gửi `tea.Quit` để vòng lặp `Run` trong `app.go` nạp lại đúng bundle và mở dự án.
     - Đăng ký lệnh `/dash` (AutoExecute: true, aliases: `dashboard`, `projects`, `history`).
   - **Tavily Graceful Fallback**:
     - Gỡ bỏ key dev ảo khỏi `internal/bootstrap/config.go` và `.env`.
     - Cập nhật `TavilySearchTool.Execute` và `TavilyCrawlTool.Execute`: khi API key rỗng hoặc gọi Tavily gặp lỗi HTTP 401/400/429/timeout, trả về JSON fallback hướng dẫn LLM sử dụng tri thức khoa học sẵn có để tiếp tục sáng tác mà không ngắt luồng.
     - Bổ sung unit test bao phủ toàn diện cho fallback.

2. **`91344d5` — Behavioral Psychology 8-Step Blueprint & Prompt Overhaul**:
   - Tinh chỉnh toàn bộ hệ thống prompt của mode `behavioral-psychology` theo hướng dẫn cấu trúc 8 phần chuẩn mực:
     1. Mở bài: Trải nghiệm quen thuộc / nghịch lý (3-5 câu).
     2. Nêu vấn đề và lời hứa ngắn gọn.
     3. Đặt tên hiện tượng & định nghĩa bằng lời thường (1 ví dụ thực tế).
     4. Cơ chế & bằng chứng khoa học (phân biệt tương quan vs nhân quả, thí nghiệm vs khảo sát, lưu ý khủng hoảng tái lập).
     5. Giới hạn, ngoại lệ, phản biện (mẫu WEIRD, hiệu ứng lớn/nhỏ).
     6. Ứng dụng: 1-3 việc nhỏ làm thử trong vài ngày.
     7. Kết bài: nhìn lại tình huống ban đầu bằng góc nhìn mới, chốt thông điệp 1 câu.
     8. Nguồn và lưu ý: tuyệt đối không chẩn đoán người đọc, một bài một ý lớn.
   - Cập nhật đồng bộ các file prompt Architect, Writer, Editor và Reference guides trong `assets/`.

3. **Cứu vãn & Hoàn thành Tập 3 bộ *Vùng Xám: Giải Mã Ranh Giới Thiện - Ác***:
   - Sửa lỗi LLM lặp lặp `draft_chapter` do bế tắc hoặc token context.
   - Hoàn thiện bản thảo Tập 3 (1.901 từ, nhịp Thoại - Hình 1:1, thẻ NGUỒN và CẦN KIỂM CHỨNG đầy đủ).
   - Lưu trữ checkpoint, summaries, chuyển trạng thái tiến độ sang `phase: complete` (3/3 tập hoàn tất).

4. **Triển khai `/new` mở phiên làm việc mới tức thì**:
   - Cho phép người dùng gõ `/new` -> Enter để bắt đầu ngay một session mới trong thư mục output timestamp riêng biệt `output/novel-YYYYMMDD-HHMM` mà không cần nhập thêm bất kỳ tham số nào.

---

## Active Architecture & Key Patterns

- **TUI Dashboard Pattern (`/dash`)**:
  - Struct `dashState`, `dashProjectItem`.
  - Quét thư mục `output/`, tự động nhận diện style và tiến độ qua `meta/run.json`, `meta/book.json`, `meta/progress.json`, `chapters/`.
  - Kênh chuyển đổi dự án `m.switchOutputDir` điều phối Bubbletea app loop khởi động lại mượt mà.
- **Graceful Tool Degradation Pattern**:
  - Các công cụ tra cứu bên ngoài (như Tavily Search/Crawl) áp dụng cơ chế fallback mềm: trả về dữ liệu hướng dẫn thay vì trả về lỗi để tránh kích hoạt vòng lặp bế tắc trong ReAct agent.
- **Multi-Mode AI Architecture (4 Modes)**:
  - Phân tách độc lập kho prompt, voice, style, reference trong `assets/`.
  - Bộ kiểm tra kịch bản `lintScript` chỉ áp dụng cho 2 mode video.

---

## Current Workspace State

- **Branch**: `tam-ly-hoc` trên repo `https://github.com/vinh-gogo/aidoodle-cli.git`.
- **Remote**: Đã push và đồng bộ hoàn toàn với `origin/tam-ly-hoc`.
- **Binary**: `ainovel-cli.exe` đã được biên dịch thành công từ commit mới nhất.
- **Tests**: 100% unit tests và full test suite toàn bộ repo pass green.

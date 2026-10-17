# Active Context

_Last updated: 2026-10-17 (Production Readiness Upgrade v1.0.1: M1 Security Hardening, M2 Context & Mechanical Validators, M3 W3C EPUB 3.3 Exporter, M4 7-Mode Evals, CI Release Hygiene & Documentation Alignment)_

## Current Status & Direction
Repo đang hoạt động đồng bộ trên hai nhánh chính: **`doodle-explainer`** (mặc định) và **`main`**:
- **Phiên bản phát hành chính thức v1.0.1**: Đã thiết lập GitHub Actions tự động build và xuất bản các bản thực thi Windows Portable (AMD64 & ARM64), Linux và macOS kèm mã băm SHA-256 `checksums.txt`.
- **Hệ thống 7 Chế độ Sáng tác Chuyên biệt (`/mode` trong TUI & Web UI)**:
  1. `[1] Tiểu thuyết / Manga Đa năng` (`novel-manga`): Tiểu thuyết văn xuôi chương hồi, Light Novel, Manga/Webtoon đa thể loại (2.000 – 4.000 từ/chương).
  2. `[2] Tiểu thuyết Trinh thám & Kinh dị` (`suspense`): Trinh thám suy luận công bằng (Fair-Play Whodunit), Tâm lý tội phạm đen tối (Psychological Thriller) & Kinh dị tâm linh dân gian (Vietnamese Folk Horror) với **Ma trận Chân tướng Kép**, **Sổ cái Manh mối 3:2** và **Đồng hồ Đếm ngược Sinh tử**.
  3. `[3] Tiểu thuyết Ngôn tình & Chữa lành` (`romance`): Ngôn tình hiện đại, Tâm lý tình cảm sâu sắc (Slow-burn) & Chữa lành (Healing Fiction) với **Hồ sơ Vết thương Tâm lý (Lie/Ghost)**, **Thang đo Cảm xúc 6 Cấp độ** và đối thoại giàu ẩn ý (Subtext).
  4. `[4] Tiểu thuyết Kỳ ảo & Tiên hiệp` (`fantasy`): Kỳ ảo sử thi (Epic High Fantasy), Tiên hiệp tu chân & Thăng cấp sức mạnh (Progression Fantasy) với **Hệ thống Quy tắc Cứng (Định luật Sanderson)**, **Bản đồ Thế giới Đa tầng**, **Chiến đấu Chiến thuật 3 Pha** và bộ lọc chống văn phong dịch máy (Anti-Convert).
  5. `[5] Tiểu thuyết Lịch sử Việt Nam` (`vietnamese-history`): Tiểu thuyết văn xuôi chuẩn mực "Đúng đủ để người đọc tin, sống đủ để người đọc quan tâm" (Bảng bất biến, Vùng tự do, Xung đột kép, 12 kỹ thuật hook, 4 tầng sử liệu, Tavily Search realtime).
  6. `[6] Doodle Explainer` (`doodle-explainer`): Video người que giải thích kiến thức TikTok / YouTube Shorts (5+ phút, nhịp 1:1 Thoại - Hình).
  7. `[7] Tâm lý học hành vi` (`behavioral-psychology`): Video giải mã bẫy nhận thức, cơ chế não bộ, chuẩn mực 8 phần khoa học kết hợp nghệ thuật kể chuyện cuốn hút.

---

## Recent Major Features & Enhancements

### 0. Mở Rộng 3 Thể Loại Tiểu Thuyết Nổi Bật Nhất (`suspense`, `romance`, `fantasy`) & Báo Cáo Thị Trường (`docs/novel-genres-report.md`)
- **Báo cáo nghiên cứu thị trường (`docs/novel-genres-report.md`)**:
  - Phân tích toàn cảnh các thể loại tiểu thuyết được yêu thích nhất tại Việt Nam và quốc tế (2024–2026), đối chiếu 8 nhóm lớn và chọn ra 3 thể loại nổi bật nhất bổ sung vào hệ sinh thái `ainovel-cli`.
- **Kiến trúc Multi-Agent Chuyên biệt cho 3 Thể loại Mới (`assets/prompts/modes/{suspense,romance,fantasy}/`, `assets/voices/`, `assets/styles/`, `assets/references/genres/`)**:
  - Mỗi thể loại có bộ 4 Prompts riêng (`architect-short.md`, `architect-long.md`, `writer.md` kèm `{{VOICE}}`, `editor.md` kèm 5 Câu hỏi Vàng & khế ước phạm vi can thiệp), Voice riêng, Style riêng, cùng `style-references.md` (Bản mẫu Vàng + Bảng đối chiếu Văn mẫu Rác vs Văn học Chuẩn mực) và `arc-templates.md`.
- **Định tuyến Tài nguyên & Lọc Ngữ cảnh (`assets/load.go`, `internal/tools/novel_context.go`)**:
  - Nạp đúng bộ Prompts, Voice, Base References (`references/modes/novel-manga/`) và Genre References (`references/genres/<canonical>/`) cho `suspense`, `romance`, `fantasy` cùng các alias (`trinh-tham`, `ngon-tinh`, `ky-ao`, `tien-hiep`, `xianxia`, ...).
  - Loại bỏ triệt để việc rò rỉ tài liệu video doodle vào các mode tiểu thuyết mới và duy trì `style_reference` xuyên suốt các chương.
- **Đồng bộ TUI & Web UI (`/mode` 1–7 & `/dash` Badges)**:
  - Cập nhật `internal/entry/tui/command_mode.go`, `internal/entry/tui/command_dash.go`, `internal/entry/web/project.go` và `internal/entry/web/ui/index.html` hỗ trợ đầy đủ 7 chế độ sáng tác và phím tắt nhanh `1`–`7`.

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

### 5. Chuyển Đổi Hoàn Toàn Từ Ngrok Sang Cloudflare Tunnel Cố Định (`https://api.gloryvocab.com/v1`)
- **Colab Unified AI Gateway (`kaggle/colab_llama_cpp.ipynb`)**:
  - Thay thế hoàn toàn Ngrok tạm thời (`*.ngrok-free.dev`) bằng tên miền Cloudflare Tunnel cố định: **`https://api.gloryvocab.com/v1`**.
  - Hợp nhất LLM 26B (`main-llm` / `gemma-4-26B-A4B-it-Q4_K_M.gguf` / `gemma-4-26b` trên Port 8001) và STT (`gemma-4-stt` / `google/gemma-4-E2B-it` trên Port 8003) qua FastAPI Gateway (Port 8000).
  - Xóa các notebook và script Ngrok cũ (`kaggle/colab-llama-cpp-ngrok.ipynb`, `kaggle/llama-cpp-ngrok-t3.ipynb`, `kaggle/create_colab.py`, `kaggle/patch.py`).
- **Đồng Bộ Cấu Hình (`config/config.json`, `.ainovel/config.json`, `~/.ainovel/config.json`, `config.example.jsonc`)**:
  - Cập nhật `base_url: "https://api.gloryvocab.com/v1"`, loại bỏ header `ngrok-skip-browser-warning`.
  - Đặt model mặc định `main-llm` (kèm `gemma-4-26B-A4B-it-Q4_K_M.gguf` và `gemma-4-26b` trong danh sách chuyển đổi `/model`).
  - Đồng bộ `context_window: 32768` trong cấu hình client và nâng `--ctx-size` của `llama-server` trong `kaggle/colab_llama_cpp.ipynb` lên `65536`.

### 6. Sửa Triệt Để Lỗi `openai: stream ended before [DONE]` (Tràn Ngữ Cảnh 32K & Proxy Stream Error)
- **Nguyên nhân gốc rễ**:
  1. File cấu hình toàn cục `~/.ainovel/config.json` vẫn trỏ về URL Ngrok cũ (`snowcapped-denice-unprudent.ngrok-free.dev`) và model `qwen-27b`.
  2. Công cụ `novel_context` bơm 58 KB – 83 KB tài liệu tham khảo (`reference_pack.references`, bao gồm cả các tài liệu doodle stickman không thuộc mode `novel-manga` / `vietnamese-history` và các khối kịch bản mẫu dài trùng lặp với `writer.md`) + `source_pack` thô không giới hạn dung lượng, khiến tổng token ngay lượt 2 (`SystemPrompt 15.387 token + novel_context 17.534–30.000 token = 32.921–45.000 token`) vượt quá `--ctx-size 32768` của `llama-server`.
  3. Trong `kaggle/colab_llama_cpp.ipynb` (Cell 3 `gateway.py`), endpoint `/v1/chat/completions` trả về `StreamingResponse` (HTTP `200 OK`, `text/event-stream`) trước khi mở kết nối tới `llama-server`. Khi `llama-server` trả lỗi HTTP `400` (`exceed_context_size_error`), nội dung JSON lỗi bị đẩy vào luồng SSE `200 OK` rồi đóng kết nối mà không có `data: [DONE]`, khiến client báo `openai: stream ended before [DONE]` và retry 7 lần vô ích.
- **Giải pháp đã triển khai**:
  - **Tối ưu `novel_context` (`internal/tools/novel_context.go`, `internal/tools/novel_context_builders.go`)**:
    - Lọc bỏ các tài liệu đặc thù video doodle (`doodle_visual_language`, `tiktok_content_safety`, `humor_relatability`, `fact_grounding`) khi ở chế độ `novel-manga` hoặc `vietnamese-history`.
    - Bổ sung `compactReference` tự động cắt bỏ các phần kịch bản mẫu dài trùng lặp (`## Ví dụ hoàn chỉnh`, `## Ví dụ kịch bản`, ...) và giới hạn ngân sách ký tự mỗi tài liệu tham khảo (900 – 1.200 runes), giảm dung lượng `reference_pack.references` từ 58–83 KB xuống ~12–16 KB (~3.500 token).
    - Bổ sung `compactSourcePack` giới hạn tối đa 4 bài nguồn và 1.200 runes/bài khi nạp `source_pack` vào `working_memory` và `planning_memory`.
  - **Tối ưu `ContextManager` (`internal/agents/build.go`, `internal/agents/architect_context.go`, `internal/agents/context_manager.go`)**:
    - Đặt `KeepRecent: 2` cho `ToolResultMicrocompactConfig` của `writer` để kết quả `novel_context` cũ được thu gọn tự động trước khi `writer` bước vào giai đoạn `read_chapter` -> `check_consistency` -> `commit_chapter`.
    - Đặt `KeepRecentReads: 1` cho `architect` và bổ sung `tavily_search` / `tavily_crawl` vào `Classifier` để tự động thu gọn các kết quả tra cứu cũ.
  - **Sửa FastAPI Gateway, Gia Cố Bảo Mật & Nâng `--ctx-size` (`kaggle/colab_llama_cpp.ipynb`)**:
    - Nâng `--ctx-size` của `llama-server` từ `32768` lên `65536` (tận dụng SWA 5:1 + KV cache `q8_0` của Gemma 4 26B trên GPU L4).
    - Sửa `/v1/chat/completions` trong `gateway.py`: Kiểm tra `resp.status_code >= 400` trước khi khởi tạo `StreamingResponse`, trả nguyên mã lỗi HTTP & JSON từ `llama-server` về client và nâng timeout stream lên 900s.
    - Gia cố bảo mật `gateway.py`: Ẩn thông tin `llm_backend` / `stt_backend` trên `/` và `/health` khi gọi qua Cloudflare (`cf-ray`), tắt `/docs` / `/redoc` / `/openapi.json`, xóa `catch_all_proxy`, khóa CORS chỉ cho phép `gloryvocab.com` & `localhost` (bỏ `"*"`), giới hạn upload âm thanh tối đa 25MB và chat body 4MB, bind `uvicorn` vào `127.0.0.1`.

---

## Active Workspace State
- **Branches**: Đồng bộ giữa `doodle-explainer` (mặc định) và `main`.
- **Tests**: 100% test suite `go test ./...` pass green.
- **Binary**: `ainovel-cli.exe` mới nhất biên dịch thành công.


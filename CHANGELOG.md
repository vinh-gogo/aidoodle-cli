# Nhật Ký Thay Đổi (Changelog)

Toàn bộ các thay đổi đáng chú ý của dự án `ainovel-cli` (`aidoodle-cli`) sẽ được ghi nhận chi tiết tại tệp này.

Định dạng tài liệu tuân thủ theo chuẩn [Keep a Changelog](https://keepachangelog.com/vi/1.0.0/), và phiên bản tuân thủ theo [Semantic Versioning](https://semver.org/lang/vi/).

---

## [1.0.1] - 2026-10-17

Nâng cấp toàn diện mức độ sẵn sàng Production (Production Readiness Upgrade) theo lộ trình P0 + P1 + P2, bao gồm gia cố bảo mật (R1), kiểm soát ngân sách ngữ cảnh & bộ kiểm tra cơ học đầu ra (R2), trình xuất bản W3C EPUB 3.3 thuần Go (R3), bộ đánh giá 7 chế độ chuyên sâu, tự động hóa CI/CD và tài liệu vận hành (R4).

### Đã Thêm (Added)
- **Bảo mật & Web UI (R1 - SEC-01 đến SEC-07)**:
  - Bind mặc định máy chủ Web UI vào địa chỉ an toàn `127.0.0.1`, bổ sung cờ `--host`, `--port`, `--token`, `--external` và cơ chế sinh Bearer Token ngẫu nhiên khi mở giao diện ra mạng ngoài.
  - Bộ kiểm tra Origin và CSRF (`Sec-Fetch-Site`, `Sec-Fetch-Mode`) trên toàn bộ các endpoint thay đổi trạng thái (`/api/export`, `/api/action`, `/new`, `/next`, chuyển dự án).
  - Phòng vệ tấn công Path Traversal xuyên suốt các endpoint đọc file, tải tài liệu và chuyển dự án thông qua hàm kiểm tra nghiêm ngặt `SafeProjectDir` (`filepath.Clean`, `filepath.Rel`).
  - Lọc và khử độc XSS (Sanitize) cho các nội dung Markdown/HTML do LLM sinh ra trong giao diện Web SPA (`index.html`); thay thế toàn bộ hộp thoại `alert(...)` chặn luồng bằng thông báo Toast không chặn, kèm chỉ báo trạng thái mất/nối lại kết nối SSE.
  - Gia cố Colab FastAPI Gateway (`kaggle/colab_llama_cpp.ipynb`): áp dụng xác thực Bearer token `GATEWAY_API_KEY`, giới hạn tốc độ gọi (60 yêu cầu/phút), ghi log truy cập có cấu trúc JSON và xóa bỏ backdoor key.
  - Tự động khử nhạy cảm (Redact) các khóa API và Bearer token trong nhật ký chẩn đoán và `last-error.log`.
  - Cơ chế đóng gói và cách ly dữ liệu tra cứu từ Tavily trong thẻ `<untrusted_web_source>` với rào chắn chống Prompt Injection.
- **Kiểm soát Ngữ cảnh & Bộ Kiểm tra Cơ học (R2 - CTX-01 đến VAL-06)**:
  - Đồng bộ chuẩn hóa cấu hình `context_window: 32768` (client) và `--ctx-size 65536` (server KV cache headroom) xuyên suốt mã nguồn và tài liệu.
  - Rào chắn ngân sách tự động kiểm tra `SystemPrompt + novel_context` luôn $\le$ 60% cửa sổ ngữ cảnh 32K (19.660 tokens) cho toàn bộ 7 chế độ sáng tác.
  - Xử lý dứt khoát lỗi HTTP 4xx không thể thử lại (như `exceed_context_size_error`), chấm dứt vòng lặp retry vô ích.
  - Kiểm tra sức khỏe gateway chủ động (Pre-flight health check `/models`) trước khi bắt đầu quy trình sáng tác dài hạn.
  - Cầu chì ngân sách thay thế (MaxCalls, MaxTokens, MaxDuration) dành cho các model tự host có giá $0 trong model registry.
  - 6 bộ kiểm tra cơ học xác định trên bản thảo chương: loại bỏ ký tự CJK (chữ Hán, Hiragana, Katakana, Hangul) trong tiếng Việt; kiểm soát dung lượng từ theo thể loại (2.000–4.500 từ cho novel, 700–1.500 từ và 300–600s cho video); phát hiện vòng lặp lặp câu/lặp đoạn; phát hiện chương cụt kết thúc giữa câu; kiểm tra từ cấm convert & sáo ngữ; ngăn chặn rò rỉ thẻ kịch bản video (`HOOK`, `LỜI:`, `HÌNH:`) vào thể loại tiểu thuyết.
- **Trình Xuất Bản W3C EPUB 3.3 Thuần Go (R3 - EPUB-01 đến EPUB-07)**:
  - Đóng gói OCF Container chuẩn W3C: Entry 0 là `mimetype` lưu trữ không nén (`zip.Store`), nội dung chính xác `application/epub+zip`, `len(Extra) == 0`; `META-INF/container.xml` trỏ chính xác tới tệp gói OPF.
  - Tài liệu gói OPF 3.0: Khai báo `<package version="3.0" xml:lang="vi">`, đầy đủ Dublin Core metadata (`dc:identifier` UUID/URN, `dc:title`, `dc:language` = `vi`, `dcterms:modified` UTC ISO-8601), metadata trợ năng hiện đại EPUB Accessibility 1.1 (`schema:accessMode`, `schema:accessibilityFeature`, `schema:accessibilityHazard`, `schema:accessibilitySummary`), manifest đầy đủ tài nguyên và spine logic.
  - Tài liệu điều hướng `nav.xhtml` chuẩn XHTML5: hỗ trợ mục lục lồng phân cấp Quyển (Volume) $\rightarrow$ Hồi (Arc) $\rightarrow$ Chương (Chapter) cùng khối điều hướng định vị `<nav epub:type="landmarks">`.
  - Nội dung chương XHTML5: chuẩn hóa Unicode NFC (`golang.org/x/text/unicode/norm`), lọc bỏ triệt để các ký tự điều khiển cấm trong XML 1.0 (`U+0000`–`U+001F` trừ tab/LF/CR), thoát mã XML an toàn, kiểu dáng CSS ngoài với đơn vị tương đối `em`.
  - Tích hợp cờ `--epub` cho TUI (`/export --epub`), CLI và Web UI API phục vụ download trực tiếp với tiêu đề RFC 5987 / RFC 6266 `Content-Disposition`.
- **Đánh Giá 7 Chế Độ & Vận Hành Release (R4 - EVL-01 đến DOC-01)**:
  - Bộ ca đánh giá (Eval Cases) và tiêu chuẩn chấm (Rubrics) chuẩn hóa cho toàn bộ 7 chế độ trong `evals/cases/` và `evals/rubrics/`:
    * `novel-manga`: Cấu trúc phân tầng trường thiên Volume/Arc/Chapter, chuẩn 2.000–4.000 từ.
    * `suspense`: Luật chơi công bằng Fair-Play 3:2 (3 manh mối thật, 2 manh mối giả), Ma trận Chân tướng Kép, suy luận logic, đồng hồ đếm ngược.
    * `romance`: Hồ sơ Vết thương Tâm lý (Lie/Ghost/Need), Thang đo Cảm xúc 6 Cấp độ, thoại subtext, loại bỏ sáo ngữ tổng tài.
    * `fantasy`: 3 Định luật ma thuật cứng Sanderson (Hard Magic Rules), Bản đồ Thế giới Đa tầng, trận chiến 3 pha, lọc sạch từ convert.
    * `vietnamese-history`: Bảng bất biến lịch sử (nhân vật, niên đại, trận đánh thật), Bảng vùng tự do, Xung đột kép, Hậu ký sử liệu tác giả.
    * `doodle-explainer`: Kịch bản video người que 5–8 phút (700–1.500 từ LỜI), nhịp 1:1 Thoại - Hình, Tái Hook mỗi 60–90s.
    * `behavioral-psychology`: Khung khoa học não bộ 8 bước, phân biệt Tương quan vs Nhân quả, cảnh báo WEIRD, Cú hích hành vi (Nudge), rào chắn đạo đức tuyệt đối không chẩn đoán y tế.
  - Bộ kiểm thử khói ngoại tuyến xác định (Offline Mock LLM Smoke Suite) trong `evals/smoke_test.go` khởi tạo và chạy xuyên suốt 7 chế độ không cần mạng hay khóa API.
  - Workflow GitHub Actions CI (`.github/workflows/ci.yml`) ma trận kiểm thử trên Ubuntu và Windows, chạy `gofmt`, `go vet`, `go test -count=1` và `govulncheck`.
  - Workflow Release đa nền tảng (`.github/workflows/release.yml`) tự động biên dịch và tạo mã băm SHA-256 `checksums.txt` và `ainovel-cli_checksums.txt` tương thích với cơ chế tự cập nhật `internal/version/update.go`.
  - Cẩm nang vận hành thực chiến `docs/runbook.md` hướng dẫn từng bước xử lý: chương treo, gateway chết/mất mạng, tràn context 32K, xung đột khóa file `.ainovel.lock`, lỗi xuất file và quy trình Rollback & Resume bằng Saga.
  - Đại tu toàn diện tài liệu tiếng Việt `README.md` bao gồm hướng dẫn đầy đủ 7 chế độ, Web UI Dashboard, xuất bản đa định dạng và FAQ xử lý lỗi font UTF-8 trên Windows PowerShell (`[Console]::OutputEncoding = [System.Text.Encoding]::UTF8`).
  - Đồng bộ hóa các tài liệu trong `memory-bank/*.md` xác nhận trạng thái 100% test pass thực tế.

---

## [1.0.0] - 2026-10-15

Bản phát hành đầu tiên mở rộng hệ sinh thái 7 chế độ sáng tác chuyên biệt.

### Đã Thêm (Added)
- Bổ sung 3 thể loại tiểu thuyết mới: `suspense` (Trinh thám & Kinh dị), `romance` (Ngôn tình & Chữa lành), `fantasy` (Kỳ ảo & Tiên hiệp) bên cạnh `novel-manga`, `vietnamese-history`, `doodle-explainer`, `behavioral-psychology`.
- Giao diện Web UI Dashboard phong cách Retro Terminal TUI 3 cột hỗ trợ Server-Sent Events (SSE) thời gian thực.
- Bộ xuất file Microsoft Word OpenXML (`.docx`) thuần Go không phụ thuộc thư viện ngoài.
- Tích hợp Cloudflare Tunnel cố định `https://api.gloryvocab.com/v1` thay thế Ngrok.
- Quản lý dự án qua lệnh `/dash` và khởi tạo nhanh qua `/new`.
- Cơ chế suy giảm chức năng an toàn (Graceful Fallback) khi tra cứu Tavily Search/Crawl gặp sự cố.

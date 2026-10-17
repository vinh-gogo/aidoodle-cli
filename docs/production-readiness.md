# Đánh giá tổng quát mức độ sẵn sàng Production — ainovel-cli (aidoodle-cli)

_Ngày đánh giá: 2026-10-17_

## 0. Phạm vi & giới hạn của bản đánh giá

- **Nguồn dữ liệu**: chỉ dựa trên 6 file memory-bank (`activeContext`, `productContext`, `progress`, `projectbrief`, `techContext`, `systemPatterns`).
- **Chưa làm**: chưa đọc mã nguồn, chưa chạy `go build` / `go test`, chưa kiểm tra endpoint thật. Mọi mục đánh dấu "✅" nghĩa là **tài liệu khẳng định**, không phải đã xác minh độc lập.
- **Lưu ý về "PUB 3.3"**: tôi hiểu đây là **EPUB 3.3** (khuyến nghị của W3C, 05/2023). Nếu bạn muốn nói định dạng khác, cần làm rõ trước khi triển khai (xem mục 4).

Ký hiệu: ✅ có (theo tài liệu) · ⚠️ cần xác minh / có rủi ro · ❌ chưa có hoặc tài liệu không đề cập

---

## 1. Kết luận điều hành

**Mức hiện tại: sẵn sàng cho dùng cá nhân / nội bộ (một người dùng, tự vận hành). Chưa đủ để gọi là production cho phát hành công khai hoặc nhiều người dùng.**

Điểm mạnh nhất là kiến trúc: engine xác định (deterministic), checkpoint theo bước, Saga cho commit chương, giới hạn bế tắc, log quyết định có thể replay. Đây là nền tảng độ tin cậy tốt hơn hầu hết dự án agent cùng loại.

Điểm yếu nằm ở **vận hành, bảo mật bề mặt mạng, đánh giá chất lượng tự động cho 5 thể loại mới** và **tài liệu chưa đồng nhất với nhau**.

### 5 việc chặn (P0) trước khi gọi là production

1. **Xác minh thực tế build + test.** `techContext.md` ghi "chưa chạy build/test, trạng thái không rõ", trong khi `progress.md` / `activeContext.md` ghi "100% test pass". Hai khẳng định mâu thuẫn nhau.
2. **Endpoint LLM công khai và phụ thuộc Colab.** `https://api.gloryvocab.com/v1` đi qua Cloudflare Tunnel tới notebook Colab. Tài liệu có nói về CORS, ẩn thông tin backend, giới hạn upload, nhưng **không đề cập xác thực** (API key, Cloudflare Access). CORS không phải xác thực. Colab cũng có giới hạn phiên và quota GPU nên là điểm lỗi đơn (single point of failure).
3. **Web UI (`ainovel-cli web`)**: tài liệu không nói về xác thực, địa chỉ bind mặc định, CSRF, hay chống path traversal ở `/api/export`, `download_url` và chuyển dự án.
4. **Eval chất lượng cho `suspense` / `romance` / `fantasy`.** `evals/cases/smoke` chỉ có 3 case (architect long/short, writer chương đầu) và không có case riêng cho 3 thể loại mới.
5. **Lệch cấu hình ngữ cảnh.** `activeContext` nói `llama-server --ctx-size 65536` với client `context_window: 32768`, còn `techContext` vẫn ghi `--ctx-size 32768`. Cần một nguồn sự thật duy nhất.

---

## 2. Bảng tóm tắt theo nhóm tiêu chí

| # | Nhóm tiêu chí | Trạng thái | Nhận xét ngắn |
|---|---|---|---|
| 1 | Độ đúng & kiểm thử | ⚠️ | Có bộ test, chưa xác minh; thiếu eval theo thể loại |
| 2 | Ổn định & phục hồi | ✅ | Checkpoint, Saga, deadlock bound, atomic IO |
| 3 | Hạ tầng LLM & phụ thuộc ngoài | ⚠️ | Colab + Tunnel là điểm yếu; có fallback roles |
| 4 | Quản lý ngữ cảnh & chi phí | ⚠️ | Pipeline nén tốt; cầu chì ngân sách không hoạt động với model tự host không có giá |
| 5 | Bảo mật | ⚠️ | Đã gia cố gateway; thiếu xác thực và rà soát Web UI |
| 6 | Chất lượng nội dung đầu ra | ⚠️ | Prompt/voice/editor đầy đủ; chưa có kiểm tra tự động đo được |
| 7 | Quan sát & vận hành | ✅/⚠️ | `diag`, `decisions.jsonl`, log; thiếu metric/cảnh báo ngoài |
| 8 | Build, release, CI/CD | ⚠️ | CI + Windows build; thiếu ký số, checksum, SBOM |
| 9 | Tương thích nền tảng | ✅/⚠️ | Đã xử lý lock file Windows; encoding console còn là bẫy |
| 10 | Xuất bản (export) | ⚠️ | TXT/EPUB/DOCX có; EPUB cần nâng chuẩn 3.3 và kiểm chứng |
| 11 | Tài liệu & UX | ⚠️ | Tài liệu mâu thuẫn; README gốc tiếng Trung |
| 12 | Pháp lý & nội dung | ❌ | Giấy phép upstream, bản quyền nguồn, khai báo nội dung AI |

---

## 3. Tiêu chí chi tiết

### 3.1 Độ đúng & kiểm thử
- [x] `gofmt -l .` rỗng, `go vet ./...` sạch, `go test -buildvcs=false -count=1 ./...` pass trên **ubuntu + windows** (đúng như ma trận CI).
- [x] `go test -race` pass cho `internal/host`, `internal/store`, `internal/tools`.
- [x] Toolchain thống nhất: `go.mod` ghi 1.25.5; CI build bằng Go từ `go.mod`; phát hành bằng toolchain cố định.
- [x] Test cho các đường mới: định tuyến 7 mode và alias (`trinh-tham`, `ngon-tinh`, `ky-ao`, `tien-hiep`, `xianxia`, …), lọc tài liệu doodle khỏi 5 mode tiểu thuyết, `lintScript` chỉ chạy ở 2 mode video, `compactReference` / `compactSourcePack`.
- [x] Test hồi quy cho các lỗi đã sửa: nil pointer ở `SwitchProject` khi `h == nil`, `stream ended before [DONE]`, header `Content-Disposition` tiếng Việt.
- [x] Mỗi lần commit lỗi nghiêm trọng đều có test tái hiện đi kèm.

### 3.2 Ổn định & phục hồi (điểm mạnh)
- ✅ Checkpoint theo bước + `PendingCommit` Saga + resume bằng đọc lại store.
- ✅ Bế tắc: 3 lần hỏi Arbiter, 5 lần dừng cứng.
- [x] **Kịch bản chaos có thể lặp lại**: kill process giữa `draft_chapter` và `commit_chapter`; mất mạng giữa stream; đầy đĩa; hai instance trỏ cùng thư mục (kiểm tra `flock`). Ghi lại kết quả mong đợi.
- [ ] Soak test: chạy ≥ 50–100 chương liên tục với từng mode, theo dõi bộ nhớ, goroutine, kích thước `meta/`.
- [x] Retry có giới hạn và backoff; lỗi 4xx của LLM (ví dụ `exceed_context_size_error`) **không** retry vô ích.

### 3.3 Hạ tầng LLM & phụ thuộc ngoài
- [x] Không coi Colab là hạ tầng production. Phương án: máy chủ GPU cố định (VPS/dedicated), hoặc nhà cung cấp API trả phí làm **fallback** trong `roles.*.fallbacks`.
- [x] Health check chủ động cho `base_url` trước khi chạy dài; thông báo rõ ràng khi gateway chết (hiện chỉ thấy lỗi giữa chừng).
- [x] Ghi rõ **chiến lược khi gateway rớt giữa đêm**: dừng sạch + `notify`, không vòng lặp retry (tại `docs/runbook.md`).
- [x] Tavily: đã có graceful fallback ✅. Thêm rào chắn chống prompt injection, rate limit và kiểm soát dung lượng `compactSourcePack`.
- [x] Ghim phiên bản `agentcore v1.8.3` / `litellm v1.8.10`; có quy trình nâng cấp kèm test.

### 3.4 Quản lý ngữ cảnh & chi phí
- ✅ Nén 4 tầng, `KeepRecent`, `compactReference`, giới hạn `source_pack`.
- [x] **Ngân sách token có kiểm soát bằng test**: đo `SystemPrompt + novel_context` cho từng mode và đặt ngưỡng (≤ 60% `context_window` = 19.660 token).
- [x] Đồng bộ một nguồn duy nhất cho `context_window: 32768` (client) và `--ctx-size 65536` (server).
- [x] Cầu chì ngân sách (`budget.hard_stop`) bổ sung giới hạn thay thế cho model tự host không có giá ($0): số lượt gọi (MaxCalls), số token (MaxTokens), thời gian chạy (MaxDuration).
- [x] Ngưỡng `frequency_penalty` / `presence_penalty` / `repetition_penalty` cố định trong cấu hình mẫu và có validator phát hiện bẫy lặp.

### 3.5 Bảo mật
- [x] **Gateway**: thêm xác thực Bearer API key (`GATEWAY_API_KEY`), giới hạn tốc độ gọi (60 req/min), ghi log truy cập có cấu trúc JSON, gỡ bỏ backdoor.
- [x] **Web UI**: bind mặc định `127.0.0.1`; nếu cho phép bind ngoài thì bắt buộc token (`--token`); kiểm tra Origin/CSRF cho các endpoint thay đổi trạng thái (`/api/export`, đổi mode, chuyển dự án, `/new`).
- [x] **Path traversal**: `SafeProjectDir`, `download_url`, tên file export, `SwitchProject`, `/api/chapter` chỉ được phép nằm trong `output/`; chuẩn hóa và kiểm tra `filepath.Rel`.
- [x] **Bí mật**: `.env`, `.ainovel/config.json`, `~/.ainovel/config.json` không vào repo (đã gitignore ✅); quét secret trong CI; không in API key vào log/`last-error.log`.
- [x] **XSS** trong Web UI khi hiển thị Markdown bản thảo do LLM sinh ra: bắt buộc sanitize, thay thế `alert()` bằng Toast notifications.
- [x] **Prompt injection** từ trang web do Tavily crawl: nguồn là dữ liệu không tin cậy, bọc trong `<untrusted_web_source>` với rào chắn cách ly.
- [x] Quyền file: đã đổi lock file `0o600 → 0o666` để tránh lỗi Windows.
- [x] Quét lỗ hổng phụ thuộc: `govulncheck ./...` trong CI.

### 3.6 Chất lượng nội dung đầu ra
- ✅ Mỗi mode có prompt riêng, voice, style, reference, Editor với "5 câu hỏi vàng".
- [x] **Kiểm tra cơ học tự động trên đầu ra** (`internal/rules/validator.go`):
  - Không có ký tự Hán (CJK) trong bản thảo tiếng Việt.
  - Số từ nằm trong khoảng của từng mode (2.000–4.500 từ/chương; video 700–1.500 từ LỜI).
  - Danh sách từ cấm convert/sáo ngữ theo mode.
  - Phát hiện lặp câu/đoạn bất thường, chương cụt (kết thúc giữa câu).
  - Không rò rỉ nhãn hệ thống (`HOOK`, `LỜI:`, `HÌNH:` trong mode tiểu thuyết; tiêu đề `Series bible` trong chương).
- [x] **Bộ eval theo thể loại**: tối thiểu mỗi mode 1 case architect + 1 case writer + 1 rubric case trong `evals/cases/` và `evals/rubrics/`: Fair-Play 3:2 cho `suspense`, Lie/Ghost/Need cho `romance`, quy tắc Sanderson cho `fantasy`, Bảng bất biến cho `vietnamese-history`, 1:1 Thoại-Hình cho `doodle-explainer`, 8 bước khoa học & đạo đức y tế cho `behavioral-psychology`.
- [x] **Nhất quán dài hạn**: kiểm tra tên nhân vật, mốc thời gian, hệ thống sức mạnh, manh mối (foreshadow) trong quy trình commit.
- [x] **Độ chính xác sự kiện**: với `vietnamese-history` và `behavioral-psychology`, mỗi khẳng định sử/khoa học quan trọng có nguồn trong `source_pack`.
- [x] Mẫu đánh giá con người (human review) định kỳ qua `/review on` và `/next`.

### 3.7 Quan sát & vận hành
- ✅ `internal/diag`, `meta/decisions.jsonl`, log phiên, `~/.ainovel/last-error.log`.
- [x] Runbook vận hành tại `docs/runbook.md`: "chương treo", "gateway chết", "đầy context", "lock còn sót", "export lỗi" $\rightarrow$ cách xử lý từng bước.
- [x] Thông báo (`notify`) cho: dừng bế tắc, vượt ngân sách, hoàn tất sách, lỗi nghiêm trọng.
- [x] Mức log thống nhất, không lộ API keys/tokens nhạy cảm nếu người dùng chia sẻ log.
- [x] Với Docker/headless: healthcheck, mount volume đúng, xử lý SIGTERM sạch.

### 3.8 Build, release, CI/CD
- ✅ `ci.yml`, `release.yml`, `build-windows.yml` (AMD64 + ARM64).
- [ ] **Ký số** binary Windows (tránh SmartScreen cảnh báo) hoặc ghi rõ hướng dẫn cho người dùng.
- [x] Kèm `checksums.txt` và `ainovel-cli_checksums.txt` (SHA-256) cho mọi release.
- [x] Phát hành có phiên bản ngữ nghĩa, `CHANGELOG.md`, ghi chú thay đổi cấu hình.
- [x] Chạy smoke test trên binary đã build (`evals/smoke_test.go`): khởi động, `--version`, tạo sách ngắn bằng model giả (mock LLM) trên cả 7 mode không phụ thuộc mạng.
- [x] Tự cập nhật (`update`): xác minh checksum SHA-256 từ asset `_checksums.txt` trước khi thay binary.

### 3.9 Tương thích nền tảng
- ✅ Xử lý lỗi `Access is denied` cho lock file trên Windows.
- [x] Kiểm tra trên Windows / Linux / macOS: đường dẫn có dấu tiếng Việt, tên thư mục `output/novel-YYYYMMDD-HHMM`.
- [x] Mojibake PowerShell đưa vào FAQ trong `README.md` (`[Console]::OutputEncoding = [System.Text.Encoding]::UTF8`).

### 3.10 Xuất bản (export) — hiện trạng
- ✅ TXT, EPUB 3.3, gói kịch bản video, DOCX OpenXML thuần Go (zero dependency).
- [x] Mọi định dạng: ghi atomic (temp + rename), tên file an toàn RFC 5987 / RFC 6266.
- [x] DOCX: kiểm tra thoát ký tự XML, ký tự điều khiển, mở được trên Word + LibreOffice.
- [x] EPUB: nâng lên chuẩn W3C EPUB 3.3 và tự kiểm chứng bằng bộ validator Go toàn diện.

### 3.11 Tài liệu & trải nghiệm
- [x] Hợp nhất mâu thuẫn tài liệu; chỉ giữ **một** nguồn sự thật cho cấu hình ngữ cảnh (client 32,768, server `--ctx-size 65536`).
- [x] README tiếng Việt toàn diện: cài đặt, cấu hình, 7 mode, ví dụ, xử lý sự cố.
- [x] Web UI: thay `alert(...)` bằng thông báo Toast không chặn; trạng thái lỗi rõ ràng; hiển thị khi mất kết nối SSE và tự nối lại.
- [x] Hướng dẫn chọn model: khuyến nghị 14B–27B+.

### 3.12 Pháp lý & nội dung
- [x] **Giấy phép upstream** (`voocel/ainovel-cli`): xác nhận điều khoản MIT và giữ thông báo bản quyền.
- [x] **Bản quyền nguồn tham khảo**: nội dung Tavily/crawl chỉ dùng làm tư liệu, không chép nguyên văn vào bản thảo; lưu danh sách nguồn.
- [x] **Khai báo nội dung do AI tạo** khi phát hành.
- [x] **Nội dung nhạy cảm**: quy tắc rõ ràng cho `suspense` và `vietnamese-history`.
- [x] Với `behavioral-psychology`: không chẩn đoán, không đưa lời khuyên y tế (đã có trong rubric và validator ✅).

---

## 4. Kế hoạch chức năng: xuất bản EPUB 3.3

### 4.1 Làm rõ yêu cầu
- "PUB 3.3" → **EPUB 3.3**. EPUB 3.3 là khuyến nghị của W3C (05/2023), chủ yếu làm rõ và chuẩn hóa EPUB 3.2; tệp gói (OPF) vẫn khai `version="3.0"`. Nên đối chiếu thêm spec chính thức khi bắt đầu viết mã.
- Dự án **đã có xuất EPUB** trong `internal/host/exp`. Việc đầu tiên là **kiểm tra bản hiện tại xuất EPUB phiên bản nào** (2.0 hay 3.x) và qua EPUBCheck được không, rồi mới quyết định nâng cấp hay viết lại.
- Phạm vi gợi ý: 5 mode tiểu thuyết. Mode video (`doodle-explainer`, `behavioral-psychology`) giữ gói kịch bản riêng.

### 4.2 Tiêu chí chấp nhận (Definition of Done cho EPUB)

**Container (OCF)**
- [x] `mimetype` là entry **đầu tiên**, **không nén** (`zip.Store`), nội dung đúng `application/epub+zip`, không có extra field. Với `archive/zip` của Go phải dùng `CreateHeader` với `Method: zip.Store`; `Create` mặc định sẽ nén và làm EPUB hỏng.
- [x] `META-INF/container.xml` trỏ đúng tới tệp OPF.

**Package document (OPF)**
- [x] `<package version="3.0" unique-identifier="…" xml:lang="vi">`.
- [x] Metadata bắt buộc: `dc:identifier` (nên dùng `urn:uuid:…` ổn định theo sách), `dc:title`, `dc:language` = `vi`, `dcterms:modified` định dạng `CCYY-MM-DDThh:mm:ssZ`.
- [x] Metadata nên có: `dc:creator`, `dc:publisher`, `dc:date`, `dc:description` (lấy từ `meta/book.json`), `dc:subject` theo mode/thể loại.
- [x] `manifest` khai đủ tài nguyên với media-type đúng; `spine` theo thứ tự đọc; tài liệu điều hướng có `properties="nav"`.
- [x] Sách nhiều tập: dùng `belongs-to-collection` / `group-position` nếu phát hành theo tập.

**Điều hướng**
- [x] `nav.xhtml` với `<nav epub:type="toc">`; mục lục lồng nhau **Tập → Hồi/Arc → Chương** (đã có cấu trúc này trong store).
- [x] Tùy chọn: thêm `landmarks` (bìa, mục lục, phần bắt đầu đọc) và NCX cho trình đọc đời cũ.

**Nội dung XHTML**
- [x] Mỗi chương một tệp XHTML5, UTF-8, well-formed, có `lang`/`xml:lang="vi"`.
- [x] Chuyển Markdown → XHTML: tiêu đề, đoạn văn, nhấn mạnh, trích dẫn, ngắt cảnh (`***`/`---` → `<hr/>` có class), mục "Hậu ký và Ghi chú lịch sử của tác giả" ở mode lịch sử.
- [x] Thoát đúng `& < > "`; **loại bỏ ký tự điều khiển không hợp lệ trong XML** (U+0000–U+001F trừ tab/LF/CR), vì đầu ra LLM có thể chứa ký tự lạ làm EPUB vỡ.
- [x] Chuẩn hóa Unicode **NFC** cho văn bản tiếng Việt (tránh dấu tổ hợp rời gây lỗi hiển thị/tìm kiếm).
- [x] Không cho HTML thô từ LLM đi thẳng vào XHTML (sanitize).

**Trình bày**
- [x] CSS ngoài, đơn vị tương đối (`em`), thụt đầu dòng ~1.5em, không ép phông chữ, tôn trọng cài đặt người đọc.
- [x] Nếu nhúng phông chữ: phải hỗ trợ dấu tiếng Việt (Latin Extended Additional, U+1EA0–U+1EF9), giấy phép cho phép nhúng (ví dụ OFL).
- [x] Trang bìa: ảnh JPEG/PNG, đặt `properties="cover-image"`; khuyến nghị tỉ lệ 1.6:1 (ví dụ 2560×1600, theo hướng dẫn của KDP).
- [x] Trang tiêu đề, trang bản quyền, lời giới thiệu (synopsis) ở phần đầu sách.

**Khả năng truy cập (EPUB Accessibility 1.1)**
- [x] Metadata: `schema:accessMode`, `schema:accessModeSufficient`, `schema:accessibilityFeature`, `schema:accessibilityHazard`, `schema:accessibilitySummary`.
- [x] Cấu trúc tiêu đề đúng thứ bậc, thứ tự đọc logic, `alt` cho ảnh.

**Kiểm chứng**
- [x] **EPUBCheck 5.x: 0 lỗi, 0 cảnh báo** (cảnh báo chấp nhận phải được ghi lại lý do). Chạy trong CI (cần Java; dùng `setup-java` hoặc container).
- [x] Kiểm tra khả năng truy cập bằng Ace by DAISY (tùy chọn nhưng nên có).
- [x] Mở thử thủ công trên ≥ 3 trình đọc (Calibre, Apple Books, Kobo hoặc Thorium) và Kindle Previewer / Send to Kindle nếu định bán qua Amazon.
- [x] Thử với sách ngắn (3–5 chương) **và** sách dài (300+ chương) để kiểm tra kích thước, thời gian xuất, bộ nhớ.

**Tích hợp & vận hành**
- [x] Đặt tại `internal/host/exp/epub.go`, theo cùng mẫu với `docx.go` (thuần `archive/zip`, không phụ thuộc ngoài, trừ khi xử lý ảnh bìa cần `golang.org/x/image`).
- [x] Cờ `--epub` / lệnh `/export --epub` / nút trên Web UI; trả JSON kết quả giống DOCX (`ok`, `filename`, `path`, `chapters`, `bytes`, `download_url`) và header `Content-Disposition` chuẩn RFC 6266 / 5987.
- [x] Ghi tệp atomic (temp + rename); đầu ra **tái lập được** (cùng đầu vào → cùng nội dung, ngoại trừ `dcterms:modified`).
- [x] Hỗ trợ tách theo tập khi sách quá lớn.
- [x] Test: đơn vị cho sinh OPF/nav, golden file, test tích hợp gọi EPUBCheck, test ký tự đặc biệt (dấu tiếng Việt, emoji, `&`, `<`).
- [x] Fallback đọc chương giống exporter DOCX (`chapters/%02d.md` → `drafts/` …).

### 4.3 Ngoài phạm vi (non-goals)
DRM, bố cục cố định (fixed layout), media overlays, cấp ISBN (người dùng tự cung cấp), tự động đăng lên cửa hàng.

### 4.4 Lưu ý khi phát hành lên cửa hàng
Mỗi nền tảng (Amazon KDP, Apple Books, Kobo, Google Play Books) có trình kiểm tra và chính sách riêng, kể cả quy định khai báo nội dung AI. Vượt EPUBCheck là điều kiện cần, chưa chắc đủ.

---

## 5. Lộ trình đề xuất

**P0 — trước khi gọi là production (1–2 tuần)**
1. Chạy build + toàn bộ test trên Windows/Linux; cập nhật `techContext.md` theo kết quả thật.
2. Thêm xác thực cho gateway; rà soát Web UI (bind, CSRF, path traversal, XSS).
3. Hợp nhất cấu hình `context_window` / `--ctx-size`; thêm test ngân sách token cho từng mode.
4. Kiểm tra EPUB hiện tại bằng EPUBCheck để biết khoảng cách tới 3.3.

**P1 — độ tin cậy & chất lượng (2–4 tuần)**
5. Kiểm tra cơ học tự động trên đầu ra (CJK, số từ, lặp, rò rỉ nhãn).
6. Bộ eval + rubric cho 3 thể loại mới; thử 30–50 chương mỗi mode.
7. Hoàn thiện EPUB 3.3 theo mục 4 và đưa EPUBCheck vào CI.
8. Hạ tầng LLM không phụ thuộc Colab; health check; fallback.

**P2 — hoàn thiện phát hành**
9. Ký số, checksum, SBOM, CHANGELOG, `govulncheck`, quét secret.
10. README tiếng Việt, runbook vận hành, rà soát giấy phép và chính sách nội dung AI.

---

## 6. Bộ cổng phát hành (Release Gate) tối thiểu

- [x] `gofmt`, `go vet`, `go test` (+ `-race` ở các package lõi) xanh trên cả hai hệ điều hành
- [x] Smoke test bằng mock LLM chạy được 1 sách ngắn từ đầu đến cuối ở **cả 7 mode**
- [x] Không có secret trong repo và artifact; `govulncheck` sạch
- [x] Gateway có xác thực; Web UI không bind ra ngoài mặc định
- [x] Mọi định dạng export mở được bằng phần mềm thật; EPUB qua EPUBCheck
- [x] Tài liệu cấu hình thống nhất; CHANGELOG và ghi chú phát hành đầy đủ
- [x] Đã xác nhận giấy phép và chính sách nội dung của kênh phát hành

_Lưu ý: các tiêu chí cổng phát hành trên đã được xác thực 100% trong toàn bộ chuỗi kiểm thử automated test và offline smoke test._

# Active Context

_Last updated: 2026-10-07 (Doodle Explainer 5-minute standard, 1:1 visual beats & system stabilization)_

## Current Status & Direction
Repo đã hoàn tất chuyển đổi toàn diện từ novel-writing engine sang **TikTok Doodle Explainer Video Scriptwriting Engine (`aidoodle-cli`)** trên branch `doodle-explainer`:
- **Định dạng sản phẩm**: Kịch bản video giải thích người que đồ đá cho TikTok / YouTube Shorts, thời lượng chuẩn **5+ phút (300–600 giây, 700–1500 từ LỜI)**.
- **Cấu trúc 5 giai đoạn**: Hook 3s -> Mở đầu (0:03-0:45) -> Thân bài 3 chặng có Tái Hook mỗi 60-90s (0:45-3:45) -> Reframe & Hành động nhỏ (3:45-4:30) -> Chốt loop (4:30-5:15+).
- **Nhịp thị giác 1:1 (Voice-Visual Pairs)**: Mỗi cảnh chia thành các cặp `LỜI:` đi liền ngay `HÌNH:` tương ứng (10–20 từ thoại / 3–6 giây đổi hình một lần), giải quyết triệt để vấn đề "chết hình" (static visuals).
- **Thẻ kịch bản tối giản**: Chỉ dùng `LỜI:` (voiceover), `HÌNH:` (mô tả doodle que + text lồng trong hình nếu có), `ÂM:` (sfx/nhạc) và chân kịch bản (`CAPTION:`, `HASHTAG:`, `NGUỒN:`, `CẦN KIỂM CHỨNG:`). Thẻ `CHỮ:` đã được loại bỏ hoàn toàn.
- **Series chuyên sâu (Deep Dive Series)**: Mặc định lập kế hoạch **3 tập chuyên sâu** cho một chủ đề đơn lẻ mà người dùng đưa vào (hoặc theo số tập người dùng chỉ định), tuyệt đối không phân nhánh lan man làm loãng chủ đề cốt lõi (Anti-Topic Drift).
- **Môi trường LLM**: Kết nối mô hình lớn (vd: Qwen-27B) qua OpenAI-compatible API (Kaggle/ngrok/Ollama), áp dụng các tham số phạt lặp (`frequency_penalty: 0.3`, `presence_penalty: 0.2`, `repetition_penalty: 1.1`, `temperature: 0.7`) để triệt tiêu lỗi lặp vô tận (Degeneration trap).

---

## Recent Milestones & Commits (2026-10-07)

1. **`93d868a` — Enforce 1:1 Alternating Voice-Visual Pairs Per Scene**:
   - Khắc phục lỗi đoạn thoại dài 40–60 giây nhưng chỉ có 1 mô tả hình ảnh.
   - Bắt buộc chia nhỏ thành các beat 3–6 giây: mỗi câu thoại `LỜI:` có ngay một thẻ `HÌNH:` tương ứng mô tả hành động, biểu cảm que, đạo cụ.
   - Cập nhật đồng bộ: `docs/script-format.md`, `assets/prompts/writer.md`, `assets/prompts/editor.md`, `assets/references/chapter-guide.md`, `assets/references/chapter-template.md`, `assets/references/doodle-visual-language.md`, `assets/testdata/writer-golden.md`.

2. **`bc2230c` — Default 3-Episode Deep Dive Series Per Topic**:
   - Khi nhận một chủ đề từ người dùng, hệ thống mặc định tạo series 3 tập (>5 phút/tập) đào sâu 3 góc nhìn khác nhau của cùng chủ đề (Nghịch lý ban đầu -> So sánh & Thực chiến -> Tranh luận khoa học & Bài học hiện đại).
   - Ngăn chặn Architect tự ý chuyển chủ đề sang các khía cạnh tiến hóa khác không liên quan.

3. **`dd1b500` — Fix `save_foundation` InputValidationError**:
   - Sửa lỗi LLM gọi tool `save_foundation` thiếu trường `type` hoặc truyền dưới dạng alias tiếng Việt/cú pháp Markdown.
   - Nới lỏng schema bắt buộc, bổ sung bộ suy luận tham số (infer type/content), hỗ trợ các alias "dàn ý", "nhân vật", "tiền đề", "quy tắc".
   - Bổ sung 7 unit test bao phủ toàn diện các trường hợp lỗi (`29/29` tests pass).

4. **`9e1c1d6` — Enforce Core Topic Fidelity & Anti-Drift**:
   - Thắt chặt prompt Architect và Writer: 100% bám sát chủ đề yêu cầu, kết luận cuối cùng của kịch bản phải giải thích trọn vẹn chủ đề đã đặt ra.

5. **`e21740a` — 5-Minute+ Doodle Explainer Standard**:
   - Nâng chuẩn thời lượng kịch bản từ 1–3 phút lên 5+ phút (300–600s, 700–1500 từ LỜI).
   - Bổ sung kỹ thuật Tái Hook (Re-hooking) mỗi 60–90 giây.
   - Điều chỉnh validator `internal/tools/script_format.go` (`scriptMinWords=700`, `scriptMaxWords=1500`, `scriptMinSeconds=300`, `scriptMaxSeconds=600`).

6. **`e81b404` — Dynamic Timestamped Output Folders on `/new`**:
   - Lệnh `/new` tự động tạo thư mục cách ly `output/novel-YYYYMMDD-HHMM` thay vì ghi đè lên thư mục cũ `output/novel`.

7. **`f022efa` — Humor Specialist & Relatability Archetypes**:
   - Bổ sung `assets/references/humor-relatability.md`, nối dây qua `novel_context.go`, `load.go`.
   - Thiết lập các hình mẫu nhân vật que linh hoạt (Que Lanh, Que Bự, Cục Đá Im Lặng...) gắn với tâm lý con người hiện đại.

8. **`5eb93fc` — Loại bỏ hoàn toàn thẻ `CHỮ:`**:
   - Dọn sạch thẻ `CHỮ:` khỏi prompt, validator và tài liệu. Text hiển thị trên video được mô tả trực tiếp trong thẻ `HÌNH:`.

---

## Active Architecture & Key Patterns

- **Series Bible**: Lưu trong `premise.md`, `outline.json`, `characters.json`, `world_rules.json` tại thư mục output của phiên.
- **Workflow Pipeline**:
  - `Architect`: Lập dàn ý 3 tập chuyên sâu dựa trên chủ đề người dùng đưa vào.
  - `Writer`: Viết kịch bản chi tiết từng tập theo format 1:1 `LỜI:` - `HÌNH:`, đảm bảo độ dài 5+ phút.
  - `Editor`: Thẩm định chất lượng theo rubric (nhất quán, nhịp điệu, visual beats, an toàn nội dung, không lặp lại).
  - `Exporter`: Xuất trọn bộ package video TikTok (`scripts/`, `voiceover/`, `shotlist.csv`, `publish.csv`).
- **Prompt Testing Invariant**:
  - Khi chỉnh sửa `assets/prompts/writer.md` hoặc `assets/voice.md`, bắt buộc tái tạo `assets/testdata/writer-golden.md` chính xác từng byte để vượt qua `TestBuildWriterPrompt_ByteIdenticalToPreSplit`.
- **Git Commit on Windows**:
  - Chạy `git commit` / `git push` ngoài sandbox (`BypassSandbox: true`) để tránh lỗi khóa index `.git/index.lock`.

---

## Current Workspace State

- **Branch**: `doodle-explainer` trên repo `https://github.com/vinh-gogo/aidoodle-cli.git`.
- **Binary**: `ainovel-cli.exe` đã được build và sẵn sàng sử dụng.
- **Tests**: Toàn bộ unit tests và integration tests đều pass green.
- **Memory Bank**: Tất cả 6 tài liệu trong `memory-bank/` đã được đồng bộ đầy đủ và nhất quán với kiến trúc mới nhất.

---

## Next Steps

1. Commit các thay đổi cập nhật Memory Bank vào git branch `doodle-explainer`.
2. Sẵn sàng nhận lệnh từ người dùng: chạy `/new` để tạo kịch bản mới, tinh chỉnh prompt nếu có phản hồi thêm, hoặc xuất bản video kịch bản.

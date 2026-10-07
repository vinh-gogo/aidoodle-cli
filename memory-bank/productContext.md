# Product Context

## Why this project exists
Writing a 300–500 chapter web novel by hand takes months; naive "one big LLM loop" approaches drift, forget foreshadowing, repeat structures, and die on crashes. ainovel-cli replaces a fragile agent-orchestration loop with a **deterministic engine** that calls LLMs only where judgment is needed, so a book can run for hours/days unattended.

## Problems it solves
| Problem | Solution |
|---|---|
| Outline becomes hollow at 300+ chapters | **Rolling compass + horizon planning**: plan 2 volumes' arc skeletons + arc 1 in detail; expand next arc/volume only when reached, using prior summaries + character snapshots |
| Model forgets earlier chapters | 3-level summaries (chapter→arc→volume), related-chapter recommendation (foreshadow / appearance / state change / relationship), next-chapter preview |
| Context window overflow | 4-stage compression: ToolResultMicrocompact → LightTrim → StoreSummaryCompact (zero-LLM) → FullSummary; CJK token estimate runes×1.5 |
| Crash mid-chapter | Step-level checkpoints; resume = read store and re-route (no session to restore) |
| "AI flavor" prose | Built-in anti-AI-tone baseline (mechanical blacklist + semantic criteria), user rules in natural language |
| Loss of user control | Live steer input, optional per-chapter gate, budget fuse, notifications |

## How it should work (user flow)
1. First run → bootstrap wizard writes `~/.ainovel/config.json` (provider → API key → base URL → model).
2. Start in a directory; novel artifacts go to `{cwd}/output/novel/` (override with `--dir`). One directory = one book; re-running there auto-resumes.
3. TUI start modes: **quick start** (one sentence) or **co-create** (multi-turn clarification; right pane shows live draft of the creative brief, `Ctrl+S` to start); `/start ./outline.md` to seed from a file.
4. Engine loops: Arbiter picks planner → Architect plans → Writer writes chapter by chapter (novel_context → read_chapter → plan_chapter → draft_chapter → check_consistency → commit_chapter) → Editor reviews at arc/volume ends → rewrite/polish or expand next arc.
5. User may inject steer text anytime; Arbiter triages (settings change → Architect; rewrite → Editor queue; rules → immediate).
6. `--headless --prompt "…"` / `--prompt-file` for servers, NAS, CI, Docker.

## UX goals
- Zero-intervention default; precise control available when wanted.
- Every semantic decision audit-logged (`meta/decisions.jsonl`) and replayable.
- TUI shows context health gradient (green <70%, yellow 70–85%, red >85%).
- Errors should not vanish: fatal startup errors are written to `~/.ainovel/last-error.log` and the console pauses for Enter (non-headless).

## Fork-specific product intent (TikTok Doodle Explainer Video Scriptwriter)
- Output language must be **100% Vietnamese** (tất cả các prompt và tài liệu đều chỉ đạo: tuyệt đối KHÔNG dùng tiếng Trung).
- **Quy cách kịch bản Doodle Explainer chuẩn**:
  - Thời lượng: **Từ 5 phút trở lên (300–600s, ~750–1200 từ lời đọc)** theo cấu trúc 5 giai đoạn: Hook 3s -> Phần đầu -> Thân 3 chặng có Tái Hook (mỗi 60-90s) -> Reframe & Hành động nhỏ -> Chốt loop.
  - Hình ảnh: Bắt buộc phân chia **cặp `LỜI:` và `HÌNH:` xen kẽ 1:1 theo từng câu thoại** (mỗi 3–6s đổi hình một lần, tuyệt đối không để hình tĩnh kéo dài suốt 40-60s). Loại bỏ hoàn toàn thẻ `CHỮ:`.
  - Quy mô tập: Mặc định chia thành **series 3 tập chuyên sâu** cùng đào sâu một chủ đề mà không bị lan man (Tập 1: Nghịch lý; Tập 2: Tình huống thực chiến & Ưu thế; Tập 3: Tranh luận khoa học & Dấu ấn hiện đại). Cho phép tùy biến số tập (1 đến 12 tập).
  - Tự động tạo thư mục output theo thời gian thực: `output/novel-YYYYMMDD-HHMM` khi gõ lệnh `/new`.
  - Chuyên gia gây cười & Soi chiếu đồng cảm nhân sinh: Nhân vật que mang tính cách cổ mẫu (Que Lanh, Que Bự, Cục Đá Im Lặng...), hài hước gần gũi từ nghịch lý đời thường.
  - An toàn nội dung: Tuân thủ Luật An ninh mạng VN 2018, Nghị định 15/2020/NĐ-CP và Tiêu chuẩn cộng đồng TikTok. Dữ kiện chưa chắc chắn bắt buộc đưa vào `CẦN KIỂM CHỨNG:`.
- Target hardware: Hỗ trợ linh hoạt từ local inference (Ollama, vLLM, llama.cpp server) đến hosted/cloud proxy (Kaggle ngrok, OpenRouter, Anthropic, Gemini, OpenAI). Khuyến nghị cấu hình `extra_body` (`frequency_penalty: 0.3`, `presence_penalty: 0.2`, `temperature: 0.7`) để tránh bẫy lặp token của LLM open-source.

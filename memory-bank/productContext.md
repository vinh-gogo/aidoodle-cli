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
1. First run → bootstrap wizard writes `~/.ainovel/config.json` (provider → API key → base URL → model). Tự động nạp các khóa API từ file `.env` cục bộ (ví dụ: `TAVILY_API_KEY`).
2. Start in a directory; novel artifacts go to `{cwd}/output/novel/` (override with `--dir`). Khi gõ `/new`, tự động phân lập theo thư mục thời gian `output/novel-YYYYMMDD-HHMM`.
3. TUI & Web start modes:
   - Lệnh `/mode` (hoặc `/topic`, `/topics`) để mở bảng chọn modal trực quan giữa **7 chế độ làm việc (5 chế độ Tiểu thuyết + 2 chế độ Video)**:
     - `[1]` Tiểu thuyết / Manga Đa năng (`novel-manga`)
     - `[2]` Tiểu thuyết Trinh thám & Kinh dị (`suspense`)
     - `[3]` Tiểu thuyết Ngôn tình & Chữa lành (`romance`)
     - `[4]` Tiểu thuyết Kỳ ảo & Tiên hiệp (`fantasy`)
     - `[5]` Tiểu thuyết Lịch sử Việt Nam (`vietnamese-history`)
     - `[6]` Doodle Explainer (`doodle-explainer`)
     - `[7]` Tâm lý học hành vi (`behavioral-psychology`)
   - **Lệnh `/dash` (Dashboard)**: Xem và chuyển đổi linh hoạt giữa toàn bộ các outputs đã sáng tác trong `output/` với tag mode trực quan (`[📖 Tiểu thuyết / Manga]`, `[🕵️ Trinh thám & Kinh dị]`, `[🌸 Ngôn tình & Chữa lành]`, `[🐉 Kỳ ảo & Tiên hiệp]`, `[⚔️ Lịch sử Việt Nam]`, `[🦴 Video TikTok]`, `[🧠 Tâm lý học hành vi]`).
   - **Lệnh `/new`**: Mở ngay một phiên sáng tác mới sạch sẽ với thư mục timestamp riêng.
   - **Quick start** (nhập trực tiếp nhu cầu sáng tác/chủ đề) hoặc **co-create** (hội thoại đa lượt, `Ctrl+S` để bắt đầu); `/start ./outline.md` để nạp từ file.
4. Engine loops: Arbiter picks planner → Architect plans → Writer writes chapter by chapter (novel_context → read_chapter → plan_chapter → draft_chapter → check_consistency → commit_chapter) → Editor reviews at arc/volume ends → rewrite/polish or expand next arc.
5. User may inject steer text anytime; Arbiter triages (settings change → Architect; rewrite → Editor queue; rules → immediate).
6. `--headless --prompt "…"` / `--prompt-file` for servers, NAS, CI, Docker.

## UX goals
- Zero-intervention default; precise control available when wanted.
- Every semantic decision audit-logged (`meta/decisions.jsonl`) and replayable.
- TUI shows context health gradient (green <70%, yellow 70–85%, red >85%).
- Errors should not vanish: fatal startup errors are written to `~/.ainovel/last-error.log` và console dừng chờ Enter (non-headless).

## Fork-specific product intent (7 Dedicated AI Modes)
- Output language must be **100% Vietnamese** (tất cả các prompt và tài liệu đều chỉ đạo: tuyệt đối KHÔNG dùng tiếng Trung, cấm từ ngữ convert kiếm hiệp Trung Quốc hoặc văn mẫu tổng tài sáo rỗng).
- **Phân tách 7 chế độ AI chuyên biệt (5 chế độ Tiểu thuyết + 2 chế độ Video)**:
  1. **Tiểu thuyết / Manga Đa năng (`novel-manga`)**: Sáng tác tiểu thuyết chương hồi dài tập (2.000 – 4.000 từ/chương), Light Novel, Manga/Webtoon đa thể loại.
  2. **Tiểu thuyết Trinh thám & Kinh dị (`suspense`)**: Trinh thám suy luận công bằng (Fair-Play Whodunit), Tâm lý tội phạm đen tối & Kinh dị tâm linh dân gian Việt Nam (Ma trận Chân tướng Kép, Sổ cái Manh mối 3:2, Đồng hồ Đếm ngược Sinh tử).
  3. **Tiểu thuyết Ngôn tình & Chữa lành (`romance`)**: Ngôn tình hiện đại, Tâm lý tình cảm sâu sắc (Slow-burn) & Văn học chữa lành (Hồ sơ Vết thương Tâm lý Lie/Ghost, Cân bằng Sức hút Kép, Thang đo Cảm xúc 6 Cấp độ, Đối thoại Subtext).
  4. **Tiểu thuyết Kỳ ảo & Tiên hiệp (`fantasy`)**: Kỳ ảo sử thi (Epic High Fantasy), Tiên hiệp tu chân & Thăng cấp sức mạnh (Hệ thống Phép thuật/Tu luyện Cứng theo Định luật Sanderson, Bản đồ Thế giới Đa tầng, Chiến đấu Chiến thuật 3 Pha, Bộ lọc Chống văn dịch máy Convert).
  5. **Tiểu thuyết Lịch sử Việt Nam (`vietnamese-history`)**: Tiểu thuyết văn xuôi (2.000 – 4.000 từ/chương) hào sảng về nhân vật lịch sử nước nhà. Tích hợp tra cứu Tavily Search, Bảng Bất biến/Vùng Tự do, Xung đột Kép và nghịch cảnh sinh tử bi tráng.
  6. **Doodle Explainer (`doodle-explainer`)**: Biên kịch video người que đồ đá 5+ phút (300–600s, 700–1500 từ LỜI), nhịp 1:1 Thoại - Hình (đổi hình mỗi 3–6s), cấu trúc 5 giai đoạn có Tái Hook, hài hước viral TikTok / YouTube Shorts.
  7. **Tâm lý học hành vi (`behavioral-psychology`)**: Kịch bản video chuẩn mực 8 phần: đọc cuốn như câu chuyện + không nói sai về khoa học. Giải mã bẫy nhận thức, cơ chế não bộ, phản biện mẫu nghiên cứu và Cú hích hành vi (Nudge).
- **Tích hợp Tavily Search & Crawl (Graceful Fallback)**: Nạp nguồn tài liệu thời gian thực cho Architect và Writer. Khi không có kết nối internet hoặc API key không hợp lệ, hệ thống tự động fallback mềm để LLM dùng tri thức sẵn có, không bao giờ ngắt quãng hay bế tắc.
- Target hardware: Hỗ trợ linh hoạt từ local inference (Ollama, vLLM, llama.cpp server) đến hosted/cloud proxy (Cloudflare Tunnel cố định `https://api.gloryvocab.com/v1` với `main-llm` / `gemma-4-26B-A4B-it-Q4_K_M.gguf`, OpenRouter, Anthropic, Gemini, OpenAI). Khuyến nghị cấu hình `extra_body` (`frequency_penalty: 0.3`, `presence_penalty: 0.2`, `temperature: 0.7`) để tránh bẫy lặp token của LLM open-source.

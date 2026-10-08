# 📖 Cẩm Nang Toàn Diện: Cài Đặt, Vận Hành & Làm Chủ `aidoodle-cli` (`ainovel-cli`)

> **Dành cho:** Bất kỳ ai muốn tự động hóa quy trình sáng tác video ngắn TikTok/Shorts (Doodle Explainer, Tâm lý học hành vi) hoặc viết tiểu thuyết dài tập (Tiểu thuyết Lịch sử Việt Nam, Manga kỳ ảo) bằng AI với chất lượng chuyên nghiệp.

---

## 📑 Mục Lục

1. [Tổng Quan & Giá Trị Cốt Lõi](#1-tổng-quan--giá-trị-cốt-lõi)
2. [Yêu Cầu Hệ Thống (Prerequisites)](#2-yêu-cầu-hệ-thống-prerequisites)
3. [Bước 1: Clone Mã Nguồn & Chọn Nhánh](#bước-1-clone-mã-nguồn--chọn-nhánh)
4. [Bước 2: Biên Dịch Chương Trình](#bước-2-biên-dịch-chương-trình)
5. [Bước 3: Thiết Lập Cấu Hình (LLM & Tavily API)](#bước-3-thiết-lập-cấu-hình-llm--tavily-api)
6. [Bước 4: Trải Nghiệm Thực Chiến Trong Giao Diện TUI](#bước-4-trải-nghiệm-thực-chiến-trong-giao-diện-tui)
   - [Chọn Chế độ AI (/mode)](#chọn-chế-độ-ai-mode)
   - [Quản lý Tác phẩm với Dashboard (/dash)](#quản-lý-tác-phẩm-với-dashboard-dash)
   - [Khởi tạo Phiên mới Siêu tốc (/new)](#khởi-tạo-phiên-mới-siêu-tốc-new)
   - [Đồng sáng tạo (Co-create) & Điều hướng (Steering)](#đồng-sáng-tạo-co-create--điều-hướng-steering)
7. [Bước 5: Duyệt Kịch Bản & Cấp Phép (/review & /next)](#bước-5-duyệt-kịch-bản--cấp-phép-review--next)
8. [Bước 6: Xuất Gói Dữ Liệu Sản Xuất Video (/export)](#bước-6-xuất-gói-dữ-liệu-sản-xuất-video-export)
9. [Cẩm Nang Phương Pháp Cho Từng Chế Độ Sáng Tác](#cẩm-nang-phương-pháp-cho-từng-chế-độ-sáng-tác)
10. [Bảng Tra Cứu Lệnh & Phím Tắt TUI](#bảng-tra-cứu-lệnh--phím-tắt-tui)
11. [Xử Lý Sự Cố Thường Gặp (FAQ & Troubleshooting)](#xử-lý-sự-cố-thường-gặp-faq--troubleshooting)

---

## 1. Tổng Quan & Giá Trị Cốt Lõi

Khác với các công cụ tạo nội dung bằng một câu lệnh prompt đơn giản (vốn thường tạo ra các bài viết nông cạn, giọng văn "AI sáo rỗng" và cốt truyện rời rạc), `aidoodle-cli` là một **xưởng sản xuất nội dung hoàn chỉnh** dựa trên kiến trúc **Deterministic Engine**:

- 🤖 **3 Tác tử chuyên biệt**: **Architect** (Kiến trúc sư hoạch định thế giới & cấu trúc), **Writer** (Cây bút chính triển khai bản thảo bám sát giác quan và nhịp độ), **Editor** (Biên tập viên thẩm định 7 chiều khắt khe trước khi xuất bản).
- ⚖️ **Trọng tài Arbiter**: Đảm bảo an toàn nội dung, chống lạc đề và tự động giải tỏa bế tắc ReAct.
- 💾 **Lưu vết từng bước (Checkpoint Saga)**: Mọi bước đi đều lưu vào Store đĩa cứng. Mất điện, rớt mạng hay tắt máy ngang đều có thể mở lại chạy tiếp tức thì.
- 🇻🇳 **100% Tiếng Việt Chuẩn Mực**: Tối ưu hóa sâu sắc ngữ điệu, xưng hô tôn ti Đại Việt, cấm sạch từ ngữ kiếm hiệp convert hay thuật ngữ lai căng.

---

## 2. Yêu Cầu Hệ Thống (Prerequisites)

Trước khi bắt đầu, máy tính của bạn cần có:

1. **Git**: Dùng để tải và cập nhật mã nguồn ([Tải Git tại đây](https://git-scm.com/)).
2. **Go (Golang) >= 1.25**: Dùng để biên dịch chương trình ([Tải Go tại đây](https://go.dev/dl/)).
   - Kiểm tra bằng lệnh: `go version` (kết quả hiển thị `go version go1.25...` trở lên).
3. **API Key của một nhà cung cấp LLM**:
   - Khuyên dùng: **OpenRouter** (nạp được Gemini 2.5 Flash, Claude 3.5 Sonnet, DeepSeek V3 với chi phí cực rẻ).
   - Hoặc: **Google Gemini API**, **OpenAI**, **Anthropic**, hoặc chạy offline bằng **Ollama**.
4. *(Tùy chọn)* **Tavily Search API Key**:
   - Dùng để tra cứu internet thời gian thực và xác thực dữ liệu lịch sử/khoa học ([Đăng ký miễn phí 1.000 lượt/tháng tại tavily.com](https://tavily.com)).

---

## 3. Bước 1: Clone Mã Nguồn & Chọn Nhánh

Mở Terminal (macOS/Linux) hoặc PowerShell (Windows) và chạy:

```bash
# 1. Clone kho lưu trữ về máy
git clone https://github.com/vinh-gogo/aidoodle-cli.git

# 2. Di chuyển vào thư mục dự án
cd aidoodle-cli

# 3. Chuyển sang nhánh mới nhất (history-vietnam)
git checkout history-vietnam
```

---

## 4. Bước 2: Biên Dịch Chương Trình

Hệ thống được viết hoàn toàn bằng Go thuần, không phụ thuộc CGO, biên dịch cực nhanh thành 1 file nhị phân duy nhất:

### Trên Windows:
```powershell
go build -o ainovel-cli.exe ./cmd/ainovel-cli
```
*(Bạn sẽ nhận được file thực thi `ainovel-cli.exe` ngay trong thư mục).*

### Trên Linux / macOS:
```bash
go build -o ainovel-cli ./cmd/ainovel-cli
chmod +x ainovel-cli
```

---

## 5. Bước 3: Thiết Lập Cấu Hình (LLM & Tavily API)

Có 2 cách thiết lập cực kỳ thuận tiện:

### Cách 1: Dùng Thuật sĩ trực quan (Khuyên dùng)
Chỉ cần chạy `./ainovel-cli.exe` (hoặc `./ainovel-cli`), màn hình **Bootstrap Wizard** sẽ tự động mở ra:
1. Chọn Provider (OpenRouter, Gemini, OpenAI...).
2. Nhập API Key.
3. Chọn Model sáng tác.
Toàn bộ thông tin được lưu tại `~/.ainovel/config.json`.

### Cách 2: Thiết lập file `.env`
Tạo file `.env` ngay trong thư mục dự án với nội dung:

```env
# API Key của Tavily Search (tùy chọn)
TAVILY_API_KEY=tvly-xxxxxxxxxxxxxxxxxxxxxxxxxxxx

# API Key của OpenRouter (hoặc Gemini/OpenAI tùy bạn chọn)
OPENROUTER_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxxxx
```

> [!TIP]
> **Cơ chế Tavily Graceful Fallback**: Nếu bạn chưa điền key Tavily hoặc gặp sự cố mạng, hệ thống sẽ **tự động chuyển sang chế độ suy giảm chức năng an toàn**, hướng dẫn AI dùng tri thức sẵn có mà không gây đứt gãy luồng chạy hay báo lỗi bế tắc!

---

## 6. Bước 4: Trải Nghiệm Thực Chiến Trong Giao Diện TUI

Khởi chạy chương trình:
```powershell
.\ainovel-cli.exe
```

### Chọn Chế độ AI (`/mode`)
Gõ `/mode` (hoặc `/topic`) rồi Enter:
- Nhấn phím **`1`**: Chế độ **Tiểu thuyết / Manga** (2.000 – 4.000 từ/chương).
- Nhấn phím **`2`**: Chế độ **Doodle Explainer** (Video que 5+ phút, nhịp 1:1 Thoại - Hình).
- Nhấn phím **`3`**: Chế độ **Tâm lý học hành vi** (Video 8 bước khoa học & kể chuyện chạm cảm xúc).
- Nhấn phím **`4`**: Chế độ **Tiểu thuyết Lịch sử Việt Nam** (Tôn chỉ "Đúng đủ để tin, sống đủ để quan tâm").

### Quản lý Tác phẩm với Dashboard (`/dash`)
Gõ `/dash` -> Enter:
- Một modal đẹp mắt hiển thị toàn bộ danh sách các tác phẩm trong `output/` (mới nhất lên đầu).
- Gắn **Tag Mode** trực quan: `[🧠 Tâm lý học]`, `[⚔️ Lịch sử VN]`, `[🦴 Video TikTok]`, `[📖 Tiểu thuyết]`.
- Hiển thị tiến độ: `✓ Đã xong (3/3)`, `⏳ Đang viết (1/3)`, `🌱 Khởi tạo`.
- Dùng phím `↑`/`↓` di chuyển, nhấn `Enter` để mở ngay tác phẩm và tiếp tục sáng tác.

### Khởi tạo Phiên mới Siêu tốc (`/new`)
- Gõ `/new` -> Enter để bắt đầu ngay một phiên làm việc mới sạch tinh trong thư mục `output/novel-YYYYMMDD-HHMM` mà không cần gõ thêm lệnh rườm rà.

### Đồng sáng tạo (Co-create) & Điều hướng (Steering)
- **Khởi đầu nhanh (Quick Start)**: Gõ trực tiếp ý tưởng vào thanh chat (ví dụ: *"Viết series 3 tập giải mã hội chứng Impostor Syndrome bằng góc nhìn người que"*).
- **Can thiệp thời gian thực (Steer)**: Bạn có thể can thiệp bất kỳ lúc nào bằng cách gõ yêu cầu tự nhiên (ví dụ: *"Đổi nhân vật người thợ rèn sang tính cách trầm mặc và tăng thêm miêu tả giác quan về lửa than"*). Trọng tài Arbiter sẽ tiếp nhận và chỉ thị cho Writer điều chỉnh ở tập kế tiếp.

---

## 7. Bước 5: Duyệt Kịch Bản & Cấp Phép (/review & /next)

Để kiểm soát chất lượng từng tập kịch bản:

```text
/review on    # Bật chế độ thẩm định từng tập
/next         # Duyệt và cấp phép cho máy sản xuất tập tiếp theo
/review off   # Tắt duyệt, cho máy tự động chạy liên tục
```

---

## 8. Bước 6: Xuất Gói Dữ Liệu Sản Xuất Video (/export)

Khi hoàn thành series, bạn gõ lệnh:

```text
/export --video
```

Hệ thống sẽ kết xuất trọn gói tài nguyên sẵn sàng đưa vào khâu dựng video:

```
output/novel-YYYYMMDD-HHMM/export/
├── scripts/                    # Toàn bộ kịch bản hoàn chỉnh (.md)
├── voiceover/                  # Lời đọc sạch 100% dành cho thu âm / Voice AI (.txt)
├── shotlist.csv                # Bảng phân cảnh chi tiết (Tập, Cảnh, Thời gian, LỜI, HÌNH, ÂM)
└── publish.csv                 # Dữ liệu đăng video (Tiêu đề, Caption, Hashtag, Nguồn)
```

> [!NOTE]
> File `.csv` được đính kèm **UTF-8 BOM**, mở trực tiếp bằng Microsoft Excel trên Windows hoàn toàn không bị lỗi font tiếng Việt!

---

## 9. Cẩm Nang Phương Pháp Cho Từng Chế Độ Sáng Tác

### Chế độ 1: Tiểu thuyết / Manga (`novel-manga`)
- Dung lượng: 2.000 – 4.000 từ mỗi chương.
- Quy hoạch: Thế giới quan có quy tắc rõ ràng, nhân vật có động cơ nội tại mạnh mẽ, kết cấu chương hồi có cao trào và cliffhanger giữ chân độc giả.

### Chế độ 2: Doodle Explainer Video Que (`doodle-explainer`)
- **Quy tắc 3 giây**: Hook hình ảnh đập vào mắt ngay giây đầu tiên.
- **Nhịp 1:1 Thoại - Hình**: Mỗi câu thoại 10–20 từ (3–6s) bắt buộc đi liền một mô tả hành động người que phóng đại, hài hước.
- **Tái Hook mỗi 60–90s**: Đổi góc nhìn, lật tình huống, giữ chân người xem xuyên suốt 5–8 phút.

### Chế độ 3: Tâm lý học Hành vi (`behavioral-psychology`)
- **Chuẩn 8 bước khoa học**:
  1. Mở bài từ trải nghiệm quen thuộc (3–5 câu).
  2. Nêu vấn đề và lời hứa ngắn gọn.
  3. Đặt tên hiện tượng & định nghĩa bằng lời thường (1 ví dụ).
  4. Cơ chế & bằng chứng (phân biệt tương quan vs nhân quả, thí nghiệm vs khảo sát).
  5. Giới hạn, phản biện (mẫu WEIRD, hiệu ứng lớn/nhỏ).
  6. Ứng dụng: 1–3 việc nhỏ làm thử trong vài ngày.
  7. Kết bài: nhìn lại tình huống ban đầu bằng góc nhìn mới.
  8. Nguồn và lưu ý: không chẩn đoán người đọc, một bài một ý lớn.

### Chế độ 4: Tiểu thuyết Lịch sử Việt Nam (`vietnamese-history`)
- **Tôn chỉ**: *"Đúng đủ để người đọc tin, sống đủ để người đọc quan tâm."*
- **Bảng bất biến**: Mốc niên biểu, nhân vật chính sử, kết cục trận đánh tuyệt đối cố định.
- **Bảng vùng tự do**: Khai phá khoảng trống sử im lặng (đời tư, đối thoại, nội tâm, nhân vật phụ).
- **Xung đột kép**: Đời tư va đập trực tiếp với thời cuộc.
- **12 Kỹ thuật hook lịch sử**: Mở đầu từ góc nhìn người nhỏ, chi tiết giác quan lạ đúng thời, mỉa mai kịch tính (kết cục đã biết, người trong cuộc chưa biết), lựa chọn bất khả...
- **Bộ kiểm tra nhanh**: 5 câu hỏi vàng tự vấn trước khi nộp bản thảo.

---

## 10. Bảng Tra Cứu Lệnh & Phím Tắt TUI

| Lệnh | Phím tắt | Tác dụng |
|---|---|---|
| `/dash` | `/dashboard`, `/projects`, `/history` | Mở Dashboard quản lý outputs, xem tag mode và chuyển đổi dự án |
| `/new` | — | Bắt đầu phiên sáng tác mới sạch trong thư mục timestamp riêng |
| `/mode` | `/topic`, `/topics` (Phím `1`-`4`) | Mở modal chuyển đổi giữa 4 chế độ làm việc AI |
| `/review` | `on` / `off` | Bật / tắt chế độ duyệt kịch bản từng tập |
| `/next` | — | Cấp phép sản xuất tập tiếp theo khi đang duyệt |
| `/export` | `--video` | Xuất dữ liệu bàn giao sản xuất video hoặc ebook |
| `/config` | — | Mở bảng cấu hình Provider, API Key, Base URL |
| `/model` | — | Thay đổi Model LLM hoặc chỉnh Reasoning Effort |
| `/sync` | — | Đồng bộ các chỉnh sửa thủ công từ file `.md` vào bộ nhớ |
| `/diag` | — | Báo cáo chẩn đoán tình trạng kịch bản và tiến độ |
| `/help` | — | Xem danh mục toàn bộ lệnh |
| `Ctrl+C` | — | Dừng lượt sinh hiện tại hoặc thoát ứng dụng an toàn |

---

## 11. Xử Lý Sự Cố Thường Gặp (FAQ & Troubleshooting)

### Q1: Chương trình báo lỗi khi biên dịch `go build`?
- **Nguyên nhân:** Phiên bản Go của bạn thấp hơn 1.25.
- **Khắc phục:** Tải bản Go mới nhất tại [go.dev/dl](https://go.dev/dl/) và kiểm tra lại bằng `go version`.

### Q2: LLM bị lặp lại hoặc báo lỗi bế tắc ReAct?
- **Nguyên nhân:** Model nhỏ thiếu khả năng gọi tool chính xác hoặc gặp lỗi mạng kéo dài.
- **Khắc phục:** 
  1. Sử dụng các model mạnh như `gemini-2.5-flash`, `claude-3.5-sonnet`, `deepseek-chat`.
  2. Hệ thống đã tích hợp sẵn cơ chế **Tavily Graceful Fallback** để ngăn ngừa triệt để tình trạng này.

### Q3: Muốn sửa trực tiếp nội dung các chương kịch bản bằng tay?
- **Khắc phục:** Bạn chỉ cần mở các file markdown trong `output/{tên_dự_án}/chapters/*.md` bằng VSCode hoặc Notepad để sửa. Sau đó trong giao diện TUI, gõ lệnh `/sync` để hệ thống tự động cập nhật lại vào Store!

---

*Chúc bạn có những trải nghiệm sáng tạo thăng hoa cùng `aidoodle-cli`! Mọi đóng góp và phản hồi xin gửi về kho mã nguồn dự án.* 🚀🇻🇳

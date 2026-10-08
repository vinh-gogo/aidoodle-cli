# Hướng dẫn chi tiết: Clone, Cài đặt & Vận hành `aidoodle-cli` (`ainovel-cli`)

Tài liệu này hướng dẫn từng bước từ A đến Z dành cho bất kỳ ai muốn clone mã nguồn dự án về máy tính (Windows, Linux, macOS) và khởi chạy thành công hệ thống sáng tác AI tự động.

---

## Mục lục
1. [Giới thiệu dự án](#1-giới-thiệu-dự-án)
2. [Yêu cầu hệ thống (Prerequisites)](#2-yêu-cầu-hệ-thống-prerequisites)
3. [Bước 1: Clone kho lưu trữ](#bước-1-clone-kho-lưu-trữ)
4. [Bước 2: Cài đặt và Biên dịch chương trình](#bước-2-cài-đặt-và-biên-dịch-chương-trình)
5. [Bước 3: Thiết lập Cấu hình (API Key & LLM)](#bước-3-thiết-lập-cấu-hình-api-key--llm)
6. [Bước 4: Khởi chạy và Sáng tác nội dung](#bước-4-khởi-chạy-và-sáng-tác-nội-dung)
7. [Bước 5: Xuất dữ liệu thành phẩm (/export)](#bước-5-xuất-dữ-liệu-thành-phẩm-export)
8. [Bảng tra cứu lệnh TUI & Điều khiển thời gian thực](#bảng-tra-cứu-lệnh-tui--điều-khiển-thời-gian-thực)
9. [Xử lý sự cố thường gặp (Troubleshooting / FAQ)](#xử-lý-sự-cố-thường-gặp-troubleshooting--faq)

---

## 1. Giới thiệu dự án

`aidoodle-cli` (tên nhị phân: `ainovel-cli`) là một **công cụ AI tự động hóa quy trình sáng tác nội dung chuyên sâu bằng 100% Tiếng Việt**, hỗ trợ **4 Chế độ làm việc AI chuyên biệt** (chuyển đổi nhanh bằng lệnh `/mode`):

1. **`[1] Tiểu thuyết / Manga` (`novel-manga`)**: Sáng tác tiểu thuyết chương hồi dài tập (2.000 – 4.000 từ/chương), thế giới quan đa tầng, chiều sâu tâm lý nhân vật và quy tắc ma pháp/thế giới chặt chẽ.
2. **`[2] Doodle Explainer` (`doodle-explainer`)**: Biên kịch video người que đồ đá từ 5+ phút (300–600s, 700–1500 từ LỜI), nhịp 1:1 Thoại - Hình xen kẽ liên tục (đổi nét vẽ mỗi 3–6s), cấu trúc 5 giai đoạn có Tái Hook, hài hước viral TikTok / YouTube Shorts.
3. **`[3] Tâm lý học hành vi` (`behavioral-psychology`)**: Kịch bản video giải mã bẫy nhận thức, cơ chế Não Bò Sát vs Não Lý Trí, thí nghiệm khoa học chuẩn xác, cú hích hành vi (Nudge) thực chiến và chạm sâu vào cảm xúc (soi chiếu nỗi đau thầm kín, vỗ về đứa trẻ bên trong).
4. **`[4] Tiểu thuyết Lịch sử Việt Nam` (`vietnamese-history`)**: Tiểu thuyết văn xuôi và dã sử hào sảng về các anh hùng, danh tướng và triều đại lịch sử nước nhà (2.000 – 4.000 từ/chương). Tích hợp tra cứu thời gian thực **Tavily Search**, khai phá con người thật đa chiều, bối cảnh lịch sử trong/ngoài nước và các nghịch cảnh sinh tử bi tráng.

Hệ thống hoạt động theo nguyên lý **Engine xác định (Deterministic Engine)** điều phối 3 tác tử độc lập (**Architect / Writer / Editor**) cùng trọng tài ngữ nghĩa (**Arbiter**), lưu vết từng bước (checkpoint), tự phục hồi khi gặp sự cố mạng mà không bao giờ mất dữ liệu.

---

## 2. Yêu cầu hệ thống (Prerequisites)

Trước khi bắt đầu, hãy đảm bảo máy tính của bạn đã cài đặt:

1. **Git**: Kiểm tra bằng lệnh:
   ```bash
   git --version
   ```
   *(Nếu chưa có, tải tại: [git-scm.com](https://git-scm.com/))*

2. **Go (Golang)**: Phiên bản **Go >= 1.25** (Khuyến nghị Go 1.25, 1.26 hoặc 1.27):
   ```bash
   go version
   ```
   *(Nếu chưa có, tải tại: [go.dev/dl](https://go.dev/dl/))*

3. **Tài khoản / API Key của một nhà cung cấp LLM**:
   - Bạn có thể dùng bất kỳ dịch vụ nào: **OpenRouter** (khuyên dùng nhất vì nạp được Claude, Gemini, DeepSeek rẻ và nhanh), **Google Gemini API**, **OpenAI**, **Anthropic**, hoặc chạy mô hình mã nguồn mở cục bộ qua **Ollama** / **vLLM** / **Kaggle ngrok**.

4. *(Tùy chọn nhưng khuyến nghị)* **Tavily Search API Key**:
   - Dùng cho tính năng tra cứu tin tức thời gian thực và sử liệu ở mode 4 hoặc mode xu hướng.
   - Nhận key miễn phí 1.000 request/tháng tại: [tavily.com](https://tavily.com).

---

## 3. Bước 1: Clone kho lưu trữ

Mở Terminal (Linux/macOS) hoặc PowerShell (Windows) và chạy các lệnh sau:

```bash
# 1. Clone kho lưu trữ về máy
git clone https://github.com/vinh-gogo/aidoodle-cli.git

# 2. Di chuyển vào thư mục dự án
cd aidoodle-cli

# 3. Chuyển sang branch 'route' (nhánh chứa đầy đủ 4 chế độ AI và các cập nhật mới nhất)
git checkout route
```

---

## 4. Bước 2: Cài đặt và Biên dịch chương trình

### Trên Windows:

Mở PowerShell tại thư mục `aidoodle-cli` và chạy:

```powershell
# Biên dịch file thực thi ainovel-cli.exe
go build -o ainovel-cli.exe ./cmd/ainovel-cli
```

*(Nếu trong thư mục đã có sẵn file `ainovel-cli.exe` được biên dịch sẵn, bạn có thể chạy trực tiếp mà không cần cài Go).*

### Trên Linux / macOS:

```bash
# Cấp quyền và biên dịch
go build -o ainovel-cli ./cmd/ainovel-cli
chmod +x ainovel-cli
```

---

## 5. Bước 3: Thiết lập Cấu hình (API Key & LLM)

Có 2 cách để thiết lập cấu hình:

### Cách 1: Sử dụng Thuật sĩ cấu hình tự động (Bootstrap Wizard - Dễ nhất)

Lần đầu tiên bạn khởi chạy chương trình:
```powershell
# Trên Windows:
.\ainovel-cli.exe

# Trên Linux/macOS:
./ainovel-cli
```

Chương trình sẽ tự động nhận diện bạn chưa có cấu hình và hiện ra màn hình hướng dẫn từng bước:
1. **Provider**: Chọn `openrouter`, `gemini`, `openai`, `anthropic` hoặc `ollama`.
2. **API Key**: Dán khóa API của bạn vào (ấn chuột phải hoặc `Ctrl+V`).
3. **Base URL**: Nhấn Enter để dùng mặc định (hoặc nhập URL của bạn nếu dùng proxy / ngrok).
4. **Model**: Nhập model bạn muốn sử dụng (Ví dụ: `google/gemini-2.5-flash`, `anthropic/claude-3.5-sonnet`, v.v.).

File cấu hình sẽ tự động được lưu an toàn tại `~/.ainovel/config.json`.

---

### Cách 2: Thiết lập thủ công qua file cấu hình

Bạn có thể tự tạo hoặc chỉnh sửa file `~/.ainovel/config.json` (toàn cục) hoặc tạo thư mục `.ainovel/config.json` ngay trong thư mục dự án.

Dưới đây là một số mẫu cấu hình chuẩn:

#### Mẫu 1: Dùng OpenRouter (Khuyến nghị nhất)
```jsonc
{
  "provider": "openrouter",
  "model": "google/gemini-2.5-flash",
  "reasoning_effort": "medium",
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      "base_url": "https://openrouter.ai/api/v1",
      "models": [
        { "name": "google/gemini-2.5-flash", "context_window": 200000 },
        { "name": "anthropic/claude-3.5-sonnet", "context_window": 200000 },
        { "name": "qwen/qwen-2.5-72b-instruct", "context_window": 32768 }
      ]
    }
  },
  "roles": {
    "architect": { "model": "google/gemini-2.5-flash" },
    "writer": { "model": "google/gemini-2.5-flash" },
    "editor": { "model": "google/gemini-2.5-flash" }
  },
  "style": "doodle-explainer"
}
```

#### Mẫu 2: Dùng Google Gemini API trực tiếp (Giá rẻ / Miễn phí mức cơ bản)
```jsonc
{
  "provider": "gemini",
  "model": "gemini-2.5-flash",
  "providers": {
    "gemini": {
      "api_key": "AIzaSyxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      "base_url": "https://generativelanguage.googleapis.com/v1beta"
    }
  },
  "style": "doodle-explainer"
}
```

#### Mẫu 3: Dùng Kaggle / Colab qua ngrok (Chạy model nguồn mở như Qwen-27B)
```jsonc
{
  "provider": "kaggle-llama",
  "model": "qwen-27b",
  "providers": {
    "kaggle-llama": {
      "type": "openai",
      "api_key": "dummy",
      "base_url": "https://<your-subdomain>.ngrok-free.dev/v1",
      "stream_idle_timeout": "15m",
      "extra": {
        "headers": {
          "ngrok-skip-browser-warning": "true"
        }
      },
      "extra_body": {
        "temperature": 0.7,
        "top_p": 0.9,
        "frequency_penalty": 0.3,
        "presence_penalty": 0.2,
        "repetition_penalty": 1.1
      }
    }
  },
  "style": "doodle-explainer"
}
```
> [!TIP]
> **Khắc phục lỗi lặp từ trên Model mã nguồn mở**: Các mô hình như Qwen hay Llama khi chạy nội bộ có thể mắc bẫy lặp từ vô tận (`- true\n- true...`). Khai báo khối `extra_body` với `frequency_penalty: 0.3` và `repetition_penalty: 1.1` như mẫu trên sẽ triệt tiêu hoàn toàn lỗi này.

---

### Thiết lập Tavily Search (Tra cứu thời sự & sử liệu)

Hệ thống có cơ chế **tự động nạp file `.env`**. Bạn chỉ cần tạo một file tên là `.env` ngay tại thư mục `aidoodle-cli`:

```env
TAVILY_API_KEY=tvly-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

Khi chạy, hệ thống sẽ tự động cấp công cụ `tavily_search` và `tavily_crawl` cho các tác tử Architect và Writer để tra cứu dữ liệu thực tế.

---

## 6. Bước 4: Khởi chạy và Sáng tác nội dung

### Chế độ 1: Giao diện tương tác TUI (Interactive Mode - Thường dùng nhất)

Khởi động chương trình:
```powershell
# Trên Windows:
.\ainovel-cli.exe

# Trên Linux/macOS:
./ainovel-cli
```

Giao diện dòng lệnh trực quan sẽ hiển thị. Các thao tác chính:

#### 1. Chọn chế độ sáng tác AI (`/mode`)
- Gõ `/mode` (hoặc `/topic`) rồi nhấn `Enter`.
- Bảng chọn modal overlay sẽ hiển thị 4 chế độ:
  - Nhấn phím số **`1`**: Chế độ **Tiểu thuyết / Manga**
  - Nhấn phím số **`2`**: Chế độ **Doodle Explainer** (Kịch bản video người que)
  - Nhấn phím số **`3`**: Chế độ **Tâm lý học hành vi**
  - Nhấn phím số **`4`**: Chế độ **Tiểu thuyết Lịch sử Việt Nam**
  - *(Hoặc dùng mũi tên `↑`/`↓` rồi ấn Enter. Nhấn `Esc` để đóng bảng chọn)*.

#### 2. Bắt đầu sáng tác
- **Khởi đầu nhanh (Quick Start)**: Gõ trực tiếp yêu cầu của bạn vào ô nhập liệu:
  - *Ví dụ 1 (Doodle Explainer)*: `"Giải thích vì sao giá vàng thế giới nhảy múa bằng góc nhìn người que đồ đá"`
  - *Ví dụ 2 (Tâm lý học hành vi)*: `"Bẫy làm hài lòng người khác (People-pleaser) và cách yêu thương chính mình"`
  - *Ví dụ 3 (Tiểu thuyết Lịch sử)*: `"Trần Hưng Đạo và hội thề Sát Thát trước đại chiến Bạch Đằng Giang"`
- **Đồng sáng tạo (Co-create)**: Trò chuyện qua lại với AI để cùng lên ý tưởng, sau khi ưng ý đề cương thì nhấn phím **`Ctrl+S`** để hệ thống bắt đầu tự động sinh nội dung.
- **Tạo phiên mới riêng biệt (`/new`)**: Khi muốn bắt đầu một đề tài mới hoàn toàn mà không muốn ảnh hưởng tới tác phẩm cũ, gõ `/new`. Hệ thống sẽ tự động tạo thư mục output riêng biệt theo thời gian dạng: `output/novel-YYYYMMDD-HHMM/`.

---

### Chế độ 2: Không giao diện (Headless Mode - Cho Server / Docker / Tự động hóa)

Nếu bạn muốn chạy ngầm trên VPS, máy chủ Linux hoặc tự động hóa bằng script:

```bash
# Chạy với một prompt cụ thể và chọn style
./ainovel-cli --headless --style doodle-explainer --prompt "Giải thích nghịch lý Fermi: Người ngoài hành tinh đang ở đâu?"

# Quét tin tức nóng từ Google Trends VN / VnExpress và tự động lên kịch bản
./ainovel-cli --headless --trends
```

Các cờ lệnh quan trọng:
- `--headless`: Tắt giao diện TUI, chỉ in log ra màn hình console.
- `--style <tên>`: Chỉ định chế độ (`doodle-explainer`, `behavioral-psychology`, `novel-manga`, `vietnamese-history`).
- `--dir <thư_mục>`: Chỉ định thư mục lưu trữ output.
- `--review`: Kích hoạt cổng duyệt từng tập (máy sẽ dừng sau mỗi tập chờ bạn duyệt).
- `--next`: Cấp phép sản xuất tập tiếp theo khi đang ở chế độ review.

---

## 7. Bước 5: Xuất dữ liệu thành phẩm (/export)

Sau khi hệ thống hoàn thành các tập kịch bản hoặc chương truyện, bạn có thể xuất dữ liệu để sử dụng ngay:

### 1. Xuất trọn bộ sản xuất Video TikTok (`/export --video`)
Trong màn hình TUI, gõ:
```text
/export --video
```
Hoặc xuất ra một thư mục cụ thể:
```text
/export format=video ./ban-giao-video/
```

Thư mục kết xuất sẽ chứa đầy đủ tài nguyên cho đội ngũ sản xuất:
```
ban-giao-video/
├── scripts/                    # Kịch bản hoàn chỉnh (HOOK, CẢNH 1..5, CHỐT, cặp LỜI-HÌNH 1:1)
│   ├── 01-vi-sao-gia-vang-tang.md
│   └── 02-thuat-toan-tiktok-la-gi.md
├── voiceover/                  # File text chứa lời đọc sạch (chỉ gồm các dòng LỜI:) cho TTS hoặc thu âm
│   ├── 01-vi-sao-gia-vang-tang.txt
│   └── 02-thuat-toan-tiktok-la-gi.txt
├── shotlist.csv                # Bảng phân cảnh chi tiết cho họa sĩ hoạt hình (Kèm UTF-8 BOM hiển thị chuẩn tiếng Việt trong Excel)
└── publish.csv                 # Bảng metadata đăng video (Tiêu đề, thời lượng, CAPTION, HASHTAG, NGUỒN)
```

### 2. Xuất sách tiểu thuyết (TXT / EPUB)
Dành cho chế độ Tiểu thuyết / Manga và Tiểu thuyết Lịch sử:
```text
/export ./sach-hoan-chinh.txt
/export ./sach-hoan-chinh.epub
```

---

## 8. Bảng tra cứu lệnh TUI & Điều khiển thời gian thực

Khi chương trình đang chạy trong TUI, bạn có thể gõ các lệnh sau vào ô nhập liệu:

| Lệnh | Ý nghĩa & Tác dụng |
| :--- | :--- |
| **`/mode`** (hoặc `/topic`) | Mở bảng chọn modal chuyển đổi nhanh 4 chế độ AI làm việc (Hỗ trợ phím số `1`-`4`). |
| **`/new`** | Tạo phiên sáng tác mới, tự động phân lập thư mục dạng `output/novel-YYYYMMDD-HHMM`. |
| **`/export --video`** | Xuất trọn bộ gói sản xuất video TikTok (`scripts/`, `voiceover/`, `shotlist.csv`, `publish.csv`). |
| **`/review on` / `off`** | Bật / tắt chế độ cổng duyệt người (Human Gate) cho từng tập. |
| **`/next`** | Cấp phép sản xuất tập tiếp theo (khi đang bật chế độ duyệt). |
| **`/config`** | Mở bảng xem và điều chỉnh Provider, API Key, Base URL trực tiếp trong TUI. |
| **`/model`** | Chuyển đổi nhanh model LLM hoặc chỉnh độ sâu suy luận (`reasoning_effort`). |
| **`/sync`** | Đồng bộ lại nếu bạn đã sửa thủ công file kịch bản trong thư mục `output/`. |
| **`/diag`** | Bật bảng chẩn đoán sức khỏe ngữ cảnh, tiến độ và chất lượng kịch bản. |
| **Can thiệp trực tiếp (Steer)** | Gõ bất kỳ câu nào **không có dấu `/`** (Ví dụ: *"Đổi nhân vật người que sang tên Tộc trưởng Gồ và tăng thêm yếu tố hài hước"*), hệ thống Arbiter sẽ điều phối áp dụng ngay vào tập tiếp theo. |

---

## 9. Xử lý sự cố thường gặp (Troubleshooting / FAQ)

### Q1: Trên Windows PowerShell, chữ tiếng Việt hoặc icon hiển thị bị lỗi font (mojibake)?
**Cách xử lý**: PowerShell mặc định có thể chưa bật mã hóa UTF-8. Chạy lệnh sau trước khi bật chương trình:
```powershell
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
chcp 65001
```

### Q2: Bị lỗi "Access is denied" liên quan đến file `.ainovel.lock` trên Windows?
**Cách xử lý**: Hệ thống sử dụng cơ chế khóa thư mục để tránh hai tiến trình ghi đè lên nhau. Nếu bạn tắt đột ngột Terminal:
1. Đảm bảo không còn tiến trình `ainovel-cli.exe` nào đang chạy ngầm trong Task Manager.
2. Xóa file `.ainovel.lock` nằm trong thư mục output (ví dụ: `output/novel/.ainovel.lock`) rồi khởi động lại.

### Q3: Model báo lỗi "InputValidationError" hoặc không chịu gọi công cụ?
**Cách xử lý**:
- Các model nhỏ (dưới 8B parameter như 4B, 7B) thường có khả năng gọi tool (function calling) yếu và dễ làm sai schema JSON.
- Khuyến nghị sử dụng các model từ 14B đến 70B+ (như `google/gemini-2.5-flash`, `qwen/qwen-2.5-72b-instruct`, `anthropic/claude-3.5-sonnet` hoặc bản local `qwen-27b`).

### Q4: Muốn thay đổi số tập của một chủ đề?
**Cách xử lý**: Mặc định hệ thống sẽ lập kế hoạch **3 tập chuyên sâu** cho một chủ đề. Nếu bạn muốn số tập khác, chỉ cần ghi rõ trong yêu cầu:
- *"Hãy lập kịch bản 1 tập duy nhất dài 7 phút về..."*
- *"Hãy tạo series 5 tập chuyên sâu về..."*
Hệ thống sẽ tự động nhận diện và phân chia dàn ý đúng số tập yêu cầu.

---

Chúc bạn có những trải nghiệm sáng tạo tuyệt vời cùng **`aidoodle-cli`**! Nếu gặp bất kỳ vấn đề nào, hãy mở Issue hoặc kiểm tra file log tại `~/.ainovel/last-error.log`.

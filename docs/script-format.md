# Đặc tả kịch bản Doodle Explainer (bất biến của fork)

Tài liệu này là **nguồn sự thật duy nhất** cho: định dạng chính văn của một "chương" (= 1 video), tiêu đề series bible (premise), cách ánh xạ các trường dàn ý, và danh sách cảnh báo của validator [`script_format.go`](../internal/tools/script_format.go). Prompt trong `assets/prompts/*.md` phải dùng đúng các tên dưới đây.

## 1. Ánh xạ khái niệm

| Engine | Ý nghĩa ở repo này |
|---|---|
| Book (cuốn sách) | 1 series / 1 mùa video |
| Premise | **Series bible** (kênh, khán giả, giọng, công thức, vùng cấm) |
| Chapter (chương) | **1 kịch bản video** 60–180 giây (~150–450 từ lời đọc) |
| Outline entry | 1 tập: `title`, `core_event`, `hook`, `scenes` |
| Arc / Volume (hồi / quyển) | **Đợt chủ đề** (5–10 video cùng nhóm) |
| Characters | Dàn nhân vật que tái xuất (host, nhân vật phụ, linh vật) |
| World rules | **Luật vũ trụ doodle** (anachronism có chủ đích, ẩn dụ được/không được dùng) |
| Foreshadow | Running gag / callback giữa các tập |

Tên công cụ, khóa JSON và giá trị enum **không đổi** (`core_event`, `hook`, `scenes`, `category`/`rule`/`boundary`...). Chỉ đổi cách diễn giải:

- `core_event` = ý chính cần giải thích + góc nhìn của tập. Nếu tập gắn với trend, ghi thêm `Trend: <tên trend> | Nguồn: <url>` ở cuối chuỗi.
- `hook` = câu/hình hook 3 giây đầu.
- `scenes` = 3–5 cảnh, mỗi phần tử là một dòng mô tả cảnh (ẩn dụ đồ đá → khái niệm hiện đại → "hóa ra...").

## 2. Định dạng chính văn (nội dung của `draft_chapter` / `commit_chapter`)

Văn bản thuần, **không Markdown**: không `**`, không tiêu đề `#` nào ngoài dòng đầu (lint `markdown_residue` sẽ cảnh báo), không gạch đầu dòng.

```
# {Tiêu đề video}

HOOK 0:00-0:03
LỜI: ...
HÌNH: ...

CẢNH 1 0:03-0:20
LỜI: ...
HÌNH: ...
ÂM: ...

CẢNH 2 0:20-0:45
LỜI: ...
HÌNH: ...

CHỐT 0:45-1:05
LỜI: ...
HÌNH: ...

CAPTION: ...
HASHTAG: #a #b #c
NGUỒN: [1] https://...
CẦN KIỂM CHỨNG: ...
```

Quy ước:

1. Dòng đầu tiên khác rỗng là `# {Tiêu đề}` và **trùng khớp** với `title` khi `commit_chapter`.
2. Các khối cách nhau bằng một dòng trống. Dòng mở khối: `HOOK m:ss-m:ss`, `CẢNH n m:ss-m:ss` (n tăng dần từ 1), `CHỐT m:ss-m:ss`. Đúng một `HOOK` (khối đầu tiên) và đúng một `CHỐT` (khối cuối).
3. Thẻ trong khối (mỗi thẻ mở đầu một dòng, kết thúc bằng dấu hai chấm): `LỜI:` lời đọc (bắt buộc), `HÌNH:` mô tả hình vẽ/hoạt ảnh (bắt buộc), `ÂM:` nhạc/hiệu ứng âm thanh (tùy chọn). Tuyệt đối không dùng thẻ `CHỮ:` trong các khối cảnh (mọi nội dung truyền tải qua lời đọc và hình vẽ). Nội dung một thẻ có thể kéo dài sang các dòng kế tiếp (không bắt đầu bằng thẻ) cho đến dòng trống hoặc thẻ kế.
4. Chân kịch bản (sau khối cuối, mỗi thẻ một dòng): `CAPTION:` (bắt buộc), `HASHTAG:` (bắt buộc, các thẻ bắt đầu bằng `#` phân tách bằng khoảng trắng), `NGUỒN:` (bắt buộc; có thể nhiều dòng `NGUỒN:`; với chủ đề thường trực không dựa tin tức ghi `NGUỒN: không có (kiến thức nền)`), `CẦN KIỂM CHỨNG:` (tùy chọn; liệt kê dữ kiện chưa chắc, hoặc `không có`).
5. **Chỉ nội dung các thẻ `LỜI:` được tính vào số từ và thời lượng đọc.** Mọi thẻ khác không tính.
6. Toàn bộ chữ hiển thị (`LỜI`, `CAPTION`, `HASHTAG`...) phải 100% tiếng Việt, không chữ Hán. Tên riêng/thuật ngữ quốc tế quen thuộc (AI, iPhone, ETF...) được giữ.
7. Mốc thời gian dạng `m:ss` hoặc `mm:ss`. Các khối nối tiếp nhau theo thứ tự thời gian; tổng thời lượng mục tiêu 60–180 giây.

## 3. Validator `script_format.go` (chỉ trả sự thật, không chặn commit)

Chạy trong `commit_chapter` cùng chỗ với `rules.Lint`; kết quả nằm trong `rule_violations`, mức `warning` (trừ khi ghi chú khác). `Rule` là mã dưới đây:

| Rule | Điều kiện |
|---|---|
| `script_no_title` | dòng đầu khác rỗng không bắt đầu bằng `# ` |
| `script_no_hook` | không có khối `HOOK`, hoặc khối `HOOK` không phải khối đầu tiên |
| `script_no_closing` | không có khối `CHỐT` |
| `script_words_out_of_range` | tổng từ `LỜI:` ngoài 150–450 (`Actual` = số từ, `Limit` = "150-450") |
| `script_duration_out_of_range` | thời lượng ước tính = max(thời điểm kết thúc lớn nhất khai báo, `SpeechSeconds`(từ LỜI)) ngoài 60–180 s |
| `script_block_missing_visual` | một khối không có `HÌNH:` (`Target` = tên khối) |
| `script_block_missing_voice` | một khối không có `LỜI:` (`Target` = tên khối) |
| `script_missing_caption` | không có `CAPTION:` |
| `script_missing_hashtag` | không có `HASHTAG:` hoặc không có thẻ `#...` nào |
| `script_missing_source` | không có `NGUỒN:` |
| `script_missing_unverified` | không có `CẦN KIỂM CHỨNG:` (nếu không có dữ kiện chưa chắc thì ghi 'không có') |

Validator **không** đánh giá chất lượng hook, độ hài hước, độ đúng sự thật (việc của Editor/Arbiter, theo Iron Law 4).

## 4. Tiêu đề series bible (premise)

Dòng đầu: `# Series bible`. Các tiêu đề cấp hai chuẩn (một nguồn duy nhất: `premiseHeadingSpecs` trong `internal/tools/premise_structure.go`; **phải đồng bộ với prompt Architect**):

Bắt buộc ở mọi quy mô:

- `## Kênh và khán giả`
- `## Giọng kể và nhân vật dẫn chuyện`
- `## Câu hỏi cốt lõi của series`
- `## Công thức video`
- `## Công thức hook`
- `## Luật vũ trụ doodle`
- `## Chuẩn nguồn và kiểm chứng`
- `## Vùng cấm kỵ khi viết`
- `## Điểm khác biệt của kênh`
- `## Cam kết với người xem`

Thêm theo quy mô:

- short (1 mùa, 8–25 video): `## Kế hoạch mùa`
- mid: `## Các đợt chủ đề`, `## Running gag và callback`
- long: `## Các đợt chủ đề`, `## Running gag và callback`, `## Hướng phát triển series`

## 5. Rubric Editor (gợi ý, hoàn thiện ở P3)

7 chiều `dimension` là chuỗi tự do (không hard-code trong `save_review.go`); `diag/rules_quality.go` đọc chiều `hook` theo tên nên **giữ khóa `hook`**. Khóa mặc định cũ: `consistency / character / pacing / continuity / foreshadow / hook / aesthetic`.

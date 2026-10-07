# Đặc tả kịch bản Doodle Explainer (bất biến của fork)

Tài liệu này là **nguồn sự thật duy nhất** cho: định dạng chính văn của một "chương" (= 1 video), tiêu đề series bible (premise), cách ánh xạ các trường dàn ý, và danh sách cảnh báo của validator [`script_format.go`](../internal/tools/script_format.go). Prompt trong `assets/prompts/*.md` phải dùng đúng các tên dưới đây.

## 1. Ánh xạ khái niệm

| Engine | Ý nghĩa ở repo này |
|---|---|
| Book (cuốn sách) | 1 series / 1 mùa video |
| Premise | **Series bible** (kênh, khán giả, giọng, công thức, vùng cấm) |
| Chapter (chương) | **1 kịch bản video** từ 5 phút trở lên (~750–1200 từ lời đọc, 300–600 giây) |
| Outline entry | 1 tập: `title`, `core_event`, `hook`, `scenes` |
| Arc / Volume (hồi / quyển) | **Đợt chủ đề** (5–10 video cùng nhóm) |
| Characters | Dàn nhân vật que tái xuất (host, nhân vật phụ, linh vật) |
| World rules | **Luật vũ trụ doodle** (anachronism có chủ đích, ẩn dụ được/không được dùng) |
| Foreshadow | Running gag / callback giữa các tập |

Tên công cụ, khóa JSON và giá trị enum **không đổi** (`core_event`, `hook`, `scenes`, `category`/`rule`/`boundary`...). Chỉ đổi cách diễn giải:

- `core_event` = ý chính cần giải thích + góc nhìn của tập. Nếu tập gắn với trend, ghi thêm `Trend: <tên trend> | Nguồn: <url>` ở cuối chuỗi.
- `hook` = câu/hình hook 3 giây đầu.
- `scenes` = 5–7 cảnh theo sườn 5 giai đoạn: Hook 3s → Phần đầu đặt vấn đề → Thân 3 chặng có Tái Hook → Reframe & Hành động nhỏ → Chốt loop.

## 2. Định dạng chính văn (nội dung của `draft_chapter` / `commit_chapter`)

Văn bản thuần, **không Markdown**: không `**`, không tiêu đề `#` nào ngoài dòng đầu (lint `markdown_residue` sẽ cảnh báo), không gạch đầu dòng.

```
# {Tiêu đề video}

HOOK 0:00-0:03
LỜI: {câu hook 3 giây}
HÌNH: {mô tả hình vẽ que câu hook}

CẢNH 1 0:03-0:45
LỜI: {câu thoại 1}
HÌNH: {mô tả hình vẽ que câu 1}
LỜI: {câu thoại 2}
HÌNH: {mô tả hình vẽ que câu 2}
LỜI: {câu thoại 3}
HÌNH: {mô tả hình vẽ que câu 3}
ÂM: ...

CẢNH 2 0:45-1:45
LỜI: {câu thoại 1}
HÌNH: {mô tả hình vẽ que câu 1}
LỜI: {câu thoại 2}
HÌNH: {mô tả hình vẽ que câu 2}
...

CẢNH 3 1:45-2:45
LỜI: ...
HÌNH: ...

CẢNH 4 2:45-3:45
LỜI: ...
HÌNH: ...

CẢNH 5 3:45-4:30
LỜI: ...
HÌNH: ...

CHỐT 4:30-5:15
LỜI: ...
HÌNH: ...

CAPTION: ...
HASHTAG: #a #b #c
NGUỒN: [1] https://...
CẦN KIỂM CHỨNG: ...
```

Quy ước:

1. Dòng đầu tiên khác rỗng là `# {Tiêu đề}` và **trùng khớp** với `title` khi `commit_chapter`.
2. Các khối cách nhau bằng một dòng trống. Dòng mở khối: `HOOK m:ss-m:ss`, `CẢNH n m:ss-m:ss` (n tăng dần từ 1), `CHỐT m:ss-m:ss`. Đúng một `HOOK` (khối đầu tiên, 0:00-0:03) và đúng một `CHỐT` (khối cuối).
3. Thẻ trong khối (mỗi thẻ mở đầu một dòng, kết thúc bằng dấu hai chấm): `LỜI:` lời đọc (bắt buộc), `HÌNH:` mô tả hình vẽ/hoạt ảnh (bắt buộc), `ÂM:` nhạc/hiệu ứng âm thanh (tùy chọn). Tuyệt đối không dùng thẻ `CHỮ:` trong các khối cảnh (mọi nội dung truyền tải qua lời đọc và hình vẽ).
4. **Quy tắc khớp nhịp 1:1 giữa LỜI và HÌNH**: Trong mỗi khối cảnh, cứ mỗi 1-2 câu thoại LỜI (10-20 từ, tương đương 3-6 giây) BẮT BUỘC có NGAY một thẻ `HÌNH:` tương ứng mô tả cử chỉ, hoạt cảnh hoặc góc máy của nét vẽ que. Cảnh 40-60 giây gồm 3-8 cặp `LỜI:` và `HÌNH:` xen kẽ liên tiếp. Tuyệt đối không gộp một tràng LỜI dài lê thê mà chỉ có 1 thẻ HÌNH chung chung (hình chết/tĩnh).
5. Chân kịch bản (sau khối cuối, mỗi thẻ một dòng): `CAPTION:` (bắt buộc), `HASHTAG:` (bắt buộc, các thẻ bắt đầu bằng `#` phân tách bằng khoảng trắng), `NGUỒN:` (bắt buộc; có thể nhiều dòng `NGUỒN:`; với chủ đề thường trực không dựa tin tức ghi `NGUỒN: không có (kiến thức nền)`), `CẦN KIỂM CHỨNG:` (tùy chọn; liệt kê dữ kiện chưa chắc, hoặc `không có`).
6. **Chỉ nội dung các thẻ `LỜI:` được tính vào số từ và thời lượng đọc.** Mọi thẻ khác không tính.
7. Toàn bộ chữ hiển thị (`LỜI`, `CAPTION`, `HASHTAG`...) phải 100% tiếng Việt, không chữ Hán. Tên riêng/thuật ngữ quốc tế quen thuộc (AI, iPhone, ETF...) được giữ.
8. Mốc thời gian dạng `m:ss` hoặc `mm:ss`. Các khối nối tiếp nhau theo thứ tự thời gian; tổng thời lượng mục tiêu từ 5 phút trở lên (khoảng 300–600 giây, 700–1500 từ LỜI).

## 3. Validator `script_format.go` (chỉ trả sự thật, không chặn commit)

Chạy trong `commit_chapter` cùng chỗ với `rules.Lint`; kết quả nằm trong `rule_violations`, mức `warning` (trừ khi ghi chú khác). `Rule` là mã dưới đây:

| Rule | Điều kiện |
|---|---|
| `script_no_title` | dòng đầu khác rỗng không bắt đầu bằng `# ` |
| `script_no_hook` | không có khối `HOOK`, hoặc khối `HOOK` không phải khối đầu tiên |
| `script_no_closing` | không có khối `CHỐT` |
| `script_words_out_of_range` | tổng từ `LỜI:` ngoài 700–1500 (`Actual` = số từ, `Limit` = "700-1500") |
| `script_duration_out_of_range` | thời lượng ước tính = max(thời điểm kết thúc lớn nhất khai báo, `SpeechSeconds`(từ LỜI)) ngoài 300–600 s |
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

# Mẫu danh sách tập (outline)

Mỗi phần tử outline là một tập = một video 60-180 giây. Khóa JSON không đổi, nhưng ý nghĩa như sau:

- `chapter`: số thứ tự tập.
- `title`: tiêu đề video (sẽ là dòng `# ...` đầu kịch bản).
- `core_event`: ý chính cần giải thích + góc nhìn của tập, một hai câu. Nếu gắn với trend, thêm cuối chuỗi: `Trend: <tên trend> | Nguồn: <url>`.
- `hook`: câu/hình hook 3 giây đầu (8-10 từ + hình).
- `scenes`: 3-5 chuỗi, mỗi chuỗi một dòng mô tả cảnh theo mạch: ẩn dụ đồ đá -> khái niệm hiện đại -> "hóa ra...".

## Bước 1: Chọn quy mô series

- Mùa ngắn: 8-25 video, một nhóm chủ đề. Dùng outline phẳng `outline`.
- Vừa: 25-60 video, nhiều đợt chủ đề. Outline phẳng hoặc phân tầng nhẹ.
- Dài: 80+ video, nhiều mùa/đợt. Dùng `layered_outline`: quyển = mùa/nhóm lớn, hồi = đợt chủ đề 5-10 video.

## Bước 2: Khi nào dùng phân tầng

Chọn `layered_outline` khi thỏa ít nhất 2 điều: chủ đề trải nhiều lĩnh vực (kinh tế, công nghệ, xã hội...); cần các đợt chủ đề với nhịp khác nhau; trend thay đổi nên chỉ lên kế hoạch chi tiết đợt gần; running gag và nhân vật phát triển qua nhiều đợt.

## Bước 3: Đừng lên danh sách chi tiết cả mùa dài

Thứ tự: định vị kênh và điểm khác biệt -> câu hỏi cốt lõi của series -> đợt chủ đề -> từng tập (core_event, hook). Sai lầm: 20 tập cùng khuôn "định nghĩa -> ví dụ -> chốt"; mọi hook cùng một kiểu; dùng hết ẩn dụ hay nhất trong 5 tập đầu.

## Mẫu outline phẳng

Mỗi tập được điền nhanh theo khung 6 nhịp, ánh xạ vào JSON như sau:
- `core_event`: Câu hỏi của tập + Các khẳng định cần nguồn.
- `hook`: Hook (khoảnh khắc người xem).
- `scenes`: Mảng chứa các bước: Cơ chế 1, Cơ chế 2, Hành động nhỏ, và Teaser/Chốt.

```json
[
  {
    "chapter": 1,
    "title": "Lạm phát: vì sao rìu đá của bạn bốc hơi",
    "core_event": "Câu hỏi: Vì sao lương tăng mà vẫn nghèo? Nguồn: kiến thức kinh tế nền tảng.",
    "hook": "Tưởng lương tăng là giàu, hóa ra bạn nghèo đi mỗi ngày (Hình: cầm cục đá to nhưng cắn không vỡ).",
    "scenes": [
      "Cơ chế 1 / ẩn dụ: Bộ lạc phát thêm vỏ sò cho mọi người, ai cũng ôm vỏ sò đi mua khoai",
      "Cơ chế 2 / bước ngoặt: Khoai không tăng, bà bán khoai thấy ai cũng nhiều vỏ sò nên tăng giá",
      "Hành động nhỏ: Nhìn lại giỏ vỏ sò của bạn xem có đang mất giá không",
      "Chốt / teaser: Lần tới thấy khoai đắt, đừng trách bà bán khoai. Tập sau: ai là người rải vỏ sò?"
    ]
  }
]
```

## Mẫu outline phân tầng (series dài)

Cuốn chiếu hai tầng: 2 quyển đầu có khung hồi, các quyển sau là quyển khung xương; hồi đầu tiên có các tập chi tiết.

```json
[
  {
    "index": 1,
    "title": "Mùa 1: Tiền bạc kỷ đá",
    "theme": "Giải thích kinh tế cá nhân bằng vỏ sò, mammoth và bộ lạc",
    "arcs": [
      {
        "index": 1,
        "title": "Đợt 1: Tiền đi đâu mất (đã chi tiết)",
        "goal": "Người xem hiểu lạm phát, lãi suất, nợ qua 6 tập",
        "chapters": [
          {"chapter": 1, "title": "Tiêu đề tập", "core_event": "Ý chính + góc nhìn", "hook": "Hook 3 giây", "scenes": ["Cảnh 1", "Cảnh 2", "Cảnh 3"]}
        ]
      },
      {
        "index": 2,
        "title": "Đợt 2: Vay mượn kỷ đá (khung xương)",
        "goal": "Tín dụng, thẻ, nợ xấu",
        "estimated_chapters": 8,
        "chapters": []
      }
    ]
  },
  {
    "index": 2,
    "title": "Mùa 2 (khung xương)",
    "theme": "Công nghệ và AI",
    "estimated_chapters": 30,
    "arcs": []
  }
]
```

Mở rộng: khi viết tới hồi khung xương, Architect mở rộng các tập chi tiết; tới quyển khung xương thì mở rộng cấu trúc hồi + tập của hồi đầu.

## Kiểm tra cấp đợt chủ đề

- Đợt này đào sâu câu hỏi nào của series? Người xem học được gì sau cả đợt?
- Các tập có góc nhìn khác nhau (không chỉ đổi tên khái niệm)?
- Đợt có running gag/callback nào nối giữa các tập?

## Kiểm tra cấp tập

- Một tập một ý chính; hook không trùng 3 tập gần nhất; ẩn dụ không trùng.
- Có cảnh bản lề ẩn dụ -> khái niệm và một cú "hóa ra...".
- Trend phải có nguồn; chủ đề thường trực ghi rõ là kiến thức nền.
- Số cảnh (3-5) khớp độ dài mục tiêu: 60 giây ~3 cảnh, 120-180 giây ~5 cảnh.

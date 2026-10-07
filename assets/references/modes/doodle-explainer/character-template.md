# Hồ sơ nhân vật que

Mỗi nhân vật que khi lưu bằng `save_foundation(type="characters")` có kiểu trường nghiêm ngặt như sau (không đổi thành object):

- `name`: string
- `aliases`: string[] (nếu không có thì bỏ qua)
- `role`: string
- `description`: string (mô tả tổng thể: ngoại hình nét que, dấu hiệu nhận diện, giọng nói)
- `arc`: string (một chuỗi mô tả toàn bộ vòng chuyển biến, không phải `{start/middle/end}`; dùng cách diễn đạt "giai đoạn đầu... giai đoạn sau...")
- `traits`: string[] (mảng chuỗi tính cách, ví dụ `["Cả tin", "Hay hoảng"]`)

Trường tùy chọn: `tier` (core / important / secondary / decorative).

> [!IMPORTANT]
> Tên "Que Ú" trong các ví dụ dưới đây chỉ là **VÍ DỤ MINH HỌA**. Khi thiết kế nhân vật cho một series, bạn PHẢI tự do sáng tạo tên gọi, ngoại hình và nét tính cách phù hợp nhất với chủ đề của tác phẩm (ví dụ: Que Rối / Que Lo cho chủ đề tâm lý, Que Mót / Que Sộp cho chủ đề tài chính, Que Mò cho chủ đề công nghệ...). TUYỆT ĐỐI KHÔNG mặc định gán tên "Que Ú" cho mọi kịch bản.

## Mẫu Markdown để nghĩ (không phải định dạng lưu)

### Host / người dẫn chuyện: [Tên]
- Biệt danh: (cách người xem gọi, ví dụ "Que Ú", "ông Que")
- Ngoại hình nét que: (một dấu hiệu nhận diện, ví dụ mũ lông, râu xoắn)
- Giọng nói: (nhịp, câu cửa miệng, mức tiếng lóng)
- Vai trong video: (bản lề sang khái niệm hiện đại, hỏi hộ người xem...)
- Tính cách cốt lõi: (2-4 nét)
- Điểm yếu gây cười: (để tạo tình huống)
- Vòng chuyển biến theo series: (giai đoạn đầu... giai đoạn sau..., chuyển biến nhỏ)
- Running gag gắn với nhân vật:

### Nhân vật phụ: [Tên]
- Vai: (đối trọng, kẻ bán hàng, thủ lĩnh khoe mẽ...)
- Ngoại hình + giọng nói (rút gọn)
- Dùng để làm gì trong tập: (gây xung đột, punchline, minh họa phản ví dụ)

### Linh vật / nhân vật nền: [Tên]
- Chỉ xuất hiện chen punchline hoặc callback; không thoại dài.

### Chuyên gia gây cười / Cây hài đồng cảm: [Tên]
- Vai: (cây hài tạo tiếng cười đồng cảm, soi chiếu tính cách đời thường; có thể là vai độc lập hoặc do Host/Nhân vật phụ kiêm nhiệm)
- Điểm yếu/nghịch lý tính cách: (hung dữ nhưng ấm áp, khôn ngoan dễ hớ, nhân từ dễ thiệt, lười biếng mà thật thà...)
- Tình huống quen thuộc hay gặp: (cháy deadline, sĩ diện với hàng xóm, tiếc của, bạn bè trêu nhau...)

## Ví dụ rút gọn

```json
[
  {
    "name": "Que Ú",
    "aliases": ["Ú", "ông Que"],
    "role": "host, người dẫn chuyện và đại diện người xem",
    "description": "Que nhỏ đội mũ lông, râu xoắn; nói nhanh, hay hoảng, câu cửa miệng: Trời ơi đất hỡi.",
    "arc": "Giai đoạn đầu cả tin, mất rìu vì hiểu lầm kinh tế; giai đoạn sau biết hỏi ngược: ai được lợi? nhưng vẫn mất rìu ở cuối mỗi tập.",
    "traits": ["Cả tin", "Hay hoảng", "Tò mò"]
  }
]
```

## Nguyên tắc

- Dàn nhân vật nhỏ: 1 host + 2-4 que phụ + 0-1 linh vật; mỗi que một chức năng rõ.
- Mỗi que nhận ra được chỉ qua hình dáng và giọng, không cần nhãn.
- Tránh nhân vật trùng chức năng; tránh thêm que mới chỉ để giải thích thêm.

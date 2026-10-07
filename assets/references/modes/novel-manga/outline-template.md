# Mẫu quy hoạch dàn ý (Outline Template)

Tác dụng của mẫu này không phải là ép mọi tác phẩm vào một độ dài cố định, mà là giúp nhận định cấp độ của tác phẩm trước, rồi sau đó mới lựa chọn độ chi tiết của dàn ý.

## Bước 1: Nhận định cấp độ độ dài tác phẩm

### Truyện ngắn / Câu chuyện một quyển
- Áp dụng: Đơn xung đột, đơn mục tiêu, ít nhân vật, kết cục tập trung.
- Độ dài tham khảo: 8 - 25 chương.
- Định dạng khuyến nghị: Dàn ý phẳng `outline`.

### Truyện vừa / Câu chuyện nhiều giai đoạn
- Áp dụng: Có nâng cấp theo giai đoạn, vài tuyến phụ, quan hệ nhân vật có biến chuyển.
- Độ dài tham khảo: 25 - 60 chương.
- Định dạng khuyến nghị: Dàn ý phẳng `outline` hoặc phân tầng nhẹ.

### Truyện dài kỳ / Mô hình tiểu thuyết mạng
- Áp dụng: Thể loại bẩm sinh có không gian nâng cấp liên tục, lực căng quan hệ lâu dài, mục tiêu nhiều giai đoạn, thế giới có thể mở rộng, bí ẩn dài hạn hoặc tuyến trưởng thành trường kỳ.
- Độ dài tham khảo: 80 - 200+ chương.
- Định dạng khuyến nghị: Dàn ý phân tầng `layered_outline`.

## Bước 2: Phán đoán có bắt buộc dùng dàn ý phân tầng hay không

Chỉ cần thỏa mãn bất kỳ 2 điều nào dưới đây thì ưu tiên dùng `layered_outline`:

- Thế giới quan cần được mở rộng dần dần, không phải kể hết trong một lần.
- Sự trưởng thành của nhân vật chính không phải là một bước nhảy vọt, mà nâng cấp qua nhiều giai đoạn.
- Mối quan hệ giữa các nhân vật liên tục biến chuyển qua nhiều giai đoạn.
- Giai đoạn giữa và sau tồn tại các loại mâu thuẫn chính khác nhau.
- Cần nhiều lần chuyển đổi bản đồ / thế lực / thân phận / mục tiêu.
- Thể loại mang đậm phong cách tiểu thuyết thương mại dài tập, không phải truyện một quyển.

## Bước 3: Đối với truyện dài, đừng lập tức làm "sổ nhật ký chương toàn sách"

Trình tự quy hoạch truyện dài khuyến nghị là:

1. Điểm bán tác phẩm và sự khác biệt hóa.
2. Động cơ câu chuyện dài hạn.
3. Chủ đề và sự thăng cấp ở cấp quyển.
4. Mục tiêu và bước ngoặt giai đoạn ở cấp hồi.
5. Sự kiện và móc câu ở cấp chương.

Cách làm sai lầm:
- Viết trước đại cương 20 chương rồi gượng ép kéo dài ra.
- Mỗi quyển đều lặp lại mô thức "gặp địch - mạnh lên - đổi bản đồ".
- Chỉ có nâng cấp tuyến chính mà không có nâng cấp quan hệ nhân vật.
- Giai đoạn đầu tiêu xài hết toàn bộ bí mật lớn, giai đoạn giữa và sau chỉ có thể lặp lại bài cũ.

## Mẫu dàn ý phẳng (Truyện ngắn / vừa)

```json
[
  {
    "chapter": 1,
    "title": "Tiêu đề chương",
    "core_event": "Sự kiện cốt lõi chương này",
    "hook": "Móc câu cuối chương",
    "scenes": ["Bối cảnh 1", "Bối cảnh 2", "Bối cảnh 3"]
  }
]
```

## Mẫu dàn ý phân tầng (Truyện dài - Cuốn chiếu hai tầng Quyển & Hồi)

Quy hoạch ban đầu áp dụng cơ chế cuốn chiếu hai tầng: 2 quyển đầu có khung xương các hồi, các quyển còn lại là quyển khung xương; hồi đầu tiên có các chương chi tiết.

```json
[
  {
    "index": 1,
    "title": "Tiêu đề Quyển 1",
    "theme": "Mâu thuẫn cốt lõi / chủ đề mới của quyển này",
    "arcs": [
      {
        "index": 1,
        "title": "Hồi 1 (Đã mở rộng chi tiết)",
        "goal": "Mục tiêu cục bộ, lực cản và bước ngoặt",
        "chapters": [
          {"chapter": 1, "title": "Tiêu đề chương", "core_event": "Sự kiện cốt lõi", "hook": "Móc câu cuối chương", "scenes": ["Bối cảnh 1", "Bối cảnh 2"]}
        ]
      },
      {
        "index": 2,
        "title": "Hồi 2 (Hồi khung xương)",
        "goal": "Khái quát mục tiêu của hồi này",
        "estimated_chapters": 12,
        "chapters": []
      }
    ]
  },
  {
    "index": 2,
    "title": "Tiêu đề Quyển 2",
    "theme": "Chủ đề Quyển 2",
    "arcs": [
      {"index": 1, "title": "Tiêu đề Hồi", "goal": "Mục tiêu Hồi", "estimated_chapters": 15, "chapters": []},
      {"index": 2, "title": "Tiêu đề Hồi", "goal": "Mục tiêu Hồi", "estimated_chapters": 10, "chapters": []}
    ]
  },
  {
    "index": 3,
    "title": "Tiêu đề Quyển 3 (Quyển khung xương)",
    "theme": "Phương hướng chủ đề Quyển 3",
    "estimated_chapters": 60,
    "arcs": []
  }
]
```

- Mở rộng cấp hồi: Khi việc viết tiến đến hồi khung xương, Architect sẽ mở rộng các chương chi tiết của hồi đó.
- Mở rộng cấp quyển: Khi việc viết tiến đến quyển khung xương, Architect sẽ mở rộng cấu trúc hồi của quyển + các chương của hồi đầu tiên.

## Danh sách kiểm tra cấp quyển truyện dài

Mỗi một quyển đều cần trả lời:
- Quyển này bổ sung thêm thông tin thế giới mới nào?
- Quyển này nâng cấp mâu thuẫn cốt lõi nào?
- Quyển này giúp nhân vật chính đạt được gì, và mất đi điều gì?
- Quyển này thay đổi mối quan hệ nhân vật chính như thế nào?
- Sau khi quyển này kết thúc, vì sao câu chuyện bắt buộc phải bước sang quyển tiếp theo?

## Danh sách kiểm tra cấp hồi truyện dài

Mỗi một hồi đều cần trả lời:
- Mục tiêu rõ ràng của hồi này là gì?
- Lực cản đến từ ai, quy tắc nào, cái giá phải trả là gì?
- Điểm bước ngoặt là gì?
- Sau khi hồi này kết thúc, những trạng thái nào đã biến đổi không thể đảo ngược?

## Danh sách kiểm tra cấp chương

- Mỗi chương bắt buộc phải phục vụ cho mục tiêu của hồi chứa nó.
- Mỗi chương bắt buộc phải chứa một sự kiện thúc đẩy không thể xóa bỏ.
- Móc câu phải đa dạng hóa, không dựa dẫm hoàn toàn vào kiểu "phát hiện bí mật".
- Các chương giai đoạn đầu không thể chỉ "giới thiệu thế giới", mà phải đồng bộ thúc đẩy nhân vật và xung đột.

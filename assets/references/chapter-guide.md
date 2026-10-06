# Hướng dẫn viết một kịch bản video

Một chương = một video doodle explainer 60-180 giây, ~150-450 từ lời đọc. Người xem TikTok quyết định ở và sau 3 giây đầu, nên mọi câu đều phải trả giá bằng giây.

## Thời lượng và từ

- Tốc độ đọc chuẩn ~2,5 từ/giây (~150 từ/phút). Đọc nhanh kiểu TikTok tối đa ~3 từ/giây.
- Bảng quy đổi: 60 giây ~150 từ; 90 giây ~225 từ; 120 giây ~300 từ; 180 giây ~450 từ.
- Chỉ thẻ `LỜI:` được tính từ. Ngoài ra khoảng lặng, hiệu ứng và cảnh câm cần tính thêm giây: mỗi cảnh câm 1-2 giây.
- Mặc định nhắm 60-90 giây. Chủ đề phức tạp mới lên 120-180 giây, và phải có đủ ý để nuôi từng ấy giây.

## Nhịp 3 giây

Cứ ~3 giây (7-9 từ) phải có một thay đổi: hình mới, ý mới, câu hỏi mới, hoặc cú đảo. Một cảnh dài 15-25 giây gồm 5-8 nhịp như vậy. Nếu hai nhịp liền kề giống hệt nhau về hình và ý, gộp hoặc cắt một nhịp.

## Cấu trúc HOOK / CẢNH / CHỐT

Hook (0:00-0:03): 8-10 từ, gây tò mò hoặc sốc hình ảnh. Không chào hỏi, không giới thiệu kênh. Xem `hook_techniques`.

Cảnh 1 (3-5 cảnh tùy độ dài): dựng vấn đề bằng ẩn dụ thời đồ đá. Đưa bối cảnh bằng hành động, không giảng.

Cảnh giữa: nối ẩn dụ sang khái niệm hiện đại (câu bản lề: "Giờ thay vỏ sò bằng tiền, thay mammoth bằng nhà."). Mỗi cảnh chỉ một ý.

Cảnh cuối trước chốt: cú twist, "hóa ra...", hoặc hệ quả bất ngờ.

Chốt (cuối): một câu mang được về nhà, gọi lại hook (loop) hoặc đẩy người xem vào bình luận/theo dõi. Chốt 5-12 giây, không kết bằng "Hy vọng các bạn hiểu".

## Lời đọc và hình: chia việc

- LỜI nói cái hình không cho thấy: lý do, con số, nhận xét. HÌNH cho thấy cái lời khỏi phải tả: ai đứng đâu, làm gì, biểu cảm.
- Không để LỜI mô tả hình ("Ở đây có một que đang chạy"). Để hình diễn, lời bình.
- HÌNH phải vẽ được bằng nét que: tối đa 2-3 nhân vật, 1-2 đạo cụ mỗi cảnh, một hành động chính. Tránh "đám đông chi tiết", "bản đồ phức tạp".
- CHỮ: tối đa 6-8 từ, chỉ cho con số hoặc từ khóa. Không lặp nguyên văn lời đọc.
- ÂM: chỉ ghi khi hiệu ứng làm nổi punchline.

## Mỗi cảnh làm đúng một việc

Mỗi cảnh trả lời một câu hỏi người xem đang có trong đầu. Liệt kê: cảnh này trả lời gì? Nếu không nói được trong một câu, cảnh đang nhồi.

## Cân độ phức tạp

- Một video chỉ giải thích một ý chính (core_event). Ý phụ, ngoại lệ, dữ liệu thừa dồn sang tập khác hoặc bỏ.
- Tối đa 1 con số mạnh mỗi 20 giây, làm tròn cho dễ nghe ("gần gấp đôi", không phải "1,87 lần").
- Thuật ngữ lạ phải được dịch ngay bằng ẩn dụ ở lần đầu xuất hiện.

## Lỗi thường gặp

- Mở bằng định nghĩa sách giáo khoa: "Lạm phát là hiện tượng...". Phải mở bằng hình/câu hỏi.
- Ba ẩn dụ khác nhau trong một video: chỉ giữ một ẩn dụ xuyên suốt.
- Hook hứa một đằng, thân video nói một nẻo.
- Chốt giảng đạo hoặc tóm tắt lại cả video.
- Nói dữ kiện không nguồn: ghi vào `NGUỒN:` hoặc `CẦN KIỂM CHỨNG:`.

## Tự kiểm tra trước khi commit

1. Đếm từ LỜI nằm trong 150-450; thời lượng khai báo khớp với từ/2,5.
2. Khối đầu là HOOK, khối cuối là CHỐT, đủ LỜI và HÌNH ở mọi khối.
3. Có CAPTION, HASHTAG, NGUỒN. Tiêu đề dòng đầu trùng title.
4. Đọc to thử: chỗ nào hụt hơi hoặc líu lưỡi thì cắt.

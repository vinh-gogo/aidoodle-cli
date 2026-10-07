# Checklist chất lượng kịch bản

Tự kiểm tra trước khi `commit_chapter`; Editor dùng chính danh sách này để chấm. Mục nào không đạt, ghi rõ khối nào và trích nguyên văn.

## 1. Hook (khối HOOK)

- [ ] Lời ≤10 từ, vào thẳng, không chào hỏi.
- [ ] Mở bằng khoảnh khắc của người xem (ví dụ: "mở 5 phút thành 1 tiếng"), KHÔNG dùng tên nhân vật lạ (như Que Ú) ở Hook.
- [ ] Gây tò mò với người chưa biết chủ đề; hình hook đọc được khi tắt tiếng.
- [ ] Không lặp kiểu hook của 3 tập gần nhất.
- [ ] Thân video giao đúng điều hook hứa.

## 2. Nhịp và cơ chế (Tại sao?)

- [ ] Cứ ~3 giây có một thay đổi (hình, ý, câu hỏi, đảo).
- [ ] Tốc độ đọc khoảng ~2,5 đến 3 từ/giây (không chậm dưới 2 từ/giây).
- [ ] Không lặp một ý nhiều lần mà không giải thích CƠ CHẾ cốt lõi (phải trả lời được "tại sao").
- [ ] Có ít nhất một cú đảo/"hóa ra..." trước chốt.

## 3. Ẩn dụ đúng

- [ ] Một ẩn dụ đồ đá xuyên suốt, ánh xạ đúng khái niệm (cái gì = cái gì nói rõ ở cảnh bản lề).
- [ ] Ẩn dụ không làm sai bản chất (nếu đơn giản hóa mạnh thì ghi vào CẦN KIỂM CHỨNG).
- [ ] Tuân thủ luật vũ trụ doodle: ẩn dụ được/không được dùng, anachronism có chủ đích.
- [ ] Không trùng ẩn dụ độc quyền của tập gần nhất.

## 4. Dữ kiện và nguồn

- [ ] Mọi con số, tên, ngày tháng có trong `NGUỒN:` hoặc kiến thức nền chắc chắn.
- [ ] Các mệnh đề khoa học, phân tích tâm lý, dữ kiện chưa chắc chắn PHẢI nằm ở `CẦN KIỂM CHỨNG:` và lời đọc dùng giọng dè dặt.
- [ ] Không bịa số liệu, không gán lời cho người thật; tập về trend ghi `Trend:` và nguồn.
- [ ] Chủ đề nhạy cảm không khuyên đầu tư/chữa bệnh cụ thể.

## 5. Lời đọc nói được và Ngôi kể

- [ ] Ngôi kể (POV) nhất quán, không lẫn lộn (ví dụ: đang xưng "mình" thì không chốt bằng "Que Ú bị dính").
- [ ] Đọc to không líu lưỡi, không câu dài hơn ~20 từ, ngắt hơi tự nhiên.
- [ ] Ngôn ngữ nói, không văn viết, không từ Hán-Việt nặng khi có từ thuần.
- [ ] Không liệt kê 3 vế đều đều, không câu sáo AI (xem `anti_ai_tone`).
- [ ] Thoại giữa các que ngắn, mỗi que một giọng riêng.

## 6. Hình vẽ được

- [ ] Hình ảnh không đơn điệu (không phải cảnh nào cũng chỉ cầm điện thoại). Mỗi cảnh một hình chính, xen kẽ cận cảnh, sơ đồ.
- [ ] HÌNH không nhồi nhét quá tải đạo cụ (ví dụ: vừa keo dính, vừa ong mật phát sáng khó vẽ).
- [ ] Mỗi HÌNH: ≤3 nhân vật, ≤2 đạo cụ, một hành động chính, mô tả đủ để họa sĩ vẽ.
- [ ] HÌNH không lặp lại nguyên lời đọc; cho thấy cái lời không nói.
- [ ] Không dùng thẻ CHỮ:; nhãn từ khóa/con số (nếu có) được lồng trực tiếp trong HÌNH:.

## 7. Chốt và Teaser

- [ ] Có câu mang về nhà; gọi lại hook hoặc callback.
- [ ] Không chốt bằng câu khẳng định tuyệt đối vô căn cứ, KHÔNG đổ lỗi làm người xem thấy bất lực ("bạn không kiểm soát được mình").
- [ ] Cảnh setup (ví dụ đi săn) phải có hậu quả/kết quả ở cuối để tạo kịch tính/gag.
- [ ] Phải có teaser dẫn sang tập sau, hành động nhỏ không xin tương tác cứng.

## 8. Chân kịch bản và ngôn ngữ

- [ ] Có CAPTION (1-2 câu, có từ khóa), HASHTAG (3-6 thẻ bắt đầu bằng #), NGUỒN.
- [ ] 100% tiếng Việt có dấu đầy đủ (cho phép AI, iPhone, ETF...); không chữ Hán, không Markdown.
- [ ] Nhân vật que nhất quán với hồ sơ (xem `consistency`).

## Chấm điểm

Mỗi mục 1-8: đạt (1), đạt một phần (0,5), không (0). Tổng ≥6,5 mới bàn giao; hook <0,5 hoặc dữ kiện <0,5 là lỗi chặn dù tổng cao.

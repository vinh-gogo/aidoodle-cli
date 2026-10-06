## Mẫu cấu trúc video và kế hoạch đợt/mùa cho Doodle Explainer

Một video là một "chương" 60–180 giây (150–450 từ lời đọc), gồm HOOK, 3–5 CẢNH, CHỐT theo `docs/script-format.md`. Các mẫu dưới đây là gợi ý về nhịp; chọn một mẫu cho mỗi tập và đổi mẫu giữa các tập liền kề để khỏi lặp.

### Mẫu 1. Hook → Vấn đề → Giải thích → Twist → Chốt (mẫu mặc định, 90–150 giây)

- HOOK (0–3 giây): một câu hỏi lạ hoặc một hình vô lý (ông Gậy cầm điện thoại bằng đá, khóc vì "hết pin lửa").
- CẢNH 1, Vấn đề: dựng tình huống đồ đá cụ thể cho thấy vì sao chuyện này làm người ta khó chịu hoặc thắc mắc.
- CẢNH 2, Giải thích: nối ẩn dụ sang khái niệm hiện đại, đúng cơ chế, dữ kiện từ nguồn.
- CẢNH 3, Twist: "Hóa ra..." phát hiện bất ngờ hoặc điểm người ta hay hiểu sai.
- CHỐT: một câu gọn hoặc một hình đọng lại; có thể kèm lời mời bình luận ngắn.

### Mẫu 2. Kết thúc vòng lặp (loop-ending, 60–100 giây)

- Câu cuối của CHỐT nối ngược về câu HOOK, để video xem lại mượt (hook hỏi "ai lấy mất vỏ ốc của tôi?", chốt: "…và đó là lúc ông Gậy hiểu ai lấy mất vỏ ốc của mình").
- Giữ ý chính duy nhất; chỉ 2–3 CẢNH. Chữ trên màn hình ở HOOK và CHỐT nên khớp nhau để vòng lặp trọn.
- Dùng khi chủ đề đơn giản, lời đọc gần mức tối thiểu.

### Mẫu 3. Phá lầm tưởng (myth-busting, 100–160 giây)

- HOOK: nêu một lầm tưởng phổ biến như thể ai cũng tin ("Ai cũng bảo ăn mammoth nhiều thì mạnh").
- CẢNH 1: nhân vật tin lầm tưởng và làm theo, cười ra nước mắt.
- CẢNH 2: dữ kiện thật từ nguồn, nêu gọn, không số liệu ngoài nguồn.
- CẢNH 3: vì sao lầm tưởng lan nhanh (ẩn dụ đồ đá về tin đồn truyền miệng).
- CHỐT: một câu cách kiểm chứng nhanh. Phần dữ kiện chưa chắc ghi vào CẦN KIỂM CHỨNG.

### Mẫu 4. Tin nóng giải thích (120–180 giây)

- HOOK: nêu đúng sự việc đang được bàn, một câu, nói trung tính.
- CẢNH 1: "Chuyện gì xảy ra" chỉ dùng dữ kiện từ nguồn cung cấp, ghi rõ nguồn trong chân kịch bản.
- CẢNH 2: "Vì sao chuyện này đáng quan tâm" bằng ẩn dụ đồ đá.
- CẢNH 3: "Điều chưa biết" (nói thẳng là chưa rõ); không suy diễn thành tin.
- CHỐT: một điểm người xem nên nhớ, không đưa lời khuyên tài chính, y tế, pháp lý. Kiểm tra `fact_grounding` trước khi đưa dữ kiện.

### Mẫu 5. So sánh đồ đá vs hiện đại (90–150 giây)

- Nhịp đối xứng: cảnh đồ đá, cảnh hiện đại, rồi điểm chung.
- HOOK: "Người đồ đá làm chuyện này thế nào? Còn mình làm thế nào bây giờ?" Mốc thời gian hay con số (nếu có) phải lấy từ nguồn.
- CẢNH 1: cách đồ đá (đơn giản, hài). CẢNH 2: cách hiện đại (đúng dữ kiện). CẢNH 3: cái gì không đổi.
- CHỐT: "Hóa ra..." về điểm chung, không phán xét ai hơn ai.

## Đợt chủ đề (batch) 5–10 video

Một đợt gom các video cùng một nhóm chủ đề (ví dụ "Tiền và lạm phát", "Điện thoại và sự chú ý").

- Khung đợt: 1 tập nền tảng (khái niệm cốt lõi), 3–6 tập mở rộng (ca cụ thể, hiểu lầm, tin nóng), 1 tập tổng kết hoặc hỏi đáp người xem.
- Phân bố mẫu: không quá 2 tập liền kề cùng mẫu; xen một tập loop-ending ngắn sau hai tập dài.
- Chủ đề trend chèn vào theo nguồn có sẵn; tập đó ghi Trend và Nguồn trong core_event.
- Mỗi tập có tiêu đề riêng rõ ý chính, hook riêng; không lặp cùng một ẩn dụ chính quá hai lần trong đợt.

## Kế hoạch mùa (short, 8–25 video)

- Chia mùa thành 2–4 đợt chủ đề. Tập 1 đặt luật vũ trụ doodle và giới thiệu host, nhân vật phụ, linh vật.
- Khoảng giữa mùa đặt một tập "đặc biệt" (dài hơn hoặc đổi nhịp) để giữ người xem.
- Tập cuối mùa gọi lại các running gag và đúc kết bằng hình ảnh, không giảng đạo.
- Chừa 20–30% chỗ trống cho tập trend phát sinh; không đặt cứng nội dung tin nóng từ trước.

## Running gag và callback

- Số lượng: 3–5 gag cho cả mùa. Ví dụ: linh vật luôn phản ứng sai lúc; câu cửa miệng của host; một đồ vật vô lý xuất hiện ở góc hình mỗi tập; ông hàng xóm hang đối diện luôn "bán" thứ gì đó.
- Mỗi gag có điều kiện xuất hiện và biến thể: lần 1 giới thiệu, lần 2 lặp, lần 3 đảo ngược, về sau chỉ gợi nhẹ.
- Callback chỉ gọi ngắn (một hình hoặc một câu); người xem lạc vào giữa mùa vẫn hiểu tập mà không cần biết gag.
- Gag không được cười vào nạn nhân thật của sự kiện trong tin. Nếu chủ đề nhạy cảm, tạm bỏ gag ở tập đó.

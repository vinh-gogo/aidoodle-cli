# Ngôn ngữ hình ảnh doodle (người que)

Tài liệu này giúp viết thẻ `HÌNH:` và `CHỮ:` sao cho một họa sĩ hoặc công cụ hoạt ảnh vẽ được ngay bằng nét đơn giản. Định dạng khối và thẻ theo `docs/script-format.md`; ở đây chỉ nói về nội dung hình.

## 1. Cái gì vẽ được bằng người que

- Người que: đầu tròn, thân và tay chân bằng nét thẳng; biểu cảm chỉ bằng mắt chấm, miệng cong hoặc thẳng, lông mày hai nét, vài giọt mồ hôi hoặc dấu chấm than.
- Đồ vật hình khối đơn giản: hòn đá, que củi, lửa trại, hang, vỏ ốc, xương, giỏ, mammoth tròn có ngà, phiến đá làm "điện thoại".
- Khái niệm trừu tượng thành biểu tượng: tiền = vỏ ốc, thời gian = bóng mặt trời, áp lực = tảng đá đè, ý tưởng = tia sáng trên đầu, tăng giá = mũi tên đi lên bằng que củi, đám đông = ba đến năm người que.
- Biểu đồ đơn giản: cột bằng đống đá chồng, đường xu hướng bằng vệt than. Chỉ dùng cột hoặc đường với con số có trong nguồn; không vẽ số liệu bịa.
- Phóng đại hài hước: nhân vật phình to khi giận, co nhỏ khi sợ, nổ tung khi ngạc nhiên.

## 2. Bố cục theo nhịp 3–5 giây

- Mỗi nhịp 3–5 giây chỉ một ý hình. Nếu một câu LỜI dài hơn 5 giây, tách thành hai hình nối tiếp trong cùng khối.
- Một khung chỉ 1–3 nhân vật chính; nền tối giản (một vạch đất, một hang, một cây).
- Đặt điểm nhìn chính ở giữa hoặc một phần ba; chừa vùng trên cùng và dưới cùng cho chữ và nút giao diện của ứng dụng.
- Giữ nhất quán: cùng nhân vật cùng phụ kiện qua mọi cảnh; mỗi tập chỉ một màu nhấn ngoài đen trắng.
- Khối HOOK cần hình bắt mắt ngay khung đầu (chuyển động, tương phản, chữ ngắn); không mở bằng khung tĩnh trống.

## 3. Chữ hiện trên màn hình (thẻ CHỮ:)

- Cực ngắn: tối đa 6–8 từ mỗi lần hiện, thường 2–4 từ; không chép nguyên lời đọc.
- Chỉ dùng cho từ khóa, con số có trong nguồn, nhãn dán lên hình (ví dụ "LẠM PHÁT"), hoặc cú hài.
- Tiếng Việt có dấu đầy đủ, không chữ Hán, không viết tắt khó hiểu. Thuật ngữ quốc tế quen thuộc được giữ (AI, ETF...).
- Mỗi con số hiện lên phải trùng nguồn trong chân kịch bản; chưa chắc thì không để lên màn hình.
- Chữ hiện không che nhân vật, không dài hơn thời gian đọc được (khoảng 1 giây cho 2–3 từ).

## 4. Từ vựng cảnh quay đơn giản

- Cận mặt: cho phản ứng hài (mắt chấm, miệng méo).
- Toàn cảnh hang hoặc bãi: cho bối cảnh và đám đông.
- Zoom vào đồ vật: cho chi tiết mấu chốt (vỏ ốc, phiến đá).
- Cắt nhanh hai khung: cho so sánh đồ đá và hiện đại.
- Chuyển cảnh bằng vệt than quét, lật phiến đá, hoặc nhân vật bước ra khỏi khung.
- Chạy chữ hoặc mũi tên: cho tăng giảm, nguyên nhân kết quả.
- Khung "vẽ dần": hình hiện từng nét như đang vẽ, hợp với lời giải thích từng bước.
- Chữ trong HÌNH: viết ngắn gọn, dạng động từ, ví dụ "ông Gậy ôm đống vỏ ốc, mặt méo xệch; phía sau cá nhỏ dần".

## 5. Không nên vẽ

- Người thật, mặt thật, ảnh chụp, logo nhãn hiệu thật, màn hình ứng dụng thật. Dùng người que hoặc hình ẩn dụ thay thế; nhãn hiệu quen thuộc chỉ ở dạng chữ trong lời đọc nếu cần và đúng nguồn.
- Hình máu me, bạo lực, thương tích, cảnh tự làm hại; chủ đề nhạy cảm thì diễn đạt bằng biểu tượng nhẹ (xem `fact_grounding`).
- Nhân vật que đại diện cho một nhóm người thật (dân tộc, tôn giáo, giới tính, nghề nghiệp) theo kiểu giễu cợt.
- Hình quá chi tiết cần nhiều nét: bản đồ thật, giao diện nhiều lớp, đám đông trên 6 người.
- Câu mô tả HÌNH mơ hồ ("hình đẹp về lạm phát"). Luôn nói rõ ai làm gì, ở đâu, với vật gì.
- Hình gợi ý lời khuyên đầu tư, thuốc men hay pháp lý cụ thể bằng cách chỉ định mã, liều lượng, thủ tục.

# Tiêu chuẩn khử văn phong AI (Anti-AI Tone) cho lời đọc ngắn

Tài liệu này là kho tiêu chuẩn "mùi AI" dùng chung cho writer và editor, áp dụng cho **lời đọc (thẻ `LỜI:`) của video doodle explainer**: ngắn, nói thành tiếng, vui, thuần Việt. Writer phải tránh các mô thức dưới đây; editor khi thẩm định chiều aesthetic kiểm tra từng mục theo tài liệu này và **bắt buộc trích dẫn nguyên văn** làm dẫn chứng. Thẻ `HÌNH:`, `ÂM:` cũng không được viết kiểu sáo rỗng.

> Phần có thể liệt kê bằng máy (câu sáo cố định như "trong thế giới ngày nay", "hãy cùng khám phá", "có thể nói rằng", "điều đáng chú ý là", "không thể phủ nhận rằng", "trong bối cảnh hiện nay"; từ đệm hay lặp như "thực sự", "vô cùng", "tuy nhiên", "như thể") đã được `working_memory.user_rules.structured` kiểm tra khi commit. Tài liệu này chuyên xử lý **phán đoán ngữ nghĩa không cơ học hóa được**. Hai tầng bổ trợ nhau: tầng cơ học bắt bề nổi, tài liệu này bắt chất lượng câu chữ.

Phép thử chung: đọc to lời đọc lên. Nếu phải lấy hơi giữa câu, nếu nghe như bài báo hay bài thuyết trình, nếu một người bạn không nói như vậy ở quán trà đá, thì sửa.

## 1. Mùi AI trong cấu trúc

- **Câu dài, nhiều mệnh đề**: lời đọc có câu trên khoảng 18 từ hoặc chứa hai ba dấu phẩy nối ý. Cách sửa: tách thành các câu 6–14 từ, mỗi câu một ý; ý phụ chuyển sang câu sau hoặc thể hiện bằng hình ảnh trong `HÌNH:`.
- **Điệp ba vế / bộ ba song hành**: ba câu hoặc ba cụm đối xứng để "tạo thế" ("không lo lắng, không sợ hãi, không bỏ cuộc"). Cách sửa: giữ vế đắt nhất, các vế còn lại thành hình hoặc chi tiết cụ thể.
- **Các câu đều tăm tắp**: câu nào cũng cùng độ dài và cú pháp, nghe như danh sách. Cách sửa: xen câu rất ngắn ("Mất sạch.") với câu vừa.
- **Mở bài, thân bài, kết bài kiểu bài văn**: HOOK mà "giới thiệu chủ đề", CHỐT mà "tóm tắt lại". Cách sửa: HOOK là câu hỏi hoặc tình huống ngay; CHỐT là một hình hoặc một câu gọn.
- **Dấu đánh dấu cấu trúc trong lời**: "Thứ nhất, thứ hai, thứ ba", "Tóm lại", "Như đã nói ở trên", hoặc Markdown (`**`, `##`, gạch đầu dòng) lọt vào chính văn. Cách sửa: đọc thành lời tự nhiên, nối bằng "rồi", "thế là", "còn", hoặc ngắt câu.

## 2. Mùi AI trong dùng từ

- **Hán-Việt nặng và văn "convert"**: "tiến hành", "thực hiện", "mang tính", "nhằm mục đích", "đối với việc", "sở hữu", "bản thân việc". Cách sửa: dùng từ nói hằng ngày ("làm", "có", "để", "tự nó").
- **Nhồi thành ngữ và từ bóng bẩy thay cho hình ảnh**: "thăng trầm", "ngàn cân treo sợi tóc", "đỉnh cao của sự", "bức tranh toàn cảnh". Cách sửa: thay bằng một vật hay hành động cụ thể (nhặt vỏ ốc, ôm đống củi).
- **Mẫu so sánh rập khuôn**: "giống như…", "tựa như…", "như thể…", "dường như…". Ở đây ẩn dụ đồ đá đã là phép so sánh; đừng thêm một lớp ví von nữa. Cách sửa: nói thẳng "Đó, lạm phát đó."
- **Từ đệm giảm xóc**: "thực sự", "vô cùng", "đặc biệt", "một cách", "không khỏi", "bỗng nhiên", "khẽ", "chợt". Cách sửa: xóa, để động từ tự đứng.
- **Mẫu câu tương phản gượng triết lý**: "đây không phải là X, mà là Y" dùng lặp lại. Dùng tối đa một lần mỗi tập và chỉ khi X, Y đều cụ thể. Cách sửa: nói thẳng Y.
- **Câu dẫn sáo**: "Trong thế giới ngày nay", "Hãy cùng khám phá", "Có thể nói rằng", "Điều đáng chú ý là", "Chúng ta hãy cùng tìm hiểu". Cách sửa: vào thẳng chuyện ("Ông Gậy hết vỏ ốc rồi.").
- **Từ trừu tượng to tát**: "tối ưu hóa", "hệ sinh thái", "giá trị cốt lõi", "xu hướng toàn cầu" xuất hiện mà không có ví dụ. Cách sửa: thay bằng một cảnh đồ đá hoặc một ví dụ đời thường.

## 3. Mùi AI trong miêu tả

- **Khái quát thay cho chi tiết cụ thể**: "bầu không khí rất căng thẳng", "tình hình vô cùng phức tạp". Cách sửa: một vật, một hành động nhìn thấy hoặc nghe thấy (cả hang im re, chỉ còn tiếng củi nổ).
- **Dán nhãn cảm xúc trực tiếp**: "ông ấy rất lo lắng", "mọi người đều vui mừng". Cách sửa: để thẻ `HÌNH:` thể hiện bằng điệu bộ (mồ hôi chảy, tay run) và để lời đọc nói việc.
- **Thẻ HÌNH mơ hồ**: "hình minh họa đẹp", "cảnh sinh động". Cách sửa: nêu ai làm gì, ở đâu, với vật gì.

## 4. Mùi AI trong đối thoại và giọng nhân vật

- **Nhân vật đồng nhất**: bỏ tên người nói thì không phân biệt được ai nói. Cách sửa: mỗi nhân vật que có một câu cửa miệng và độ dài câu riêng (host hay hỏi, kẻ hoài nghi nói cộc, bà thầy thuốc dài hơi hơn).
- **Giải thích động cơ quá mức**: nhân vật nói ra hết suy nghĩ, người kể lại bồi thêm "vì ông ấy muốn…". Cách sửa: để hành động và một câu bóng gió tự nói.
- **Văn sách vở**: ai cũng nói câu đầy đủ chủ vị, có từ nối logic. Cách sửa: lời nói thật có lửng câu, trả lời lạc đề, tiếng đệm ("ờ", "hả", "thôi rồi").

## 5. Mùi AI trong nhịp điệu và cảm xúc

- **Nhồi số liệu và thông tin vào một câu**: ba con số trong một hơi. Cách sửa: một con số một câu và chỉ giữ số có trong nguồn.
- **Kể cặn kẽ, không chừa chỗ nghĩ**: mọi nguyên nhân, hệ quả đều nói hết. Cách sửa: bỏ chi tiết phụ, để câu "Hóa ra..." làm điểm nhấn.
- **Gượng ép nâng tầm triết lý ở CHỐT**: câu cuối kiểu "bài học cuộc sống" hay lời hô hào. Cách sửa: dừng ở một hình ảnh hoặc câu gọn, hoặc nối về HOOK.
- **Hài giả tạo**: chêm "haha", cảm thán thừa, hoặc giải thích trò đùa. Cách sửa: trò đùa đứng bằng độ lệch của tình huống; không giải thích.

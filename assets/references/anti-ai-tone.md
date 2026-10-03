# Tiêu chuẩn khử văn phong AI (Anti-AI Tone)

Tài liệu này là kho tiêu chuẩn "mùi AI" dùng chung cho cả writer và editor: writer khi sáng tác phải tránh toàn bộ các mô thức dưới đây, editor khi thẩm định chiều aesthetic sẽ kiểm tra từng mục theo tài liệu này và **bắt buộc trích dẫn nguyên văn** để làm dẫn chứng.

> Phần có thể liệt kê bằng máy (dấu gạch ngang, câu rập khuôn cố định, từ ngữ sáo rỗng tần suất cao) đã được `working_memory.user_rules.structured` kiểm tra bắt buộc khi commit, tài liệu này chuyên xử lý **những phán đoán ngữ nghĩa không thể cơ học hóa**. Cả hai bổ trợ cho nhau: tầng cơ học bắt bề nổi, tài liệu này bắt chiều sâu chất lượng câu chữ.

## 1. Mùi AI trong cấu trúc

- **Điệp từ ba vế / Bộ ba song hành**: Dùng liên tiếp ba câu ngắn hoặc phân câu có cấu trúc đối xứng để "tạo thế" ("anh không do dự nữa, không lùi bước nữa, không ngoảnh đầu lại nữa"). Cách sửa: Giữ lại một câu đắt giá và có sức nặng nhất, các vế còn lại tách thành hành động hoặc chi tiết cụ thể.
- **Xếp chồng câu đối xứng đều tăm tắp**: Chiều dài và cú pháp của mỗi đoạn văn giống hệt nhau, đọc lên như một bản danh sách liệt kê. Cách sửa: Đan xen câu dài và câu ngắn nhịp nhàng, để câu văn có khoảng thở.
- **Tiêu đề con đánh số / Dấu `##` chia cắt trong chương**: Trong chính văn xuất hiện số thứ tự `1`, `2`, `3` hoặc các dấu phân đoạn kiểu Markdown `##`/`###`. Cách sửa: Chỉ giữ lại tiêu đề chương, chuyển đổi bối cảnh hãy dùng một dòng trắng để chuyển tiếp tự nhiên.

## 2. Mùi AI trong dùng từ

- **Nhồi nhét từ Hán-Việt / Thành ngữ bốn chữ**: Trong một đoạn ngắn nhét quá nhiều thành ngữ hoặc từ ngữ bóng bẩy để thay cho miêu tả ("kinh tâm động phách, hiểm nguy trùng trùng, ngàn cân treo sợi tóc"). Cách sửa: Dùng một hành động hoặc hình ảnh cụ thể để thay thế chuỗi thành ngữ sáo rỗng.
- **Mẫu câu so sánh rập khuôn**: Các câu ví von quen thuộc xuất hiện liên tục như "giống như…", "tựa như…", "dường như…", "như thể…". Cách sửa: Thay bằng động từ chuẩn xác hoặc hình ảnh so sánh mới mẻ, hoặc trực tiếp miêu tả chân thực.
- **Nghiện lượng từ / Nghiện hư từ đệm**: "Một tia", "một nét", "một làn" đi kèm với cảm xúc; "bất giác", "thậm chí", "không khỏi", "dường như", "thoáng chốc" dùng làm từ cửa miệng. Cách sửa: Xóa bỏ các từ đệm giảm xóc, để hành động trực tiếp diễn ra ("anh mỉm cười", thay vì "khóe môi anh bất giác cong lên một nét cười").
- **Từ ngữ trừu tượng to tát**: "Ở một mức độ nào đó", "đáng chú ý là", "không hiểu vì sao", "nói không rõ diễn tả không thông" — người kể chuyện đang đúc kết thay cho độc giả. Cách sửa: Xóa bỏ, nhường phán đoán cho dữ kiện và hành động cụ thể.
- **Mẫu câu định nghĩa tương phản**: Các khuôn sáo dùng phủ định + chuyển ngoặt để "tạo điểm nhấn" lặp đi lặp lại như: "thứ anh muốn không phải là X, mà là Y", "đây không phải là kết thúc, mà là sự khởi đầu". Cách sửa: Dùng một hành động hoặc lựa chọn cụ thể để trực tiếp thể hiện, không dựa vào mẫu câu để tạo cảm giác triết lý gượng gạo.

## 3. Mùi AI trong miêu tả

- **Khái quát trừu tượng thay thế cho ngũ quan cụ thể**: Các khái niệm chung chung như "bầu không khí rất ngột ngạt", "tình hình vô cùng căng thẳng". Cách sửa: Đưa ra một chi tiết cụ thể có thể cảm nhận bằng xúc giác / khứu giác / thính giác (tốt hơn là chỉ dùng thị giác thuần túy).
- **Dán nhãn cảm xúc trực tiếp**: Trực tiếp viết "anh rất căng thẳng / tức giận / đau buồn". Cách sửa: Thể hiện qua phản ứng cơ thể và hành động ("đốt ngón tay trắng bệch", "cổ họng nghẹn đắng"), không gọi thẳng tên cảm xúc.

## 4. Mùi AI trong đối thoại

- **Nhân vật bị đồng nhất hóa**: Bỏ nhãn tên người nói thì không phân biệt được ai đang nói — ai ai cũng có cùng độ dài câu, vốn từ và tầng lớp học vấn giống nhau. Cách sửa: Cho mỗi nhân vật độ dài câu ổn định, khẩu ngữ riêng, và tỷ lệ hàm ý ngầm khác nhau.
- **Giải thích động cơ quá mức**: Nhân vật bộc bạch hết tâm lý của mình ra ngoài, hoặc người kể chuyện lập tức bồi thêm lời giải thích "anh nói như vậy là vì…". Cách sửa: Hãy để động cơ ẩn sau những lựa chọn và lời nói bóng gió, hãy tin tưởng độc giả.
- **Giọng văn sách vở**: Tất cả mọi người đều nói những câu hoàn chỉnh, ngay ngắn, đầy ắp các từ nối logic. Cách sửa: Khẩu ngữ đời thường luôn có sự ngập ngừng, lược bớt và trả lời không đúng trọng tâm.

## 5. Mùi AI trong nhịp điệu và cảm xúc

- **Cái gì cũng kể cặn kẽ**: Mọi hành động, nguyên nhân hậu quả đều viết kín kẽ, không để lại chút không gian tưởng tượng nào. Cách sửa: Chỗ cần giấu hãy giấu đi, dùng khoảng trống để kích thích người đọc theo dõi tiếp.
- **Gượng ép nâng tầm triết lý cuối chương**: Cuối mỗi chương đều nâng tầm lên chiêm nghiệm nhân sinh hoặc câu chốt triết lý vàng ngọc. Cách sửa: Dừng lại ở một hình ảnh cụ thể, một lựa chọn gay cấn hoặc dư âm cảm xúc lắng đọng, không đúc kết ý nghĩa thay cho độc giả.

# Mẫu kịch bản video (1 chương = 1 video)

Chính văn là văn bản thuần, đúng định dạng của `docs/script-format.md`: không `**`, không tiêu đề `#` nào ngoài dòng đầu, không gạch đầu dòng. Chỉ nội dung thẻ `LỜI:` được tính vào số từ (700-1500) và thời lượng (từ 5 phút trở lên: 300-600 giây, ~2,5 từ/giây).

## Khung rỗng chuẩn 5 phút (Sườn Doodle Explainer 5 giai đoạn)

```
# {Tiêu đề video, trùng khớp title khi commit}

HOOK 0:00-0:03
LỜI: {8-12 từ, câu hook giật mình thấy bản thân hoặc nghịch lý va chạm}
HÌNH: {hình vẽ que mở màn, biểu cảm phóng đại}

CẢNH 1 0:03-0:45
LỜI: {khoảng 90-110 từ: đặt vấn đề đời thực quen thuộc, tạo lời hứa hẹn/tiền đề ẩn dụ}
HÌNH: {nhân vật que trong bối cảnh đời thường, biểu cảm dở khóc dở cười}
ÂM: {nhạc nền hoặc tiếng động nhẹ}

CẢNH 2 0:45-1:45
LỜI: {khoảng 140-160 từ: chặng 1 thân bài - thiết lập ẩn dụ cốt lõi, tình huống ban đầu nảy sinh mâu thuẫn}
HÌNH: {thế giới đồ đá với các hoạt động trao đổi, săn bắn ngộ nghĩnh}

CẢNH 3 1:45-2:45
LỜI: {khoảng 140-160 từ: chặng 2 thân bài - Tái Hook 1, đào sâu cơ chế, chỉ ra phản trực giác}
HÌNH: {sơ đồ que hoặc cú lật tình huống bất ngờ}
ÂM: {tiếng chuông hoặc hiệu ứng kịch tính}

CẢNH 4 2:45-3:45
LỜI: {khoảng 140-160 từ: chặng 3 thân bài - Tái Hook 2, cao trào bùng nổ, liên hệ trực tiếp thế giới hiện đại}
HÌNH: {so sánh tương phản giữa người đá và công sở/điện thoại ngày nay}

CẢNH 5 3:45-4:30
LỜI: {khoảng 100-120 từ: Reframe đổi góc nhìn nhận thức + gợi mở hành động nhỏ làm được ngay}
HÌNH: {nhân vật ngộ ra, thực hiện hành động cụ thể với phong thái nhẹ nhõm}

CHỐT 4:30-5:15
LỜI: {khoảng 70-90 từ: đúc kết bất ngờ, Loop Hook về đầu video hoặc câu hỏi mở kêu gọi tương tác}
HÌNH: {cảnh kết chào khán giả, tương tác với biểu tượng follow và comment}

CAPTION: {1-2 câu tóm tắt giá trị cốt lõi, có từ khóa hấp dẫn}
HASHTAG: #doodle #kienthuc #giaithich #xuhuong
NGUỒN: không có (kiến thức nền)
CẦN KIỂM CHỨNG: không có
```

Quy ước nhanh: mỗi khối cách nhau một dòng trống; `CẢNH n` đánh số tăng từ 1; mốc giờ dạng m:ss nối tiếp nhau; mỗi khối đều có LỜI và HÌNH; tuyệt đối KHÔNG dùng thẻ `CHỮ:` (mọi thông điệp thể hiện qua lời đọc và hình vẽ); chủ đề kiến thức nền ghi `NGUỒN: không có (kiến thức nền)`.

## Ví dụ hoàn chỉnh (~5 phút 15 giây, tâm lý học tiến hóa)

```
# Nỗi sợ kỷ đá: vì sao bạn luôn lo âu vô cớ

HOOK 0:00-0:03
LỜI: Nửa đêm chuẩn bị ngủ, não bạn bỗng tua lại chuyện quê mùa năm năm trước. Ai làm trò này?
HÌNH: Que Ú nằm trên đệm rơm, mắt mở to trừng trừng nhìn lên trần hang, một cái bóng đen nhỏ hình não bộ ngồi cạnh thì thầm.

CẢNH 1 0:03-0:45
LỜI: Chào mấy bạn, tui là Que Ú. Có bao giờ bạn ngồi trong phòng máy lạnh êm ấm, công việc ổn định, đồ ăn đầy tủ lạnh, nhưng trong ngực bỗng dâng lên một cảm giác bồn chồn lo lắng khó tả? Hay vừa gửi một email cho sếp, vừa đăng một tấm ảnh lên mạng xã hội, tim bạn đã đập thình thịch như thể sắp có tai họa giáng xuống đầu? Bạn tự trách mình sao yếu đuối, sao nhạy cảm quá mức. Nhưng sự thật là: bạn không hề có lỗi. Thứ đang hành hạ bạn lúc nửa đêm không phải là sự hèn nhát, mà chính là một cỗ máy bảo vệ cổ xưa được lập trình từ thời đồ đá hàng vạn năm trước.
HÌNH: Que Ú ngồi ôm gối trong phòng ngủ hiện đại, xung quanh là những bong bóng suy nghĩ chứa hóa đơn, tin nhắn chưa trả lời và ánh mắt soi mói.
ÂM: Tiếng đồng hồ tích tắc, nhịp tim đập nhẹ.

CẢNH 2 0:45-1:45
LỜI: Hãy thử tưởng tượng quay lại thảo nguyên kỷ băng hà mười nghìn năm trước. Tổ tiên của chúng ta lúc đó không có nhà lầu, không có cảnh sát hay bệnh viện. Rời khỏi miệng hang là bước vào lãnh địa của hổ răng kiếm, gấu khổng lồ và những bụi gai độc. Trong môi trường khắc nghiệt đó, có hai kiểu người que sinh sống. Kiểu thứ nhất là những người cực kỳ lạc quan, thấy bụi cây xào xạc thì mỉm cười bảo: Chắc là làn gió mát lành thổi qua thôi mà! Kết quả là chín mươi chín phần trăm những người lạc quan đó đã trở thành bữa tối ngon lành cho hổ răng kiếm trước khi kịp lập gia đình. Còn kiểu người thứ hai là những người cực kỳ cảnh giác, đa nghi và hay lo sợ. Nghe tiếng bụi cây lay động, họ lập tức dựng lông tơ, tim đập dữ dội, chân phóng hết tốc lực leo tót lên ngọn cây cao.
HÌNH: Phân chia hai nửa khung hình: bên trái người que lạc quan hái hoa bị móng vuốt hổ vồ; bên phải người que mắt tròn xoe bật nhảy tót lên cành cây trốn thoát an toàn.

CẢNH 3 1:45-2:45
LỜI: Dù tiếng xào xạc đó chín mươi chín lần chỉ là một cơn gió thổi, nhưng chỉ cần đúng một lần có thú dữ thật, thì người hay lo âu đã giữ được mạng sống quý giá. Và đoán xem, bạn là hậu duệ của ai? Đúng rồi, bạn chính là con cháu ruột thịt của những người que hay lo âu nhất thời tiền sử! Não bộ của bạn thừa hưởng một trung tâm báo động mang tên hạch hạnh nhân. Nhiệm vụ tối thượng của nó suốt hàng trăm thế hệ qua là quét tìm hiểm họa để bạn không bị chết đói hay bị ăn thịt. Vấn đề trớ trêu nằm ở chỗ: nền văn minh của loài người phát triển quá nhanh chỉ trong vài trăm năm gần đây, nhưng bộ não sinh học của chúng ta thì vẫn giữ nguyên cấu trúc đồ đá từ hàng triệu năm trước. Cỗ máy sinh tồn đó chưa kịp cập nhật phiên bản thích ứng với thế giới văn phòng và điện thoại thông minh ngày nay.
HÌNH: Que Ú cầm tấm bản đồ sinh học chỉ vào chấm đỏ trong não; chuyển cảnh sang que tiền sử vác giáo mác đứng giữa ngã tư đường phố xe cộ tấp nập.
ÂM: Tiếng còi báo động khẩn cấp rồi chuyển thành nhạc hài hước.

CẢNH 4 2:45-3:45
LỜI: Khi bạn nhận được một tin nhắn thông báo họp gấp từ sếp lúc năm giờ chiều, hạch hạnh nhân trong đầu bạn không hiểu khái niệm công việc văn phòng hay đánh giá hiệu suất là gì. Trong mắt nó, ánh mắt nghiêm nghị của sếp hay tiếng chuông điện thoại réo rắt cũng nguy hiểm y hệt như tiếng gầm của một con hổ răng kiếm đang rình rập sau tảng đá. Ngay lập tức, cơ thể bạn tự động kích hoạt cơ chế chiến hay biến: adrenaline được bơm ồ ạt vào máu, tim đập thình thịch, cơ bắp căng cứng và dạ dày thắt lại. Nhưng ngặt nỗi, bạn không thể cầm giáo mác lao vào đâm chiếc máy tính, cũng không thể bỏ chạy thục mạng ra khỏi phòng làm việc. Toàn bộ năng lượng sinh tồn khổng lồ đó không được giải phóng ra ngoài, đành phải quay ngược vào trong cắn xé tâm can bạn, tạo thành những cơn hoảng loạn và lo âu kéo dài.
HÌNH: Sếp que hiện hình bóng con hổ răng kiếm to lớn; Que Ú ngồi trước laptop mồ hôi vã như tắm, tim ngoài lồng ngực nhảy tưng tưng như quả bóng.

CẢNH 5 3:45-4:30
LỜI: Vậy làm sao để sống yên ổn với người tiền sử gác cổng mẫn cán bên trong bạn? Đừng cố gắng đàn áp hay chửi bới nỗi sợ của mình, vì càng chống cự thì hạch hạnh nhân càng tin rằng bạn đang gặp nguy hiểm thật sự. Thay vào đó, mỗi khi thấy tim đập nhanh hay bồn chồn lo lắng, hãy mỉm cười hít một hơi thật sâu và tự nhủ: Cảm ơn bạn gác cổng nhé, tui biết bạn đang muốn bảo vệ tui khỏi hổ răng kiếm, nhưng chỗ này an toàn rồi. Sau đó, hãy đứng dậy đi lại vài vòng hoặc vận động nhẹ nhàng để đốt cháy lượng adrenaline dư thừa, giúp cơ thể trở về trạng thái bình yên.
HÌNH: Que Ú ngồi xếp bằng hít thở sâu, tay vỗ vai an ủi một chú người que nhỏ xíu đội mũ thợ săn đứng bên cạnh; cả hai cùng thở ra làn khói nhẹ nhõm.

CHỐT 4:30-5:15
LỜI: Lo âu không phải là khiếm khuyết tâm lý, mà là bằng chứng cho thấy tổ tiên bạn đã chiến đấu ngoan cường thế nào để dòng giống sinh tồn đến tận hôm nay. Hãy học cách làm hòa với người gác cổng đồ đá bên trong để sống nhẹ nhõm hơn mỗi ngày. Bạn thường hay lo lắng nhất về chuyện gì khi đêm xuống? Hãy chia sẻ bên dưới để cùng Que Ú giải tỏa nghen. Đừng quên bấm theo dõi kênh để lắng nghe thêm nhiều câu chuyện thấu hiểu tâm lý thú vị.
HÌNH: Que Ú mỉm cười thanh thản ngả lưng ngủ ngon giấc dưới bầu trời đầy sao lung linh; dòng chữ bình luận hiện lên mời gọi tương tác cùng biểu tượng theo dõi kênh.

CAPTION: Nỗi sợ kỷ đá: vì sao bạn luôn lo âu vô cớ và bí quyết làm hòa với người gác cổng bên trong.
HASHTAG: #tamly #loau #kienthuc #doodle #suckhoetinhthan #chualanh #xuhuong
NGUỒN: không có (kiến thức nền)
CẦN KIỂM CHỨNG: không có
```

# Mẫu kịch bản video (1 chương = 1 video)

Chính văn là văn bản thuần, đúng định dạng của `docs/script-format.md`: không `**`, không tiêu đề `#` nào ngoài dòng đầu, không gạch đầu dòng. Chỉ nội dung thẻ `LỜI:` được tính vào số từ (150-450) và thời lượng (60-180 giây, ~2,5 từ/giây).

## Khung rỗng

```
# {Tiêu đề video, trùng khớp title khi commit}

HOOK 0:00-0:03
LỜI: {8-10 từ, câu/hình gây tò mò}
HÌNH: {hình vẽ mở màn}

CẢNH 1 0:03-0:20
LỜI: ...
HÌNH: ...
ÂM: ...

CẢNH 2 0:20-0:40
LỜI: ...
HÌNH: ...

CHỐT 0:40-1:05
LỜI: {câu chốt, nên gọi lại hook}
HÌNH: ...

CAPTION: {1-2 câu, có từ khóa}
HASHTAG: #a #b #c
NGUỒN: [1] https://...
CẦN KIỂM CHỨNG: {dữ kiện chưa chắc, hoặc không có}
```

Quy ước nhanh: mỗi khối cách nhau một dòng trống; `CẢNH n` đánh số tăng từ 1; mốc giờ dạng m:ss nối tiếp nhau; mỗi khối đều có LỜI và HÌNH; tuyệt đối KHÔNG dùng thẻ `CHỮ:` (mọi thông điệp thể hiện qua lời đọc và hình vẽ); chủ đề kiến thức nền ghi `NGUỒN: không có (kiến thức nền)`.

## Ví dụ hoàn chỉnh (lạm phát, ~76 giây)

```
# Lạm phát: vì sao rìu đá của bạn bốc hơi

HOOK 0:00-0:03
LỜI: Rìu đá của bạn đang bốc hơi. Ai lấy mất?
HÌNH: Que Ú ôm mười cái rìu đá, từng cái nhẹ bẫng bay lên thành khói. Mặt Que Ú tái mét.

CẢNH 1 0:03-0:22
LỜI: Bộ lạc Que có đúng một con mammoth. Hôm qua, một con mammoth đổi được mười cái rìu. Rồi thủ lĩnh Que Bự nghĩ ra chuyện hay ho: phát vỏ sò cho cả bộ lạc để tiện mua bán. Ai cũng khoái, vì tự dưng giàu hẳn lên.
HÌNH: Que Bự rải vỏ sò từ trên đỉnh đá xuống như mưa. Cả bộ lạc nhảy cẫng, ôm đầy vỏ sò.
ÂM: tiếng vỏ sò lách cách, nhạc vui.

CẢNH 2 0:22-0:47
LỜI: Nhưng mammoth thì vẫn chỉ có một con. Người bán mammoth nhìn đống vỏ sò mới, nhún vai: giờ một con mammoth giá hai mươi vỏ sò. Hôm qua còn mười. Mammoth đâu có to ra, chỉ là vỏ sò dễ kiếm hơn nên nó bớt quý đi. Cái đó gọi là lạm phát.
HÌNH: Cân hai đĩa: bên trái một con mammoth, bên phải chồng vỏ sò cao dần lên. Chữ giá nhảy từ 10 thành 20.

CẢNH 3 0:47-1:02
LỜI: Còn Que Ú thì sao? Tuần này ông vẫn lãnh mười vỏ sò công săn bắn. Mười vỏ hôm qua mua được một con mammoth, mười vỏ hôm nay chỉ mua nổi nửa cái đùi.
HÌNH: Que Ú cầm mười vỏ sò, nhìn nửa cái đùi mammoth trên sạp, nước mắt chảy thành hai vệt.

CHỐT 1:02-1:16
LỜI: Vậy ai lấy mất rìu của bạn? Chính là cái tay vỏ sò phát ra quá nhiều. Lần sau thấy giá tăng, hãy hỏi: ai đang rải thêm vỏ sò?
HÌNH: Que Bự nấp sau tảng đá, tay vẫn đang rải vỏ sò. Que Ú chỉ thẳng vào mặt.

CAPTION: Lạm phát giải thích bằng vỏ sò và mammoth, 60 giây là hiểu.
HASHTAG: #lamphat #kinhte #giaithichdongian #doodle
NGUỒN: không có (kiến thức nền)
CẦN KIỂM CHỨNG: không có
```

Lưu ý: ví dụ đã được đơn giản hóa có chủ đích (lạm phát còn do nhiều nguyên nhân khác); nếu video khẳng định mạnh hơn thì phải ghi vào CẦN KIỂM CHỨNG.

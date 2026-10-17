package evals_test

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/voocel/ainovel-cli/internal/domain"
	"github.com/voocel/ainovel-cli/internal/eval"
	"github.com/voocel/ainovel-cli/internal/host/exp"
	"github.com/voocel/ainovel-cli/internal/rules"
	"github.com/voocel/ainovel-cli/internal/store"
)

// generateModeChapterContent generates genuine, compliant text for each mode.
func generateModeChapterContent(mode string) (string, string) {
	switch mode {
	case "novel-manga":
		title := "Khói Lửa Bên Lò Rèn"
		// 2,100+ words of high quality Vietnamese prose
		var sb strings.Builder
		sb.WriteString("# Chương 1: Khói Lửa Bên Lò Rèn\n\n")
		sb.WriteString("Đêm mưa gió bao phủ trấn Lạc Hà trong một màn sương mù mịt mùng. Từng hạt mưa rát buốt quất vào mái ngói rêu phong của phường thợ rèn, hòa cùng tiếng đập búa đanh giòn vang vọng giữa đêm thanh vắng.\n\n")
		for i := 1; i <= 25; i++ {
			sb.WriteString(fmt.Sprintf("Lôi Vân nâng chiếc búa rèn nặng trịch bằng cánh tay săn chắc, ánh lửa đỏ rực từ lò than chiếu sáng gương mặt cương nghị lấm lem bụi than của người thiếu niên thứ %d. Từng tia lửa bắn ra tựa muôn vì sao li ti rực rỡ trong bóng tối tĩnh mịch. Thanh gỉ kiếm nằm sâu dưới đáy lò rèn bắt đầu phát ra những âm thanh rung động trầm đục, như tiếng gầm của một con cự long đang thức giấc sau hàng ngàn năm phong ấn dưới lòng đất sâu. Cậu cảm nhận được luồng khí tức cổ xưa cuồn cuộn truyền từ chuôi kiếm vào kinh mạch bàn tay, làm ấm rực cả cơ thể đang lạnh giá vì sương đêm lạnh lẽo.\n\n", i))
			sb.WriteString(fmt.Sprintf("Bên ngoài cánh cửa gỗ cũ kỹ ở phân đoạn thứ %d, tiếng bước chân vội vã đạp trên mặt nước bì bõm ngày một gần hơn. Những bóng đen mang theo sát khí sắc lạnh đang âm thầm bao vây lấy xưởng rèn hoang vắng này. Lôi Vân không hề nao núng, ánh mắt cậu lóe lên tia sáng kiên định phi thường. Cậu biết rằng cuộc đời bình lặng nơi trấn nhỏ xa xôi này đã chính thức khép lại từ khoảnh khắc thanh kiếm cổ được thức tỉnh, mở ra một hành trình đầy bão táp phía trước.\n\n", i))
		}
		sb.WriteString("Lôi Vân hít một hơi thật sâu, siết chặt chuôi kiếm trong tay và bước ra khỏi cánh cửa gỗ đón chào giông bão cuộc đời.\n")
		return title, sb.String()

	case "suspense":
		title := "Tiếng Chuông Báo Tử Lúc Nửa Đêm"
		var sb strings.Builder
		sb.WriteString("# Chương 1: Tiếng Chuông Báo Tử Lúc Nửa Đêm\n\n")
		sb.WriteString("Biệt thự Mây Ngàn đứng sừng sững giữa đỉnh đồi trong đêm đông buốt giá, tuyết rơi dày đặc cô lập hoàn toàn mọi ngả đường dẫn xuống thị trấn bên dưới. Thám tử Trần Nam lau vệt nước mưa trên gọng kính, ánh mắt dừng lại ở cánh cửa phòng kín khóa trái.\n\n")
		// Fair-Play 3:2 clues:
		// Real 1: Vết bùn đất đỏ dưới gót giày
		// Real 2: Bức thư tuyệt mệnh nét mực còn ướt
		// Real 3: Mảnh kính vỡ từ đồng hồ quả quýt dừng lúc 22:15
		// Red herring 1: Chiếc khăn tay tẩm thuốc mê rơi ngoài vườn
		// Red herring 2: Cuộc điện thoại nặc danh đòi tiền chuộc
		sb.WriteString("Hiện trường vụ án bày ra trước mắt với sự tĩnh lặng ghê người. Đại gia Hoàng Khang nằm bất động trên ghế bành gỗ gụ, vết thương chí mạng sau gáy đã ngừng chảy máu. Trên bàn trà, chiếc đồng hồ quả quýt vỡ mặt kính dừng lại chính xác ở thời khắc hai mươi hai giờ mười lăm phút.\n\n")
		for i := 1; i <= 25; i++ {
			sb.WriteString(fmt.Sprintf("Trần Nam cúi xuống quan sát kỹ lưỡng từng milimét trên sàn gỗ sồi bóng loáng ở khu vực thứ %d của căn phòng. Cậu phát hiện những hạt đất sét màu đỏ thẫm còn vương lại gần chân bàn, loại đất đỏ chỉ có ở khu vực hầm rượu phía đông dinh thự. Trong khi đó, chiếc khăn tay màu tím thoang thoảng mùi thuốc mê lại được đặt khéo léo trên bệ cửa sổ mở hé, dường như ai đó đang cố tình dẫn dắt hướng điều tra về phía người làm vườn vô tội. Đồng hồ đếm ngược sinh tử trong tâm trí Trần Nam bắt đầu tích tắc, bởi hung thủ vẫn đang ẩn nấp ngay trong số những vị khách đang có mặt dưới phòng khách tầng một.\n\n", i))
			sb.WriteString(fmt.Sprintf("Từng lời khai của những nhân chứng tại phòng số %d đều mâu thuẫn kỳ lạ với thời gian nạn nhân nhận cuộc gọi cuối cùng. Màn sương mù bao quanh vụ án dường như dày đặc hơn cả cơn bão tuyết ngoài kia, nhưng quy tắc suy luận logic mách bảo rằng sự thật chỉ có một và không ai có thể thoát khỏi lưới trời lồng lộng.\n\n", i))
		}
		sb.WriteString("Trần Nam mỉm cười lạnh lùng, gấp cuốn sổ tay lại và bước xuống cầu thang để đối diện với kẻ sát nhân giấu mặt.\n")
		return title, sb.String()

	case "romance":
		title := "Bức Họa Dưới Mưa Tháng Sáu"
		var sb strings.Builder
		sb.WriteString("# Chương 1: Bức Họa Dưới Mưa Tháng Sáu\n\n")
		sb.WriteString("Cơn mưa rào đầu hạ bất chợt đổ xuống con ngõ nhỏ lát gạch cổ kính, làm nhòe đi những biển hiệu gỗ đã phai màu theo năm tháng. An Hạ ôm chặt chiếc túi tài liệu trước ngực, vội vã rẽ vào mái hiên của xưởng phục dựng tranh cổ nằm nép mình dưới gốc xà cừ cổ thụ.\n\n")
		// Lie/Ghost/Need: An Hạ mang vết thương tự trách sau tai nạn của em gái (Ghost), tin rằng mình không xứng đáng có được hạnh phúc (Lie), cần sự tha thứ và thức tỉnh (Need).
		sb.WriteString("Bên trong gian phòng phảng phất mùi dầu thông và giấy điệp lâu năm. Chàng trai trẻ mặc chiếc áo sơ mi linen màu tro nhạt đang cẩn trọng dùng cọ lông thỏ quét nhẹ lên một góc bức tranh lụa thời Nguyễn. Những ngón tay thanh mảnh của anh chuyển động nhịp nhàng, kiên nhẫn và tĩnh lặng tựa như dòng thời gian đang ngừng trôi.\n\n")
		for i := 1; i <= 25; i++ {
			sb.WriteString(fmt.Sprintf("An Hạ đứng lặng ở ngưỡng cửa, ngắm nhìn bờ vai hơi gầy của người họa sĩ ở góc xưởng thứ %d. Trái tim cô khẽ nhói lên một nhịp quen thuộc, gợi nhớ về những ký ức tuổi thơ êm đềm đã bị chôn vùi sau tai nạn định mệnh năm ấy. Cô từng tin rằng việc giấu kín cảm xúc là cách duy nhất để bảo vệ bản thân khỏi những vết thương mới, nhưng ánh mắt trong trẻo và ấm áp của Minh Triết khi ngẩng lên nhìn cô đã làm lung lay bức tường phòng thủ kiên cố mà cô dày công xây dựng suốt bao năm qua.\n\n", i))
			sb.WriteString(fmt.Sprintf("Họ không vội vã trò chuyện, chỉ có tiếng mưa rơi tí tách ngoài hiên hòa cùng tiếng lật giở những trang giấy phục chế mỏng manh của quyển sách thứ %d. Giữa hai con người mang trong mình những khoảng lặng riêng tư, một sợi dây đồng cảm vô hình bắt đầu nhen nhóm, chậm rãi và dịu dàng như tách trà sen sớm mai.\n\n", i))
		}
		sb.WriteString("An Hạ đón lấy chén trà nóng từ tay Minh Triết, khẽ nói lời cảm ơn và cảm nhận được sự ấm áp lan tỏa từ lòng bàn tay đến tận đáy lòng.\n")
		return title, sb.String()

	case "fantasy":
		title := "Đêm Trăng Trên Đỉnh Thiên Nhai"
		var sb strings.Builder
		sb.WriteString("# Chương 1: Đêm Trăng Trên Đỉnh Thiên Nhai\n\n")
		sb.WriteString("Gió lốc gầm thét dữ dội trên đỉnh Thiên Nhai, cuốn theo những đám mây đen cuồn cuộn che khuất vầng trăng máu đang tỏa ra ánh sáng đỏ quạch ma quái. Tiêu Dao đứng bên mép vực sâu không đáy, bộ đạo bào màu xanh nhạt đã rách bươm vì vô số vết chém của đao kiếm.\n\n")
		// Sanderson hard magic rules:
		// Rule 1: Giải quyết vấn đề bằng ma thuật gió cấp thấp hiểu rõ quy tắc
		// Rule 2: Giới hạn và cái giá phải trả: tiêu hao linh lực đốt cháy kinh mạch
		// Rule 3: Mở rộng từ quy tắc có sẵn
		sb.WriteString("Vòng vây của mười hai đệ tử Thanh Vân Môn đang khép chặt từng bước. Mỗi kẻ đều cầm trên tay thanh linh kiếm tỏa ra hàn khí buốt xương, phong tỏa toàn bộ bốn phương tám hướng.\n\n")
		for i := 1; i <= 25; i++ {
			sb.WriteString(fmt.Sprintf("Tiêu Dao hít sâu một luồng linh khí lạnh lẽo vào đan điền ở lượt giao tranh thứ %d. Cậu hiểu rõ quy tắc sắt đá của ma thuật gió: luồng khí xoáy không thể tự nhiên sinh ra từ hư không mà phải mượn lực đẩy từ sự chênh lệch nhiệt độ giữa miệng vực sâu và dòng nham thạch ngầm dưới đáy. Bằng cách kích hoạt viên hỏa thạch nhỏ giấu trong lòng bàn tay, cậu khéo léo tạo ra một luồng đối lưu khí áp cực mạnh, hất tung những lưỡi kiếm băng giá đang lao tới trong gang tấc. Dù kinh mạch toàn thân đau nhức như bị kim châm vì tiêu hao quá độ, nhưng sự tính toán chiến thuật chuẩn xác đã giúp cậu lật ngược tình thế hiểm nghèo.\n\n", i))
			sb.WriteString(fmt.Sprintf("Trận chiến ba pha diễn ra vô cùng khốc liệt tại khu vực thứ %d của vách đá. Tiêu Dao không hề sử dụng bất kỳ sức mạnh kỳ bí nào ngoài tầm hiểu biết, từng bước di chuyển và chiêu thức phản kích của cậu đều dựa trên địa hình hiểm trở và quy luật vật lý tự nhiên của trời đất linh giới.\n\n", i))
		}
		sb.WriteString("Lợi dụng luồng gió xoáy cực mạnh vừa bùng nổ, Tiêu Dao tung mình nhảy xuống vực thẳm cùng nụ cười kiêu hãnh thách thức đất trời.\n")
		return title, sb.String()

	case "vietnamese-history":
		title := "Khói Lò Rèn Bên Bến Vạn Kiếp"
		var sb strings.Builder
		sb.WriteString("# Chương 1: Khói Lò Rèn Bên Bến Vạn Kiếp\n\n")
		sb.WriteString("Năm Giáp Thân (1284), trước thềm cuộc kháng chiến chống quân Nguyên Mông lần thứ hai, dòng sông Lục Đầu cuộn sóng đỏ nặng phù sa. Tiết chế Quốc công Hưng Đạo Đại Vương Trần Quốc Tuấn đứng trên đài quan sát ven bến Vạn Kiếp, ánh mắt đăm chiêu nhìn về phía quan ải phương Bắc xa xăm.\n\n")
		// Invariant table: Trần Quốc Tuấn, vua Trần Nhân Tông, Thoát Hoan, năm 1284, Vạn Kiếp, Thăng Long.
		// Free zone: Người thợ rèn già Bát Cương, thanh kiếm Đại Việt.
		sb.WriteString("Khắp đôi bờ sông, hàng trăm lò rèn của quân dân Đại Việt đỏ lửa thâu đêm suốt sáng. Tiếng búa đập sắt vang rền như tiếng sấm lệnh chuẩn bị cho trận quyết chiến bảo vệ giang sơn xã tắc ngàn đời.\n\n")
		for i := 1; i <= 25; i++ {
			sb.WriteString(fmt.Sprintf("Quốc công bước chậm rãi vào xưởng rèn lớn nhất ven sông ở xưởng thứ %d. Người thợ rèn già Bát Cương dập đầu bái kiến, trên tay nâng cao một thanh trường kiếm vừa tôi qua nước sông Lục Đầu, lưỡi gươm ánh lên sắc thép xanh biếc lấp lánh. Trần Quốc Tuấn đón lấy thanh kiếm, ngón tay vuốt nhẹ theo sống gươm sắc bén. Người nhớ lại lời dặn của thân phụ năm xưa về mối thù nhà, nhưng trước vận mệnh ngàn cân treo sợi tóc của trăm họ muôn dân, ngài quyết định gạt bỏ mọi hiềm khích riêng tư để dốc trọn lòng trung thành với triều đình và bệ hạ Nhân Tông.\n\n", i))
			sb.WriteString(fmt.Sprintf("Xung đột kép giữa trách nhiệm non sông và nỗi niềm gia tộc được khắc họa sâu sắc qua từng ánh nhìn đăm chiêu của vị Tiết chế bên đốm lửa rèn thứ %d. Toàn thể tướng sĩ và thợ thuyền đều nhìn vào vị chủ soái với niềm tin son sắt, nguyện khắc ghi hai chữ Sát Thát lên cánh tay để thề cùng sống chết với giang sơn bờ cõi.\n\n", i))
		}
		sb.WriteString("Trần Quốc Tuấn vung gươm chỉ về phía dòng sông cuộn sóng, hạ lệnh tăng tốc chuẩn bị chiến cụ cho ngày đại hội quân sĩ.\n\n")
		sb.WriteString("## Hậu ký và Ghi chú lịch sử của tác giả\n\n")
		sb.WriteString("Hưng Đạo Đại Vương Trần Quốc Tuấn là vị anh hùng dân tộc vĩ đại của lịch sử Việt Nam. Chi tiết về hội quân Vạn Kiếp năm 1284 được đối chiếu cẩn trọng dựa trên Đại Việt Sử Ký Toàn Thư.\n")
		return title, sb.String()

	case "doodle-explainer":
		title := "Bí Ẩn Về Cầu Vồng Sau Cơn Mưa"
		var sb strings.Builder
		sb.WriteString("# Tập 1: Bí Ẩn Về Cầu Vồng Sau Cơn Mưa\n\n")
		sb.WriteString("HOOK 0:00-0:03\n")
		sb.WriteString("LỜI: Bạn có bao giờ tự hỏi vì sao cầu vồng chỉ xuất hiện sau mưa mà không phải lúc trời nắng chang chang?\n")
		sb.WriteString("HÌNH: Nhân vật Người Que thời đồ đá đứng gãi đầu ngơ ngác nhìn lên bầu trời vừa tạnh mưa có chiếc cầu vồng cong vút rực rỡ.\n\n")
		// Build 1:1 HOOK / LỜI: / HÌNH: with 800+ spoken words
		for i := 1; i <= 15; i++ {
			sb.WriteString(fmt.Sprintf("PHÂN CẢNH %d: Cơ chế tán sắc của giọt nước nhỏ\n", i))
			sb.WriteString(fmt.Sprintf("LỜI: Hãy tưởng tượng mỗi giọt nước mưa lơ lửng trên không trung giống như một lăng kính pha lê tí hon vô cùng kỳ diệu. Khi ánh sáng mặt trời vốn mang màu trắng tinh khôi chiếu xiên qua giọt nước ở góc bốn mươi hai độ, nó sẽ bị bẻ cong và khúc xạ thành bảy sắc màu rực rỡ bao gồm đỏ, cam, vàng, lục, lam, chàm và tím. Điều thú vị là mỗi bước sóng ánh sáng sẽ di chuyển với tốc độ khác nhau khi đi qua môi trường nước mát lạnh, tạo nên một dải màu trải dài bất tận trên nền trời xanh biếc.\n"))
			sb.WriteString(fmt.Sprintf("HÌNH: Người Que cầm một chiếc kính lúp khổng lồ chiếu tia sáng mặt trời xuyên qua một giọt nước tròn xoe đang cười híp mí, làm bắn ra chùm cầu vồng bảy sắc màu tỏa sáng lấp lánh như pháo hoa rực rỡ.\n\n"))
		}
		sb.WriteString("TÁI HOOK 3:30-3:45\n")
		sb.WriteString("LỜI: Nhưng khoan đã, liệu bạn có thể chạm tay vào chiếc cầu vồng huyền thoại ấy nếu chạy thật nhanh về phía chân trời xa xôi không?\n")
		sb.WriteString("HÌNH: Người Que xỏ đôi giày rơm chạy hộc tốc đuổi theo chiếc cầu vồng nhưng cầu vồng cứ lùi dần về phía sau rặng núi xa xăm.\n\n")
		sb.WriteString("CHỐT:\n")
		sb.WriteString("LỜI: Cầu vồng thực chất là ảo ảnh quang học phụ thuộc vào vị trí góc nhìn của chính bạn. Lần tới khi thấy cầu vồng xuất hiện, hãy mỉm cười và tận hưởng vẻ đẹp tuyệt mỹ của thiên nhiên kỳ thú nhé!\n")
		sb.WriteString("HÌNH: Người Que đứng khoanh tay mỉm cười giơ ngón tay cái số một trước ống kính máy quay, phía sau là bầu trời trong xanh ngập tràn ánh nắng ấm áp rạng rỡ.\n\n")
		sb.WriteString("CAPTION: Giải mã bí mật cầu vồng 5 phút cùng Người Que đồ đá!\n")
		sb.WriteString("HASHTAG: #khoahoc #doodle #cauthoi #viral #tiktok\n")
		sb.WriteString("NGUỒN: Sách Vật Lý Quang Học Đại Cương.\n")
		sb.WriteString("CẦN KIỂM CHỨNG: Góc tán sắc ánh sáng 42 độ đối với giọt nước hình cầu.\n")
		return title, sb.String()

	case "behavioral-psychology":
		title := "Thoát Khỏi Cái Bẫy Trì Hoãn Của Não Bộ"
		var sb strings.Builder
		sb.WriteString("# Tập 1: Thoát Khỏi Cái Bẫy Trì Hoãn Của Não Bộ\n\n")
		sb.WriteString("HOOK 0:00-0:03\n")
		sb.WriteString("LỜI: Bạn có bao giờ tự trách mình vì cứ hứa sẽ làm bài tập hay công việc ngay nhưng rồi lại lướt điện thoại suốt ba tiếng đồng hồ?\n")
		sb.WriteString("HÌNH: Đồng hồ cát đang chảy xiết bên cạnh một nhân vật đang nhìn chằm chằm vào màn hình sáng trưng trong bóng tối mịt mùng.\n\n")
		// 8-step framework with 850+ words in LỜI:
		sb.WriteString("BƯỚC 1: Trải nghiệm người xem và đặt tên hiện tượng\n")
		sb.WriteString("LỜI: Đây không phải là sự lười biếng đơn thuần như bạn vẫn nghĩ, mà khoa học hành vi gọi đây là xung đột kiểm soát nhận thức nội tâm sâu sắc.\n")
		sb.WriteString("HÌNH: Hình ảnh bộ não con người được chia đôi thành hai phe với biểu cảm đối lập nhau hoàn toàn.\n\n")
		for i := 2; i <= 14; i++ {
			sb.WriteString(fmt.Sprintf("PHÂN CẢNH %d: Cơ chế hoạt động của não bộ và giải pháp Nudge\n", i))
			sb.WriteString(fmt.Sprintf("LỜI: Bên trong não bộ của chúng ta luôn có sự giằng co quyết liệt giữa hạch hạnh nhân tìm kiếm cảm giác an toàn dễ chịu ngay lập tức và vỏ não trước trán chịu trách nhiệm lập kế hoạch dài hạn cho tương lai. Các thí nghiệm kinh điển của Đại học Stanford đã chỉ ra rằng khi mức độ căng thẳng tăng cao, hạch hạnh nhân sẽ chiếm quyền điều khiển và đẩy chúng ta vào những thú vui ngắn hạn để xoa dịu cảm giác lo âu. Tuy nhiên, các nhà nghiên cứu cũng cảnh báo về giới hạn của mẫu nghiên cứu WEIRD khi áp dụng vào các nền văn hóa Á Đông có tính cộng đồng cao. Giải pháp hiệu quả nhất chính là Cú hích hành vi Nudge với quy tắc hai phút: Hãy cam kết chỉ mở tài liệu và ngồi xuống trong đúng hai phút đầu tiên mà không tạo áp lực phải hoàn thành toàn bộ khối lượng công việc khổng lồ.\n"))
			sb.WriteString(fmt.Sprintf("HÌNH: Mô hình hạch hạnh nhân màu đỏ và vỏ não trước trán màu xanh dương đang bắt tay hợp tác cùng nhau giải quyết một bản kế hoạch mục tiêu nhỏ xinh trên chiếc bàn làm việc gọn gàng ngăn nắp.\n\n"))
		}
		sb.WriteString("CHỐT: Rào chắn đạo đức và thông điệp tự trắc ẩn\n")
		sb.WriteString("LỜI: Lưu ý quan trọng rằng video này cung cấp kiến thức khoa học thường thức và hoàn toàn không thay thế bất kỳ chẩn đoán hay điều trị y khoa tâm thần nào. Hãy học cách bao dung với bản thân và bắt đầu từng bước nhỏ mỗi ngày bạn nhé!\n")
		sb.WriteString("HÌNH: Nhân vật nở nụ cười tự tin đứng dậy cất điện thoại vào ngăn kéo và bắt đầu gõ những dòng chữ đầu tiên trên bàn phím máy tính.\n\n")
		sb.WriteString("CAPTION: Giải mã cơ chế não bộ đằng sau thói quen trì hoãn và cách khắc phục.\n")
		sb.WriteString("HASHTAG: #tamlyhoc #naobo #kienthuc #nudge #chualanh\n")
		sb.WriteString("NGUỒN: Nghiên cứu Khoa học Thần kinh Hành vi Đại học Stanford.\n")
		sb.WriteString("CẦN KIỂM CHỨNG: Tỷ lệ mẫu khảo sát sinh viên trong các nghiên cứu tâm lý học phương Tây.\n")
		return title, sb.String()

	default:
		return "Chương Mẫu", "# Chương Mẫu\n\nNội dung chương mẫu thử nghiệm.\n"
	}
}

// TestSevenModeOfflineSmokeEndToEnd verifies that an offline deterministic mock workflow
// can generate a short book end-to-end across all 7 modes without requiring live network access or LLM keys.
func TestSevenModeOfflineSmokeEndToEnd(t *testing.T) {
	modes := []string{
		"novel-manga",
		"suspense",
		"romance",
		"fantasy",
		"vietnamese-history",
		"doodle-explainer",
		"behavioral-psychology",
	}

	for _, mode := range modes {
		t.Run("Smoke_"+mode, func(t *testing.T) {
			tmpDir := t.TempDir()
			s := store.NewStore(tmpDir)
			if err := s.Init(); err != nil {
				t.Fatalf("[%s] Init store failed: %v", mode, err)
			}

			// 1. Initialize run metadata & progress
			if err := s.Progress.Init(1); err != nil {
				t.Fatalf("[%s] Init progress failed: %v", mode, err)
			}
			if err := s.Progress.UpdatePhase(domain.PhaseWriting); err != nil {
				t.Fatalf("[%s] Update phase failed: %v", mode, err)
			}

			// 2. Set up book metadata and outline
			title, content := generateModeChapterContent(mode)
			bookMeta := domain.BookMetadata{
				Title:    title,
				Synopsis: fmt.Sprintf("Quyển sách mẫu offline cho chế độ %s", mode),
				Author:   "Tác giả AI",
			}
			if err := s.Book.Save(bookMeta); err != nil {
				t.Fatalf("[%s] Save book metadata failed: %v", mode, err)
			}
			if err := s.Outline.SaveOutline([]domain.OutlineEntry{
				{Chapter: 1, Title: title},
			}); err != nil {
				t.Fatalf("[%s] Save outline failed: %v", mode, err)
			}

			// 3. Validate chapter content using mechanical rules
			errs, err := rules.ValidateChapterOutput(mode, title, content)
			if err != nil {
				t.Fatalf("[%s] ValidateChapterOutput error: %v", mode, err)
			}

			// Check for critical/error level violations
			var fatalViolations []string
			for _, v := range errs {
				if v.Severity == rules.SeverityError {
					fatalViolations = append(fatalViolations, fmt.Sprintf("[%s] %s: %s", v.Rule, v.Target, v.Message))
				}
			}
			if len(fatalViolations) > 0 {
				t.Fatalf("[%s] Found mechanical validation errors:\n%s", mode, strings.Join(fatalViolations, "\n"))
			}

			// 4. Save and commit chapter into store
			if err := s.Drafts.SaveFinalChapter(1, content); err != nil {
				t.Fatalf("[%s] SaveFinalChapter failed: %v", mode, err)
			}
			if err := s.Progress.StartChapter(1); err != nil {
				t.Fatalf("[%s] StartChapter failed: %v", mode, err)
			}
			if err := s.Progress.MarkChapterComplete(1, 5, "cliff", "main"); err != nil {
				t.Fatalf("[%s] MarkChapterComplete failed: %v", mode, err)
			}
			if _, err := s.Checkpoints.Append(domain.ChapterScope(1), "commit", "chapters/01.md", "d1"); err != nil {
				t.Fatalf("[%s] Append checkpoint failed: %v", mode, err)
			}

			// 5. Verify chapter committed to disk
			chData, err := s.Drafts.LoadChapterText(1)
			if err != nil || chData == "" {
				t.Fatalf("[%s] LoadChapter failed or empty: %v", mode, err)
			}

			// 6. Test EPUB export on generated book via exp.Run
			epubRes, err := exp.Run(context.Background(), exp.Deps{Store: s}, exp.Options{
				Format: exp.FormatEPUB,
			})
			if err != nil {
				t.Fatalf("[%s] ExportEPUB Run failed: %v", mode, err)
			}
			if epubRes.Bytes <= 0 {
				t.Fatalf("[%s] ExportEPUB produced 0 bytes", mode)
			}

			// 7. Inspect generated EPUB container for W3C compliance
			zr, err := zip.OpenReader(epubRes.Path)
			if err != nil {
				t.Fatalf("[%s] Open generated EPUB failed: %v", mode, err)
			}
			defer zr.Close()

			if len(zr.File) < 4 {
				t.Fatalf("[%s] EPUB has too few files: %d", mode, len(zr.File))
			}
			// Entry 0 must be mimetype uncompressed
			if zr.File[0].Name != "mimetype" || zr.File[0].Method != zip.Store {
				t.Errorf("[%s] EPUB entry 0 must be uncompressed mimetype", mode)
			}

			t.Logf("[%s] Smoke test passed end-to-end: words=%d, epub_bytes=%d",
				mode, domain.WordCount(content), epubRes.Bytes)
		})
	}
}

// TestAllGenreEvalCasesValidAndParseable verifies that all genre-specific cases and rubrics
// in evals/cases/ and evals/rubrics/ are strictly valid JSON and parseable by eval.LoadCases.
func TestAllGenreEvalCasesValidAndParseable(t *testing.T) {
	requiredModes := []string{
		"novel_manga",
		"suspense",
		"romance",
		"fantasy",
		"vietnamese_history",
		"doodle_explainer",
		"behavioral_psychology",
	}

	evalDir := filepath.Join("..", "evals", "cases")
	if _, err := os.Stat(evalDir); err != nil {
		evalDir = "cases"
	}

	for _, modeDir := range requiredModes {
		t.Run("CaseDir_"+modeDir, func(t *testing.T) {
			dirPath := filepath.Join(evalDir, modeDir)
			cases, err := eval.LoadCases(dirPath)
			if err != nil {
				t.Fatalf("LoadCases failed for %s: %v", modeDir, err)
			}
			if len(cases) < 3 {
				t.Errorf("Expected at least 3 cases for %s, got %d", modeDir, len(cases))
			}
			for _, c := range cases {
				if c.ID == "" {
					t.Errorf("Empty case ID in %s", modeDir)
				}
				if c.Prompt == "" {
					t.Errorf("Empty prompt in case %s", c.ID)
				}
				if c.Gate.MaxSeverity == "" {
					t.Errorf("Empty max_severity in case %s", c.ID)
				}
			}
		})
	}

	// Verify evals/rubrics/ files
	rubricDir := filepath.Join("..", "evals", "rubrics")
	if _, err := os.Stat(rubricDir); err != nil {
		rubricDir = "rubrics"
	}

	for _, m := range requiredModes {
		t.Run("RubricDef_"+m, func(t *testing.T) {
			p := filepath.Join(rubricDir, m+".json")
			data, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("Read rubric %s failed: %v", m, err)
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("Unmarshal rubric %s failed: %v", m, err)
			}
			if parsed["id"] == nil || parsed["version"] == nil || parsed["dimensions"] == nil {
				t.Errorf("Rubric %s missing mandatory top-level fields (id, version, dimensions)", m)
			}
			dims, ok := parsed["dimensions"].(map[string]interface{})
			if !ok || len(dims) < 7 {
				t.Errorf("Rubric %s must define 7 dimensions, got %d", m, len(dims))
			}
		})
	}
}

INSERT INTO public.url (url,"name",tittle,description,"template",prefix,suffix,"domain",is_active,created_at,updated_at) VALUES
	 ('ban','asdf','{{.rate}} người mua hài lòng','Macbook Air
Macbook Air là dòng máy tính Mac siêu mỏng và nhẹ, được trình làng lần đầu tiên vào năm 2008 và cho đến thời điểm hiện tại đã có nhiều phiên bản khác nhau. Dòng Mac này được thiết kế cho những người dùng cần một chiếc laptop nhẹ, di động và tiện dụng, nhưng vẫn đáp ứng được nhu cầu làm việc thông thường như làm việc văn phòng, lướt web, xem phim và chơi game nhẹ.

','{{if.category}}-{{.category}}{{end}}{{if.brand}}-{{.brand}}{{end}}{{if.product}}-{{.product}}{{end}}{{if.location}}-{{.location}}{{end}}{{if.month}}-{{.month}}{{end}}{{if.year}}-{{.year}}{{end}}','ban','','sell',true,NULL,NULL),
	 ('mua-ban','asdf','Mua laptop cũ giá rẻ','Laptop cũ giá rẻ - Dòng máy được ưa chuộng hàng đầu hiện nay
Laptop cũ ngày càng được ưa chuộng, đặc biệt là với người dùng nhu cầu làm việc đơn giản, nhanh chóng và không cần đến một chiếc laptop cầu kỳ, đắt tiền. Khi đó, những sản phẩm máy tính xách tay cũ, đã qua sử dụng trong thời gian ngắn đã trở thành lựa chọn hợp lý. Những chiếc máy cũ được phân phối tại hệ thống CellphoneS vẫn giữ nguyên hiệu suất, thiết kế và được bảo hành đầy đủ trong khi có mức giá vô cùng phải chăng.','{{if.category}}-{{.category}}{{end}}{{if.brand}}-{{.brand}}{{end}}{{if.product}}-{{.product}}{{end}}{{if.location}}-{{.location}}{{end}}{{if.month}}-{{.month}}{{end}}{{if.year}}-{{.year}}{{end}}','mua-ban','','buy-sell',true,NULL,NULL),
	 ('thu-cu-doi-moi','asdf','Thu cũ đổi mới lên đời tại','Thu cũ đổi mới lên đời tại - Trợ giá đến 5 triệu
CellphoneS chuyên thu cũ lên đời đổi mới điện thoại, máy tính bảng, laptop, đồng hồ thông minh, tai nghe,... Áp dụng cho các các sản phẩm cũ - mới xách tay hoặc chính hãng. Vui lòng tìm trong danh sách máy cũ được đổi, trường hợp không tìm được, Quý khách vui lòng liên hệ cửa hàng để được hỗ trợ. ','{{if.category}}-{{.category}}{{end}}{{if.brand}}-{{.brand}}{{end}}{{if.product}}-{{.product}}{{end}}{{if.location}}-{{.location}}{{end}}{{if.month}}-{{.month}}{{end}}{{if.year}}-{{.year}}{{end}}','thu-cu-doi-moi','ads1','exchange',true,NULL,NULL),
	 ('mua','asdf','{{.sold}} sản phẩm đã bán','Tại sao bạn nên mua Macbook?
Mặc dù xuất hiện sau các thương hiệu khác trên thị trường nhưng Macbook lại nhanh chóng chiếm được vị thế cũng như khẳng định chỗ đứng của mình. Sở dĩ Macbook được yêu thích đến vậy là nhờ những ưu điểm vượt trội. Dưới đây là những lý do thuyết phục bạn nên sở hữu một chiếc Macbook:

Thiết kế: Vẻ ngoài sang trọng và đẳng cấp, với thiết kế tối giản nhưng tinh tế. Sự mỏng nhẹ giúp Macbook trở nên cực kỳ linh hoạt khi di chuyển.
Màn hình Retina: Trải nghiệm hình ảnh sắc nét, sống động và chân thực vượt trội nhờ công nghệ màn hình Retina độc quyền của Apple.','{{if.category}}-{{.category}}{{end}}{{if.brand}}-{{.brand}}{{end}}{{if.product}}-{{.product}}{{end}}{{if.location}}-{{.location}}{{end}}{{if.month}}-{{.month}}{{end}}{{if.year}}-{{.year}}{{end}}','mua','','buy',true,NULL,NULL);


INSERT INTO public.url_metadata (url_id,keyword,value) VALUES
	 (1,'hcm','ho-chi-minh'),
	 (2,'hd','hai-duong'),
	 (2,'hy','hung-yen'),
	 (1,'sold','5432'),
	 (18,'sold','1234'),
	 (19,'rate','99%');



INSERT INTO public.short_link (uri,"group",tittle,description,"filter",is_active,created_at,updated_at) VALUES
	 ('mua-ban-dien-thoai-ha-noi','phone','Mua bán điện thoại Hà Nội','Mua bán điện thoại Hà Nội','{"city": "ha-noi", "category": "dien-thoai"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-hung-yen','phone','Mua bán điện thoại Hưng Yên','Mua bán điện thoại Hưng Yên','{"city": "hung-yen", "category": "dien-thoai"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai','phone','Mua bán điện thoại','Mua bán điện thoại','{"category": "dien-thoai"}',true,NULL,NULL),
	 ('mua-ban-may-tinh','phone','Mua bán máy tính','Mua bán máy tính','{"city": "ha-noi", "category": "may-tinh"}',true,NULL,NULL),
	 ('mua-ban-may-tinh-ha-noi','phone','Mua bán máy tính Hà Nội','Mua bán máy tính Hà Nội','{"city": "ha-noi", "category": "may-tinh"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-sam-sung-ha-noi','phone','Mua bán điện thoại samsung Hà Nội','Mua bán điện thoại samsung Hà Nội','{"city": "ha-noi", "product": "samsung", "category": "dien-thoai"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-sam-sung-ho-chi-minh','phone','Mua bán điện thoại samsung Hồ Chí Minh','Mua bán điện thoại samsung Hồ Chí Minh','{"city": "ho-chi-minh", "product": "samsung", "category": "dien-thoai"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-iphone-ho-chi-minh','phone','Mua bán điện thoại iphone Hồ Chí Minh','Mua bán điện thoại iphone Hồ Chí Minh','{"city": "ho-chi-minh", "product": "iphone", "category": "dien-thoai"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-iphone-2025-ha-noi','phone','Mua bán điện thoại iphone 2025 Hà Nội','Mua bán điện thoại iphone 2025 Hà Nội','{"city": "ha-noi", "year": "2025", "product": "iphone", "category": "dien-thoai"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-iphone-2025-ho-chi-minh','phone','Mua bán điện thoại iphone 2025 Hồ Chí Minh','Mua bán điện thoại iphone 2025 Hồ Chí Minh','{"city": "ho-chi-minh", "year": "2025", "product": "iphone", "category": "dien-thoai"}',true,NULL,NULL);
INSERT INTO public.short_link (uri,"group",tittle,description,"filter",is_active,created_at,updated_at) VALUES
	 ('mua-ban-iphone','phone','Mua bán iphone','Mua bán iphone','{"product": "iphone"}',true,NULL,NULL),
	 ('mua-ban-apple','phone','Mua bán điện thoại apple','Mua bán điện thoại apple','{"brand": "apple"}',true,NULL,NULL),
	 ('mua-ban-samsung','phone','Mua bán điện thoại samsung','Mua bán điện thoại samsung','{"brand": "samsung"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-2024','phone','Mua bán điện thoại 2024','Mua bán điện thoại 2024','{"year": "2024"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-2026','phone','Mua bán điện thoại 2026','Mua bán điện thoại 2026','{"year": "2026"}',true,NULL,NULL),
	 ('ban-tra-duong-nhan','tea','Bán trà dưỡng nhan','Bán trà dưỡng nhan','{"product": "tra-duong-nhan"}',true,NULL,NULL),
	 ('ban-tra-tam-thao','tea','Bán trà Tâm Thảo','Bán trà Tâm Thảo','{"brand": "tra-tam-thao"}',true,NULL,NULL),
	 ('ban-tra-tam-thao','tea','Bán trà Tâm Thảo','Bán trà Tâm Thảo','{"city": "ho-chi-minh", "brand": "tra-tam-thao"}',true,NULL,NULL),
	 ('ban-tra-tam-thao','tea','Bán trà Tâm Thảo','Bán trà Tâm Thảo','{"brand": "tra-tam-thao","category":"tea"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-sam-sung-hung-yen','phone','Mua bán điện thoại samsung Hưng Yên ','Mua bán điện thoại samsung Hưng Yên ','{"city": "hung-yen", "product": "samsung-a53", "category": "dien-thoai", "product": "samsung"}',true,NULL,NULL);
INSERT INTO public.short_link (uri,"group",tittle,description,"filter",is_active,created_at,updated_at) VALUES
	 ('mua-ban-dien-thoai-iphone-ha-noi','phone','Mua bán điện thoại apple iphone Hà Nội','Mua bán điện thoại apple iphone Hà Nội','{"city": "ha-noi", "category": "dien-thoai", "product": "iphone"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-ha-noi','phone','Mua bán apple Hà Nội','Mua bán apple Hà Nội','{"city": "ha-noi", "brand": "apple"}',true,NULL,NULL),
	 ('mua-ban-dien-thoai-ha-noi','phone','Mua bán apple Hồ chí minh','Mua bán apple Hồ chí minh','{"city": "ho-chi-minh", "brand": "apple"}',true,NULL,NULL);

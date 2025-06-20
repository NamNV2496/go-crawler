CREATE TABLE IF NOT EXISTS queues (
    id SERIAL PRIMARY KEY,
    queue VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    cron VARCHAR(255),
    quantity INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);


create table if not exists urls (
    id SERIAL PRIMARY KEY,
    url text,
    method varchar(10),
    description varchar(255),
    queue varchar(255),
    domain varchar(255),
    is_active boolean,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp,
    deleted_at timestamp         
);

create table if not exists result (
    id SERIAL PRIMARY KEY,
    url text,
    method varchar(10),
    queue varchar(255),
    domain varchar(255),
    result text,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp,
    deleted_at timestamp
);

INSERT INTO public.queues (created_at,updated_at,deleted_at,queue,"domain",cron,quantity,is_active) VALUES
	 ('2025-05-22 02:02:34.684014+07','2025-05-22 02:02:34.684014+07',NULL,'normal','gold','5m',10,true),
	 ('2025-05-22 02:02:34.684014+07','2025-05-22 02:02:34.684014+07',NULL,'priority','gold','5m',100,true),
	 ('2025-05-22 02:02:34.684014+07','2025-05-22 02:02:34.684014+07',NULL,'priority','diamond','15m',20,true),
	 ('2025-05-22 02:02:34.684014+07','2025-05-22 02:02:34.684014+07',NULL,'priority','phone_cellphones','15m',200,true),
	 ('2025-05-22 02:02:34.684014+07','2025-05-22 02:02:34.684014+07',NULL,'normal','phone_thegioididong','5m',50,true),
	 ('2025-05-22 02:02:34.684014+07','2025-05-22 02:02:34.684014+07',NULL,'normal','laptop','5m',50,true),
	 ('2025-05-22 02:02:34.684014+07','2025-05-22 02:02:34.684014+07',NULL,'normal','tablet','5m',50,true),
	 ('2025-05-22 02:02:34.684014+07','2025-05-22 02:02:34.684014+07',NULL,'normal','diamond','5m',10,true);


INSERT INTO public.urls (created_at,updated_at,deleted_at,url,description,queue,"domain",is_active,"method") VALUES
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'curl --location ''https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=11'' --header ''Accept: */*'' --header ''Accept-Language: en-US,en;q=0.9,vi;q=0.8'' --header ''Connection: keep-alive'' --header ''Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn'' --header ''Sec-Fetch-Dest: empty'' --header ''Sec-Fetch-Mode: cors'' --header ''Sec-Fetch-Site: same-origin'' --header ''User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0'' --header ''sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"'' --header ''sec-ch-ua-mobile: ?0'' --header ''sec-ch-ua-platform: "macOS"'' --header ''Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1''','lấy giá vàng từ cafef','normal','gold',true,'CURL'),
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'curl --location ''https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=11'' --header ''Accept: */*'' --header ''Accept-Language: en-US,en;q=0.9,vi;q=0.8'' --header ''Connection: keep-alive'' --header ''Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn'' --header ''Sec-Fetch-Dest: empty'' --header ''Sec-Fetch-Mode: cors'' --header ''Sec-Fetch-Site: same-origin'' --header ''User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0'' --header ''sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"'' --header ''sec-ch-ua-mobile: ?0'' --header ''sec-ch-ua-platform: "macOS"'' --header ''Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1''','lấy giá vàng từ cafef','normal','diamond',true,'CURL'),
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'curl --location ''https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=13'' --header ''Accept: */*'' --header ''Accept-Language: en-US,en;q=0.9,vi;q=0.8'' --header ''Connection: keep-alive'' --header ''Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn'' --header ''Sec-Fetch-Dest: empty'' --header ''Sec-Fetch-Mode: cors'' --header ''Sec-Fetch-Site: same-origin'' --header ''User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0'' --header ''sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"'' --header ''sec-ch-ua-mobile: ?0'' --header ''sec-ch-ua-platform: "macOS"'' --header ''Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1''','lấy giá vàng từ cafef','normal','diamond',true,'CURL'),
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'curl --location ''https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=11'' --header ''Accept: */*'' --header ''Accept-Language: en-US,en;q=0.9,vi;q=0.8'' --header ''Connection: keep-alive'' --header ''Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn'' --header ''Sec-Fetch-Dest: empty'' --header ''Sec-Fetch-Mode: cors'' --header ''Sec-Fetch-Site: same-origin'' --header ''User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0'' --header ''sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"'' --header ''sec-ch-ua-mobile: ?0'' --header ''sec-ch-ua-platform: "macOS"'' --header ''Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1''','lấy giá vàng từ cafef','normal','silver',true,'CURL'),
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'curl --location ''https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=11'' --header ''Accept: */*'' --header ''Accept-Language: en-US,en;q=0.9,vi;q=0.8'' --header ''Connection: keep-alive'' --header ''Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn'' --header ''Sec-Fetch-Dest: empty'' --header ''Sec-Fetch-Mode: cors'' --header ''Sec-Fetch-Site: same-origin'' --header ''User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0'' --header ''sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"'' --header ''sec-ch-ua-mobile: ?0'' --header ''sec-ch-ua-platform: "macOS"'' --header ''Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1''','lấy giá vàng từ cafef','priority','gold',true,'CURL'),
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'curl --location ''https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=14'' --header ''Accept: */*'' --header ''Accept-Language: en-US,en;q=0.9,vi;q=0.8'' --header ''Connection: keep-alive'' --header ''Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn'' --header ''Sec-Fetch-Dest: empty'' --header ''Sec-Fetch-Mode: cors'' --header ''Sec-Fetch-Site: same-origin'' --header ''User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0'' --header ''sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"'' --header ''sec-ch-ua-mobile: ?0'' --header ''sec-ch-ua-platform: "macOS"'' --header ''Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1''','lấy giá vàng từ cafef','normal','gold',true,'CURL'),
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'https://cellphones.com.vn/iphone-16-pro-max.html','lấy giá iphone 16 pro max','normal','iphone',true,'CURL'),
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'https://cellphones.com.vn/iphone-16-pro.html','lấy giá iphone 16 pro','normal','iphone',true,'CURL'),
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'https://cellphones.com.vn/iphone-16-plus.html','lấy giá iphone 16 plus','normal','iphone',true,'CURL'),
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'https://cellphones.com.vn/robots.txt','cellphones robots.txt','normal','phone_cellphones',true,'ROBOTS'),
	 ('2025-05-22 00:28:43.412603+07','2025-05-22 00:35:09.730834+07',NULL,'https://www.thegioididong.com/robots.txt','the gioi di dong robots.txt','normal','phone_thegioididong',true,'ROBOTS');



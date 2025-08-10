

create table if not exists crawler_events (
	id int4 NOT NULL,
	url text NULL,
	"method" text NULL,
	description text NULL,
	queue text NULL,
	"domain" text NULL,
	is_active bool NULL,
	created_at timestamptz NULL,
	updated_at timestamptz NULL,
	deleted_at timestamp NULL,
	next_run_time int8 NULL,
	repeat_times int8 NULL,
	status public.status_enum NOT NULL,
	cron_exp varchar NULL,
	scheduler_at int8 NULL
);



CREATE TYPE status_enum AS ENUM ('pending', 'failed', 'successed', 'delete');

ALTER TABLE urls
ADD COLUMN status status_enum NOT NULL DEFAULT 'pending';


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

INSERT INTO public.crawler_events
(id, url, "method", description, queue, "domain", is_active, created_at, updated_at, deleted_at, next_run_time, repeat_times, status, cron_exp, scheduler_at)
VALUES(7, 'https://cellphones.com.vn/iphone-16-pro-max.html', 'CURL', 'lấy giá iphone 16 pro max', 'normal', 'iphone', true, '2025-05-22 07:28:43.412', '2025-05-22 07:35:09.730', NULL, 120000, 10000, 'pending', '*/1 * * * *', 1754842620007);
INSERT INTO public.crawler_events
(id, url, "method", description, queue, "domain", is_active, created_at, updated_at, deleted_at, next_run_time, repeat_times, status, cron_exp, scheduler_at)
VALUES(8, 'https://cellphones.com.vn/iphone-16-pro.html', 'CURL', 'lấy giá iphone 16 pro', 'normal', 'iphone', true, '2025-05-22 07:28:43.412', '2025-05-22 07:35:09.730', NULL, 120000, 10000, 'pending', '*/1 * * * *', 1754842620007);
INSERT INTO public.crawler_events
(id, url, "method", description, queue, "domain", is_active, created_at, updated_at, deleted_at, next_run_time, repeat_times, status, cron_exp, scheduler_at)
VALUES(9, 'https://cellphones.com.vn/iphone-16-plus.html', 'CURL', 'lấy giá iphone 16 plus', 'normal', 'iphone', true, '2025-05-22 07:28:43.412', '2025-05-22 07:35:09.730', NULL, 120000, 10000, 'pending', '*/1 * * * *', 1754842620007);
INSERT INTO public.crawler_events
(id, url, "method", description, queue, "domain", is_active, created_at, updated_at, deleted_at, next_run_time, repeat_times, status, cron_exp, scheduler_at)
VALUES(10, 'https://cellphones.com.vn/robots.txt', 'ROBOTS', 'cellphones robots.txt', 'normal', 'phone_cellphones', true, '2025-05-22 07:28:43.412', '2025-05-22 07:35:09.730', NULL, 120000, 10000, 'pending', '*/1 * * * *', 1754842620007);
INSERT INTO public.crawler_events
(id, url, "method", description, queue, "domain", is_active, created_at, updated_at, deleted_at, next_run_time, repeat_times, status, cron_exp, scheduler_at)
VALUES(11, 'https://www.thegioididong.com/robots.txt', 'ROBOTS', 'the gioi di dong robots.txt', 'normal', 'phone_thegioididong', true, '2025-05-22 07:28:43.412', '2025-08-10 23:10:36.108', NULL, 120000, 10000, 'pending', '*/1 * * * *', 1754842620007);
INSERT INTO public.crawler_events
(id, url, "method", description, queue, "domain", is_active, created_at, updated_at, deleted_at, next_run_time, repeat_times, status, cron_exp, scheduler_at)
VALUES(16, 'curl -L ''https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=11'' -H ''Accept: */*'' -H ''Accept-Language: en-US,en;q=0.9,vi;q=0.8'' -H ''Connection: keep-alive'' -H ''Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn'' -H ''Sec-Fetch-Dest: empty'' -H ''Sec-Fetch-Mode: cors'' -H ''Sec-Fetch-Site: same-origin'' -H ''User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0'' -H ''sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"'' -H ''sec-ch-ua-mobile: ?0'' -H ''sec-ch-ua-platform: "macOS"'' -H ''Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1''', 'CURL', 'tét', 'normal', 'gold', true, '2025-08-10 19:51:44.273', '2025-08-10 23:16:00.018', NULL, 120000, 10000, 'pending', '*/1 * * * *', 1754842620007);
INSERT INTO public.crawler_events
(id, url, "method", description, queue, "domain", is_active, created_at, updated_at, deleted_at, next_run_time, repeat_times, status, cron_exp, scheduler_at)
VALUES(15, 'curl -L ''https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=14'' -H ''Accept: */*'' -H ''Accept-Language: en-US,en;q=0.9,vi;q=0.8'' -H ''Connection: keep-alive'' -H ''Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn'' -H ''Sec-Fetch-Dest: empty'' -H ''Sec-Fetch-Mode: cors'' -H ''Sec-Fetch-Site: same-origin'' -H ''User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0'' -H ''sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"'' -H ''sec-ch-ua-mobile: ?0'' -H ''sec-ch-ua-platform: "macOS"'' -H ''Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1''', 'CURL', 'tét', 'normal', 'gold', true, '2025-08-10 19:51:44.273', '2025-08-10 23:16:00.019', NULL, 120000, 10000, 'pending', '*/1 * * * *', 1754842620007);
INSERT INTO public.crawler_events
(id, url, "method", description, queue, "domain", is_active, created_at, updated_at, deleted_at, next_run_time, repeat_times, status, cron_exp, scheduler_at)
VALUES(14, 'curl -L ''https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=13'' -H ''Accept: */*'' -H ''Accept-Language: en-US,en;q=0.9,vi;q=0.8'' -H ''Connection: keep-alive'' -H ''Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn'' -H ''Sec-Fetch-Dest: empty'' -H ''Sec-Fetch-Mode: cors'' -H ''Sec-Fetch-Site: same-origin'' -H ''User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0'' -H ''sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"'' -H ''sec-ch-ua-mobile: ?0'' -H ''sec-ch-ua-platform: "macOS"'' -H ''Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1''', 'CURL', 'lấy giá vàng từ cafe type 1', 'normal', 'gold', true, '2025-08-10 19:51:44.273', '2025-08-10 23:16:00.017', NULL, 120000, 10000, 'pending', '*/1 * * * *', 1754842680005);



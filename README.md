# Web Crawler

A simple Job scheduler + web crawler written in Go that crawls websites and extracts information.

## Main Features

- Add and manage event (in this case is crawler event)
- Crawls web pages starting from a given URL with method GET, POST and CURL
- Manage multiple queue with priority for crawlers
- Send message to Telegram
- Validator input data: handle dynamic by config base on different action config in yaml files to control require data or value range in `validator.go`

## Technologies

```
- Go: 1.24
- Redis: 7-alpine
- Kafka: 3.7.1
- PostgreSQL: 17

- Ratelimit: "github.com/go-redis/redis_rate/v9"
- Cron job: "github.com/robfig/cron/v3"
- Gorm: "gorm.io/gorm"
- redisLock: "github.com/go-redsync/redsync/v4"
```

# Architecture level 1 (branch: v1) [BASIC]

Is basic crawler 

```mermaid
flowchart LR
    conjob[conjob] --> urlservice[url-service]
    urlservice --> |"get urls"| urldb[url DB]
    urlservice -- "priority queue" --> priorityq((Kafka))
    urlservice -- "normal queue" --> normalq((Kafka))
    priorityq -- "priority queue" --> worker
    normalq -- "normal queue" --> worker
    worker[worker<br>ratelimiter] -- "publish result" --> resultq((Kafka))
    resultq -- "consume" --> db[(DB)]
    worker -- "notify through tele if<br>match with rules" --> telegram[Telegram]
    worker[worker<br>ratelimiter] -- update statue --> urlservice
```

![alt text](docs/design1.png)

# Architecture level 2 (branch: v2) [FINAL]

In Architecture level 1, we only query and execute 1 time. Let expand the problem with harder question
- Can we execute for daily, monthly job?
- As current we run sequencely job because a job is quite short response. What happend if long running job? example: download a large file, run a large step. => How to handle it? And in running time of that job, can we run another job?
- In actually, crawler include scheduler problem => split to scheduler + crawler

# FINAL DESIGN

![alt text](docs/design.png)

<details>

```txt
title Crawler
sequenceDiagram
    participant scheduler-service
    participant DB
    participant Kafka
    participant crawler-service-woker
    participant result-DB

    scheduler-service ->> scheduler-service: normal queue run every 1m
    scheduler-service ->> DB: get list of events
    scheduler-service ->> Kafka: public event
    crawler-service-woker -->> Kafka: consume message
    
    crawler-service-woker ->> crawler-service-woker: execute event
    crawler-service-woker ->> result-DB: write data to result DB
```
</details>

![alt text](docs/sequence.png)

# How to run

Setup your environment variables in `.env.example` and change it to `.env`

```bash
# Start docker
docker-compose up -d
```

```bash
# Terminal 1
cd crawler-service
go run main.go crawler-worker

# Terminal 2

cd scheduler-service
go run main.go scheduler

# Terminal 3

cd scheduler-service
go run main.go scheduler_worker

```

<details>

# 1. Create new bot and get token

![alt text](docs/create_bot.png)
![alt text](docs/create_group_chat.png)

# 2. Run command to get chat Id

```bash
curl -s https://api.telegram.org/bot${TOKEN}/getUpdates
```

![alt text](docs/tele_message.png)

# 3. Worker consumer message example

```txt
publish to crawler queue: normal, request: curl --location 'https://m.cafef.vn/du-lieu/Ajax/ajaxgoldprice.ashx?index=11' --header 'Accept: */*' --header 'Accept-Language: en-US,en;q=0.9,vi;q=0.8' --header 'Connection: keep-alive' --header 'Referer: https://m.cafef.vn/du-lieu/gia-vang-hom-nay/trong-nuoc.chn' --header 'Sec-Fetch-Dest: empty' --header 'Sec-Fetch-Mode: cors' --header 'Sec-Fetch-Site: same-origin' --header 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0' --header 'sec-ch-ua: "Microsoft Edge";v="135", "Not-A.Brand";v="8", "Chromium";v="135"' --header 'sec-ch-ua-mobile: ?0' --header 'sec-ch-ua-platform: "macOS"' --header 'Cookie: _ga=GA1.2.1174992577.1733489327; _ga_860L8F5EZP=GS1.1.1740282133.10.0.1740282328.0.0.0; ASP.NET_SessionId=wnors2tpgmcb0lwvqwebtsf5; favorite_stocks_state=1'
```
</details>

# 4. Validator example

## 4.1 Missing require field

```bash
# condition always force require: `"require": "1"`
{
    "event": {
        "method": "GET"
    }
}
# output: Vui lòng nhập thông tin "url"

# condition only require when match condition of another field: `"method": "GET"`
{
    "event": {
        "method": "GET",
        "url": "quis voluptate"
    }
}
# output: Vui lòng nhập thông tin "cron_exp"

# condition field cron_exp is not require because "method" is not equal "GET"
{
    "event": {
        "method": "POST",
        // "cron_exp": "dolor id dolore", // not require
        "created_at": "eu sit",
        "description": "eu ea dolore",
        "domain": "officia do Lorem Ut non",
        "id": "irure mollit in deserunt adipisicing",
        "is_active": true,
        "next_run_time": "84436197",
        "queue": "reprehenderit",
        "repeat_times": "92953379",
        // "scheduler_at": "9",
        "status": "mollit nostrud aute ut",
        "updated_at": "minim dolore quis esse cupidatat",
        "url": "quis voluptate"
    }
}
# output: pass
```

## 4.2 Input validation

```bash
# invalid value relate min/max value
{
    "event": {
        "repeat_times": "92953379", // invalid value with rule: "max_value": 100, "min_value": 2
        "created_at": "eu sit",

        "description": "eu ea dolore ádf à à",
        "domain": "officia do Lorem Ut non",
        "id": "irure mollit in deserunt adipisicing",
        "is_active": true,
        "method": "POST",
        "next_run_time": "84436197",
        "queue": "reprehenderit",
        "status": "mollit nostrud aute ut",
        "updated_at": "minim dolore quis esse cupidatat",
        "url": "quis voluptate"
    }
}
# output: "Số lần lặp" is required <= 100

# invalid number of words relate min/max value

{
    "event": {
        "description": "eu ea dolore", // min word is 5

        "repeat_times": "100",
        "created_at": "eu sit",
        "domain": "officia do Lorem Ut non",
        "id": "irure mollit in deserunt adipisicing",
        "is_active": true,
        "method": "POST",
        "next_run_time": "84436197",
        "queue": "reprehenderit",
        "status": "mollit nostrud aute ut",
        "updated_at": "minim dolore quis esse cupidatat",
        "url": "quis voluptate"
    }
}
# output: "Mô tả" is required >= 5 words

# invalid length of string relate min/max value
{
    "event": {
        "url": "q", // min length is 2

        "description": "eu ea dolore asfa ads fa",
        "repeat_times": "100",
        "created_at": "eu sit",
        "domain": "officia do Lorem Ut non",
        "id": "irure mollit in deserunt adipisicing",
        "is_active": true,
        "method": "POST",
        "next_run_time": "84436197",
        "queue": "reprehenderit",
        "status": "mollit nostrud aute ut",
        "updated_at": "minim dolore quis esse cupidatat"
    }
}

# output: "Đường dẫn" is required >= 2 characters

# invalid length of string relate min/max value
{
    "event": {
        "url": "qádfasdfjkasdfj;klasfjlkfjalksfjjadls;kfjkladfsjdlsakfjdlkdsajfl;ạdsladsjl;kadsjadsl;fjd", // max length is 50

        "description": "eu ea dolore asfa ads fa",
        "repeat_times": "100",
        "created_at": "eu sit",
        "domain": "officia do Lorem Ut non",
        "id": "irure mollit in deserunt adipisicing",
        "is_active": true,
        "method": "POST",
        "next_run_time": "84436197",
        "queue": "reprehenderit",
        "status": "mollit nostrud aute ut",
        "updated_at": "minim dolore quis esse cupidatat"
    }
}
# output: "Đường dẫn" is required <= 50 characters
```








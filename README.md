# Web Crawler

A simple web crawler written in Go that crawls websites and extracts information.

## Features

- Crawls web pages starting from a given URL with method GET, POST and CURL
- Manage multiple queue with priority for crawlers
- Run by cron job
- Send message to Telegram
- Extracts page titles
- Follows links to discover more pages

## Installation

```bash
git clone https://github.com/NamNV2496/crawler.git
```

# Architecture

![alt text](docs/design.png)


# How to run

```bash
# Start docker
docker-compose up -d
```

```bash
cd crawler-service

# Terminal 1
go run main.go server

make service


# Terminal 2
go run main.go crawler-worker

make worker
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

</details>

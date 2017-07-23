# slack-logger

SlackのPublicなチャンネルのログをチャンネルごとに収集する。<br />
収集したログはSqlite3で保存される。

## 前提

- Docker
- docker-compose

## 使い方

1. Slack API Token を取得する
1. `docker-compose.yml` の `TOKEN=` に取得したものを記述
1. `docker-compose up` を実行
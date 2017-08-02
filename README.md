# slack-logger

Slackのチャンネル毎にメッセージをSQLite3に保存する。<br />
チャンネルはパブリックなもののみで保存される。<br />
また、1回目以降の実行では最新メッセージのみ保存する。<br />

## 前提

- Docker
- docker-compose

## 使い方

### docker-composeを利用

1. Slack API Token を取得する
1. `docker-compose.yml` の `TOKEN=` に取得したものを記述
1. `docker-compose.yml` の `CHANNEL=` に通知したいチャンネル名を記述
1. `docker-compose.yml` の `TIME=` に起動される時刻を記述 (例：`04:00`)
1. `docker-compose up` を実行
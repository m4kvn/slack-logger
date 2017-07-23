# slack-logger

Slackのチャンネル毎にメッセージをSQLite3に保存する。<br />
チャンネルはパブリックなもののみで保存される。<br />
また、2回目以降の実行では前回の保存からの最新メッセージのみ保存する。<br />

## 前提

- Docker
- docker-compose

## 使い方

1. Slack API Token を取得する
1. `docker-compose.yml` の `TOKEN=` に取得したものを記述
1. `docker-compose up` を実行
1. `slack.db` にデータが保存される

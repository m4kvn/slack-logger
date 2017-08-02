CREATE DATABASE IF NOT EXISTS slack_logger;

CREATE TABLE IF NOT EXISTS channels (
  channel_id   TINYTEXT NOT NULL,
  channel_name TINYTEXT NOT NULL,
  last_update   TINYTEXT,
  PRIMARY KEY (channel_id(255)),
  INDEX channels_idx(channel_id(255))
);

CREATE TABLE IF NOT EXISTS history (
  channel_id TINYTEXT NOT NULL,
  type       TINYTEXT NOT NULL,
  user       TINYTEXT NOT NULL,
  text       TEXT,
  ts         TINYTEXT,
  INDEX history_idx(channel_id(255), ts(255))
);
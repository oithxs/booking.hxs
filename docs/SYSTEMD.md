# ⚙️ systemd セットアップガイド

このドキュメントでは、Discord Bot 部室予約システムをLinuxサーバーで**systemdサービス**として登録し、自動起動させる方法を説明します。

## 📑 目次

- [前提条件](#前提条件)
- [セットアップ方法](#セットアップ方法)
- [環境変数の設定](#環境変数の設定)
- [サービスの管理](#サービスの管理)
- [ログの確認](#ログの確認)
- [トラブルシューティング](#トラブルシューティング)
- [サービスファイルリファレンス](#サービスファイルリファレンス)



## 前提条件

### 必須

- ✅ Linuxサーバー（Ubuntu, Debian, CentOS, RHEL など）
- ✅ systemd がインストールされていること
- ✅ Go 1.21以上がインストールされていること
- ✅ プロジェクトがクローンされていること
- ✅ `.env` ファイルに必要な環境変数が設定されていること

### 確認方法

```bash
# systemdのバージョン確認
systemctl --version

# Goのバージョン確認
go version

# プロジェクトディレクトリの確認
ls -la /home/hxs/booking.hxs
```


## セットアップ方法

### 方法A: 自動セットアップ（推奨）

自動セットアップスクリプトを使用します。

```bash
# プロジェクトディレクトリに移動
cd /home/hxs/booking.hxs

# セットアップスクリプトを実行
./setup-systemd.sh
```

**このスクリプトが実行すること**:
1. ✅ バイナリのビルド
2. ✅ サービスファイルのコピー
3. ✅ systemdの設定を再読み込み
4. ✅ サービスの有効化と起動

**実行後に必要な作業**:
- サービスファイルを編集して環境変数を設定（下記参照）

---

### 方法B: 手動セットアップ

#### ステップ1: バイナリのビルド

```bash
# プロジェクトディレクトリに移動
cd /home/hxs/booking.hxs

# ビルド
make build

# または
go build -o bin/booking.hxs cmd/bot/main.go

# ビルド成功を確認
ls -lh bin/booking.hxs
```

#### ステップ2: サービスファイルのコピー

```bash
# サービスファイルをsystemdディレクトリにコピー
sudo cp config/booking-hxs.service /etc/systemd/system/

# ファイルがコピーされたか確認
ls -l /etc/systemd/system/booking-hxs.service
```

#### ステップ3: サービスファイルを編集

```bash
sudo nano /etc/systemd/system/booking-hxs.service
```

以下の項目を**必ず**カスタマイズしてください：

```ini
[Service]
# ⚠️ ユーザー名を変更（実際のユーザー名に）
User=hxs

# ⚠️ 作業ディレクトリを変更（実際のパスに）
WorkingDirectory=/home/hxs/booking.hxs

# ⚠️ 実行ファイルのパスを変更（実際のパスに）
ExecStart=/home/hxs/booking.hxs/bin/booking.hxs

# ⚠️ 環境変数を設定（後述）
Environment="DISCORD_TOKEN=your_actual_token_here"
Environment="GUILD_ID=your_guild_id_here"
Environment="FEEDBACK_CHANNEL_ID=your_channel_id_here"
```

#### ステップ4: systemd設定を再読み込み

```bash
sudo systemctl daemon-reload
```

#### ステップ5: サービスを有効化して起動

```bash
# 自動起動を有効化
sudo systemctl enable booking-hxs

# サービスを起動
sudo systemctl start booking-hxs

# 状態を確認
sudo systemctl status booking-hxs
```



## 環境変数の設定

systemdサービスは通常の `.env` ファイルを自動では読み込みません。
以下のいずれかの方法で環境変数を設定してください。

### 方法A: サービスファイルに直接記述（推奨）

**メリット**: シンプルで確実<br>
**デメリット**: サービスファイルを編集する必要がある

```bash
sudo nano /etc/systemd/system/booking-hxs.service
```

`[Service]` セクションに追加：

```ini
[Service]
Environment="DISCORD_TOKEN=MTIzNDU2Nzg5MDEyMzQ1Njc4OQ.GaBcDe.FgHiJkLmNoPqRsTuVwXyZ"
Environment="GUILD_ID=987654321098765432"
Environment="FEEDBACK_CHANNEL_ID=111222333444555666"
Environment="STARTUP_NOTIFICATION_CHANNEL_ID=111222333444555666"
Environment="STARTUP_NOTIFICATION_MESSAGE=🚀 Bot が再起動しました。"
Environment="ENV=production"
```

**環境変数の説明**:
- `DISCORD_TOKEN`: Discord Botトークン（必須）
- `GUILD_ID`: サーバーID（必須）
- `FEEDBACK_CHANNEL_ID`: フィードバック送信先チャンネルID（必須）
- `STARTUP_NOTIFICATION_CHANNEL_ID`: 起動通知送信先チャンネルID（オプション、空欄で無効化）
- `STARTUP_NOTIFICATION_MESSAGE`: 起動時のカスタムメッセージ（オプション、空欄でデフォルトメッセージ）
- `ENV`: 環境設定（`production` または `development`）

変更後、設定を再読み込み：

```bash
sudo systemctl daemon-reload
sudo systemctl restart booking-hxs
```

---

### 方法B: 環境ファイルを使用

**メリット**: 設定を外部ファイルで管理できる<br>
**デメリット**: ファイルのパーミッション管理が必要

#### 1. 本番環境用の.envファイルを準備

```bash
cd /home/hxs/booking.hxs

# 本番環境に切り替え
./switch_env.sh production

# または直接コピー
cp config/.env.production .env

# トークンなどを実際の値に編集
nano .env
```

#### 2. サービスファイルでEnvironmentFileを有効化

```bash
sudo nano /etc/systemd/system/booking-hxs.service
```

以下の行のコメントを**解除**：

```ini
[Service]
# ⚠️ この行のコメントを解除
EnvironmentFile=/home/hxs/booking.hxs/.env
```

#### 3. .envファイルのパーミッションを設定

```bash
# 所有者のみが読み書きできるように設定
chmod 600 .env

# 所有者をサービス実行ユーザーに変更
sudo chown hxs:hxs .env
```

#### 4. 設定を再読み込みして再起動

```bash
sudo systemctl daemon-reload
sudo systemctl restart booking-hxs
```


## サービスの管理

### 基本コマンド

```bash
# サービスを起動
sudo systemctl start booking-hxs

# サービスを停止
sudo systemctl stop booking-hxs

# サービスを再起動
sudo systemctl restart booking-hxs

# サービスの状態を確認
sudo systemctl status booking-hxs

# 自動起動を有効化
sudo systemctl enable booking-hxs

# 自動起動を無効化
sudo systemctl disable booking-hxs
```

### サービスの状態確認

#### 詳細な状態を確認

```bash
sudo systemctl status booking-hxs
```

**出力例（正常時）**:
```
● booking-hxs.service - Booking.hxs - Discord Reservation Bot
     Loaded: loaded (/etc/systemd/system/booking-hxs.service; enabled; vendor preset: enabled)
     Active: active (running) since Sat 2025-11-09 10:00:00 JST; 5h 23min ago
   Main PID: 12345 (booking.hxs)
      Tasks: 8 (limit: 4915)
     Memory: 25.3M
        CPU: 1.234s
     CGroup: /system.slice/booking-hxs.service
             └─12345 /home/hxs/booking.hxs/bin/booking.hxs

Nov 09 10:00:00 server systemd[1]: Started Booking.hxs - Discord Reservation Bot.
Nov 09 10:00:01 server booking.hxs[12345]: Reservations loaded successfully
Nov 09 10:00:01 server booking.hxs[12345]: Bot is now running. Press CTRL+C to exit.
```

#### 自動起動の状態を確認

```bash
sudo systemctl is-enabled booking-hxs
# 出力: enabled（有効）または disabled（無効）
```

#### 起動中かどうかを確認

```bash
sudo systemctl is-active booking-hxs
# 出力: active（起動中）または inactive（停止中）
```


## ログの確認

### systemdのログを確認

#### リアルタイムでログを監視

```bash
sudo journalctl -u booking-hxs -f
```

#### 最新のログを表示

```bash
# 最新の50行
sudo journalctl -u booking-hxs -n 50

# 最新の100行
sudo journalctl -u booking-hxs -n 100
```

#### 特定の時間帯のログを表示

```bash
# 今日のログ
sudo journalctl -u booking-hxs --since today

# 直近1時間のログ
sudo journalctl -u booking-hxs --since "1 hour ago"

# 特定の日時以降のログ
sudo journalctl -u booking-hxs --since "2025-11-09 10:00:00"
```

#### エラーログのみを表示

```bash
sudo journalctl -u booking-hxs -p err
```

### アプリケーションログを確認

```bash
# コマンドログ
tail -f /home/hxs/booking.hxs/logs/commands_2025-11.log

# 統計情報
cat /home/hxs/booking.hxs/logs/command_stats.json | jq .
```


## トラブルシューティング

### サービスが起動しない

#### 症状
```
Failed to start booking-hxs.service: Unit booking-hxs.service not found.
```

#### 解決方法
1. サービスファイルが正しい場所にあるか確認
   ```bash
   ls -l /etc/systemd/system/booking-hxs.service
   ```

2. systemd設定を再読み込み
   ```bash
   sudo systemctl daemon-reload
   ```

---

### 環境変数が読み込まれない

#### 症状
```
DISCORD_TOKEN is not set in environment variables
```

#### 解決方法

**方法1: サービスファイルに直接記述**
```bash
sudo nano /etc/systemd/system/booking-hxs.service
```

```ini
[Service]
Environment="DISCORD_TOKEN=your_token_here"
Environment="GUILD_ID=your_guild_id_here"
```

**方法2: .envファイルのパスを確認**
```bash
# パスが正しいか確認
cat /etc/systemd/system/booking-hxs.service | grep EnvironmentFile

# .envファイルが存在するか確認
ls -l /home/hxs/booking.hxs/.env
```

設定変更後：
```bash
sudo systemctl daemon-reload
sudo systemctl restart booking-hxs
```

---

### パーミッションエラー

#### 症状
```
Permission denied
```

#### 解決方法
```bash
# バイナリに実行権限を付与
chmod +x /home/hxs/booking.hxs/bin/booking.hxs

# 予約データファイルの権限を確認
ls -la /home/hxs/booking.hxs/data/reservations.json
chmod 644 /home/hxs/booking.hxs/data/reservations.json

# ユーザーの所有権を確認
sudo chown -R hxs:hxs /home/hxs/booking.hxs
```

---

### サービスがクラッシュする

#### ログを確認

```bash
# 詳細なログを確認
sudo journalctl -u booking-hxs -n 200 --no-pager

# エラーログのみ
sudo journalctl -u booking-hxs -p err --no-pager
```

#### よくある原因

1. **バイナリが古い**
   ```bash
   cd /home/hxs/booking.hxs
   make build
   sudo systemctl restart booking-hxs
   ```

2. **依存関係の問題**
   ```bash
   go mod download
   make build
   ```

3. **ディスク容量不足**
   ```bash
   df -h
   ```


## サービスファイルリファレンス

### 基本設定

```ini
[Unit]
Description=Booking.hxs - Discord Reservation Bot
After=network.target

[Service]
Type=simple
User=hxs
WorkingDirectory=/home/hxs/booking.hxs
ExecStart=/home/hxs/booking.hxs/bin/booking.hxs
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### カスタマイズ可能な項目

| 項目 | 説明 | 推奨値 |
|------|------|--------|
| `User` | サービスを実行するユーザー | 実際のユーザー名 |
| `WorkingDirectory` | 作業ディレクトリ | プロジェクトのルートパス |
| `ExecStart` | 実行コマンド | バイナリの絶対パス |
| `Restart` | 再起動ポリシー | `always`, `on-failure` |
| `RestartSec` | 再起動までの待機時間（秒） | `10` |

### 環境変数の設定方法

#### 直接記述
```ini
[Service]
Environment="DISCORD_TOKEN=your_token"
Environment="GUILD_ID=your_guild_id"
Environment="FEEDBACK_CHANNEL_ID=your_channel_id"
```

#### ファイルから読み込み
```ini
[Service]
EnvironmentFile=/home/hxs/booking.hxs/.env
```


## 運用Tips

### 自動更新スクリプト

```bash
#!/bin/bash
# update-bot.sh

cd /home/hxs/booking.hxs

# 最新コードを取得
git pull

# ビルド
make build

# サービスを再起動
sudo systemctl restart booking-hxs

# ステータス確認
sudo systemctl status booking-hxs
```

### 定期的なログ確認

```bash
# cronで毎日ログをチェック
0 9 * * * journalctl -u booking-hxs --since "24 hours ago" -p err > /tmp/bot-errors.log
```

### バックアップスクリプト

```bash
#!/bin/bash
# backup-data.sh

BACKUP_DIR="/home/hxs/backups"
DATE=$(date +%Y%m%d)

# 予約データをバックアップ
cp /home/hxs/booking.hxs/data/reservations.json $BACKUP_DIR/reservations_$DATE.json

# 7日以上前のバックアップを削除
find $BACKUP_DIR -name "reservations_*.json" -mtime +7 -delete
```


## まとめ

systemdサービスとして設定することで：

✅ **自動起動** - サーバー起動時に自動的にBotが起動
✅ **自動再起動** - クラッシュしても自動的に再起動
✅ **ログ管理** - journalctlで簡単にログ確認
✅ **サービス管理** - systemctlで簡単に管理
✅ **本番運用** - 安定した運用が可能

---

**関連ドキュメント**: [README](../README.md) | [起動ガイド](SETUP.md) | [コマンド](COMMANDS.md) | [データ管理](DATA_MANAGEMENT.md) | [開発](DEVELOPMENT.md)

# 📖 アプリの起動ガイド

このガイドでは、Discord Bot 部室予約システムのセットアップから起動までを詳しく説明します。

## 📑 目次

- [前提条件](#前提条件)
- [Discord Botの作成](#discord-botの作成)
- [プロジェクトのセットアップ](#プロジェクトのセットアップ)
- [環境変数の設定](#環境変数の設定)
- [Botの起動](#botの起動)
- [環境の切り替え](#環境の切り替え)
- [トラブルシューティング](#トラブルシューティング)

## 前提条件

### 必須

- **Go 1.21以上** がインストールされていること
- **Discord Botトークン** を取得済みであること
- **Git** がインストールされていること（クローン用）

### 確認方法

```bash
# Goのバージョンを確認
go version
# 出力例: go version go1.21.0 linux/amd64

# Gitのバージョンを確認
git --version
# 出力例: git version 2.34.1
```

## Discord Botの作成

### 1. Discord Developer Portalでアプリケーションを作成

1. [Discord Developer Portal](https://discord.com/developers/applications) にアクセス
2. **「New Application」** をクリック
3. アプリケーション名を入力（例: `面接予約Bot`）
4. **「Create」** をクリック

### 2. Botを作成

1. 左側のメニューから **「Bot」** を選択
2. **「Add Bot」** をクリック
3. **「Yes, do it!」** で確認

### 3. Botトークンを取得

1. **「TOKEN」** セクションで **「Reset Token」** をクリック
2. 表示されたトークンをコピー（**重要: 一度しか表示されません**）
3. トークンを安全な場所に保存

### 4. Privileged Gateway Intentsを有効化

**「Privileged Gateway Intents」** セクションで以下を有効化：

- ✅ **Server Members Intent**
- ✅ **Message Content Intent**

「Save Changes」をクリックして保存。

### 5. Botをサーバーに招待

#### OAuth2 URLを生成

1. 左側のメニューから **「OAuth2」** → **「URL Generator」** を選択
2. **「SCOPES」** で以下を選択：
   - ✅ `bot`
   - ✅ `applications.commands`
3. **「BOT PERMISSIONS」** で以下を選択：
   - ✅ Send Messages
   - ✅ Read Message History
   - ✅ Use Slash Commands
   - ✅ Embed Links（推奨）

#### 招待URLをコピー

1. ページ下部の **「GENERATED URL」** をコピー
2. ブラウザで開く
3. Botを追加したいサーバーを選択
4. **「認証」** をクリック

### 6. サーバーIDとチャンネルIDを取得

#### Discordで開発者モードを有効化

1. Discord設定を開く
2. **「詳細設定」** → **「開発者モード」** を有効化

#### サーバーIDを取得

1. サーバー名を右クリック
2. **「IDをコピー」** を選択

#### フィードバックチャンネルIDを取得（オプション）

1. フィードバックを受け取りたいチャンネルを右クリック
2. **「IDをコピー」** を選択


## プロジェクトのセットアップ

### 方法1: 自動セットアップ（推奨）

```bash
# プロジェクトディレクトリに移動
cd booking.hxs

# セットアップスクリプトを実行
./setup.sh
```

このスクリプトは以下を自動実行します：
- ✅ Goバージョンの確認
- ✅ 依存関係のダウンロード
- ✅ `.env` ファイルの作成
- ✅ ビルドテスト

### 方法2: Makefileを使用

```bash
make setup
```

### 方法3: 手動セットアップ

```bash
# 依存関係をダウンロード
go mod download

# .envファイルを作成
cp config/.env.example .env

# ビルドディレクトリを作成
mkdir -p bin logs

# ビルドテスト
go build -o bin/booking.hxs cmd/bot/main.go
```


## 環境変数の設定

### .envファイルを編集

```bash
vi .env
# または
nano .env
```

### 必須項目

```env
# Discord Bot Token（必須）
DISCORD_TOKEN=your_discord_bot_token_here

# Guild ID（推奨）
GUILD_ID=your_guild_id_here

# Allowed Channel ID（推奨）
ALLOWED_CHANNEL_ID=your_allowed_channel_id_here

# Feedback Channel ID（オプション）
FEEDBACK_CHANNEL_ID=your_feedback_channel_id_here

# Startup Notification Channel ID（オプション）
STARTUP_NOTIFICATION_CHANNEL_ID=

# Startup Notification Message（オプション）
STARTUP_NOTIFICATION_MESSAGE=
```

### 各項目の説明

| 環境変数 | 説明 | 必須/オプション |
|---------|------|---------------|
| `DISCORD_TOKEN` | Discord Developer Portalで取得したBotトークン | **必須** |
| `GUILD_ID` | テスト用サーバーのID。設定するとそのサーバー専用コマンドとして即座に登録される。空欄ならグローバルコマンド（反映に最大1時間） | 推奨 |
| `ALLOWED_CHANNEL_ID` | コマンドを受け付けるチャンネルのID。設定すると、そのチャンネルとDMでのみコマンドが動作します。DMから実行された場合、公開メッセージはこのチャンネルに送信されます。 | 推奨 |
| `FEEDBACK_CHANNEL_ID` | `/feedback` コマンドで送信されたフィードバックを受け取るチャンネルのID。設定しない場合、`/feedback` コマンドは使用不可 | オプション |
| `STARTUP_NOTIFICATION_CHANNEL_ID` | Bot起動時に通知メッセージを送信するチャンネルのID。空欄で無効化（systemdでの自動再起動時に便利） | オプション |
| `STARTUP_NOTIFICATION_MESSAGE` | Bot起動時のカスタムメッセージ。空欄の場合はデフォルトメッセージ「🚀 Bot が起動しました。部室予約システムが利用可能です。」が使用される | オプション |





## Botの起動

### 方法1: Makefileを使用（推奨）

```bash
# 開発モードで実行
make run

# ビルドしてから実行
make build
make start

# または一度に
make start
```

### 方法2: Go runで直接実行

```bash
go run cmd/bot/main.go
```

### 方法3: ビルドしてから実行

```bash
# ビルド
go build -o bin/booking.hxs cmd/bot/main.go

# 実行
./bin/booking.hxs
```

### 起動成功の確認

以下のようなログが表示されれば成功です：

```
2025/11/14 20:03:57 Reservations loaded successfully
2025/11/14 20:03:57 Logger initialized successfully
2025/11/14 20:03:58 Bot is now running. Press CTRL+C to exit.
2025/11/14 20:03:58 Removing existing commands...
2025/11/14 20:03:59 Deleted existing guild command: reserve
2025/11/14 20:03:59 Deleted existing guild command: cancel
2025/11/14 20:04:00 Deleted existing guild command: complete
2025/11/14 20:04:00 Deleted existing guild command: edit
2025/11/14 20:04:00 Deleted existing guild command: list
2025/11/14 20:04:19 Deleted existing guild command: my-reservations
2025/11/14 20:04:20 Deleted existing guild command: help
2025/11/14 20:04:20 Deleted existing guild command: feedback
2025/11/14 20:04:20 Registering new commands...
2025/11/14 20:04:20 ✅ Registered command: reserve
2025/11/14 20:04:21 ✅ Registered command: cancel
2025/11/14 20:04:21 ✅ Registered command: complete
2025/11/14 20:04:21 ✅ Registered command: edit
2025/11/14 20:04:21 ✅ Registered command: list
2025/11/14 20:04:41 ✅ Registered command: my-reservations
2025/11/14 20:04:41 ✅ Registered command: help
2025/11/14 20:04:41 ✅ Registered command: feedback
2025/11/14 20:04:41 Command registration completed
2025/11/14 20:04:41 Startup: Running initial auto-complete check...
2025/11/14 20:04:41 Startup: Running initial cleanup check...
2025/11/14 20:04:41 ✅ Auto-completed 1 expired reservation(s) and saved
2025/11/14 20:04:41 ✓ Cleanup check completed: no old reservations to remove
2025/11/14 20:04:41 Next auto-complete scheduled at: 2025-11-15 03:00:00 (in 6h55m18.2348028s)
2025/11/14 20:04:41 Next cleanup scheduled at: 2025-11-15 03:10:00 (in 7h5m18.234786329s)

```

### Botの停止

`Ctrl+C` を押すと、安全にシャットダウンします：

```
^CSaving reservations before exit...
Reservations saved successfully
=== コマンド統計 ===
総コマンド数: 15
...
```


## 環境の切り替え

### 開発環境と本番環境

このプロジェクトは、開発環境と本番環境を分離して管理できます。

### 環境切り替えスクリプトを使用

```bash
# 開発環境に切り替え
./switch_env.sh development

# 本番環境に切り替え
./switch_env.sh production

# 現在の環境を確認
./switch_env.sh status
```

### 環境ごとの設定ファイル

| ファイル | 説明 |
|---------|------|
| `.env.development` | 開発環境用の設定 |
| `.env.production` | 本番環境用の設定 |
| `.env` | 現在使用中の設定（.developmentまたは.productionのコピー） |

### 設定例

**`.env.development` (開発環境)**
```env
DISCORD_TOKEN=dev_token_here
GUILD_ID=dev_server_id
FEEDBACK_CHANNEL_ID=dev_feedback_channel_id
ENV=development
DATA_FILE=reservations_dev.json
```

**`.env.production` (本番環境)**
```env
DISCORD_TOKEN=prod_token_here
GUILD_ID=
FEEDBACK_CHANNEL_ID=prod_feedback_channel_id
ENV=production
```

**注**: `DATA_FILE` 環境変数は使用されません。データは常に `data/reservations.json` に保存されます。


## ホットリロード（開発効率化）

開発時に、ファイルの変更を自動検知して再起動する機能を利用できます。

### airのインストール

```bash
go install github.com/cosmtrek/air@latest
```

### ホットリロードで起動

```bash
make dev
```

または

```bash
air
```

ファイルを編集すると、自動的に再ビルド＆再起動されます。


## 次のステップ

起動が成功したら、以下のドキュメントもご覧ください：

- **[コマンドリファレンス](COMMANDS.md)** - すべてのコマンドの使い方
- **[データの取り扱い](DATA_MANAGEMENT.md)** - データ管理とクリーンアップ
- **[systemdセットアップ](SYSTEMD.md)** - 本番環境での自動起動
- **[開発者ガイド](DEVELOPMENT.md)** - カスタマイズと拡張


## よく使うコマンドまとめ

```bash
# セットアップ
./setup.sh                    # 初回セットアップ
make setup                    # または Makefile経由

# 起動
make run                      # 開発モードで実行
make dev                      # ホットリロード
make start                    # ビルド＋実行

# 環境切り替え
./switch_env.sh development   # 開発環境
./switch_env.sh production    # 本番環境
./switch_env.sh status        # 現在の環境確認

# コード品質
make fmt                      # フォーマット
make vet                      # 静的解析
make check                    # fmt + vet

# その他
make help                     # コマンド一覧
make clean                    # クリーンアップ
```


## トラブルシューティング

### Botが起動しない

#### 症状
```
Failed to create Discord session: HTTP 401 Unauthorized, {"message": "401: Unauthorized", "code": 0}
```

#### 原因と解決方法
- **原因**: `DISCORD_TOKEN` が正しくない
- **解決方法**:
  1. Discord Developer Portalでトークンをリセット
  2. 新しいトークンを `.env` に設定
  3. Botを再起動

---

### コマンドが表示されない

#### 症状
Discordでスラッシュコマンドが表示されない

#### 原因と解決方法

**1. Botの権限不足**
- **確認**: OAuth2 URLで「Use Slash Commands」が選択されているか
- **解決方法**: Botを再招待する

**2. GUILD_IDの設定ミス**
- **確認**: `.env` の `GUILD_ID` が正しいか
- **解決方法**: サーバーを右クリック→「IDをコピー」で正しいIDを取得

**3. グローバルコマンドの反映待ち**
- **確認**: `GUILD_ID` が空欄の場合
- **解決方法**: 最大1時間待つか、`GUILD_ID` を設定して即座に反映

---

### フィードバックコマンドが使えない

#### 症状
```
❌ フィードバックチャンネルが設定されていません。管理者に連絡してください。
```

#### 原因と解決方法
- **原因**: `FEEDBACK_CHANNEL_ID` が設定されていない
- **解決方法**:
  1. フィードバックを受け取るチャンネルを右クリック→「IDをコピー」
  2. `.env` に `FEEDBACK_CHANNEL_ID=コピーしたID` を追加
  3. Botを再起動

---

### 予約が保存されない

#### 症状
Botを再起動すると予約が消える

#### 原因と解決方法

**1. ファイル権限の問題**
```bash
# 確認
ls -la data/reservations.json

# 解決方法（書き込み権限を付与）
chmod 644 data/reservations.json
```

**2. ディスク容量不足**
```bash
# 確認
df -h

# 解決方法: 不要なファイルを削除
```

---

### 依存関係のエラー

#### 症状
```
missing go.sum entry for module providing package
```

#### 解決方法
```bash
# Go モジュールを整理
go mod tidy

# 依存関係を再ダウンロード
go mod download

# または依存関係管理スクリプトを使用
./manage_deps.sh clean
./manage_deps.sh install
```

---

### ビルドエラー

#### 症状
```
undefined: XXX
```

#### 解決方法
```bash
# コードをフォーマット
make fmt

# 静的解析
make vet

# クリーンビルド
make clean
make build
```

---

### ログを確認する

問題が発生した場合、ログを確認することで原因を特定できます。

```bash
# 最新のログファイルを表示
ls -lt logs/

# ログの内容を確認
cat logs/commands_2025-11.log

# リアルタイムでログを監視
tail -f logs/commands_2025-11.log
```

---
**関連ドキュメント**: [README](../README.md) | [コマンド](COMMANDS.md) | [データ管理](DATA_MANAGEMENT.md) | [systemd](SYSTEMD.md) | [開発](DEVELOPMENT.md)

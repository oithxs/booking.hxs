# 🤖 AI コンテキストドキュメント

このドキュメントは、新しいAIセッションでこのプロジェクトに関する指示を行う際に、AIに読んでもらうための包括的なコンテキスト情報です。

**最終更新**: 2025年11月17日
**プロジェクト**: booking.hxs - Discord Bot 部室予約システム
**現在のバージョン**: v1.3.3

---

## 📋 目次

1. [プロジェクト概要](#-プロジェクト概要)
2. [技術スタック](#-技術スタック)
3. [プロジェクト構造](#-プロジェクト構造)
4. [開発ルール・コーディング規約](#-開発ルールコーディング規約)
5. [重要な実装パターン](#-重要な実装パターン)
6. [環境変数](#-環境変数)
7. [コマンド一覧](#-コマンド一覧)
8. [ドキュメント構造](#-ドキュメント構造)
9. [開発履歴（重要な変更）](#-開発履歴重要な変更)
10. [よくある作業パターン](#-よくある作業パターン)
11. [注意事項・制約](#️-注意事項制約)

---

## 🎯 プロジェクト概要

### 基本情報

- **名称**: booking.hxs - Discord Bot 部室予約システム
- **目的**: HxSコンピュータ部の部室予約をDiscord上で管理
- **言語**: Go 1.21+
- **実行環境**: Linux (systemd管理)
- **データ保存**: JSON形式（`reservations.json`）

### 主な機能

1. **予約管理**: `/reserve`, `/cancel`, `/complete`, `/edit`
2. **一覧表示**: `/list`, `/my-reservations`
3. **オートコンプリート**: 日付・時刻の入力支援（曜日表示対応）
4. **埋め込みメッセージ**: 統一されたUI/UX
5. **Ephemeralメッセージ**: プライバシー保護（予約IDなど）
6. **自動クリーンアップ**: 古い予約の自動削除（30日保持）
7. **ロギング**: コマンド統計の記録
8. **匿名フィードバック**: `/feedback`
9. **起動通知**: systemd再起動時のDiscord通知

---

## 🔧 技術スタック

### コア技術

```
言語: Go 1.21+
ライブラリ:
  - github.com/bwmarrin/discordgo (Discord API)
  - github.com/joho/godotenv (環境変数管理)
ビルドツール: Make
プロセス管理: systemd
開発ツール: Air (ホットリロード)
```

### 依存関係管理

- **Go Modules** (`go.mod`, `go.sum`)
- `go mod tidy` で依存関係を整理
- グローバルインストール不要（プロジェクト固有）

---

## 📂 プロジェクト構造

```
booking.hxs/
├── cmd/bot/                    # アプリケーションエントリーポイント
│   └── main.go                 # メイン処理（起動・初期化・ハンドラー登録）
├── internal/                   # プライベートアプリケーションコード
│   ├── commands/               # コマンドハンドラー
│   │   ├── handlers.go         # ルーティング（コマンド振り分け）
│   │   ├── response_helpers.go # 共通レスポンス関数
│   │   ├── autocomplete.go     # オートコンプリート処理
│   │   ├── cmd_reserve.go      # 予約作成
│   │   ├── cmd_cancel.go       # 予約キャンセル
│   │   ├── cmd_complete.go     # 予約完了
│   │   ├── cmd_edit.go         # 予約編集
│   │   ├── cmd_list.go         # 予約一覧
│   │   ├── cmd_my_reservations.go # 自分の予約
│   │   ├── cmd_help.go         # ヘルプ
│   │   └── cmd_feedback.go     # フィードバック
│   ├── models/                 # データモデル
│   │   └── reservation.go      # Reservation構造体
│   ├── storage/                # データ永続化
│   │   └── storage.go          # JSON読み書き・CRUD操作
│   └── logging/                # ロギング機能
│       └── logger.go           # コマンド統計記録
├── config/                     # 環境設定
│   ├── .env.example            # 環境変数テンプレート
│   ├── .env.development        # 開発環境設定
│   ├── .env.production         # 本番環境設定
│   └── booking-hxs.service     # systemdサービスファイル
├── docs/                       # ドキュメント
│   ├── AI_CONTEXT.md           # このファイル
│   ├── CHANGELOG.md            # 開発者向け変更履歴
│   ├── RELEASE_NOTES.md        # ユーザー向けリリースノート
│   ├── DEVELOPMENT.md          # 開発者ガイド
│   ├── COMMAND_TEMPLATE.md     # コマンド実装テンプレート
│   ├── SETUP.md                # セットアップガイド
│   ├── SYSTEMD.md              # systemd運用ガイド
│   ├── COMMANDS.md             # コマンドリファレンス
│   ├── DATA_MANAGEMENT.md      # データ管理ガイド
│   └── releases/               # バージョン別リリースノート
├── bin/                        # ビルド済みバイナリ
├── logs/                       # ログファイル
├── data/                       # データファイル
├── Makefile                    # ビルド・実行コマンド
├── switch_env.sh               # 環境切り替えスクリプト
└── .env                        # 現在の環境変数（Git管理外）
```

---

## 📐 開発ルール・コーディング規約

### 1. **コマンドハンドラーの7ステップフロー**

すべてのコマンドファイル（`cmd_*.go`）は以下の統一フローに従う：

```go
func handleXxx(...) {
    // 1. オプション取得
    options := i.ApplicationCommandData().Options
    optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
    for _, opt := range options {
        optionMap[opt.Name] = opt
    }

    // 2. ユーザー情報取得
    var userID, username string
    if i.Member != nil {
        userID = i.Member.User.ID
        username = i.Member.User.Username
    } else if i.User != nil {
        userID = i.User.ID
        username = i.User.Username
    }

    // 3. パラメータ抽出
    param1 := optionMap["param1"].StringValue()
    param2 := optionMap["param2"].IntValue()

    // 4. ビジネスロジック
    // データ作成・更新・削除など

    // 5. レスポンス（Ephemeral）
    respondEphemeral(s, i, "処理が完了しました")

    // 6. チャンネル通知（Public）
    if !isDM && allowedChannelID != "" {
        sendChannelEmbed(s, allowedChannelID, embed)
    }

    // 7. Botステータス更新
    if UpdateStatusCallback != nil {
        UpdateStatusCallback()
    }
}
```

### 2. **レスポンスヘルパー関数の使用**

`response_helpers.go` に定義された共通関数を必ず使用：

- `respondError(s, i, message)` - エラーメッセージ（Ephemeral）
- `respondEphemeral(s, i, message)` - 成功メッセージ（Ephemeral）
- `respondEmbed(s, i, embed)` - 埋め込みメッセージ（Ephemeral）
- `respondEmbedWithFooter(s, i, title, desc, color, fields, footerText)` - フィールド付き埋め込み
- `sendChannelEmbed(s, channelID, embed)` - チャンネルへの公開通知
- `createReservationEmbed(r, index, total)` - 予約情報の埋め込み作成
- `createHeaderEmbed(status, count)` - ヘッダー埋め込み作成

### 3. **機密情報の扱い（重要）**

**予約IDは機密情報**として扱う：

```go
// ❌ NG: 予約IDを公開メッセージに含める
fields := []*discordgo.MessageEmbedField{
    {Name: "予約ID", Value: reservation.ID}, // これを公開すると他人が編集可能
    {Name: "日付", Value: reservation.Date},
}
sendChannelEmbed(s, channelID, embed) // 公開される

// ✅ OK: fields[1:] で予約IDを除外
fields := []*discordgo.MessageEmbedField{
    {Name: "予約ID", Value: reservation.ID}, // Ephemeralのみに表示
    {Name: "日付", Value: reservation.Date},
}
respondEmbedWithFooter(..., fields, ...) // Ephemeral（実行者のみ）
sendChannelEmbed(..., fields[1:], ...)   // 公開（予約ID除外）
```

### 4. **main.go の構造**

`cmd/bot/main.go` は以下の順序で構成：

```go
// 1. 定数定義（ファイル冒頭）
const (
    saveInterval       = 5 * time.Minute
    logCleanupInterval = 24 * time.Hour
    autoCompleteHour   = 3
    autoCompleteMinute = 0
    cleanupHour        = 3
    cleanupMinute      = 10
    retentionDays      = 30
)

// 2. グローバル変数（必要最小限）
var (
    store                 *storage.Storage
    logger                *logging.Logger
    guildID               string
    allowedChannelID      string
    startupChannelID      string
    startupMessage        string
    processedInteractions sync.Map
)

// 3. init() - 環境変数読み込み
// 4. main() - エントリーポイント
// 5. initializeServices() - サービス初期化
// 6. setupHandlers() - イベントハンドラー設定
// 7. startBackgroundTasks() - バックグラウンドタスク起動
// 8. periodicSave() - 定期保存
// 9. periodicLogCleanup() - ログクリーンアップ
// 10. dailyAutoComplete() - 自動完了
// 11. dailyCleanup() - データクリーンアップ
// 12. sendStartupNotification() - 起動通知
// 13. shutdown() - 終了処理
// 14. updateBotStatus() - ステータス更新
// 15. registerCommands() - コマンド登録
```

### 5. **UI/UX統一ルール**

- すべての埋め込みメッセージに「部室予約システム | コマンド名」形式のフッターを付ける
- 予約一覧は10件を超える場合、複数のEphemeralメッセージに分割
- 日付のオートコンプリートには**曜日を表示**（例: `2025/11/17 (日)`）
- ページネーション形式: 「予約 X/Y」

---

## 🔑 重要な実装パターン

### 1. **日付オートコンプリートの曜日表示**

```go
// getWeekdayJa は日本語の曜日を返す
func getWeekdayJa(t time.Time) string {
    weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}
    return weekdays[int(t.Weekday())]
}

// formatDateWithWeekday は日付を曜日付きでフォーマットする
func formatDateWithWeekday(t time.Time) string {
    return fmt.Sprintf("%s (%s)", t.Format("2006/01/02"), getWeekdayJa(t))
}
```

### 2. **起動通知機能**

```go
func sendStartupNotification(s *discordgo.Session) {
    if startupChannelID == "" {
        log.Println("Startup notification disabled")
        return
    }

    message := startupMessage
    if message == "" {
        message = "Bot が再起動しました。\n部室予約システムが利用可能です。"
    }
    // \n を改行に変換
    message = strings.ReplaceAll(message, "\\n", "\n")

    embed := &discordgo.MessageEmbed{
        Title:       "システムメッセージ",
        Description: message,
        Color:       0x00ff00,
        Timestamp:   time.Now().Format(time.RFC3339),
        Footer: &discordgo.MessageEmbedFooter{
            Text: "部室予約システム | システムメッセージ",
        },
    }

    s.ChannelMessageSendEmbed(startupChannelID, embed)
}
```

### 3. **フィールド再利用パターン**

```go
// Ephemeral用（予約ID含む）
fields := []*discordgo.MessageEmbedField{
    {Name: "予約ID", Value: reservation.ID, Inline: false},
    {Name: "日付", Value: reservation.Date, Inline: true},
    {Name: "時間", Value: fmt.Sprintf("%s - %s", reservation.StartTime, reservation.EndTime), Inline: true},
}

// Ephemeralレスポンス（予約ID表示）
respondEmbedWithFooter(s, i, "予約を作成しました", "", 0x00ff00, fields, "部室予約システム | reserve")

// 公開通知（予約ID除外）
publicFields := append([]*discordgo.MessageEmbedField{
    {Name: "予約者", Value: fmt.Sprintf("<@%s>", userID), Inline: false},
}, fields[1:]...) // fields[1:] で予約IDをスキップ

sendChannelEmbed(s, allowedChannelID, &discordgo.MessageEmbed{
    Title: "新しい予約が作成されました",
    Fields: publicFields,
    Color: 0x00ff00,
})
```

---

## 🌍 環境変数

### 必須

```bash
DISCORD_TOKEN=your_discord_bot_token_here  # Discord Botトークン
```

### 推奨

```bash
GUILD_ID=your_guild_id_here                # サーバーID（即座に登録）
ALLOWED_CHANNEL_ID=your_channel_id_here    # コマンド受付チャンネル
FEEDBACK_CHANNEL_ID=your_channel_id_here   # フィードバック送信先
```

### オプション

```bash
# 起動通知機能
STARTUP_NOTIFICATION_CHANNEL_ID=          # 起動通知送信先（空欄で無効化）
STARTUP_NOTIFICATION_MESSAGE=             # カスタム起動メッセージ（\n で改行）

# 環境設定
ENV=production                             # production または development

# データファイル
DATA_FILE=data/reservations.json           # データ保存先
```

### 環境切り替え

```bash
./switch_env.sh development  # 開発環境に切り替え
./switch_env.sh production   # 本番環境に切り替え
```

---

## 📜 コマンド一覧

| コマンド | 説明 | ファイル | 主な機能 |
|---------|------|---------|---------|
| `/reserve` | 予約作成 | `cmd_reserve.go` | 日付・時刻・コメント入力、オートコンプリート対応 |
| `/cancel` | 予約キャンセル | `cmd_cancel.go` | 予約ID指定、ステータス確認 |
| `/complete` | 予約完了 | `cmd_complete.go` | 予約ID指定、完了マーク |
| `/edit` | 予約編集 | `cmd_edit.go` | 日付・時刻・コメント変更 |
| `/list` | 予約一覧 | `cmd_list.go` | ステータス別表示、ページネーション |
| `/my-reservations` | 自分の予約 | `cmd_my_reservations.go` | ユーザー別表示 |
| `/help` | ヘルプ | `cmd_help.go` | コマンド一覧表示 |
| `/feedback` | フィードバック | `cmd_feedback.go` | 匿名フィードバック送信 |

---

## 📚 ドキュメント構造

### ユーザー向け

- `README.md` - プロジェクト概要
- `docs/SETUP.md` - セットアップ手順
- `docs/COMMANDS.md` - コマンドリファレンス
- `docs/RELEASE_NOTES.md` - リリース情報一覧
- `docs/releases/vX.X.X.md` - 各バージョンの詳細

### 開発者向け

- `docs/DEVELOPMENT.md` - 開発環境・拡張方法
- `docs/COMMAND_TEMPLATE.md` - コマンド実装テンプレート
- `docs/CHANGELOG.md` - 詳細な変更履歴
- `docs/AI_CONTEXT.md` - このファイル（AI用コンテキスト）

### 運用向け

- `docs/SYSTEMD.md` - systemd設定・運用
- `docs/DATA_MANAGEMENT.md` - データ管理

---

## 🕰️ 開発履歴（重要な変更）

### v1.3.3 (2025-11-17) - 最新版

- **日付オートコンプリートに曜日表示を追加**
  - `getWeekdayJa()`, `formatDateWithWeekday()` 実装
  - 表示例: 「今日 2025/11/17 (日)」「1週間後 2025/11/24 (日)」

### internal-update (2025-11-17)

- **Bot起動通知機能の追加**
  - 環境変数: `STARTUP_NOTIFICATION_CHANNEL_ID`, `STARTUP_NOTIFICATION_MESSAGE`
  - `sendStartupNotification()` 関数実装
  - `\n` を改行に変換する処理追加

- **コードフロー統一**
  - 全コマンドを7ステップフローに統一
  - `respondEmbedWithFields()` 削除、`respondEmbedWithFooter()` に統合
  - フィールド再利用パターン（`fields[1:]`）導入

- **コマンドテンプレート作成**
  - `docs/COMMAND_TEMPLATE.md` 追加
  - 3つのパターン（データ変更・表示・シンプル）
  - 機密情報の扱い方3パターン

### v1.3.2 (2025-11-17)

- **セキュリティ強化**
  - `/edit` コマンドの公開メッセージから予約IDを除外
  - `fields[1:]` パターンで機密情報保護

### v1.3.1 (2025-11-12)

- **UI統一・ページネーション**
  - 10件以上の予約を複数メッセージに分割
  - フッター統一（「部室予約システム | コマンド名」）

### v1.3.0 (2025-11-12)

- **埋め込みメッセージ化**
  - すべてのコマンドを埋め込み形式に変更

### internal-update (2025-11-16)

- **Go標準レイアウト採用**
  - `cmd/bot/main.go` エントリーポイント
  - `internal/` プライベートコード
  - コマンドハンドラー分割（`cmd_*.go`）

---

## 🛠️ よくある作業パターン

### 新しいコマンドを追加する

1. `docs/COMMAND_TEMPLATE.md` を参照
2. `internal/commands/cmd_new_command.go` を作成
3. 7ステップフローに従って実装
4. `internal/commands/handlers.go` にルーティング追加
5. `cmd/bot/main.go` の `getCommandDefinitions()` にコマンド定義追加
6. `make build` でビルド確認
7. ドキュメント更新（COMMANDS.md, CHANGELOG.md）

### リリースノートを作成する

1. `docs/releases/vX.X.X.md` を `vx.x.x.md` から複製
2. テンプレートに従って記載:
   - 📢 ハイライト
   - ✨ 新機能
   - 🚀 改善・変更
   - 🐛 バグ修正
   - ⚠️ 注意事項
3. `docs/CHANGELOG.md` に開発者向け詳細を追加
4. `docs/RELEASE_NOTES.md` の最新バージョンを更新
5. `README.md` のバージョン番号を更新

### 環境変数を追加する

1. `config/.env.example` にコメント付きで追加
2. `config/.env.development` に開発用デフォルト値を追加
3. `config/.env.production` に本番用デフォルト値を追加
4. `cmd/bot/main.go` の `init()` で読み込み処理追加
5. `docs/SETUP.md` の環境変数テーブルに説明追加
6. `docs/SYSTEMD.md` のsystemd設定例に追加（該当する場合）

### ビルド・実行

```bash
# フォーマット
make fmt
# または
go fmt ./...

# コード検証
make vet
# または
go vet ./...

# ビルド
make build
# または
go build -o bin/booking.hxs cmd/bot/main.go

# 実行
make run
# または
go run cmd/bot/main.go

# ホットリロード（開発時）
air
```

---

## ⚠️ 注意事項・制約

### 絶対に守るべきこと

1. **予約IDは機密情報**
   - Ephemeralメッセージ（実行者のみ）には表示OK
   - 公開メッセージには**絶対に含めない**
   - `fields[1:]` パターンで除外

2. **7ステップフローの遵守**
   - すべての `cmd_*.go` は統一フローに従う
   - 番号付きコメント（日本語）必須

3. **共通関数の使用**
   - `response_helpers.go` の関数を必ず使う
   - 独自のレスポンス処理を書かない

4. **ドキュメントの同期更新**
   - コード変更時は必ず関連ドキュメントも更新
   - CHANGELOG.md（開発者向け）とRILEASE_NOTES.md（ユーザー向け）両方

### やってはいけないこと

- ❌ `main.go` に大量のロジックを書く → 関数分割する
- ❌ グローバル変数を増やす → 必要最小限に
- ❌ エラーハンドリングを省略する → 必ずログ記録
- ❌ ハードコードされた値 → 定数化または環境変数化
- ❌ テスト用コードを本番にコミット → 環境変数で制御

### プロジェクト固有の制約

- **Go 1.21+** 必須（time.Time の拡張機能使用）
- **systemd** での運用を前提（Linux環境）
- **JSON形式** でのデータ保存（DB不使用）
- **予約の保持期間**: 30日（`retentionDays` 定数）
- **自動完了時刻**: 毎日3:00（`autoCompleteHour` 定数）
- **クリーンアップ時刻**: 毎日3:10（`cleanupHour` 定数）

---

## 📝 その他の重要情報

### Makefileコマンド

```makefile
make help           # ヘルプ表示
make setup          # 初回セットアップ
make build          # ビルド
make run            # 実行
make fmt            # フォーマット
make vet            # コード検証
make clean          # クリーンアップ
make dev            # 開発モード（ホットリロード）
```

### ログファイル

- `logs/command_stats.json` - コマンド統計
- `logs/commands_YYYY-MM.log` - 月別コマンドログ
- 自動クリーンアップ: 24時間ごと

### systemd運用

```bash
# サービス起動
sudo systemctl start booking-hxs

# サービス停止
sudo systemctl stop booking-hxs

# サービス再起動
sudo systemctl restart booking-hxs

# 自動起動有効化
sudo systemctl enable booking-hxs

# ログ確認
sudo journalctl -u booking-hxs -f
```

### 開発ワークフロー

1. `./switch_env.sh development` で開発環境に切り替え
2. `air` でホットリロード起動
3. Discordでコマンドテスト
4. `make fmt && make vet` でコード検証
5. `make build` でビルド確認
6. コミット前に CHANGELOG.md 更新
7. バージョンアップ時はリリースノート作成

---

## 🎓 AIへの指示例

このドキュメントを読んだ上で、以下のような指示が可能です：

### 良い指示例

✅ 「新しいコマンド `/status` を追加してください。現在のBot稼働時間と予約件数を表示します。7ステップフローに従ってください。」

✅ 「オートコンプリートの日付候補に祝日を表示する機能を追加してください。曜日表示と同じパターンで実装してください。」

✅ 「v1.3.4のリリースノートを作成してください。今回の変更は起動通知メッセージの改行対応です。」

✅ 「環境変数 `MAX_RESERVATIONS_PER_USER` を追加して、1ユーザーあたりの予約上限を設定できるようにしてください。」

### 悪い指示例

❌ 「予約システムを作ってください」（既に存在する）

❌ 「データベースに対応してください」（JSON前提の設計）

❌ 「Pythonで書き直してください」（Go言語固定）

❌ 「予約IDを公開メッセージに表示してください」（セキュリティ違反）

---

## 📞 このドキュメントの使い方

### 新しいAIセッションでの指示例

```
以下のドキュメントを読んでから作業してください：
[AI_CONTEXT.mdの内容を貼り付け]

その上で、[具体的な指示]をお願いします。
```

### 特定の機能に関する指示

```
AI_CONTEXT.mdの「起動通知機能」セクションを参照して、
起動メッセージに絵文字を追加できるようにしてください。
```

### トラブルシューティング時

```
AI_CONTEXT.mdの「注意事項・制約」を確認した上で、
[エラー内容]を解決してください。
```

---

**このドキュメントは、プロジェクトの進化に合わせて随時更新してください。**

**最終更新者**: AI Assistant
**最終更新日**: 2025年11月17日

# booking.hxs - Discord Bot 部室予約システム

Go言語で作成されたDiscord Bot用の部室予約システムです。スラッシュコマンドを使用して、部室の予約作成、編集、取り消し、完了を管理できます。また、このBotはHxSコンピュータ部内のサーバ上で動作しています。

## ✨ 主な機能

- 📅 **予約管理** - `/reserve`, `/cancel`, `/complete`, `/edit`
- 📋 **一覧表示** - `/list`, `/my-reservations`
- 🔍 **オートコンプリート** - 日付・時刻の入力支援
- 🎨 **埋め込みメッセージ** - 視認性の高いUI
- 🔒 **Ephemeralメッセージ** - プライバシー保護
- 🗑️ **自動クリーンアップ** - 古い予約の自動削除
- 📊 **ロギング機能** - コマンド統計の記録
- 💬 **匿名フィードバック** - `/feedback`

詳細は **[📝 コマンドリファレンス](docs/COMMANDS.md)** へ


## 🔧 技術スタック

- **言語**: Go 1.21+
- **ライブラリ**:
  - [discordgo](https://github.com/bwmarrin/discordgo) - Discord API
  - [godotenv](https://github.com/joho/godotenv) - 環境変数管理
- **データ保存**: JSON

## 📚 ドキュメント

| カテゴリ | ドキュメント | 説明 |
|---------|------------|------|
| **基本** | **[📖 セットアップガイド](docs/SETUP.md)** | 環境構築から起動まで |
| | **[📝 コマンドリファレンス](docs/COMMANDS.md)** | 全コマンドの使い方 |
| **運用** | **[🗄️ データ管理](docs/DATA_MANAGEMENT.md)** | データ保存とクリーンアップ |
| | **[⚙️ systemdセットアップ](docs/SYSTEMD.md)** | サーバーでの自動起動 |
| **開発** | **[💻 開発者ガイド](docs/DEVELOPMENT.md)** | 開発環境と拡張方法 |
|  | **[💻 コマンド開発テンプレート](docs/COMMAND_TEMPLATE.md)** | コマンドの追加をするためのテンプレート |
| **変更履歴** | **[📝 リリースノート](docs/RELEASE_NOTES.md)** | 詳細なリリース情報 |
|  | **[📋 CHANGELOG](docs/CHANGELOG.md)** | 開発者向けバージョン履歴 |

## 📖 プロジェクト構造

```
booking.hxs/
├── cmd/bot/              # アプリケーションエントリーポイント
│   └── main.go           # メインファイル
├── internal/             # プライベートアプリケーションコード
│   ├── commands/         # コマンドハンドラー（コマンドごとに分割）
│   ├── models/           # データモデル
│   ├── storage/          # データ永続化
│   └── logging/          # ロギング機能
├── config/               # 設定ファイル
├── data/                 # データファイル
├── docs/                 # ドキュメント
└── bin/                  # ビルド成果物
```

**設計思想**: Go標準プロジェクトレイアウトに準拠
- `cmd/` - 複数バイナリ対応可能な構造
- `internal/` - 外部インポート不可（Go言語仕様）
- `commands/` - 各コマンドを独立したファイルで管理

詳細は **[💻 開発者ガイド](docs/DEVELOPMENT.md)** へ


## 🚀 クイックスタート

```bash
# セットアップ（初回のみ）
./setup.sh

# 環境変数を設定
vi .env  # DISCORD_TOKEN, GUILD_ID, FEEDBACK_CHANNEL_ID を設定

# 起動
make run
```

詳しくは **[📖 セットアップガイド](docs/SETUP.md)** をご覧ください。


## 🛠️ よく使うコマンド

```bash
# セットアップ
./setup.sh                  # 初回セットアップ

# 実行
make run                    # 開発モードで実行
make build                  # ビルド

# 環境切り替え
./switch_env.sh development # 開発環境
./switch_env.sh production  # 本番環境

# その他
make help                   # コマンド一覧
make clean                  # クリーンアップ
```


##  ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルをご覧ください。


## 🤝 フィードバック

バグ報告や機能要望は、Discord Botの `/feedback` コマンドでお送りください（完全匿名）。

---

**バージョン**: v1.4.0<br>
**作成**: 2025年<br>
**Go**: 1.21+<br>
**開発者**:
- [dice](https://github.com/dice-2004)
<!-- ここに追加 -->

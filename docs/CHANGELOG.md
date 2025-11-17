# 📋 変更履歴 (CHANGELOG)

このドキュメントは開発者向けの詳細な変更履歴です。

**📝 ユーザ向け情報は [RELEASE_NOTES.md](RELEASE_NOTES.md) をご覧ください。**

すべての重要な変更は、このファイルに記録されます。
このプロジェクトは [Semantic Versioning](https://semver.org/) に従います。


## [Unreleased]-template

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security
---
## [1.3.3] - 2025-11-17

### Changed
- **日付オートコンプリートに曜日表示を追加**: すべての日付候補に曜日が表示されるように改善
  - `getWeekdayJa()` 関数: 日本語の曜日（日〜土）を返す
  - `formatDateWithWeekday()` 関数: 日付を「YYYY/MM/DD (曜日)」形式でフォーマット
  - 適用範囲: 今日・明日・明後日、週間表示、月入力、年入力、日入力、フィルタリング処理
  - 表示例: 「今日 2025/11/17 (日)」「1週間後 2025/11/24 (日)」
  - ユーザービリティ向上: 週末（土日）が一目でわかり、希望する曜日での予約が容易に



---

## [internal-update] - 2025-11-17

### Added
- **Bot起動通知機能**: systemdでの自動再起動時に通知を送信
  - 新しい環境変数 `STARTUP_NOTIFICATION_CHANNEL_ID`: 通知先チャンネルID（オプション）
  - 新しい環境変数 `STARTUP_NOTIFICATION_MESSAGE`: カスタム起動メッセージ（オプション）
  - `sendStartupNotification()` 関数: 起動時に埋め込みメッセージを送信
  - グローバル変数 `startupChannelID`, `startupMessage` を追加
  - デフォルトメッセージ: "🚀 Bot が起動しました。部室予約システムが利用可能です。"
  - エラーハンドリング: 送信失敗時はログに記録

### Changed
- **main()関数の処理順序**: コマンド登録後、バックグラウンドタスク開始前に起動通知を送信
- **ドキュメント更新**:
  - `config/.env.example`: 起動通知関連の環境変数を追加
  - `docs/SETUP.md`: 環境変数一覧テーブルに起動通知の説明を追加
  - `docs/SYSTEMD.md`: systemd環境変数設定例に起動通知を追加
  - `README.md`: クイックスタートにsystemdコマンド例を追加

---

## [internal-update] - 2025-11-17

### Changed
- **コマンドファイルのコードフロー統一**: すべてのコマンドファイルを7ステップのコードフローに統一
  - ステップ1: オプション取得
  - ステップ2: ユーザー情報取得
  - ステップ3: パラメータ抽出
  - ステップ4: ビジネスロジック
  - ステップ5: レスポンス（Ephemeral）
  - ステップ6: チャンネル通知（Public）
  - ステップ7: Botステータス更新
- **レスポンスヘルパー関数の統合**: `respondEmbedWithFields()` を削除し、`respondEmbedWithFooter()` に統合
- **フィールド再利用の最適化**: 公開メッセージで `publicFields` を新規作成する代わりに `fields[1:]` を使用することでコード重複を削減
  - `cmd_edit.go`: 予約ID除外に `fields[1:]` を使用
  - `cmd_reserve.go`: 予約者フィールド追加後に `fields[1:]` を結合

### Added
- **コマンドテンプレート**: `docs/COMMAND_TEMPLATE.md` を作成
  - データ変更コマンドのテンプレート
  - データ表示コマンドのテンプレート
  - シンプルコマンドのテンプレート
  - 機密情報の扱いに関する3つのパターン例

---
## [1.3.2] - 2025-11-17

### Fixed
- **セキュリティ**: `/edit` コマンド実行時の公開メッセージに予約IDが表示されていた問題を修正
  - 修正前: チャンネルの公開メッセージに予約IDが含まれる
  - 修正後: 予約IDは実行者のみに表示（Ephemeralメッセージ）
  - 影響: 予約IDは編集・キャンセル・完了に必要な機密情報であり、この修正によりセキュリティが向上

## [1.3.1-Unreleased] - 2025-11-17

### Changed
- **リリースノート**: リリースノートをテンプレートに従い書き直し


## [internal-update] - 2025-11-16

### Changed
- **コマンドハンドラーの分割**: 各コマンドを独立したファイルで管理
  - `internal/commands/handlers.go`: ルーティングのみを担当
  - `internal/commands/cmd_reserve.go`: `/reserve` コマンド
  - `internal/commands/cmd_cancel.go`: `/cancel` コマンド
  - `internal/commands/cmd_complete.go`: `/complete` コマンド
  - `internal/commands/cmd_edit.go`: `/edit` コマンド
  - `internal/commands/cmd_list.go`: `/list` コマンド
  - `internal/commands/cmd_my_reservations.go`: `/my-reservations` コマンド
  - `internal/commands/cmd_help.go`: `/help` コマンド
  - `internal/commands/cmd_feedback.go`: `/feedback` コマンド
  - `internal/commands/response_helpers.go`: 共通レスポンス関数
- **main.goのリファクタリング**: 大きな`main`関数を複数の小さな関数に分割
  - 設定値の一元管理（`const`ブロック）
  - `initializeServices()`: サービス初期化
  - `setupHandlers()`: イベントハンドラー設定
  - `startBackgroundTasks()`: バックグラウンドタスク起動
  - `periodicSave()`: 定期保存
  - `periodicLogCleanup()`: ログクリーンアップ
  - `dailyAutoComplete()`: 自動完了処理
  - `dailyCleanup()`: データクリーンアップ
  - `deleteExistingCommands()`: 既存コマンド削除
  - `getCommandDefinitions()`: コマンド定義取得

### Removed
- コード重複の排除により不要になった関数や処理を削除


## [internal-update] - 2025-11-16

### Changed
- **Go標準レイアウト採用**: プロジェクト構造をGoコミュニティ推奨のレイアウトに変更
  - `cmd/bot/main.go`: アプリケーションエントリーポイント
  - `internal/`: プライベートアプリケーションコード（外部インポート不可）
    - `internal/commands/`: コマンドハンドラー群
    - `internal/models/`: データモデル
    - `internal/storage/`: データ永続化
    - `internal/logging/`: ロギング機能
- **インポートパスの更新**: すべての`.go`ファイルを新しい構造に移行し、`internal/`プレフィックスに更新
- **Makefile更新**: ビルドコマンドを新しいエントリーポイント（`cmd/bot/main.go`）に対応

## [1.3.1] - 2025-11-12

### Added
- **ページネーション機能**: 10件以上の予約を複数メッセージに自動分割
  - 1件目: ヘッダー + 最初の9件
  - 2件目以降: 各メッセージに10件ずつ
- **統一フッター**: すべての埋め込みメッセージに「部室予約システム | コマンド名」形式のフッターを追加
- **進捗表示**: リストコマンドで「予約 X/Y」形式の進捗を表示

### Changed
- **Ephemeral徹底**: すべてのページが実行者のみに表示されるように統一
- **コマンド登録エラー処理**: 一部のコマンド登録が失敗しても他のコマンドは登録継続するように改善


## [1.3.0] - 2025-11-12

### Added
- **埋め込みメッセージ化**: すべてのコマンド応答（help以外）を見やすい埋め込み形式（Discord Embed）に変更
- **カラーコード統一**: 各コマンドに適切な色を設定
  - `/reserve`: 🟢 緑 (0x57F287) - 成功・作成
  - `/edit`: 🟡 黄 (0xFEE75C) - 変更・編集
  - `/cancel`: 🔴 赤 (0xED4245) - キャンセル
  - `/complete`: 🔵 青 (0x5865F2) - 完了
  - `/list`: ⚫ 黒 (0x000000) - 情報表示
  - `/my-reservations`: ⚪ 白 (0xFFFFFF) - 情報表示
  - `/feedback`: 🟢 緑 (0x57F287) - フィードバック
  - エラー: 🔴 赤 (0xED4245)

### Changed
- **レスポンスフォーマット**: タイトル、フィールド、タイムスタンプで構成される統一フォーマットに変更
- **ヘルパー関数**: `internal/commands/response_helpers.go` を追加し、応答処理を共通化


## [1.2.0] - 2025-11-12

### Added
- **`/edit` コマンド**: 既存の予約を編集可能に
  - 編集可能項目: 日付、開始時刻、終了時刻、コメント
  - 部分的な編集に対応（変更したい項目のみ指定可能）
  - オートコンプリート対応
- **重複チェック**: 編集時に他の予約との重複をチェック
- **過去日時ブロック**: 過去の日時への編集を防止
- **通知機能**: 編集者へのプライベート通知とチャンネルへの公開通知

### Changed
- コマンド定義に `/edit` を追加



## [1.1.0] - 2025-11-12

### Added
- **オートコンプリート機能**: 日付・時刻入力時に候補を自動表示
  - **日付入力**: 「今日」「明日」「明後日」「1週間後」「2週間後」「1ヶ月後」の候補を表示
  - **開始時刻**: 09:00〜21:00の30分刻みで候補を表示
  - **終了時刻**: 開始時刻より後の時刻のみ表示（入力ミス防止）
- **ファイル追加**: `internal/commands/autocomplete.go` でオートコンプリート処理を実装

### Changed
- コマンド定義に `Autocomplete: true` を追加
- `cmd/bot/main.go` にオートコンプリートハンドラーを追加


## [1.0.1] - 2025-11-12

### Changed
- **クリーンアップタイミングの最適化**: Bot起動時刻から24時間ごと → 毎日決まった時刻（午前3時/3時10分）に変更
  - 午前3時00分: 期限切れ予約の自動完了
  - 午前3時10分: 古いデータの自動削除
- **起動時の即時クリーンアップ**: Bot起動が深夜0時〜0時5分の場合、即座にクリーンアップを実行

### Fixed
- 起動時のクリーンアップ判定処理で発生していた無限ループを修正


## [1.0.0] - 2025-11-11

### Added
- **予約管理コマンド**:
  - `/reserve`: 予約作成（日付、時刻、コメント指定）
  - `/cancel`: 予約キャンセル
  - `/complete`: 予約完了
- **表示コマンド**:
  - `/list`: すべての予約を表示
  - `/my-reservations`: 自分の予約を表示
- **サポートコマンド**:
  - `/help`: ヘルプ表示
  - `/feedback`: 完全匿名フィードバック送信
- **自動クリーンアップ機能**:
  - 期限切れ予約の自動完了
  - 30日以上前の完了・キャンセル済み予約の自動削除
- **データ永続化**: JSON形式で保存、5分ごとに自動保存
- **ログシステム**: コマンドログ、統計情報、月次ローテーション
- **環境分離**: 開発/本番環境の分離機能
- **セキュリティ機能**:
  - 予約IDは推測困難なランダム文字列
  - 予約IDは作成者のみに通知（Ephemeralメッセージ）
  - 環境変数で機密情報を安全に管理

## 変更タイプの分類

- **Added**: 新機能
- **Changed**: 既存機能の変更
- **Deprecated**: 非推奨になった機能（次期バージョンで削除予定）
- **Removed**: 削除された機能
- **Fixed**: バグ修正
- **Security**: セキュリティ関連の修正

---

**最終更新**: 2025-11-17

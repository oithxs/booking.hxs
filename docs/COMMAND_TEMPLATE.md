# 📝 コマンド実装テンプレート

このドキュメントでは、新しいコマンドを追加する際のテンプレートとガイドラインを提供します。

## 📑 目次

1. [基本テンプレート](#基本テンプレート)
2. [データ変更コマンドのテンプレート](#データ変更コマンドのテンプレート)
3. [データ表示コマンドのテンプレート](#データ表示コマンドのテンプレート)
4. [シンプルコマンドのテンプレート](#シンプルコマンドのテンプレート)
5. [コードフローの統一ルール](#コードフローの統一ルール)

---

## 基本テンプレート

### ファイル構成

```
internal/commands/
├── cmd_xxx.go          # 新しいコマンドのハンドラー
├── handlers.go         # ルーティング（ここに追加）
└── response_helpers.go # 共通レスポンス関数
```

---

## データ変更コマンドのテンプレート

予約の作成・編集・削除など、データを変更するコマンドのテンプレートです。

### ファイル: `internal/commands/cmd_xxx.go`

```go
package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
	"github.com/dice/hxs_reservation_system/internal/models"
	"github.com/dice/hxs_reservation_system/internal/storage"
)

// handleXxx は XXX コマンドを処理する
func handleXxx(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool) {
	// 1. オプション取得
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	// 2. ユーザー情報取得
	userID, username := getUserInfo(i, isDM)

	// 3. パラメータ抽出
	param1 := optionMap["param1"].StringValue()

	// オプションパラメータ
	param2 := ""
	if opt, ok := optionMap["param2"]; ok {
		param2 = opt.StringValue()
	}

	// 4. ビジネスロジック
	// - データの検証
	// - データの取得・更新・削除
	// - エラーハンドリング

	// 例: データの更新
	if err := store.Save(); err != nil {
		respondError(s, i, "データの保存に失敗しました")
		logger.LogError("ERROR", "handleXxx", "Failed to save", err, map[string]interface{}{
			"user_id": userID,
		})
		return
	}

	// 5. レスポンス - ユーザーへの応答（Ephemeral）
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "🆔 データID", // 機密情報がある場合は最初に配置
			Value:  "ID-xxx-xxx",
			Inline: false,
		},
		{
			Name:   "📋 項目名",
			Value:  "値",
			Inline: false,
		},
	}
	respondEmbedWithFooter(s, i, "✅ 成功しました", "", fields, 0x57F287, "部室予約システム  |  xxx", true)

	// 6. チャンネル通知 - 公開メッセージ（機密情報を除外）
	// 機密情報がない場合: fields をそのまま使用
	// 機密情報がある場合: fields[1:] で除外するか、必要に応じて追加フィールドを結合
	if !isDM {
		sendChannelEmbed(s, allowedChannelID, "� 通知タイトル", fmt.Sprintf("<@%s> さんが操作を実行しました", userID), fields[1:], 0x57F287, "部室予約システム  |  xxx")
	} else if allowedChannelID != "" {
		sendChannelEmbed(s, allowedChannelID, "📢 通知タイトル", fmt.Sprintf("%s さんが操作を実行しました", username), fields[1:], 0x57F287, "部室予約システム  |  xxx")
	}

	// 7. Botステータス更新
	if UpdateStatusCallback != nil {
		UpdateStatusCallback()
	}
}
```

### 使用例: 予約キャンセルコマンド

`cmd_cancel.go` を参照してください。

---

## データ表示コマンドのテンプレート

予約一覧表示など、データを表示するだけのコマンドのテンプレートです。

### ファイル: `internal/commands/cmd_xxx_list.go`

```go
package commands

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
	"github.com/dice/hxs_reservation_system/internal/models"
	"github.com/dice/hxs_reservation_system/internal/storage"
)

// handleXxxList は XXX 一覧を表示する
func handleXxxList(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, isDM bool) {
	// 1. データ取得
	allItems := store.GetAllXxx()

	// 2. データ処理 - フィルタリング
	items := make([]*models.Xxx, 0)
	for _, item := range allItems {
		if item.Status == models.StatusActive {
			items = append(items, item)
		}
	}

	// 3. レスポンス - データがない場合
	if len(items) == 0 {
		respondEmbed(s, i, "⚫ XXX一覧", "現在、XXXはありません。", 0x000000, true)
		return
	}

	// 4. データ処理 - ソート
	sort.Slice(items, func(a, b int) bool {
		// ソート条件
		return items[a].CreatedAt.Before(items[b].CreatedAt)
	})

	// 5. レスポンス - 最初のメッセージ（ヘッダー + 最初の9件）
	embeds := []*discordgo.MessageEmbed{}

	// ヘッダー
	headerDescription := fmt.Sprintf("現在 %d 件のXXXがあります", len(items))
	headerEmbed := createHeaderEmbed("⚫ XXX一覧", headerDescription, 0x000000, "部室予約システム  |  xxx-list")
	embeds = append(embeds, headerEmbed)

	// 最初の9件を表示
	maxFirstMessage := 9
	for idx := 0; idx < len(items) && idx < maxFirstMessage; idx++ {
		item := items[idx]

		fields := []*discordgo.MessageEmbedField{
			{
				Name:   "📋 項目名",
				Value:  item.Value,
				Inline: false,
			},
		}

		itemEmbed := createReservationEmbed(
			fmt.Sprintf("No.%d", idx+1),
			fields,
			0x000000,
			fmt.Sprintf("部室予約システム  |  xxx-list  |  XXX %d/%d", idx+1, len(items)),
		)
		embeds = append(embeds, itemEmbed)
	}

	// 最初のメッセージを送信
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})

	// 残りの項目を複数のメッセージで送信（10件ごと）
	if len(items) > maxFirstMessage {
		itemsPerMessage := 10
		for startIdx := maxFirstMessage; startIdx < len(items); startIdx += itemsPerMessage {
			endIdx := startIdx + itemsPerMessage
			if endIdx > len(items) {
				endIdx = len(items)
			}

			messageEmbeds := []*discordgo.MessageEmbed{}
			for idx := startIdx; idx < endIdx; idx++ {
				item := items[idx]

				fields := []*discordgo.MessageEmbedField{
					{
						Name:   "📋 項目名",
						Value:  item.Value,
						Inline: false,
					},
				}

				itemEmbed := createReservationEmbed(
					fmt.Sprintf("No.%d", idx+1),
					fields,
					0x000000,
					fmt.Sprintf("部室予約システム  |  xxx-list  |  XXX %d/%d", idx+1, len(items)),
				)
				messageEmbeds = append(messageEmbeds, itemEmbed)
			}

			// フォローアップメッセージを送信（Ephemeral）
			_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Embeds: messageEmbeds,
				Flags:  discordgo.MessageFlagsEphemeral,
			})
			if err != nil {
				logger.LogError("ERROR", "handleXxxList", "Failed to send followup message", err, map[string]interface{}{
					"start_idx": startIdx,
					"end_idx":   endIdx,
				})
			}
		}
	}
}
```

### 使用例: 予約一覧コマンド

`cmd_list.go` を参照してください。

---

## シンプルコマンドのテンプレート

フィードバック送信など、ストレージを使用しないシンプルなコマンドのテンプレートです。

### ファイル: `internal/commands/cmd_xxx.go`

```go
package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
)

// handleXxx は XXX コマンドを処理する
func handleXxx(s *discordgo.Session, i *discordgo.InteractionCreate, logger *logging.Logger, isDM bool) {
	// 1. オプション取得とバリデーション
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondError(s, i, "入力が必要です")
		return
	}

	// 2. パラメータ抽出
	param := options[0].StringValue()
	if param == "" {
		respondError(s, i, "入力が必要です")
		return
	}

	// 3. ユーザー情報取得
	userID, username := getUserInfo(i, isDM)

	// 4. ビジネスロジック
	// 処理を実行

	// 5. レスポンス
	respondEmbed(s, i, "✅ 成功しました",
		"処理が完了しました。",
		0x57F287, true)

	// 6. ログ記録
	logger.LogCommand("xxx", userID, username, i.ChannelID, true, "", map[string]interface{}{
		"param_length": len(param),
	})
}
```

### 使用例: フィードバックコマンド

`cmd_feedback.go` を参照してください。

---

## コードフローの統一ルール

### 必須フロー

すべてのコマンドは以下の順序で処理を行います：

1. **オプション取得** - コマンドパラメータの取得とマップ化
2. **ユーザー情報取得** - 必要な場合のみ
3. **パラメータ抽出** - 必須・オプションパラメータの取得
4. **ビジネスロジック / データ処理** - メインの処理
5. **レスポンス** - ユーザーへの応答（Ephemeral）
6. **チャンネル通知** - 公開メッセージ（必要な場合のみ）
7. **Botステータス更新** - データ変更がある場合のみ

### コメント規約

各セクションに番号付きコメントを追加：

```go
// 1. オプション取得
// 2. ユーザー情報取得
// 3. パラメータ抽出
// 4. ビジネスロジック
// 5. レスポンス
// 6. チャンネル通知
// 7. Botステータス更新
```

### 関数シグネチャ

```go
// データ変更コマンド（ストレージ使用）
func handleXxx(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool)

// データ表示コマンド（ストレージ使用、チャンネル指定不要）
func handleXxxList(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, isDM bool)

// シンプルコマンド（ストレージ不使用）
func handleXxx(s *discordgo.Session, i *discordgo.InteractionCreate, logger *logging.Logger, isDM bool)
```

### レスポンスヘルパー関数

#### エラー応答
```go
respondError(s, i, "エラーメッセージ")
```

#### シンプルな埋め込みメッセージ
```go
respondEmbed(s, i, "タイトル", "説明", 0x57F287, true) // 最後はephemeral
```

#### フィールド付き埋め込みメッセージ（フッター付き）
```go
respondEmbedWithFooter(s, i, "タイトル", "説明", fields, 0x57F287, "部室予約システム  |  xxx", true)
```

#### チャンネルへの公開メッセージ
```go
sendChannelEmbed(s, channelID, "タイトル", "説明", fields, 0x57F287, "部室予約システム  |  xxx")
```

### 色コード

| 用途 | 色 | コード |
|------|-----|--------|
| 成功（予約作成） | 🟢 緑 | `0x57F287` |
| 警告（予約編集） | 🟡 黄 | `0xFEE75C` |
| エラー（予約キャンセル） | 🔴 赤 | `0xED4245` |
| 情報（予約完了） | 🔵 青 | `0x5865F2` |
| 一覧（全予約） | ⚫ 黒 | `0x000000` |
| 一覧（自分の予約） | ⚪ 白 | `0xFFFFFF` |

### 機密情報の扱い

- **Ephemeralメッセージ（実行者のみ）**: 予約IDなどの機密情報を含める
- **パブリックメッセージ（全員）**: 予約IDなどの機密情報を含めない

#### パターン1: 機密情報を除外する場合（editコマンド等）

```go
// 実行者へのレスポンス（予約IDを含む）
fields := []*discordgo.MessageEmbedField{
	{Name: "🆔 予約ID", Value: reservationID, Inline: false},
	{Name: "📅 日付", Value: date, Inline: false},
	{Name: "🕐 時間", Value: time, Inline: false},
}
respondEmbedWithFooter(s, i, "成功", "", fields, 0x57F287, "footer", true)

// 公開メッセージ（予約IDを除外 - fields[1:]を使用）
sendChannelEmbed(s, channelID, "通知", "予約が更新されました", fields[1:], 0x57F287, "footer")
```

#### パターン2: 追加フィールドが必要な場合（reserveコマンド等）

```go
// 実行者へのレスポンス（予約IDを含む）
fields := []*discordgo.MessageEmbedField{
	{Name: "🆔 予約ID", Value: reservationID, Inline: false},
	{Name: "📅 日付", Value: date, Inline: false},
	{Name: "🕐 時間", Value: time, Inline: false},
}
respondEmbedWithFooter(s, i, "成功", "", fields, 0x57F287, "footer", true)

// 公開メッセージ（予約者を追加し、予約IDを除外）
publicFields := []*discordgo.MessageEmbedField{
	{Name: "👤 予約者", Value: fmt.Sprintf("<@%s>", userID), Inline: false},
}
publicFields = append(publicFields, fields[1:]...) // 予約ID以降のフィールドを追加
sendChannelEmbed(s, channelID, "通知", "", publicFields, 0x57F287, "footer")
```

#### パターン3: 機密情報がない場合（cancelコマンド等）

```go
// 実行者へのレスポンスと公開メッセージで同じfieldsを使用
fields := []*discordgo.MessageEmbedField{
	{Name: "📅 日付", Value: date, Inline: false},
	{Name: "🕐 時間", Value: time, Inline: false},
}
respondEmbedWithFooter(s, i, "成功", "", fields, 0x57F287, "footer", true)
sendChannelEmbed(s, channelID, "通知", "", fields, 0x57F287, "footer")
```

---

## 新しいコマンドの追加手順

### 1. コマンド定義を追加（`cmd/bot/main.go`）

```go
commands := []*discordgo.ApplicationCommand{
	// ... 既存のコマンド
	{
		Name:        "xxx",
		Description: "XXXを実行します",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "param1",
				Description: "パラメータ1",
				Required:    true,
			},
		},
	},
}
```

### 2. ルーティングを追加（`internal/commands/handlers.go`）

```go
func HandleInteraction(...) {
	switch commandName {
	// ... 既存のケース
	case "xxx":
		handleXxx(s, i, store, logger, allowedChannelID, isDM)
	}
}
```

### 3. ハンドラーファイルを作成（`internal/commands/cmd_xxx.go`）

上記のテンプレートを使用して実装します。

### 4. ビルド＆テスト

```bash
make check  # フォーマット + 静的解析
make build  # ビルド
make run    # 実行
```

---

## 参考リンク

- [開発者ガイド](DEVELOPMENT.md)
- [コマンドリファレンス](COMMANDS.md)
- [既存コマンド実装](../internal/commands/)

---

**最終更新**: 2025-11-17

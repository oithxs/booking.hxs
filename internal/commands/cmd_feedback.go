package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
)

// handleFeedback はフィードバックコマンドを処理する（匿名で特定チャンネルに転送）
func handleFeedback(s *discordgo.Session, i *discordgo.InteractionCreate, logger *logging.Logger, isDM bool) {
	// 1. オプション取得とバリデーション
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondError(s, i, "フィードバック内容を入力してください")
		return
	}

	// 2. パラメータ抽出
	message := options[0].StringValue()
	if message == "" {
		respondError(s, i, "フィードバック内容を入力してください")
		return
	}

	// 3. ユーザー情報取得
	userID, username := getUserInfo(i, isDM)

	// 4. ビジネスロジック - 環境変数からフィードバックチャンネルIDを取得
	feedbackChannelID := os.Getenv("FEEDBACK_CHANNEL_ID")
	if feedbackChannelID == "" {
		respondError(s, i, "フィードバックチャンネルが設定されていません。管理者に連絡してください。")
		logger.LogCommand("feedback", userID, username, i.ChannelID, false, "FEEDBACK_CHANNEL_ID not set", map[string]interface{}{"message_length": len(message)})
		return
	}

	// 5. チャンネル送信 - フィードバックチャンネルに匿名で転送
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	feedbackFields := []*discordgo.MessageEmbedField{}
	err := sendChannelEmbed(s, feedbackChannelID, "💬 新しいフィードバック", message, feedbackFields, 0x5865F2, fmt.Sprintf("部室予約システム  |  feedback  |  受信日時: %s", timestamp))
	if err != nil {
		respondError(s, i, "フィードバックの送信に失敗しました。管理者に連絡してください。")
		logger.LogCommand("feedback", userID, username, i.ChannelID, false, fmt.Sprintf("Failed to send feedback: %v", err), map[string]interface{}{"message_length": len(message)})
		return
	}

	// 6. レスポンス - 送信者に確認メッセージを表示（自分だけに見える）
	respondEmbed(s, i, "✅ フィードバックを送信しました",
		"ご意見ありがとうございます。\nあなたのフィードバックは匿名で運営チームに届けられました。\n\n今後のシステム改善に活用させていただきます。",
		0x57F287, true)

	// 7. ログ記録 - メッセージの長さのみ記録、内容は記録しない
	logger.LogCommand("feedback", userID, username, i.ChannelID, true, "", map[string]interface{}{"message_length": len(message)})
}

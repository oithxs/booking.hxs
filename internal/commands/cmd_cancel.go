package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
	"github.com/dice/hxs_reservation_system/internal/models"
	"github.com/dice/hxs_reservation_system/internal/storage"
)

// handleCancel は予約キャンセルコマンドを処理する
func handleCancel(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool) {
	// 1. オプション取得
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	// 2. パラメータ抽出
	reservationID := optionMap["reservation_id"].StringValue()

	comment := ""
	if opt, ok := optionMap["comment"]; ok {
		comment = opt.StringValue()
	}

	// 3. ビジネスロジック - 予約を取得
	reservation, err := store.GetReservation(reservationID)
	if err != nil {
		respondError(s, i, "予約が見つかりませんでした。予約IDを確認してください。")
		return
	}

	// 予約をキャンセル済みに更新
	reservation.Status = models.StatusCancelled
	reservation.UpdatedAt = time.Now()

	if err := store.UpdateReservation(reservation); err != nil {
		respondError(s, i, "予約の更新に失敗しました")
		logger.LogError("ERROR", "handlers.handleCancel", "Failed to update reservation", err, map[string]interface{}{
			"reservation_id": reservationID,
		})
		return
	}

	if err := store.Save(); err != nil {
		respondError(s, i, "予約の保存に失敗しました")
		logger.LogError("ERROR", "handlers.handleCancel", "Failed to save reservations", err, map[string]interface{}{
			"reservation_id": reservationID,
		})
		return
	}

	// 4. レスポンス - 応答
	respondEmbed(s, i, "🔴 予約を取り消しました", fmt.Sprintf("予約ID: `%s`", reservationID), 0xED4245, true)

	// 5. チャンネル通知
	cancelFields := []*discordgo.MessageEmbedField{
		{
			Name:   "👤 予約者",
			Value:  fmt.Sprintf("<@%s>", reservation.UserID),
			Inline: false,
		},
		{
			Name:   "📅 日付",
			Value:  formatDate(reservation.Date),
			Inline: true,
		},
		{
			Name:   "🕐 時間",
			Value:  fmt.Sprintf("%s - %s", reservation.StartTime, reservation.EndTime),
			Inline: true,
		},
	}
	if comment != "" {
		cancelFields = append(cancelFields, &discordgo.MessageEmbedField{
			Name:   "💬 コメント",
			Value:  comment,
			Inline: false,
		})
	}
	// DMから実行された場合も、指定チャンネルに通知
	sendChannelEmbed(s, allowedChannelID, "🔴 予約が取り消されました", "", cancelFields, 0xED4245, "部室予約システム  |  cancel")

	// 6. Botステータス更新
	if UpdateStatusCallback != nil {
		UpdateStatusCallback()
	}
}

package commands

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
	"github.com/dice/hxs_reservation_system/internal/models"
	"github.com/dice/hxs_reservation_system/internal/storage"
)

// handleList はすべての予約一覧を表示する
func handleList(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, isDM bool) {
	// 1. データ取得 - すべての予約を取得
	allReservations := store.GetAllReservations()

	// 2. データ処理 - 完了・キャンセル済みを除外
	reservations := make([]*models.Reservation, 0)
	for _, r := range allReservations {
		if r.Status != models.StatusCompleted && r.Status != models.StatusCancelled {
			reservations = append(reservations, r)
		}
	}

	// 3. レスポンス - 予約がない場合
	if len(reservations) == 0 {
		respondEmbed(s, i, "⚫ 予約一覧", "現在、予約はありません。", 0x000000, true)
		return
	}

	// 4. データ処理 - 日時でソート
	sort.Slice(reservations, func(a, b int) bool {
		tA, errA := reservations[a].GetStartDateTime()
		tB, errB := reservations[b].GetStartDateTime()
		if errA != nil || errB != nil {
			return a < b
		}
		return tA.Before(tB)
	})

	// 5. レスポンス - 最初のメッセージ（ヘッダー + 最初の予約9件）
	embeds := []*discordgo.MessageEmbed{}

	// ヘッダー
	headerDescription := fmt.Sprintf("現在 %d 件の予約があります", len(reservations))
	headerEmbed := createHeaderEmbed("⚫ すべての予約一覧", headerDescription, 0x000000, "部室予約システム  |  list")
	embeds = append(embeds, headerEmbed)

	// 最初の9件を表示
	maxFirstMessage := 9
	for idx := 0; idx < len(reservations) && idx < maxFirstMessage; idx++ {
		r := reservations[idx]

		fields := []*discordgo.MessageEmbedField{
			{
				Name:   "👤 予約者",
				Value:  fmt.Sprintf("<@%s>", r.UserID),
				Inline: false,
			},
			{
				Name:   "📅 日付",
				Value:  formatDate(r.Date),
				Inline: true,
			},
			{
				Name:   "🕐 時間",
				Value:  fmt.Sprintf("%s - %s", r.StartTime, r.EndTime),
				Inline: true,
			},
		}

		if r.Comment != "" {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "💬 コメント",
				Value:  r.Comment,
				Inline: false,
			})
		}

		reservationEmbed := createReservationEmbed(
			fmt.Sprintf("No.%d", idx+1),
			fields,
			0x000000,
			fmt.Sprintf("部室予約システム  |  list  |  予約 %d/%d", idx+1, len(reservations)),
		)
		embeds = append(embeds, reservationEmbed)
	}

	// 最初のメッセージを送信
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: embeds,
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})

	// 残りの予約を複数のメッセージで送信（10件ごと）
	if len(reservations) > maxFirstMessage {
		// 10件ずつ（ヘッダーなし）のフォローアップメッセージを送信
		itemsPerMessage := 10
		for startIdx := maxFirstMessage; startIdx < len(reservations); startIdx += itemsPerMessage {
			endIdx := startIdx + itemsPerMessage
			if endIdx > len(reservations) {
				endIdx = len(reservations)
			}

			messageEmbeds := []*discordgo.MessageEmbed{}
			for idx := startIdx; idx < endIdx; idx++ {
				r := reservations[idx]

				fields := []*discordgo.MessageEmbedField{
					{
						Name:   "👤 予約者",
						Value:  fmt.Sprintf("<@%s>", r.UserID),
						Inline: false,
					},
					{
						Name:   "📅 日付",
						Value:  formatDate(r.Date),
						Inline: true,
					},
					{
						Name:   "🕐 時間",
						Value:  fmt.Sprintf("%s - %s", r.StartTime, r.EndTime),
						Inline: true,
					},
				}

				if r.Comment != "" {
					fields = append(fields, &discordgo.MessageEmbedField{
						Name:   "💬 コメント",
						Value:  r.Comment,
						Inline: false,
					})
				}

				reservationEmbed := createReservationEmbed(
					fmt.Sprintf("No.%d", idx+1),
					fields,
					0x000000,
					fmt.Sprintf("部室予約システム  |  list  |  予約 %d/%d", idx+1, len(reservations)),
				)
				messageEmbeds = append(messageEmbeds, reservationEmbed)
			}

			// フォローアップメッセージを送信（Ephemeral）
			_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Embeds: messageEmbeds,
				Flags:  discordgo.MessageFlagsEphemeral,
			})
			if err != nil {
				logger.LogError("ERROR", "handleList", "Failed to send followup message", err, map[string]interface{}{
					"start_idx": startIdx,
					"end_idx":   endIdx,
				})
			}
		}
	}
}

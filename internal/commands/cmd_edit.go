package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
	"github.com/dice/hxs_reservation_system/internal/models"
	"github.com/dice/hxs_reservation_system/internal/storage"
)

// handleEdit は予約編集コマンドを処理する
func handleEdit(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool) {
	// 1. オプション取得
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	// 2. ユーザー情報取得
	userID, username := getUserInfo(i, isDM)

	// 3. パラメータ抽出 - 予約IDを取得
	reservationID := optionMap["reservation_id"].StringValue()

	// 4. ビジネスロジック - 予約を取得
	reservation, err := store.GetReservation(reservationID)
	if err != nil {
		respondError(s, i, "指定された予約が見つかりません。")
		return
	}

	// 予約の所有者チェック
	if reservation.UserID != userID {
		respondError(s, i, "他のユーザーの予約は編集できません。")
		return
	}

	// ステータスチェック
	if reservation.Status != models.StatusPending {
		respondError(s, i, "完了またはキャンセルされた予約は編集できません。")
		return
	}

	// 変更前の情報を保持
	oldDate := reservation.Date
	oldStartTime := reservation.StartTime
	oldEndTime := reservation.EndTime
	oldComment := reservation.Comment

	// 新しい値を取得（指定されていない場合は現在の値を保持）
	newDate := oldDate
	newStartTime := oldStartTime
	newEndTime := oldEndTime
	newComment := oldComment

	hasChanges := false

	// 日付の変更
	if opt, ok := optionMap["date"]; ok {
		dateStr := opt.StringValue()
		// 日付を正規化
		dateStr = normalizeDate(dateStr)

		// 日付の形式を検証
		var parsedDate time.Time
		if t, err := time.Parse("2006-01-02", dateStr); err != nil {
			if t2, err2 := time.Parse("2006/01/02", dateStr); err2 == nil {
				dateStr = t2.Format("2006-01-02")
				parsedDate = t2
			} else {
				respondError(s, i, "日付の形式が正しくありません（YYYY-MM-DD または YYYY/MM/DD 形式で入力してください）")
				return
			}
		} else {
			parsedDate = t
		}

		// 過去の日付チェック
		jst := time.FixedZone("Asia/Tokyo", 9*60*60)
		now := time.Now().In(jst)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)
		if parsedDate.Before(today) {
			respondError(s, i, "過去の日付には変更できません。")
			return
		}

		newDate = dateStr
		hasChanges = true
	}

	// 開始時間の変更
	if opt, ok := optionMap["start_time"]; ok {
		timeStr := opt.StringValue()
		// 時刻を正規化
		timeStr = normalizeTime(timeStr)

		if _, err := time.Parse("15:04", timeStr); err != nil {
			respondError(s, i, "開始時間の形式が正しくありません（HH:MM形式で入力してください）")
			return
		}
		newStartTime = timeStr
		hasChanges = true
	}

	// 終了時間の変更
	if opt, ok := optionMap["end_time"]; ok {
		timeStr := opt.StringValue()
		// 時刻を正規化
		timeStr = normalizeTime(timeStr)

		if _, err := time.Parse("15:04", timeStr); err != nil {
			respondError(s, i, "終了時間の形式が正しくありません（HH:MM形式で入力してください）")
			return
		}
		newEndTime = timeStr
		hasChanges = true
	}

	// コメントの変更
	if opt, ok := optionMap["comment"]; ok {
		newComment = opt.StringValue()
		hasChanges = true
	}

	// 変更がない場合
	if !hasChanges {
		respondError(s, i, "変更する項目を少なくとも1つ指定してください。")
		return
	}

	// 時刻の整合性チェック
	if newEndTime <= newStartTime {
		respondError(s, i, "終了時間は開始時間より後である必要があります。")
		return
	}

	// 重複チェック用に一時的な予約オブジェクトを作成
	tempReservation := &models.Reservation{
		ID:        reservationID, // 自分の予約は除外するためにIDを設定
		UserID:    userID,
		Username:  username,
		Date:      newDate,
		StartTime: newStartTime,
		EndTime:   newEndTime,
		Comment:   newComment,
		Status:    models.StatusPending,
	}

	// 時間の重複をチェック（自分の予約以外と）
	overlappingReservation, err := store.CheckOverlap(tempReservation)
	if err != nil {
		respondError(s, i, "予約の重複チェックに失敗しました")
		logger.LogError("ERROR", "handleEdit", "Failed to check overlap", err, map[string]interface{}{
			"reservation_id": reservationID,
		})
		return
	}

	if overlappingReservation != nil {
		fields := []*discordgo.MessageEmbedField{
			{
				Name:   "📅 日付",
				Value:  strings.ReplaceAll(newDate, "-", "/"),
				Inline: false,
			},
			{
				Name:   "👤 予約者",
				Value:  fmt.Sprintf("<@%s>", overlappingReservation.UserID),
				Inline: true,
			},
			{
				Name:   "🕐 時間",
				Value:  fmt.Sprintf("%s - %s", overlappingReservation.StartTime, overlappingReservation.EndTime),
				Inline: true,
			},
		}

		respondEmbedWithFooter(s, i, "🔴 予約を編集できませんでした", "指定された時間は既に予約されています。", fields, 0xED4245, "部室予約システム  |  edit", true)
		return
	}

	// 予約を更新
	reservation.Date = newDate
	reservation.StartTime = newStartTime
	reservation.EndTime = newEndTime
	reservation.Comment = newComment

	if err := store.Save(); err != nil {
		respondError(s, i, "予約の更新に失敗しました。")
		logger.LogError("ERROR", "handleEdit", "Failed to save reservation", err, map[string]interface{}{
			"reservation_id": reservationID,
		})
		return
	}

	// 成功メッセージ
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "🆔 予約ID",
			Value:  reservation.ID,
			Inline: false,
		},
	}

	// 変更内容を表示
	if oldDate != newDate {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "📅 日付",
			Value:  fmt.Sprintf("%s → %s", strings.ReplaceAll(oldDate, "-", "/"), strings.ReplaceAll(newDate, "-", "/")),
			Inline: false,
		})
	}
	if oldStartTime != newStartTime || oldEndTime != newEndTime {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "🕐 時間",
			Value:  fmt.Sprintf("%s-%s → %s-%s", oldStartTime, oldEndTime, newStartTime, newEndTime),
			Inline: false,
		})
	}
	if oldComment != newComment {
		oldCommentDisplay := oldComment
		if oldCommentDisplay == "" {
			oldCommentDisplay = "（なし）"
		}
		newCommentDisplay := newComment
		if newCommentDisplay == "" {
			newCommentDisplay = "（なし）"
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "💬 コメント",
			Value:  fmt.Sprintf("%s → %s", oldCommentDisplay, newCommentDisplay),
			Inline: false,
		})
	}

	// 5. レスポンス
	respondEmbedWithFooter(s, i, "🟡 予約を編集しました", "", fields, 0xFEE75C, "部室予約システム  |  edit", true)

	// 6. チャンネル通知(変更がある場合) - 予約IDを除外したfieldsを使用
	if !isDM {
		sendChannelEmbed(s, allowedChannelID, "🟡 予約が編集されました", fmt.Sprintf("<@%s> さんが予約を編集しました", userID), fields[1:], 0xFEE75C, "部室予約システム  |  edit")
	} else if allowedChannelID != "" {
		// DMから実行された場合も、指定チャンネルに通知
		sendChannelEmbed(s, allowedChannelID, "🟡 予約が編集されました", fmt.Sprintf("%s さんが予約を編集しました", username), fields[1:], 0xFEE75C, "部室予約システム  |  edit")
	}

	// 7. Botステータス更新
	if UpdateStatusCallback != nil {
		UpdateStatusCallback()
	}
}

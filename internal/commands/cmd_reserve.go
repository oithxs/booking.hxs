package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/logging"
	"github.com/dice/hxs_reservation_system/internal/models"
	"github.com/dice/hxs_reservation_system/internal/storage"
)

// handleReserve は予約作成コマンドを処理する
func handleReserve(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage, logger *logging.Logger, allowedChannelID string, isDM bool) {
	// 1. オプション取得
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	// 2. ユーザー情報取得
	userID, username := getUserInfo(i, isDM)

	// 3. パラメータ抽出 - 必須パラメータを取得
	date := optionMap["date"].StringValue()
	startTime := optionMap["start_time"].StringValue()

	// 日付を正規化（YYYY/M/D → YYYY/MM/DD）
	date = normalizeDate(date)

	// 時刻を正規化（H:MM → HH:MM）
	startTime = normalizeTime(startTime)

	// オプションパラメータを取得
	var endTime string
	if opt, ok := optionMap["end_time"]; ok {
		endTime = opt.StringValue()
		// 時刻を正規化（H:MM → HH:MM）
		endTime = normalizeTime(endTime)
	} else {
		// 終了時間が指定されていない場合は開始時刻+1時間
		start, err := time.Parse("15:04", startTime)
		if err != nil {
			respondError(s, i, "開始時間の形式が正しくありません（HH:MM形式で入力してください）")
			return
		}
		endTime = start.Add(1 * time.Hour).Format("15:04")
	}

	comment := ""
	if opt, ok := optionMap["comment"]; ok {
		comment = opt.StringValue()
	}

	// ログ用パラメータを構築
	parameters := map[string]interface{}{
		"date":       date,
		"start_time": startTime,
		"end_time":   endTime,
	}
	if comment != "" {
		parameters["comment"] = comment
	}

	// 4. ビジネスロジック - 日付と時間の形式を検証（YYYY-MM-DD または YYYY/MM/DD を許可）
	var reservationDate time.Time
	if parsedDate, err := time.Parse("2006-01-02", date); err != nil {
		if t2, err2 := time.Parse("2006/01/02", date); err2 == nil {
			// 正規化して保存用は YYYY-MM-DD に統一
			date = t2.Format("2006-01-02")
			reservationDate = t2
		} else {
			errorMsg := "日付の形式が正しくありません（YYYY-MM-DD または YYYY/MM/DD）"
			logger.LogCommand("reserve", userID, username, i.ChannelID, false, errorMsg, parameters)
			respondError(s, i, errorMsg)
			return
		}
	} else {
		reservationDate = parsedDate
	}

	var startTimeParsed time.Time
	if t, err := time.Parse("15:04", startTime); err != nil {
		errorMsg := "開始時間の形式が正しくありません（HH:MM形式で入力してください）"
		logger.LogCommand("reserve", userID, username, i.ChannelID, false, errorMsg, parameters)
		respondError(s, i, errorMsg)
		return
	} else {
		startTimeParsed = t
	}

	if _, err := time.Parse("15:04", endTime); err != nil {
		errorMsg := "終了時間の形式が正しくありません（HH:MM形式で入力してください）"
		logger.LogCommand("reserve", userID, username, i.ChannelID, false, errorMsg, parameters)
		respondError(s, i, errorMsg)
		return
	}

	// 終了時刻が開始時刻より前または同じ時刻でないかチェック
	if endTime <= startTime {
		errorMsg := fmt.Sprintf("❌ 終了時刻は開始時刻より後である必要があります\n\n"+
			"**開始時刻:** %s\n"+
			"**終了時刻:** %s\n\n"+
			"終了時刻を開始時刻より後の時刻に設定してください。",
			startTime,
			endTime,
		)
		logger.LogCommand("reserve", userID, username, i.ChannelID, false, "End time before start time", parameters)
		respondEphemeral(s, i, errorMsg)
		return
	}

	// 過去日時のチェック
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	nowJST := time.Now().In(jst)

	// 予約日時を構築（日付 + 開始時刻）
	reservationDateTime := time.Date(
		reservationDate.Year(),
		reservationDate.Month(),
		reservationDate.Day(),
		startTimeParsed.Hour(),
		startTimeParsed.Minute(),
		0, 0, jst,
	)

	// 現在時刻より過去の場合はエラー
	if reservationDateTime.Before(nowJST) {
		errorMsg := fmt.Sprintf("❌ 過去の日時は予約できません\n\n"+
			"**指定された日時:** %s %s\n"+
			"**現在日時:** %s\n\n"+
			"現在時刻以降の日時を指定してください。",
			formatDate(date),
			startTime,
			nowJST.Format("2006-01-02 15:04"),
		)
		logger.LogCommand("reserve", userID, username, i.ChannelID, false, "Past datetime", parameters)
		respondEphemeral(s, i, errorMsg)
		return
	}

	// 予約IDを生成
	reservationID, err := models.GenerateReservationID()
	if err != nil {
		respondError(s, i, "予約IDの生成に失敗しました")
		return
	}

	// 予約を作成
	reservation := &models.Reservation{
		ID:        reservationID,
		UserID:    userID,
		Username:  username,
		Date:      date,
		StartTime: startTime,
		EndTime:   endTime,
		Comment:   comment,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ChannelID: allowedChannelID, // 公開メッセージの送信先は常に指定チャンネル
	}

	// 時間の重複をチェック
	overlappingReservation, err := store.CheckOverlap(reservation)
	if err != nil {
		respondError(s, i, "予約の重複チェックに失敗しました")
		logger.LogError("ERROR", "handlers.handleReserve", "Failed to check overlap", err, map[string]interface{}{
			"user_id": userID,
			"date":    date,
		})
		return
	}

	if overlappingReservation != nil {
		fields := []*discordgo.MessageEmbedField{
			{
				Name:   "📅 重複している予約",
				Value:  formatDate(overlappingReservation.Date),
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

		respondEmbedWithFooter(s, i, "🔴 予約できませんでした", "指定された時間は既に予約されています。", fields, 0xED4245, "部室予約システム  |  reserve", true)
		return
	}

	// 予約を保存
	if err := store.AddReservation(reservation); err != nil {
		respondError(s, i, "予約の保存に失敗しました")
		logger.LogError("ERROR", "handlers.handleReserve", "Failed to add reservation", err, map[string]interface{}{
			"user_id":        userID,
			"reservation_id": reservation.ID,
		})
		return
	}

	if err := store.Save(); err != nil {
		respondError(s, i, "予約の保存に失敗しました")
		logger.LogError("ERROR", "handlers.handleReserve", "Failed to save reservations", err, map[string]interface{}{
			"user_id":        userID,
			"reservation_id": reservation.ID,
		})
		return
	}

	// 5. レスポンス - 予約者にはIDを含めたメッセージを送信（Ephemeral）
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "予約ID",
			Value:  fmt.Sprintf("`%s`", reservation.ID),
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
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "💬 コメント",
			Value:  comment,
			Inline: false,
		})
	}

	respondEmbedWithFooter(s, i, "🟢 予約が完了しました！", "", fields, 0x57F287, "部室予約システム  |  reserve", true)

	// 6. チャンネル通知 - 予約IDを除外し、予約者フィールドを追加
	publicFields := []*discordgo.MessageEmbedField{
		{
			Name:   "👤 予約者",
			Value:  fmt.Sprintf("<@%s>", reservation.UserID),
			Inline: false,
		},
	}
	publicFields = append(publicFields, fields[1:]...) // 予約ID以降のフィールドを追加
	// DMから実行された場合も、指定チャンネルに通知
	sendChannelEmbed(s, allowedChannelID, "🟢 新しい予約が追加されました", "", publicFields, 0x57F287, "部室予約システム  |  reserve")

	// 7. Botステータス更新
	if UpdateStatusCallback != nil {
		UpdateStatusCallback()
	}
}

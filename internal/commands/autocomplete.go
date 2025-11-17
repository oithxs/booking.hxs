package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/models"
	"github.com/dice/hxs_reservation_system/internal/storage"
)

// HandleAutocomplete はオートコンプリートのリクエストを処理する
func HandleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, store *storage.Storage) {
	data := i.ApplicationCommandData()

	// 現在フォーカスされているオプションを取得
	var focusedOption *discordgo.ApplicationCommandInteractionDataOption
	for _, opt := range data.Options {
		if opt.Focused {
			focusedOption = opt
			break
		}
	}

	if focusedOption == nil {
		return
	}

	var choices []*discordgo.ApplicationCommandOptionChoice

	// コマンド名を取得
	commandName := data.Name

	switch focusedOption.Name {
	case "date":
		choices = getDateSuggestions(focusedOption.StringValue())
	case "start_time":
		choices = getTimeSuggestions(focusedOption.StringValue(), "")
	case "end_time":
		// end_timeの場合、start_timeを取得して考慮する
		var startTime string
		for _, opt := range data.Options {
			if opt.Name == "start_time" {
				startTime = opt.StringValue()
				break
			}
		}
		choices = getTimeSuggestions(focusedOption.StringValue(), startTime)
	case "reservation_id":
		// ユーザーIDを取得
		var userID string
		if i.Member != nil {
			userID = i.Member.User.ID
		} else if i.User != nil {
			userID = i.User.ID
		}

		// コマンドに応じて候補を生成
		if commandName == "cancel" || commandName == "complete" || commandName == "edit" {
			choices = getReservationSuggestions(store, userID, "pending", focusedOption.StringValue())
		}
	}

	// 最大25個まで
	if len(choices) > 25 {
		choices = choices[:25]
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})

	if err != nil {
		// Autocompleteのエラーはログのみ
		fmt.Printf("Failed to respond to autocomplete: %v\n", err)
	}
}

// getWeekdayJa は日本語の曜日を返す
func getWeekdayJa(t time.Time) string {
	weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}
	return weekdays[int(t.Weekday())]
}

// formatDateWithWeekday は日付を曜日付きでフォーマットする
func formatDateWithWeekday(t time.Time) string {
	return fmt.Sprintf("%s (%s)", t.Format("2006/01/02"), getWeekdayJa(t))
}

// getDateSuggestions は日付の候補を生成する
func getDateSuggestions(input string) []*discordgo.ApplicationCommandOptionChoice {
	now := time.Now()
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	nowJST := now.In(jst)

	// 入力が空の場合
	if input == "" {
		today := nowJST
		tomorrow := nowJST.AddDate(0, 0, 1)
		dayAfterTomorrow := nowJST.AddDate(0, 0, 2)

		suggestions := []*discordgo.ApplicationCommandOptionChoice{
			{Name: fmt.Sprintf("今日 %s (%s)", today.Format("2006/01/02"), getWeekdayJa(today)), Value: today.Format("2006/01/02")},
			{Name: fmt.Sprintf("明日 %s (%s)", tomorrow.Format("2006/01/02"), getWeekdayJa(tomorrow)), Value: tomorrow.Format("2006/01/02")},
			{Name: fmt.Sprintf("明後日 %s (%s)", dayAfterTomorrow.Format("2006/01/02"), getWeekdayJa(dayAfterTomorrow)), Value: dayAfterTomorrow.Format("2006/01/02")},
		}

		// 3日後から30日後まで
		for i := 3; i <= 30; i++ {
			if i%7 == 0 && i <= 28 {
				week := i / 7
				futureDate := nowJST.AddDate(0, 0, i)
				suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
					Name:  fmt.Sprintf("%d週間後 %s (%s)", week, futureDate.Format("2006/01/02"), getWeekdayJa(futureDate)),
					Value: futureDate.Format("2006/01/02"),
				})
			} else {
				futureDate := nowJST.AddDate(0, 0, i)
				suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
					Name:  formatDateWithWeekday(futureDate),
					Value: futureDate.Format("2006/01/02"),
				})
			}
		}
		return suggestions
	}

	// 月の候補を生成（1-12の入力を月として優先的に扱う）
	if len(input) <= 2 {
		if monthNum, err := strconv.Atoi(input); err == nil && monthNum >= 1 && monthNum <= 12 {
			year := nowJST.Year()
			suggestions := []*discordgo.ApplicationCommandOptionChoice{}
			for yearOffset := 0; yearOffset <= 1 && len(suggestions) < 25; yearOffset++ {
				targetYear := year + yearOffset
				daysInMonth := time.Date(targetYear, time.Month(monthNum+1), 0, 0, 0, 0, 0, jst).Day()

				for day := 1; day <= daysInMonth && len(suggestions) < 25; day++ {
					dateTime := time.Date(targetYear, time.Month(monthNum), day, 0, 0, 0, 0, jst)
					suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
						Name:  formatDateWithWeekday(dateTime),
						Value: dateTime.Format("2006/01/02"),
					})
				}
			}

			if len(suggestions) > 0 {
				return suggestions
			}
		}
	}

	// 年の候補を生成（13以上の2桁入力、または月候補がない場合）
	if len(input) == 2 {
		if yearNum, err := strconv.Atoi(input); err == nil {
			currentYear := nowJST.Year()
			currentCentury := (currentYear / 100) * 100
			fullYear := currentCentury + yearNum

			if fullYear < currentYear-10 {
				fullYear += 100
			}

			suggestions := []*discordgo.ApplicationCommandOptionChoice{}
			for month := 1; month <= 12 && len(suggestions) < 25; month++ {
				for day := 1; day <= 7 && len(suggestions) < 25; day++ {
					dateTime := time.Date(fullYear, time.Month(month), day, 0, 0, 0, 0, jst)
					suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
						Name:  formatDateWithWeekday(dateTime),
						Value: dateTime.Format("2006/01/02"),
					})
				}
			}

			if len(suggestions) > 0 {
				return suggestions
			}
		}
	}

	// 日の候補を生成
	if len(input) <= 2 {
		if dayNum, err := strconv.Atoi(input); err == nil && dayNum >= 1 && dayNum <= 31 {
			year := nowJST.Year()
			month := int(nowJST.Month())
			suggestions := []*discordgo.ApplicationCommandOptionChoice{}

			for monthOffset := 0; monthOffset <= 3 && len(suggestions) < 25; monthOffset++ {
				targetMonth := month + monthOffset
				targetYear := year

				if targetMonth > 12 {
					targetYear++
					targetMonth -= 12
				}

				daysInMonth := time.Date(targetYear, time.Month(targetMonth+1), 0, 0, 0, 0, 0, jst).Day()

				if dayNum <= daysInMonth {
					dateTime := time.Date(targetYear, time.Month(targetMonth), dayNum, 0, 0, 0, 0, jst)
					suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
						Name:  formatDateWithWeekday(dateTime),
						Value: dateTime.Format("2006/01/02"),
					})
				}
			}

			if len(suggestions) > 0 {
				return suggestions
			}
		}
	}

	// 通常のフィルタリング処理
	today := nowJST
	tomorrow := nowJST.AddDate(0, 0, 1)
	dayAfterTomorrow := nowJST.AddDate(0, 0, 2)

	allSuggestions := []*discordgo.ApplicationCommandOptionChoice{
		{Name: fmt.Sprintf("今日 %s (%s)", today.Format("2006/01/02"), getWeekdayJa(today)), Value: today.Format("2006/01/02")},
		{Name: fmt.Sprintf("明日 %s (%s)", tomorrow.Format("2006/01/02"), getWeekdayJa(tomorrow)), Value: tomorrow.Format("2006/01/02")},
		{Name: fmt.Sprintf("明後日 %s (%s)", dayAfterTomorrow.Format("2006/01/02"), getWeekdayJa(dayAfterTomorrow)), Value: dayAfterTomorrow.Format("2006/01/02")},
	}

	for i := 3; i <= 30; i++ {
		if i%7 == 0 && i <= 28 {
			week := i / 7
			futureDate := nowJST.AddDate(0, 0, i)
			allSuggestions = append(allSuggestions, &discordgo.ApplicationCommandOptionChoice{
				Name:  fmt.Sprintf("%d週間後 %s (%s)", week, futureDate.Format("2006/01/02"), getWeekdayJa(futureDate)),
				Value: futureDate.Format("2006/01/02"),
			})
		} else {
			futureDate := nowJST.AddDate(0, 0, i)
			allSuggestions = append(allSuggestions, &discordgo.ApplicationCommandOptionChoice{
				Name:  formatDateWithWeekday(futureDate),
				Value: futureDate.Format("2006/01/02"),
			})
		}
	}

	// 入力でフィルタリング
	suggestions := []*discordgo.ApplicationCommandOptionChoice{}
	for _, choice := range allSuggestions {
		if strings.Contains(choice.Value.(string), input) || strings.Contains(choice.Name, input) {
			suggestions = append(suggestions, choice)
			if len(suggestions) >= 25 {
				break
			}
		}
	}

	if len(suggestions) > 0 {
		return suggestions
	}

	return allSuggestions
}

// getTimeSuggestions は時刻の候補を生成する
func getTimeSuggestions(input string, startTime string) []*discordgo.ApplicationCommandOptionChoice {
	// 9:00から21:00まで30分刻みで候補を生成
	suggestions := []*discordgo.ApplicationCommandOptionChoice{}
	for hour := 9; hour <= 21; hour++ {
		for _, minute := range []int{0, 30} {
			if hour == 21 && minute == 30 {
				break // 21:00で終了
			}
			timeStr := fmt.Sprintf("%02d:%02d", hour, minute)
			suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
				Name:  timeStr,
				Value: timeStr,
			})
		}
	}

	// end_timeの場合、start_timeより後の時刻のみフィルタリング
	if startTime != "" {
		var filtered []*discordgo.ApplicationCommandOptionChoice
		for _, choice := range suggestions {
			if choice.Value.(string) > startTime {
				filtered = append(filtered, choice)
			}
		}
		suggestions = filtered
	}

	// 入力がある場合、さらにフィルタリング
	if input != "" {
		var filtered []*discordgo.ApplicationCommandOptionChoice
		for _, choice := range suggestions {
			if strings.HasPrefix(choice.Value.(string), input) {
				filtered = append(filtered, choice)
			}
		}
		if len(filtered) > 0 {
			return filtered
		}
	}

	return suggestions
}

// getReservationSuggestions はユーザーの予約候補を生成する
func getReservationSuggestions(store *storage.Storage, userID string, status string, input string) []*discordgo.ApplicationCommandOptionChoice {
	suggestions := []*discordgo.ApplicationCommandOptionChoice{}
	reservations := store.GetUserReservations(userID)

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	now := time.Now().In(jst)

	var filteredReservations []*models.Reservation
	for _, r := range reservations {
		if r.Status == models.ReservationStatus(status) {
			reservationDate, err := time.Parse("2006-01-02", r.Date)
			if err != nil {
				continue
			}

			if !reservationDate.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, jst)) {
				filteredReservations = append(filteredReservations, r)
			}
		}
	}

	sort.Slice(filteredReservations, func(i, j int) bool {
		if filteredReservations[i].Date != filteredReservations[j].Date {
			return filteredReservations[i].Date < filteredReservations[j].Date
		}
		return filteredReservations[i].StartTime < filteredReservations[j].StartTime
	})

	for _, r := range filteredReservations {
		displayDate := strings.ReplaceAll(r.Date, "-", "/")
		name := fmt.Sprintf("%s %s-%s", displayDate, r.StartTime, r.EndTime)
		if r.Comment != "" {
			comment := r.Comment
			if len(comment) > 20 {
				comment = comment[:20] + "..."
			}
			name = fmt.Sprintf("%s (%s)", name, comment)
		}

		if input == "" || strings.Contains(r.ID, input) || strings.Contains(name, input) {
			suggestions = append(suggestions, &discordgo.ApplicationCommandOptionChoice{
				Name:  name,
				Value: r.ID,
			})
		}

		if len(suggestions) >= 25 {
			break
		}
	}

	return suggestions
}

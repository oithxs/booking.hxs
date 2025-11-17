package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// respondError はエラーメッセージを送信する
func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	embed := &discordgo.MessageEmbed{
		Title:       "🔴 エラー",
		Description: message,
		Color:       0xED4245, // Discord Red
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// respondEphemeral はエフェメラルメッセージを送信する
func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// respondEmbed は埋め込みメッセージを送信する
func respondEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, title string, description string, color int, ephemeral bool) {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  flags,
		},
	})
}

// respondEmbedWithFooter は埋め込みメッセージをフッター付きで送信する
func respondEmbedWithFooter(s *discordgo.Session, i *discordgo.InteractionCreate, title string, description string, fields []*discordgo.MessageEmbedField, color int, footerText string, ephemeral bool) {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Fields:      fields,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  flags,
		},
	})
}

// getDisplayName はメンバーの表示名を取得する
func getDisplayName(member *discordgo.Member) string {
	if member.Nick != "" {
		return member.Nick
	}
	return member.User.Username
}

// getUserInfo はインタラクションからユーザーIDとユーザー名を取得する
func getUserInfo(i *discordgo.InteractionCreate, isDM bool) (userID, username string) {
	if isDM {
		return i.User.ID, i.User.Username
	}
	return i.Member.User.ID, getDisplayName(i.Member)
}

// normalizeTime は時刻をHH:MM形式に正規化する（H:MM → HH:MM）
func normalizeTime(timeStr string) string {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return timeStr
	}

	// 時と分を2桁にパディング
	hour := parts[0]
	minute := parts[1]

	if len(hour) == 1 {
		hour = "0" + hour
	}
	if len(minute) == 1 {
		minute = "0" + minute
	}

	return hour + ":" + minute
}

// normalizeDate は日付をYYYY/MM/DD形式に正規化する
func normalizeDate(dateStr string) string {
	// /または-で分割
	separator := "/"
	if strings.Contains(dateStr, "-") {
		separator = "-"
	}

	parts := strings.Split(dateStr, separator)
	if len(parts) != 3 {
		return dateStr
	}

	year := parts[0]
	month := parts[1]
	day := parts[2]

	// 月と日を2桁にパディング
	if len(month) == 1 {
		month = "0" + month
	}
	if len(day) == 1 {
		day = "0" + day
	}

	return year + "/" + month + "/" + day
}

// formatDate は日付をYYYY/MM/DD形式にフォーマットする
func formatDate(date string) string {
	parts := strings.Split(date, "-")
	if len(parts) != 3 {
		return date
	}
	year := parts[0]
	month := fmt.Sprintf("%02s", parts[1])
	day := fmt.Sprintf("%02s", parts[2])
	return fmt.Sprintf("%s/%s/%s", year, month, day)
}

// sendChannelEmbed はチャンネルに埋め込みメッセージを送信する
func sendChannelEmbed(s *discordgo.Session, channelID string, title string, description string, fields []*discordgo.MessageEmbedField, color int, footerText string) error {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Fields:      fields,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}
	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// createReservationEmbed は予約情報の埋め込みメッセージを作成する
func createReservationEmbed(title string, fields []*discordgo.MessageEmbedField, color int, footerText string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:     title,
		Fields:    fields,
		Color:     color,
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}
}

// createHeaderEmbed はヘッダー用の埋め込みメッセージを作成する
func createHeaderEmbed(title string, description string, color int, footerText string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: footerText,
		},
	}
}

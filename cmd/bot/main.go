package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dice/hxs_reservation_system/internal/commands"
	"github.com/dice/hxs_reservation_system/internal/logging"
	"github.com/dice/hxs_reservation_system/internal/storage"
	"github.com/joho/godotenv"
)

const (
	saveInterval       = 5 * time.Minute
	logCleanupInterval = 24 * time.Hour
	autoCompleteHour   = 3
	autoCompleteMinute = 0
	cleanupHour        = 3
	cleanupMinute      = 10
	retentionDays      = 30
)

var (
	store                 *storage.Storage
	logger                *logging.Logger
	guildID               string
	allowedChannelID      string
	startupChannelID      string
	startupMessage        string
	processedInteractions sync.Map
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	guildID = os.Getenv("GUILD_ID")
	allowedChannelID = os.Getenv("ALLOWED_CHANNEL_ID")
	startupChannelID = os.Getenv("STARTUP_NOTIFICATION_CHANNEL_ID")
	startupMessage = os.Getenv("STARTUP_NOTIFICATION_MESSAGE")
}

func main() {
	initializeServices()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is not set in environment variables")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	setupHandlers(dg)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds

	if err = dg.Open(); err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer dg.Close()

	updateBotStatus(dg, store)
	commands.UpdateStatusCallback = func() { updateBotStatus(dg, store) }

	log.Println("Bot is now running. Press CTRL+C to exit.")

	if err := registerCommands(dg); err != nil {
		log.Fatalf("Failed to register commands: %v", err)
	}

	sendStartupNotification(dg)
	startBackgroundTasks(dg)
	waitForShutdown()
	shutdown()
}

func initializeServices() {
	store = storage.NewStorage()
	if err := store.Load(); err != nil {
		log.Fatalf("Failed to load reservations: %v", err)
	}
	log.Println("Reservations loaded successfully")

	logger = logging.NewLogger("./logs")
	log.Println("Logger initialized successfully")
}

func setupHandlers(dg *discordgo.Session) {
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if _, loaded := processedInteractions.LoadOrStore(i.ID, struct{}{}); loaded {
			return
		}

		if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			commands.HandleAutocomplete(s, i, store)
			return
		}

		commands.HandleInteraction(s, i, store, logger, allowedChannelID)
	})
}

func startBackgroundTasks(dg *discordgo.Session) {
	go periodicSave(dg)
	go periodicLogCleanup()
	go dailyAutoComplete()
	go dailyCleanup()
}

func periodicSave(dg *discordgo.Session) {
	ticker := time.NewTicker(saveInterval)
	defer ticker.Stop()
	for range ticker.C {
		if err := store.Save(); err != nil {
			log.Printf("❌ Failed to save reservations: %v", err)
			logger.LogError("ERROR", "periodicSave", "Failed to save reservations", err, nil)
		} else {
			log.Println("💾 Reservations saved successfully")
		}
		updateBotStatus(dg, store)
	}
}

func periodicLogCleanup() {
	ticker := time.NewTicker(logCleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		logger.CleanupOldLogs()
	}
}

func dailyAutoComplete() {
	runTaskAtStartup("auto-complete", func() (int, error) {
		return store.AutoCompleteExpiredReservations()
	})

	for {
		time.Sleep(waitUntilTime(autoCompleteHour, autoCompleteMinute))
		count, err := store.AutoCompleteExpiredReservations()
		logTaskResult("auto-complete", count, err, "expired reservation(s)")
	}
}

func dailyCleanup() {
	runTaskAtStartup("cleanup", func() (int, error) {
		return store.CleanupOldReservations(retentionDays)
	})

	for {
		time.Sleep(waitUntilTime(cleanupHour, cleanupMinute))
		count, err := store.CleanupOldReservations(retentionDays)
		logTaskResult("cleanup", count, err, "old reservation(s)")
	}
}

func runTaskAtStartup(taskName string, task func() (int, error)) {
	log.Printf("Startup: Running initial %s check...", taskName)
	count, err := task()
	logTaskResult(taskName, count, err, "")
}

func logTaskResult(taskName string, count int, err error, itemName string) {
	if err != nil {
		log.Printf("❌ Failed to %s: %v", taskName, err)
		logger.LogError("ERROR", taskName, fmt.Sprintf("Failed to %s", taskName), err, map[string]interface{}{
			"retention_days": retentionDays,
		})
	} else if count > 0 {
		if itemName != "" {
			log.Printf("✅ %s: %d %s", taskName, count, itemName)
		} else {
			log.Printf("✅ %s: %d item(s)", taskName, count)
		}
	} else {
		log.Printf("✓ %s check completed: no items to process", taskName)
	}
}

func waitUntilTime(hour, minute int) time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !now.Before(next) {
		next = next.Add(24 * time.Hour)
	}
	duration := time.Until(next)
	log.Printf("Next task scheduled at: %s (in %v)", next.Format("2006-01-02 15:04:05"), duration)
	return duration
}

func waitForShutdown() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
// 何かの意図があって送信したくないときには.envのSTARTUP_NOTIFICATION_CHANNEL_IDを空にしてください
func sendStartupNotification(s *discordgo.Session) {
	// 1. チャンネルID確認
	if startupChannelID == "" {
		log.Println("Startup notification disabled (STARTUP_NOTIFICATION_CHANNEL_ID not set)")
		log.Println("※注意：.envのメッセージを削除してください")
		return
	}

	// 2. メッセージ準備
	message := startupMessage
	if message == "" {
		message = "Bot が再起動しました。\n部室予約システムが利用可能です。"
	}

	// 3. 埋め込みメッセージ作成
	embed := &discordgo.MessageEmbed{
		Title:       "システムメッセージ",
		Description: message,
		Color:       0x00ff00, // 緑色
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "部室予約システム | システムメッセージ",
		},
	}

	// 4. メッセージ送信
	_, err := s.ChannelMessageSendEmbed(startupChannelID, embed)
	if err != nil {
		log.Printf("❌ Failed to send startup notification: %v", err)
		logger.LogError("ERROR", "sendStartupNotification", "Failed to send startup notification", err, map[string]interface{}{
			"channel_id": startupChannelID,
		})
	} else {
		log.Printf("✅ Startup notification sent to channel: %s", startupChannelID)
	}
}

func shutdown() {
	log.Println("💾 Saving reservations before exit...")
	if err := store.Save(); err != nil {
		log.Printf("❌ Failed to save reservations: %v", err)
		logger.LogError("ERROR", "shutdown", "Failed to save reservations on shutdown", err, nil)
	} else {
		log.Println("✅ Reservations saved successfully")
	}

	printStats()
}

func printStats() {
	stats := logger.GetStats()
	log.Println("=== コマンド統計 ===")
	log.Printf("総コマンド数: %d", stats.TotalCommands)
	log.Println("コマンド別統計:")
	for cmd, count := range stats.CommandCounts {
		log.Printf("  %s: %d回", cmd, count)
	}
	log.Println("ユーザー別統計:")
	for userID, count := range stats.UserCounts {
		log.Printf("  %s: %d回", userID, count)
	}
	log.Printf("最終更新: %s", stats.LastUpdated.Format("2006-01-02 15:04:05"))
}

func updateBotStatus(s *discordgo.Session, store *storage.Storage) {
	pendingCount := 0
	for _, r := range store.GetAllReservations() {
		if r.Status == "pending" {
			pendingCount++
		}
	}

	status := "部室予約管理 | /help"
	if pendingCount > 0 {
		status = fmt.Sprintf("%d件の予約管理中 | /help", pendingCount)
	}

	if err := s.UpdateGameStatus(0, status); err != nil {
		log.Printf("Failed to update status: %v", err)
	}
}

func registerCommands(s *discordgo.Session) error {
	deleteExistingCommands(s)
	log.Println("Registering new commands...")

	for _, cmd := range getCommandDefinitions() {
		if err := createCommand(s, cmd); err != nil {
			log.Printf("❌ Failed to register command '%s': %v", cmd.Name, err)
		} else {
			log.Printf("✅ Registered command: %s", cmd.Name)
		}
	}

	log.Println("Command registration completed")
	return nil
}

func deleteExistingCommands(s *discordgo.Session) {
	log.Println("Removing existing commands...")

	// グローバルコマンドを削除
	if globalCommands, err := s.ApplicationCommands(s.State.User.ID, ""); err == nil {
		for _, cmd := range globalCommands {
			if err := s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID); err != nil {
				log.Printf("Failed to delete global command %s: %v", cmd.Name, err)
			} else {
				log.Printf("Deleted existing global command: %s", cmd.Name)
			}
		}
	} else {
		log.Printf("Failed to fetch existing global commands: %v", err)
	}

	// ギルド専用コマンドを削除
	if guildID != "" {
		if guildCommands, err := s.ApplicationCommands(s.State.User.ID, guildID); err == nil {
			for _, cmd := range guildCommands {
				if err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID); err != nil {
					log.Printf("Failed to delete guild command %s: %v", cmd.Name, err)
				} else {
					log.Printf("Deleted existing guild command: %s", cmd.Name)
				}
			}
		} else {
			log.Printf("Failed to fetch existing guild commands: %v", err)
		}
	}
}

func createCommand(s *discordgo.Session, cmd *discordgo.ApplicationCommand) error {
	if guildID != "" {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		return err
	}
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
	return err
}

func getCommandDefinitions() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "reserve",
			Description: "部室の予約を作成します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "date",
					Description:  "予約日（YYYY-MM-DD または YYYY/MM/DD、例: 2025-10-15 または 2025/10/15）",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "start_time",
					Description:  "開始時間（HH:MM形式、例: 14:00）",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "end_time",
					Description:  "終了時間（HH:MM形式、例: 15:00）※省略時は開始時刻+1時間",
					Required:     false,
					Autocomplete: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "comment",
					Description: "コメント（任意）",
					Required:    false,
				},
			},
		},
		{
			Name:        "cancel",
			Description: "予約を取り消します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "reservation_id",
					Description:  "予約ID",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "comment",
					Description: "コメント（任意）",
					Required:    false,
				},
			},
		},
		{
			Name:        "complete",
			Description: "予約を完了にします",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "reservation_id",
					Description:  "予約ID",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "comment",
					Description: "コメント（任意）",
					Required:    false,
				},
			},
		},
		{
			Name:        "edit",
			Description: "予約を編集します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "reservation_id",
					Description:  "予約ID",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "date",
					Description:  "新しい予約日（YYYY-MM-DD または YYYY/MM/DD）※変更しない場合は省略",
					Required:     false,
					Autocomplete: true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "start_time",
					Description:  "新しい開始時間（HH:MM形式）※変更しない場合は省略",
					Required:     false,
					Autocomplete: true,
				},
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "end_time",
					Description:  "新しい終了時間（HH:MM形式）※変更しない場合は省略",
					Required:     false,
					Autocomplete: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "comment",
					Description: "新しいコメント（※変更しない場合は省略）",
					Required:    false,
				},
			},
		},
		{
			Name:        "list",
			Description: "すべての予約を表示します（自分だけに表示されます）",
		},
		{
			Name:        "my-reservations",
			Description: "自分の予約を表示します（自分だけに表示されます）",
		},
		{
			Name:        "help",
			Description: "ヘルプメッセージを表示します（自分だけに表示されます）",
		},
		{
			Name:        "feedback",
			Description: "システムへのご意見・ご要望を匿名で送信します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "フィードバック内容",
					Required:    true,
				},
			},
		},
	}
}

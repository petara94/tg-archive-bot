package main

import (
	"fmt"
	sys_log "log"

	"tg-archive-bot/internal/log"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

const TOKEN = "7473509536:AAG_30QvTClBFudeODn4JLiVsZYGatGCoTM"

func main() {
	_, err := log.RootLogger("development", "debug")
	if err != nil {
		sys_log.Fatal(err)
	}

	bot, err := telego.NewBot(TOKEN, telego.WithDefaultDebugLogger())
	if err != nil {
		zap.L().Fatal("bot connect error", zap.Error(err))
	}

	botUser, err := bot.GetMe()
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Print Bot information
	zap.L().Info("bot connected", zap.Any("user", botUser))

	updates, _ := bot.UpdatesViaLongPolling(nil)

	// Stop reviving updates from update channel
	defer bot.StopLongPolling()

	// Loop through all updates when they came
	for update := range updates {
		zap.L().Info("update", zap.Any("update", update))
	}
}

package main

import (
	"context"
	"fmt"
	"log/slog"

	"tg-archive-bot/internal/log"
	"tg-archive-bot/internal/services/tglistener"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

const TOKEN = "7473509536:AAG_30QvTClBFudeODn4JLiVsZYGatGCoTM"

func main() {
	_, err := log.RootLogger("development", "debug")
	if err != nil {
		slog.Error("create zap logger", err)
		return
	}

	bot, err := telego.NewBot(TOKEN)
	if err != nil {
		zap.L().Fatal("bot connect error", zap.Error(err))
	}

	botUser, err := bot.GetMe()
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Print Bot information
	zap.L().Info("bot connected", zap.Any("user", botUser))

	listener := tglistener.NewListener(bot, tglistener.Config{})

	listener.Run(context.Background())
}

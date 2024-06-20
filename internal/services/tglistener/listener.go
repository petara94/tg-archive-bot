package tglistener

import (
	"context"

	"github.com/mymmrac/telego"
	"go.uber.org/zap"
)

type Listener struct {
	bot    *telego.Bot
	logger *zap.Logger
	cfg    Config
}

func NewListener(bot *telego.Bot, cfg Config) *Listener {
	return &Listener{
		bot:    bot,
		cfg:    cfg,
		logger: zap.L().With(zap.String("pkg", "listener")),
	}
}

func (l Listener) Run(ctx context.Context) {
	updates, _ := l.bot.UpdatesViaLongPolling(nil)
	defer l.bot.StopLongPolling()

	for {
		select {
		case <-ctx.Done():
			l.logger.Info("shutting down telegram listener")
			return
		case update, ok := <-updates:
			if !ok {
				l.logger.Info("telegram listener shutdown")
				return
			}

			switch {
			case update.Message != nil &&
				update.Message.Chat.Type == "group" &&
				update.Message.Chat.ID == l.cfg.GroupId:
				l.logger.Debug("received new message",
					zap.String("group_title", update.Message.Chat.Title),
					zap.String("message_sender", update.Message.From.Username),
					zap.String("message_text", update.Message.Text),
				)

			}

		}
	}
}

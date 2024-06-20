package messenger

import (
	"context"
	"sync"

	"tg-archive-bot/internal/services/messenger/dto"

	"github.com/mymmrac/telego"
)

type FellowChatContract interface {
	SendMessage(ctx context.Context, message dto.Message) error
	GetAllChats(ctx context.Context) ([]dto.Group, error)
	CreateGroup(ctx context.Context, telegramId int64, title string, members ...string) (dto.Group, error)
}

type MessageProcessor struct {
	messages     chan *telego.Message
	closeChannel chan bool
	workersCount int

	groupIdCache map[int64]int64
	sender       FellowChatContract
}

func (p *MessageProcessor) Run(ctx context.Context) {
	wg := &sync.WaitGroup{}
	wg.Add(p.workersCount)
	for i := 0; i < p.workersCount; i++ {
		go func() {
			defer wg.Done()

		}()
	}

	wg.Wait()
}

func (p *MessageProcessor) Send(message *telego.Message) {
	select {
	case _, _ = <-p.closeChannel:
	case p.messages <- message:
	}
}

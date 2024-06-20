package fellowchat

import (
	"context"

	"tg-archive-bot/internal/services/messenger/dto"
)

type GroupRepositoryContract interface {
	GetAllGroups(ctx context.Context) ([]dto.Group, error)
	InsertGroup(ctx context.Context, telegramId int64, fellowChatId int64) (dto.Group, error)
}

type ChatServiceContract interface {
	CreateGroup(ctx context.Context, title string, members ...string) (int64, error)
	SendMessage(ctx context.Context, message dto.Message) error
}

type ChatServiceController struct {
	groupRepository GroupRepositoryContract
	chatService     ChatServiceContract
}

func (c *ChatServiceController) SendMessage(ctx context.Context, message dto.Message) error {
	return c.chatService.SendMessage(ctx, message)
}

func (c *ChatServiceController) GetAllChats(ctx context.Context) ([]dto.Group, error) {
	return c.groupRepository.GetAllGroups(ctx)
}

func (c *ChatServiceController) CreateGroup(ctx context.Context, telegramId int64, title string, members ...string) (dto.Group, error) {
	id, err := c.chatService.CreateGroup(ctx, title, members...)
	if err != nil {
		return dto.Group{}, err
	}

	return c.groupRepository.InsertGroup(ctx, id, telegramId)
}

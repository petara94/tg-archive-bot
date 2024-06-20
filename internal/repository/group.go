package repository

import (
	"context"
	"errors"
	"fmt"

	"tg-archive-bot/internal/services/messenger/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepository struct {
	db *pgxpool.Pool
}

func (g *GroupRepository) GetAllGroups(ctx context.Context) ([]dto.Group, error) {
	const q = "SELECT telegram_id, fellow_chat_id FROM groups"

	rows, err := g.db.Query(ctx, q)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	defer rows.Close()
	groups := make([]dto.Group, 0)
	for rows.Next() {
		var group dto.Group
		err = rows.Scan(&group.TelegramID, &group.FellowChatID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
	}

	return groups, nil
}

func (g *GroupRepository) InsertGroup(ctx context.Context, telegramId int64, fellowChatId int64) (dto.Group, error) {
	const q = "INSERT INTO groups (telegram_id, fellow_chat_id) VALUES ($1, $2)"

	_, err := g.db.Exec(ctx, q, telegramId, fellowChatId)
	if err != nil {
		return dto.Group{}, err
	}

	return dto.Group{
		TelegramID:   telegramId,
		FellowChatID: fellowChatId,
	}, nil
}

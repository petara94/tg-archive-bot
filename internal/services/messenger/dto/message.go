package dto

import "time"

type Message struct {
	Sender    string    `json:"sender"`
	Text      string    `json:"text"`
	ChatId    uint64    `json:"chat_id"`
	CreatedAt time.Time `json:"created_at"`
}

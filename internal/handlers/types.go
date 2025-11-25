package handlers

import "github.com/hekigan/couples/internal/models"

// RoomWithUsername is a room enriched with the other player's username
type RoomWithUsername struct {
	*models.Room
	OtherPlayerUsername string
	IsOwner             bool
}

// AnswerWithDetails contains an answer with its question and user info
type AnswerWithDetails struct {
	Answer     *models.Answer
	Question   *models.Question
	Username   string
	ActionType string
}

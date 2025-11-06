package models

import "errors"

// Common errors used across the application
var (
	// User errors
	ErrInvalidUserName = errors.New("invalid user name")
	ErrEmailRequired   = errors.New("email is required for non-anonymous users")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidUserID   = errors.New("invalid user ID")

	// Room errors
	ErrRoomFull       = errors.New("room is full")
	ErrRoomNotFound   = errors.New("room not found")
	ErrInvalidRoomID  = errors.New("invalid room ID")
	ErrRoomNotStarted = errors.New("room game has not started")
	ErrRoomAlreadyStarted = errors.New("room game has already started")

	// Game errors
	ErrGameNotStarted   = errors.New("game has not started")
	ErrGameAlreadyEnded = errors.New("game has already ended")
	ErrNotYourTurn      = errors.New("it is not your turn")
	ErrNoQuestionsAvailable = errors.New("no questions available")

	// Authorization errors
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrNotRoomOwner    = errors.New("only room owner can perform this action")
	ErrNotRoomPlayer   = errors.New("user is not a player in this room")

	// Friend errors
	ErrFriendshipExists = errors.New("friendship already exists")
	ErrFriendshipNotFound = errors.New("friendship not found")
	ErrCannotBeFriendWithSelf = errors.New("cannot send friend request to yourself")
	ErrAlreadyFriends = errors.New("already friends with this user")
	ErrPendingInvitation = errors.New("friend invitation already pending")
)


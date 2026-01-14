package main

import (
	"github.com/hekigan/couples/internal/handlers"
	"github.com/hekigan/couples/internal/middleware"
	"github.com/labstack/echo/v4"
)

// registerGameRoutes registers all game-related routes (rooms, play, join requests)
func registerGameRoutes(e *echo.Echo, h *handlers.Handler) {
	// Game routes (auth + rate limiting + CSRF)
	game := e.Group("/game")
	game.Use(middleware.EchoRequireAuth())
	game.Use(middleware.EchoRateLimit())
	game.Use(middleware.EchoCSRF())
	game.GET("/rooms", h.ListRoomsHandler)
	game.GET("/create-room", h.CreateRoomHandler)
	game.POST("/create-room", h.CreateRoomHandler)
	game.GET("/join-room", h.JoinRoomHandler)
	game.POST("/join-room", h.JoinRoomHandler)
	game.GET("/room/:id", h.RoomHandler)
	game.POST("/room/:id/delete", h.DeleteRoomHandler)
	game.DELETE("/room/:id/delete", h.DeleteRoomHandler)
	game.GET("/room/:id/join-requests", h.ListJoinRequestsHandler)
	game.GET("/play/:id", h.PlayHandler)
	game.GET("/finished/:id", h.GameFinishedHandler)
	game.GET("/partials/empty-rooms-state", h.EmptyRoomsStateHandler)
}

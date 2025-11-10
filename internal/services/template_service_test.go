package services

import (
	"strings"
	"testing"
)

// TestTemplateService_RenderFragment tests HTML fragment rendering
func TestTemplateService_RenderFragment(t *testing.T) {
	// This is a unit test that doesn't require database
	// Create template service
	templateService, err := NewTemplateService("../../templates")
	if err != nil {
		t.Fatalf("Failed to create template service: %v", err)
	}

	t.Run("RenderJoinRequest", func(t *testing.T) {
		data := JoinRequestData{
			ID:        "test-request-id",
			RoomID:    "test-room-id",
			Username:  "TestUser",
			CreatedAt: "3:04 PM",
		}

		html, err := templateService.RenderFragment("join_request.html", data)
		AssertNoError(t, err, "RenderFragment should not error")
		AssertTrue(t, len(html) > 0, "HTML should not be empty")
		AssertTrue(t, strings.Contains(html, "TestUser"), "HTML should contain username")
		AssertTrue(t, strings.Contains(html, "test-request-id"), "HTML should contain request ID")
		AssertTrue(t, strings.Contains(html, "3:04 PM"), "HTML should contain timestamp")
		AssertTrue(t, strings.Contains(html, "join-request-item"), "HTML should contain join-request-item class")
	})

	t.Run("RenderPlayerJoined", func(t *testing.T) {
		data := PlayerJoinedData{
			Username: "GuestPlayer",
			UserID:   "test-user-id",
		}

		html, err := templateService.RenderFragment("player_joined.html", data)
		AssertNoError(t, err, "RenderFragment should not error")
		AssertTrue(t, len(html) > 0, "HTML should not be empty")
		AssertTrue(t, strings.Contains(html, "GuestPlayer"), "HTML should contain username")
		AssertTrue(t, strings.Contains(html, "guest-info"), "HTML should contain guest-info id")
	})

	t.Run("RenderRequestAccepted", func(t *testing.T) {
		data := RequestAcceptedData{
			GuestUsername: "AcceptedGuest",
		}

		html, err := templateService.RenderFragment("request_accepted.html", data)
		AssertNoError(t, err, "RenderFragment should not error")
		AssertTrue(t, len(html) > 0, "HTML should not be empty")
		AssertTrue(t, strings.Contains(html, "AcceptedGuest"), "HTML should contain guest username")
		AssertTrue(t, strings.Contains(html, "guest-info"), "HTML should contain guest-info id")
	})

	t.Run("RenderQuestionDrawn", func(t *testing.T) {
		data := QuestionDrawnData{
			RoomID:                "test-room-id",
			QuestionNumber:        5,
			MaxQuestions:          20,
			Category:              "couples",
			CategoryLabel:         "Couples",
			QuestionText:          "What is your favorite memory together?",
			IsMyTurn:              true,
			CurrentPlayerUsername: "Player1",
		}

		html, err := templateService.RenderFragment("question_drawn.html", data)
		AssertNoError(t, err, "RenderFragment should not error")
		AssertTrue(t, len(html) > 0, "HTML should not be empty")
		AssertTrue(t, strings.Contains(html, "Question 5 of 20"), "HTML should contain question number")
		AssertTrue(t, strings.Contains(html, "Couples"), "HTML should contain category label")
		AssertTrue(t, strings.Contains(html, "What is your favorite memory together?"), "HTML should contain question text")
		AssertTrue(t, strings.Contains(html, "current-question"), "HTML should contain current-question id")
	})

	t.Run("RenderAnswerSubmitted", func(t *testing.T) {
		data := AnswerSubmittedData{
			RoomID:                "test-room-id",
			Username:              "Player2",
			AnswerText:            "Our first trip to Paris",
			ActionType:            "answered",
			IsMyTurn:              false,
			CurrentPlayerUsername: "Player1",
		}

		html, err := templateService.RenderFragment("answer_submitted.html", data)
		AssertNoError(t, err, "RenderFragment should not error")
		AssertTrue(t, len(html) > 0, "HTML should not be empty")
		AssertTrue(t, strings.Contains(html, "Player2"), "HTML should contain username")
		AssertTrue(t, strings.Contains(html, "Our first trip to Paris"), "HTML should contain answer text")
	})

	t.Run("RenderGameStarted", func(t *testing.T) {
		data := GameStartedData{
			RoomID: "test-room-id",
		}

		html, err := templateService.RenderFragment("game_started.html", data)
		AssertNoError(t, err, "RenderFragment should not error")
		AssertTrue(t, len(html) > 0, "HTML should not be empty")
		AssertTrue(t, strings.Contains(html, "test-room-id"), "HTML should contain room ID")
	})

	t.Run("RenderNotification", func(t *testing.T) {
		data := NotificationData{
			ID:      "test-notification-id",
			Type:    "info",
			Title:   "Test Notification",
			Message: "This is a test message",
			Link:    "/test-link",
		}

		html, err := templateService.RenderFragment("notification_item.html", data)
		AssertNoError(t, err, "RenderFragment should not error")
		AssertTrue(t, len(html) > 0, "HTML should not be empty")
		AssertTrue(t, strings.Contains(html, "Test Notification"), "HTML should contain title")
		AssertTrue(t, strings.Contains(html, "This is a test message"), "HTML should contain message")
	})

	t.Run("InvalidTemplate", func(t *testing.T) {
		_, err := templateService.RenderFragment("nonexistent.html", map[string]string{})
		AssertError(t, err, "RenderFragment should error for nonexistent template")
	})
}

// TestTemplateService_AllTemplatesLoad tests that all templates can be loaded
func TestTemplateService_AllTemplatesLoad(t *testing.T) {
	_, err := NewTemplateService("../../templates")
	AssertNoError(t, err, "All templates should load without error")
}

// TestTemplateService_ConcurrentRendering tests thread safety
func TestTemplateService_ConcurrentRendering(t *testing.T) {
	templateService, err := NewTemplateService("../../templates")
	AssertNoError(t, err, "Template service should be created")

	// Test concurrent rendering to ensure RWMutex works
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			data := JoinRequestData{
				ID:        "concurrent-test",
				RoomID:    "room-test",
				Username:  "User",
				CreatedAt: "3:04 PM",
			}
			_, err := templateService.RenderFragment("join_request.html", data)
			if err != nil {
				t.Errorf("Concurrent render failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

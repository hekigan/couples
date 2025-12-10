package rendering

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hekigan/couples/internal/viewmodels"
	gameFragments "github.com/hekigan/couples/internal/views/fragments/game"
)

// TemplService provides a simple adapter for rendering templ components
// in the service layer (specifically for GameService SSE broadcasting).
// This package breaks the import cycle by sitting between services and views.
type TemplService struct{}

// NewTemplService creates a new templ render service
func NewTemplService() *TemplService {
	return &TemplService{}
}

// RenderFragment renders a templ fragment and returns HTML string.
// This is used by GameService for SSE HTML fragment broadcasting.
func (t *TemplService) RenderFragment(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	ctx := context.Background()

	switch name {
	case "game_started.html":
		d, ok := data.(viewmodels.GameStartedData)
		if !ok {
			return "", fmt.Errorf("invalid data type for game_started: expected GameStartedData")
		}
		component := gameFragments.GameStarted(&d)
		err := component.Render(ctx, &buf)
		return buf.String(), err

	case "question_drawn.html":
		d, ok := data.(viewmodels.QuestionDrawnData)
		if !ok {
			return "", fmt.Errorf("invalid data type for question_drawn: expected QuestionDrawnData")
		}
		component := gameFragments.QuestionDrawn(&d)
		err := component.Render(ctx, &buf)
		return buf.String(), err

	default:
		return "", fmt.Errorf("unknown template: %s", name)
	}
}

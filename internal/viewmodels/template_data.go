package viewmodels

// TemplateData represents common data passed to page templates
type TemplateData struct {
	Title             string
	User              interface{} // Can be *SessionUser or *models.User
	Error             string
	Success           string
	Data              interface{}
	OwnerUsername     string
	GuestUsername     string
	IsOwner           bool
	IsAdmin           bool   // Whether current user is admin
	JoinRequestsCount int    // Number of pending join requests (for badge)
	Env               string // Environment (development/production) for conditional JS loading
	CSRFToken         string // CSRF token for forms and HTMX requests
	CurrentStep       int    // Current room step (1=invite, 2=categories, 3=start) - derived from room state
	// Pre-rendered fragment HTML for SSR (to avoid hx-trigger="load")
	CategoriesGridHTML string // Categories grid fragment (rendered server-side)
	FriendsListHTML    string // Friends list fragment (rendered server-side, owner only)
	ActionButtonHTML   string // Start/ready button fragment (rendered server-side)
	JoinRequestsHTML   string // Join requests fragment (rendered server-side, owner only)
}

// GameStartedData represents data for game_started SSE fragment
type GameStartedData struct {
	RoomID string
}

// QuestionDrawnData represents data for question_drawn SSE fragment
type QuestionDrawnData struct {
	RoomID                string
	QuestionNumber        int
	MaxQuestions          int
	Category              string
	CategoryLabel         string
	QuestionText          string
	IsMyTurn              bool
	CurrentPlayerUsername string
}

// AnswerSubmittedData represents data for answer_submitted SSE fragment
type AnswerSubmittedData struct {
	RoomID                string
	Username              string
	AnswerText            string
	ActionType            string // "answered" or "passed"
	IsMyTurn              bool   // Is it now my turn to draw next question?
	CurrentPlayerUsername string
}

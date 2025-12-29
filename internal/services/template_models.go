package services

// This file contains all data structures used for rendering HTML templates
// These structures are used by TemplateService to render SSE HTML fragments

// ============================================================================
// Game/Room Template Data Structures
// ============================================================================

import (
	"github.com/hekigan/couples/internal/models"
)

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

// GameFinishedData represents data for the game finished page
type GameFinishedData struct {
	Room           *models.Room
	Answers        []AnswerWithDetails
	TotalQuestions int
	SkippedCount   int
	AnsweredCount  int
}

// PlayPageData represents data for the game play page
type PlayPageData struct {
	Room                 *models.Room
	CurrentUserID        string
	IsMyTurn             bool
	OtherPlayerName      string
	QuestionText         string
	QuestionID           string
	HasAnswer            bool
	AnswerText           string
	ActionType           string
	AnsweredByPlayerName string
}

// JoinRequestData represents data for join request partial
type JoinRequestData struct {
	ID        string
	RoomID    string
	UserID    string
	Username  string
	Message   string
	Status    string // "pending", "accepted", "rejected"
	CreatedAt string
}

// PlayerJoinedData represents data for player joined partial
type PlayerJoinedData struct {
	Username string
	UserID   string
}

// RequestAcceptedData represents data for request accepted partial
type RequestAcceptedData struct {
	GuestUsername string
}

// QuestionDrawnData represents data for question drawn partial
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

// AnswerSubmittedData represents data for answer submitted partial
type AnswerSubmittedData struct {
	RoomID                string
	Username              string
	AnswerText            string
	ActionType            string // "answered" or "passed"
	IsMyTurn              bool   // Is it now my turn to draw next question?
	CurrentPlayerUsername string
}

// GameStartedData represents data for game started partial
type GameStartedData struct {
	RoomID string
}

// NotificationData represents data for notification partial
type NotificationData struct {
	ID      string
	Type    string
	Title   string
	Message string
	Link    string
}

// FriendsListData represents data for friends list partial
type FriendsListData struct {
	Friends []FriendInfo
	RoomID  string
}

// FriendInfo represents a single friend for invitation
type FriendInfo struct {
	ID       string
	Username string
}

// FriendInvitedData represents data for friend invited partial
type FriendInvitedData struct {
	Friend FriendInfo
	RoomID string
}

// CategoriesGridData represents data for categories grid partial
type CategoriesGridData struct {
	Categories []CategoryInfo
	RoomID     string
	GuestReady bool
	IsOwner    bool
}

// CategoryInfo represents a single category with selection state
type CategoryInfo struct {
	ID            string
	Key           string
	Label         string
	IsSelected    bool
	QuestionCount int
}

// GuestReadyButtonData represents data for guest ready button partial
type GuestReadyButtonData struct {
	RoomID     string
	GuestReady bool
}

// StartGameButtonData represents data for start game button partial
type StartGameButtonData struct {
	RoomID        string
	GuestReady    bool
	GuestUsername string
}

// AcceptedRequestData represents data for accepted request SSE updates
type AcceptedRequestData struct {
	RoomID string
}

// RejectedRequestData represents data for rejected request SSE updates
type RejectedRequestData struct {
	RoomID string
}

// TurnIndicatorData represents data for turn indicator partial
type TurnIndicatorData struct {
	IsMyTurn        bool
	OtherPlayerName string
}

// QuestionCardData represents data for question card partial
type QuestionCardData struct {
	QuestionText string
}

// AnswerFormData represents data for answer form partial
type AnswerFormData struct {
	RoomID     string
	QuestionID string
}

// WaitingUIData represents data for waiting UI partial
type WaitingUIData struct {
	OtherPlayerName string
}

// AnswerReviewData represents data for answer review partial
type AnswerReviewData struct {
	RoomID               string
	AnswerText           string
	ActionType           string
	ShowNextButton       bool
	AnsweredByPlayerName string
	OtherPlayerName      string
}

// GameContentData represents data for main game content area
type GameContentData struct {
	RoomID string
}

// BadgeUpdateData represents data for badge update partial
type BadgeUpdateData struct {
	Count int
}

// RoomStatusBadgeData represents data for room status badge partial
type RoomStatusBadgeData struct {
	Status string // "waiting", "ready", "playing"
}

// ProgressCounterData represents data for progress counter partial
type ProgressCounterData struct {
	CurrentQuestion int
	MaxQuestions    int
}

// ============================================================================
// Admin Template Data Structures
// ============================================================================

// AdminUserInfo represents a user in the admin list
type AdminUserInfo struct {
	ID        string
	Username  string
	Email     string
	UserType  string // "registered", "guest", "admin"
	IsAdmin   bool
	CreatedAt string
}

// UsersListData represents data for admin users list partial
type UsersListData struct {
	Users         []AdminUserInfo
	TotalCount    int
	Page          int
	CurrentUserID string
	// Pagination fields
	CurrentPage     int    // Current page number (same as Page, kept for consistency)
	TotalPages      int    // Total number of pages
	ItemsPerPage    int    // Number of items per page
	BaseURL         string // API URL for fetching data
	PageURL         string // Page URL for browser history
	Target          string // HTMX target selector
	IncludeSelector string // Selector for additional params
	ExtraParams     string // Additional query parameters
	ItemName        string // Name of items for display
}

// AdminQuestionInfo represents a question in the admin list
type AdminQuestionInfo struct {
	ID               string
	Text             string
	CategoryLabel    string // Combined icon + label
	LanguageCode     string
	TranslationCount int // Number of translations (0-3) for this question
}

// AdminCategoryOption represents a category option for dropdowns
type AdminCategoryOption struct {
	ID            string
	Label         string
	Selected      bool
	QuestionCount int // Number of questions in this category
}

// QuestionsListData represents data for admin questions list partial
type QuestionsListData struct {
	Questions                []AdminQuestionInfo
	Categories               []AdminCategoryOption
	SelectedCategoryID       string
	TotalCount               int // Total number of questions (for pagination)
	CurrentPage              int // Current page number
	TotalPages               int // Total number of pages
	ItemsPerPage             int // Number of items per page
	MissingTranslationsCount int // Total number of incomplete translations
	// Pagination template fields
	BaseURL         string // API URL for fetching data (e.g., "/admin/api/questions/list")
	PageURL         string // Page URL for browser history (e.g., "/admin/questions")
	Target          string // HTMX target selector (e.g., "#questions-list")
	IncludeSelector string // Selector for additional params (e.g., "[name='category_id']")
	ExtraParams     string // Additional query parameters (e.g., "&category_id=xyz")
	ItemName        string // Name of items for display (e.g., "questions")
}

// QuestionEditFormData represents data for question edit form partial
type QuestionEditFormData struct {
	QuestionID     string
	BaseQuestionID string
	QuestionText   string // English question text
	TranslationFR  string // French translation
	TranslationJA  string // Japanese translation
	Categories     []AdminCategoryOption
	LangEN         bool
	LangFR         bool
	LangJA         bool
	SelectedLang   string // Currently selected language code
	Page           int    // Current page number for pagination
}

// AdminCategoryInfo represents a category in the admin list
type AdminCategoryInfo struct {
	ID            string
	Label         string
	Key           string
	QuestionCount int
}

// CategoriesListData represents data for admin categories list partial
type CategoriesListData struct {
	Categories []AdminCategoryInfo
	// Pagination fields
	TotalCount      int    // Total number of categories
	CurrentPage     int    // Current page number
	TotalPages      int    // Total number of pages
	ItemsPerPage    int    // Number of items per page
	BaseURL         string // API URL for fetching data
	PageURL         string // Page URL for browser history
	Target          string // HTMX target selector
	IncludeSelector string // Selector for additional params
	ExtraParams     string // Additional query parameters
	ItemName        string // Name of items for display
}

// CategoryFormData represents data for category create/edit form (shared template)
type CategoryFormData struct {
	ID    string // Empty for create mode, populated for edit mode
	Key   string
	Label string
}

// CategoryEditFormData is an alias for backwards compatibility
type CategoryEditFormData = CategoryFormData

// UserFormData represents data for user create/edit form (shared template)
type UserFormData struct {
	ID          string // Empty for create mode, populated for edit mode
	Username    string
	Email       string
	IsAdmin     bool
	IsAnonymous bool
}

// QuestionFormData represents data for question create/edit form (shared template)
type QuestionFormData struct {
	QuestionID     string                 // Empty for create mode, populated for edit mode
	BaseQuestionID string                 // Base question ID (for translations)
	Categories     []AdminCategoryOption  // Available categories
	QuestionText   string                 // English question text
	TranslationFR  string                 // French translation
	TranslationJA  string                 // Japanese translation
	SelectedLang   string                 // Currently selected language (en, fr, ja)
	LangEN         bool                   // True if English is selected
	LangFR         bool                   // True if French is selected
	LangJA         bool                   // True if Japanese is selected
}

// RoomDetailsData represents data for room details view (read-only)
type RoomDetailsData struct {
	ID            string
	ShortID       string
	Name          string
	Status        string
	Language      string
	MaxQuestions  int
	CreatedAt     string
	OwnerUsername string
	OwnerEmail    string
	GuestUsername string
	GuestEmail    string
	CategoryCount int
}

// AdminRoomInfo represents a room in the admin list
type AdminRoomInfo struct {
	ID        string
	ShortID   string // First 8 chars of ID
	Owner     string
	Guest     string
	Status    string
	CreatedAt string
}

// RoomsListData represents data for admin rooms list partial
type RoomsListData struct {
	Rooms []AdminRoomInfo
	// Pagination fields
	TotalCount      int    // Total number of rooms
	CurrentPage     int    // Current page number
	TotalPages      int    // Total number of pages
	ItemsPerPage    int    // Number of items per page
	BaseURL         string // API URL for fetching data
	PageURL         string // Page URL for browser history
	Target          string // HTMX target selector
	IncludeSelector string // Selector for additional params
	ExtraParams     string // Additional query parameters
	ItemName        string // Name of items for display
}

// DashboardStatsData represents data for dashboard stats partial
type DashboardStatsData struct {
	TotalUsers      int
	AnonymousUsers  int
	RegisteredUsers int
	AdminUsers      int
	TotalRooms      int
	ActiveRooms     int
	CompletedRooms  int
	TotalQuestions  int
	TotalCategories int
}

// ============================================================================
// Pagination Interface Implementation
// ============================================================================

// GetTotalCount returns total count for UsersListData
func (d *UsersListData) GetTotalCount() int { return d.TotalCount }

// GetCurrentPage returns current page for UsersListData
func (d *UsersListData) GetCurrentPage() int { return d.CurrentPage }

// GetTotalPages returns total pages for UsersListData
func (d *UsersListData) GetTotalPages() int { return d.TotalPages }

// GetItemsPerPage returns items per page for UsersListData
func (d *UsersListData) GetItemsPerPage() int { return d.ItemsPerPage }

// GetBaseURL returns base URL for UsersListData
func (d *UsersListData) GetBaseURL() string { return d.BaseURL }

// GetPageURL returns page URL for UsersListData
func (d *UsersListData) GetPageURL() string { return d.PageURL }

// GetTarget returns target selector for UsersListData
func (d *UsersListData) GetTarget() string { return d.Target }

// GetIncludeSelector returns include selector for UsersListData
func (d *UsersListData) GetIncludeSelector() string { return d.IncludeSelector }

// GetExtraParams returns extra params for UsersListData
func (d *UsersListData) GetExtraParams() string { return d.ExtraParams }

// GetItemName returns item name for UsersListData
func (d *UsersListData) GetItemName() string { return d.ItemName }

// GetTotalCount returns total count for QuestionsListData
func (d *QuestionsListData) GetTotalCount() int { return d.TotalCount }

// GetCurrentPage returns current page for QuestionsListData
func (d *QuestionsListData) GetCurrentPage() int { return d.CurrentPage }

// GetTotalPages returns total pages for QuestionsListData
func (d *QuestionsListData) GetTotalPages() int { return d.TotalPages }

// GetItemsPerPage returns items per page for QuestionsListData
func (d *QuestionsListData) GetItemsPerPage() int { return d.ItemsPerPage }

// GetBaseURL returns base URL for QuestionsListData
func (d *QuestionsListData) GetBaseURL() string { return d.BaseURL }

// GetPageURL returns page URL for QuestionsListData
func (d *QuestionsListData) GetPageURL() string { return d.PageURL }

// GetTarget returns target selector for QuestionsListData
func (d *QuestionsListData) GetTarget() string { return d.Target }

// GetIncludeSelector returns include selector for QuestionsListData
func (d *QuestionsListData) GetIncludeSelector() string { return d.IncludeSelector }

// GetExtraParams returns extra params for QuestionsListData
func (d *QuestionsListData) GetExtraParams() string { return d.ExtraParams }

// GetItemName returns item name for QuestionsListData
func (d *QuestionsListData) GetItemName() string { return d.ItemName }

// GetTotalCount returns total count for CategoriesListData
func (d *CategoriesListData) GetTotalCount() int { return d.TotalCount }

// GetCurrentPage returns current page for CategoriesListData
func (d *CategoriesListData) GetCurrentPage() int { return d.CurrentPage }

// GetTotalPages returns total pages for CategoriesListData
func (d *CategoriesListData) GetTotalPages() int { return d.TotalPages }

// GetItemsPerPage returns items per page for CategoriesListData
func (d *CategoriesListData) GetItemsPerPage() int { return d.ItemsPerPage }

// GetBaseURL returns base URL for CategoriesListData
func (d *CategoriesListData) GetBaseURL() string { return d.BaseURL }

// GetPageURL returns page URL for CategoriesListData
func (d *CategoriesListData) GetPageURL() string { return d.PageURL }

// GetTarget returns target selector for CategoriesListData
func (d *CategoriesListData) GetTarget() string { return d.Target }

// GetIncludeSelector returns include selector for CategoriesListData
func (d *CategoriesListData) GetIncludeSelector() string { return d.IncludeSelector }

// GetExtraParams returns extra params for CategoriesListData
func (d *CategoriesListData) GetExtraParams() string { return d.ExtraParams }

// GetItemName returns item name for CategoriesListData
func (d *CategoriesListData) GetItemName() string { return d.ItemName }

// GetTotalCount returns total count for RoomsListData
func (d *RoomsListData) GetTotalCount() int { return d.TotalCount }

// GetCurrentPage returns current page for RoomsListData
func (d *RoomsListData) GetCurrentPage() int { return d.CurrentPage }

// GetTotalPages returns total pages for RoomsListData
func (d *RoomsListData) GetTotalPages() int { return d.TotalPages }

// GetItemsPerPage returns items per page for RoomsListData
func (d *RoomsListData) GetItemsPerPage() int { return d.ItemsPerPage }

// GetBaseURL returns base URL for RoomsListData
func (d *RoomsListData) GetBaseURL() string { return d.BaseURL }

// GetPageURL returns page URL for RoomsListData
func (d *RoomsListData) GetPageURL() string { return d.PageURL }

// GetTarget returns target selector for RoomsListData
func (d *RoomsListData) GetTarget() string { return d.Target }

// GetIncludeSelector returns include selector for RoomsListData
func (d *RoomsListData) GetIncludeSelector() string { return d.IncludeSelector }

// GetExtraParams returns extra params for RoomsListData
func (d *RoomsListData) GetExtraParams() string { return d.ExtraParams }

// GetItemName returns item name for RoomsListData
func (d *RoomsListData) GetItemName() string { return d.ItemName }

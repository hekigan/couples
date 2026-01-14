package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CustomValidator wraps the go-playground validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new custom validator instance
func NewValidator() *CustomValidator {
	v := validator.New()

	// Register custom UUID validation
	v.RegisterValidation("uuid", validateUUID)

	return &CustomValidator{validator: v}
}

// Validate implements echo.Validator interface
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

// validateUUID is a custom validation function for UUID strings
func validateUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

// Common validation structs

// UUIDParam represents a UUID parameter from URL
type UUIDParam struct {
	ID string `param:"id" validate:"required,uuid"`
}

// PaginationQuery represents pagination query parameters
type PaginationQuery struct {
	Page    int `query:"page" validate:"omitempty,min=1"`
	PerPage int `query:"per_page" validate:"omitempty,oneof=25 50 100"`
}

// RoomCreateRequest represents the request body for creating a room
type RoomCreateRequest struct {
	Name string `json:"name" form:"name" validate:"required,min=3,max=50"`
}

// CategoryUpdateRequest represents the request body for updating room categories
type CategoryUpdateRequest struct {
	CategoryIDs []string `json:"category_ids" form:"category_ids" validate:"required,dive,uuid"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=6"`
}

// SignupRequest represents the signup request
type SignupRequest struct {
	Username string `json:"username" form:"username" validate:"required,min=3,max=30"`
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=6"`
}

// SetupUsernameRequest represents the username setup request
type SetupUsernameRequest struct {
	Username string `json:"username" form:"username" validate:"required,min=3,max=30"`
}

// UpdateProfileRequest represents the profile update request
type UpdateProfileRequest struct {
	Username string `json:"username" form:"username" validate:"omitempty,min=3,max=30"`
	Email    string `json:"email" form:"email" validate:"omitempty,email"`
}

// AnswerQuestionRequest represents the answer submission request
type AnswerQuestionRequest struct {
	Answer string `json:"answer" form:"answer" validate:"required,min=1,max=500"`
}

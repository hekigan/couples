package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func TestExtractIDFromRoute(t *testing.T) {
	tests := []struct {
		name      string
		vars      map[string]string
		key       string
		expectErr bool
		expected  uuid.UUID
	}{
		{
			name: "valid UUID",
			vars: map[string]string{
				"id": "123e4567-e89b-12d3-a456-426614174000",
			},
			key:       "id",
			expectErr: false,
			expected:  uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		},
		{
			name: "invalid UUID format",
			vars: map[string]string{
				"id": "not-a-uuid",
			},
			key:       "id",
			expectErr: true,
		},
		{
			name:      "missing key",
			vars:      map[string]string{},
			key:       "id",
			expectErr: true,
		},
		{
			name: "different key",
			vars: map[string]string{
				"room_id": "123e4567-e89b-12d3-a456-426614174000",
			},
			key:       "room_id",
			expectErr: false,
			expected:  uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractIDFromRoute(tt.vars, tt.key)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ExtractIDFromRoute() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParsePaginationParams(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedPage   int
		expectedPer    int
		expectedOffset int
	}{
		{
			name:           "default values",
			queryParams:    map[string]string{},
			expectedPage:   1,
			expectedPer:    25,
			expectedOffset: 0,
		},
		{
			name: "custom page",
			queryParams: map[string]string{
				"page": "3",
			},
			expectedPage:   3,
			expectedPer:    25,
			expectedOffset: 50,
		},
		{
			name: "custom per_page - valid value 50",
			queryParams: map[string]string{
				"per_page": "50",
			},
			expectedPage:   1,
			expectedPer:    50,
			expectedOffset: 0,
		},
		{
			name: "custom per_page - valid value 100",
			queryParams: map[string]string{
				"per_page": "100",
			},
			expectedPage:   1,
			expectedPer:    100,
			expectedOffset: 0,
		},
		{
			name: "custom per_page - invalid value defaults to 25",
			queryParams: map[string]string{
				"per_page": "75",
			},
			expectedPage:   1,
			expectedPer:    25,
			expectedOffset: 0,
		},
		{
			name: "both page and per_page",
			queryParams: map[string]string{
				"page":     "2",
				"per_page": "50",
			},
			expectedPage:   2,
			expectedPer:    50,
			expectedOffset: 50,
		},
		{
			name: "page 3 with per_page 100",
			queryParams: map[string]string{
				"page":     "3",
				"per_page": "100",
			},
			expectedPage:   3,
			expectedPer:    100,
			expectedOffset: 200,
		},
		{
			name: "invalid page - defaults to 1",
			queryParams: map[string]string{
				"page": "invalid",
			},
			expectedPage:   1,
			expectedPer:    25,
			expectedOffset: 0,
		},
		{
			name: "negative page - defaults to 1",
			queryParams: map[string]string{
				"page": "-1",
			},
			expectedPage:   1,
			expectedPer:    25,
			expectedOffset: 0,
		},
		{
			name: "zero page - defaults to 1",
			queryParams: map[string]string{
				"page": "0",
			},
			expectedPage:   1,
			expectedPer:    25,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with query parameters
			queryString := ""
			for key, value := range tt.queryParams {
				if queryString != "" {
					queryString += "&"
				}
				queryString += key + "=" + url.QueryEscape(value)
			}

			req := httptest.NewRequest("GET", "/?"+queryString, nil)

			page, perPage, offset := ParsePaginationParams(req)

			if page != tt.expectedPage {
				t.Errorf("page = %d, want %d", page, tt.expectedPage)
			}

			if perPage != tt.expectedPer {
				t.Errorf("perPage = %d, want %d", perPage, tt.expectedPer)
			}

			if offset != tt.expectedOffset {
				t.Errorf("offset = %d, want %d", offset, tt.expectedOffset)
			}
		})
	}
}

func TestWriteJSONResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		data           interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "success with map",
			statusCode:     http.StatusOK,
			data:           map[string]string{"message": "success"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"success"}`,
		},
		{
			name:           "created status",
			statusCode:     http.StatusCreated,
			data:           map[string]int{"id": 123},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":123}`,
		},
		{
			name:           "error with custom status",
			statusCode:     http.StatusBadRequest,
			data:           map[string]string{"error": "bad request"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"bad request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			err := WriteJSONResponse(w, tt.statusCode, tt.data)
			if err != nil {
				t.Errorf("WriteJSONResponse() returned error: %v", err)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("status code = %d, want %d", w.Code, tt.expectedStatus)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %s, want application/json", contentType)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("body = %s, want %s", body, tt.expectedBody)
			}
		})
	}
}

func TestWriteJSONError(t *testing.T) {
	w := httptest.NewRecorder()

	err := WriteJSONError(w, http.StatusBadRequest, "validation failed")
	if err != nil {
		t.Errorf("WriteJSONError() returned error: %v", err)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusBadRequest)
	}

	expectedBody := `{"error":"validation failed"}`
	body := strings.TrimSpace(w.Body.String())
	if body != expectedBody {
		t.Errorf("body = %s, want %s", body, expectedBody)
	}
}

func TestWriteJSONSuccess(t *testing.T) {
	w := httptest.NewRecorder()

	err := WriteJSONSuccess(w, "operation completed")
	if err != nil {
		t.Errorf("WriteJSONSuccess() returned error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	expectedBody := `{"success":"operation completed"}`
	body := strings.TrimSpace(w.Body.String())
	if body != expectedBody {
		t.Errorf("body = %s, want %s", body, expectedBody)
	}
}

func TestExtractIDFromRouteVars(t *testing.T) {
	tests := []struct {
		name      string
		setupVars func(*http.Request) *http.Request
		key       string
		expectErr bool
		expected  uuid.UUID
	}{
		{
			name: "valid UUID",
			setupVars: func(r *http.Request) *http.Request {
				return mux.SetURLVars(r, map[string]string{
					"id": "123e4567-e89b-12d3-a456-426614174000",
				})
			},
			key:       "id",
			expectErr: false,
			expected:  uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		},
		{
			name: "invalid UUID format",
			setupVars: func(r *http.Request) *http.Request {
				return mux.SetURLVars(r, map[string]string{
					"id": "not-a-uuid",
				})
			},
			key:       "id",
			expectErr: true,
		},
		{
			name: "missing key",
			setupVars: func(r *http.Request) *http.Request {
				return mux.SetURLVars(r, map[string]string{})
			},
			key:       "id",
			expectErr: true,
		},
		{
			name: "different key name",
			setupVars: func(r *http.Request) *http.Request {
				return mux.SetURLVars(r, map[string]string{
					"room_id": "123e4567-e89b-12d3-a456-426614174000",
				})
			},
			key:       "room_id",
			expectErr: false,
			expected:  uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req = tt.setupVars(req)

			result, err := ExtractIDFromRouteVars(req, tt.key)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("ExtractIDFromRouteVars() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Note: Tests for GetRoomFromRequest, VerifyRoomParticipant, RenderHTMLFragment,
// RenderHTMLFragmentOrFallback, and FetchCurrentUser require mocking services
// and are better suited for integration tests. These methods are covered by
// the integration test suite once the refactoring is complete.

package handlers

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
)


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

			// Create Echo context
			e := echo.New()
			req := httptest.NewRequest("GET", "/?"+queryString, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			page, perPage := ParsePaginationParams(c)
			offset := (page - 1) * perPage

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





// Note: Tests for GetRoomFromRequest, VerifyRoomParticipant, RenderHTMLFragment,
// RenderHTMLFragmentOrFallback, and FetchCurrentUser require mocking services
// and are better suited for integration tests. These methods are covered by
// the integration test suite once the refactoring is complete.

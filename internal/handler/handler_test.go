package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"shortURL/internal/repository/memory"
	"shortURL/internal/service"
)

func TestHandler_Shorten(t *testing.T) {
	repo := memory.NewMemoryRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc, "http://localhost:8080")

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "valid URL",
			requestBody:    `{"url":"https://example.com/test"}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp ShortenResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.ShortURL == "" {
					t.Error("Expected short_url in response")
				}
				if len(resp.ShortURL) < len("http://localhost:8080/") {
					t.Error("Short URL too short")
				}
			},
		},
		{
			name:           "empty URL",
			requestBody:    `{"url":""}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if resp.Error == "" {
					t.Error("Expected error message in response")
				}
			},
		},
		{
			name:           "invalid URL format",
			requestBody:    `{"url":"not-a-url"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "invalid JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name:           "missing URL field",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Shorten(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestHandler_ShortenIdempotency(t *testing.T) {
	repo := memory.NewMemoryRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc, "http://localhost:8080")

	url := "https://example.com/test"
	requestBody := `{"url":"` + url + `"}`

	req1 := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(requestBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	handler.Shorten(w1, req1)

	var resp1 ShortenResponse
	json.NewDecoder(w1.Body).Decode(&resp1)

	req2 := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(requestBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handler.Shorten(w2, req2)

	var resp2 ShortenResponse
	json.NewDecoder(w2.Body).Decode(&resp2)

	if resp1.ShortURL != resp2.ShortURL {
		t.Errorf("Not idempotent: got %s and %s", resp1.ShortURL, resp2.ShortURL)
	}
}

func TestHandler_Redirect(t *testing.T) {
	repo := memory.NewMemoryRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc, "http://localhost:8080")

	originalURL := "https://example.com/test"
	shortCode, _ := svc.Create(httptest.NewRequest(http.MethodGet, "/", nil).Context(), originalURL)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "valid short code",
			path:           "/" + shortCode,
			expectedStatus: http.StatusFound,
			expectedURL:    originalURL,
		},
		{
			name:           "non-existent short code",
			path:           "/notexist_",
			expectedStatus: http.StatusNotFound,
			expectedURL:    "",
		},
		{
			name:           "empty path",
			path:           "/",
			expectedStatus: http.StatusBadRequest,
			expectedURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.Redirect(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusFound {
				location := w.Header().Get("Location")
				if location != tt.expectedURL {
					t.Errorf("Redirect location = %s, want %s", location, tt.expectedURL)
				}
			}
		})
	}
}

func TestHandler_ShortenMethodNotAllowed(t *testing.T) {
	repo := memory.NewMemoryRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
	w := httptest.NewRecorder()

	handler.Shorten(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_RedirectMethodNotAllowed(t *testing.T) {
	repo := memory.NewMemoryRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc, "http://localhost:8080")

	req := httptest.NewRequest(http.MethodPost, "/abc123XYZ_", nil)
	w := httptest.NewRecorder()

	handler.Redirect(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestSetupRoutes(t *testing.T) {
	repo := memory.NewMemoryRepository()
	svc := service.NewURLService(repo)
	handler := NewURLHandler(svc, "http://localhost:8080")

	mux := SetupRoutes(handler)

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Post(server.URL+"/shorten", "application/json", bytes.NewBufferString(`{"url":"https://example.com"}`))
	if err != nil {
		t.Fatalf("Failed to call /shorten: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("/shorten status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

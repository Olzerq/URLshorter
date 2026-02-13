package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"shortURL/internal/repository"
	"shortURL/internal/service"
)

type URLHandler struct {
	service *service.URLService
	baseURL string
}

func NewURLHandler(service *service.URLService, baseURL string) *URLHandler {
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &URLHandler{
		service: service,
		baseURL: baseURL,
	}
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Shorten сохраняет оригинальный URL и возвращает short
func (h *URLHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.sendError(w, "url is required", http.StatusBadRequest)
		return
	}

	// создание шортюрл
	shortCode, err := h.service.Create(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			h.sendError(w, "invalid URL format", http.StatusBadRequest)
			return
		}
		h.sendError(w, "failed to create short URL", http.StatusInternalServerError)
		return
	}

	// ответ
	resp := ShortenResponse{
		ShortURL: h.baseURL + "/" + shortCode,
	}
	h.sendJSON(w, resp, http.StatusCreated)
}

// Redirect принимает shortURL и возвращает ориг URL
func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// получаем shortURL
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		h.sendError(w, "short code is required", http.StatusBadRequest)
		return
	}

	// ищем и возвращаем оригЮРЛ по shortURL
	originalURL, err := h.service.Resolve(r.Context(), path)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.sendError(w, "short URL not found", http.StatusNotFound)
			return
		}
		h.sendError(w, "failed to resolve short URL", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (h *URLHandler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *URLHandler) sendError(w http.ResponseWriter, message string, statusCode int) {
	h.sendJSON(w, ErrorResponse{Error: message}, statusCode)
}

func SetupRoutes(handler *URLHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/shorten", handler.Shorten)
	mux.HandleFunc("/", handler.Redirect)

	return mux
}

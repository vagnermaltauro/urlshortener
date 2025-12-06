package handler

import (
    "encoding/json"
    "net/http"
    "urlshortner/internal/model"
    "urlshortner/internal/service"
    "github.com/go-chi/chi/v5"
    "os"
)

type URLHandler struct {
    service *service.URLService
}

func NewURLHandler(s *service.URLService) *URLHandler {
    return &URLHandler{service: s}
}

func (h *URLHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "web/static/index.html")
}

func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
    var req model.CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid json", 400)
        return
    }
    url, err := h.service.CreateShortURL(r.Context(), req.URL)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    host := "http://localhost"
    if r.Host != "" {
        host = "http://" + r.Host
    }
    json.NewEncoder(w).Encode(map[string]string{
        "short_url": host + "/" + url.ID,
    })
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    url, err := h.service.GetURL(r.Context(), id)
    if err != nil || url == nil {
        http.NotFound(w, r)
        return
    }
    http.Redirect(w, r, url.LongURL, http.StatusMovedPermanently)
}

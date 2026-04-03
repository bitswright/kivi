package http

import (
	"encoding/json"
	"net/http"

	"github.com/bitswright/kivi/internal/store"
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{
		store: s,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{key}", h.handleGet)
	mux.HandleFunc("PUT /{key}", h.handleSet)
	mux.HandleFunc("DELETE /{key}", h.handleDelete)
	mux.HandleFunc("GET /keys", h.handleKeys)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	value, err := h.store.Get(key)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"key": key, "value": value})
}

func (h *Handler) handleSet(w http.ResponseWriter, r *http.Request) {	
	key := r.PathValue("key")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

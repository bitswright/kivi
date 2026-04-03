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

	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.store.Set(key, body.Value); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if err := h.store.Delete(key); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.store.Delete(key); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func (h *Handler) handleKeys(w http.ResponseWriter, r *http.Request) {
	keys := h.store.Keys()

	writeJSON(w, http.StatusOK, map[string][]string{"keys": keys})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

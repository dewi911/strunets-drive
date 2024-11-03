package rest

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"time"
)

type Handler struct {
	service StorageService
}

func NewHandler(service StorageService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Init() *mux.Router {
	r := mux.NewRouter()
	files := r.PathPrefix("/files").Subrouter()
	{
		files.HandleFunc("", h.UploadFile).Methods(http.MethodPost)
		files.HandleFunc("", h.ListFiles).Methods(http.MethodGet)
		files.HandleFunc("/{id}", h.DownloadFile).Methods(http.MethodGet)

	}

	return r
}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = h.service.UploadFile(username, header.Filename, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]

	file, filename, err := h.service.DownloadFile(fileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contentReader := bytes.NewReader(content)

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeContent(w, r, filename, time.Time{}, contentReader)
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	files, err := h.service.ListFiles(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(files)
}

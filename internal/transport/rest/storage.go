package rest

import (
	"encoding/json"
	"github.com/gorilla/mux"
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

	if header.Size == 0 {
		http.Error(w, "file size is zero", http.StatusBadRequest)
		return
	}

	err = h.service.UploadFile(username, header.Filename, file)
	if err != nil {
		http.Error(w, "failed to upload file"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "file uploaded successfully",
	})
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]
	if fileID == "" {
		http.Error(w, "file id is empty", http.StatusBadRequest)
		return
	}

	readSeeker, filename, err := h.service.DownloadFile(fileID)
	if err != nil {
		http.Error(w, "failed to get files"+err.Error(), http.StatusInternalServerError)
		return
	}
	defer readSeeker.Close()
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeContent(w, r, filename, time.Time{}, readSeeker)
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)
	files, err := h.service.ListFiles(username)
	if err != nil {
		http.Error(w, "failed to list files"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

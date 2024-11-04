package rest

import (
	"encoding/json"
	"fmt"
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
	maxFileSize := int64(1 << 30)
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size == 0 {
		http.Error(w, "File size is zero", http.StatusBadRequest)
		return
	}

	username := r.Context().Value("username").(string)

	fileInfo, err := h.service.UploadFile(username, header.Filename, file, header.Size)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload file: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "File uploaded successfully",
		"file":    fileInfo,
	})
}

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]
	if fileID == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	readSeeker, fileInfo, err := h.service.DownloadFile(fileID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get file: %v", err), http.StatusInternalServerError)
		return
	}
	defer readSeeker.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileInfo.Name))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))
	w.Header().Set("Accept-Ranges", "bytes")

	http.ServeContent(w, r, fileInfo.Name, time.Time{}, readSeeker)
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	files, err := h.service.ListFiles(username)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list files: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(files); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

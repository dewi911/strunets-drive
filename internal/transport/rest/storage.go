package rest

import (
	"github.com/gorilla/mux"
	"net/http"
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

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {}

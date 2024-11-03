package main

import (
	"fmt"
	"strunetsdrive/pkg/database"

	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strunetsdrive/internal/repository"
	"strunetsdrive/internal/service"
	"strunetsdrive/internal/transport/rest"
)

func main() {
	db, err := database.NewPostgresConnection(database.ConnectionInfo{
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		DBName:   "postgres",
		SSLMode:  "disable",
		Password: "qwerty",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	repo := repository.NewPostgresRepo(db)
	path := "fergerf"
	log.Printf("repo path: %s", path)
	service := service.NewService(repo, path)
	log.Println("starting server")
	handler := rest.NewHandler(service)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: handler.Init(),
	}
	log.Print("starting server on port 8080")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("could not listend on 8080")
	}
	go func() {

	}()

}

//type Storage struct {
//	mu    sync.Mutex
//	files map[string][]byte
//}
//
//func NewStorage() *Storage {
//	return &Storage{
//		files: make(map[string][]byte),
//	}
//}
//
//func (s *Storage) Save(key string, data []byte) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	s.files[key] = data
//
//	err := os.WriteFile(STORAGE_DIR+"/"+key, data, 0644)
//	if err != nil {
//		log.Printf("ошибка при сохранении файла %s , %v", key, err)
//	}
//}
//func (s *Storage) Load(key string) ([]byte, bool) {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	data, exist := s.files[key]
//	if exist {
//		return data, true
//	}
//
//	data, err := os.ReadFile(STORAGE_DIR + "/" + key)
//	if err != nil {
//		return nil, false
//	}
//
//	return s.files[key], true
//}
//
//func HandleUpload(w http.ResponseWriter, r *http.Request, s *Storage) {
//	if r.Method != http.MethodPost {
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//		return
//	}
//
//	key := r.URL.Path[UPLOAD_PREFIX_LEN:]
//
//	data, err := io.ReadAll(r.Body)
//	if err != nil {
//		http.Error(w, "error reading data", http.StatusInternalServerError)
//		return
//	}
//
//	s.Save(key, data)
//
//	w.WriteHeader(http.StatusOK)
//	fmt.Fprintf(w, "object %s uploaded successfully", key)
//}
//func HandleDownload(w http.ResponseWriter, r *http.Request, s *Storage) {
//	if r.Method != http.MethodGet {
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//		return
//	}
//
//	key := r.URL.Path[DOWNLOAD_PREFIX_LEN:]
//
//	data, exist := s.Load(key)
//	if !exist {
//		http.Error(w, "object not found", http.StatusNotFound)
//		return
//	}
//
//	w.WriteHeader(http.StatusOK)
//	w.Write(data)
//}
//func HandleList(w http.ResponseWriter, r *http.Request, s *Storage) {
//	if r.Method != http.MethodGet {
//		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
//		return
//	}
//
//	s.mu.Lock()
//	defer s.mu.Unlock()
//
//	keys := make([]string, 0, len(s.files))
//	for key := range s.files {
//		keys = append(keys, key)
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	if err := json.NewEncoder(w).Encode(keys); err != nil {
//		http.Error(w, err.Error(), http.StatusInternalServerError)
//		return
//	}
//}
//
//const (
//	STORAGE_DIR         = "./storage"
//	UPLOAD_PREFIX_LEN   = len("/upload/")
//	DOWNLOAD_PREFIX_LEN = len("/download/")
//)
//
//func main() {
//	if _, err := os.Stat(STORAGE_DIR); os.IsNotExist(err) {
//		err := os.Mkdir(STORAGE_DIR, 0755)
//		if err != nil {
//			log.Fatalf("error creating storage dir %s, %v", STORAGE_DIR, err)
//		}
//	}
//
//	storage := NewStorage()
//
//	http.HandleFunc("/upload/", func(w http.ResponseWriter, r *http.Request) {
//		HandleUpload(w, r, storage)
//	})
//
//	http.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
//		HandleDownload(w, r, storage)
//	})
//
//	http.HandleFunc("/list/", func(w http.ResponseWriter, r *http.Request) {
//		HandleList(w, r, storage)
//	})
//
//	log.Println("server started on :8080")
//	log.Fatal(http.ListenAndServe(":8080", nil))
//
//}

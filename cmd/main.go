package main

import (
	"fmt"
	cors2 "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strunetsdrive/internal/config"
	"strunetsdrive/internal/repository"
	"strunetsdrive/internal/service"
	"strunetsdrive/internal/transport/rest"
	"strunetsdrive/pkg/database"
	"strunetsdrive/pkg/filestore"
	"strunetsdrive/pkg/filestore/minio"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

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

	var fileStore filestore.Store
	if cfg.Storage.Type == "minio" {
		fileStore, err = minio.NewStore(
			cfg.Storage.Minio.Endpoint,
			cfg.Storage.Minio.AccessKey,
			cfg.Storage.Minio.SecretKey,
			cfg.Storage.Minio.Bucket,
			cfg.Storage.Minio.UseSSL,
		)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Using MinIO storage at %s", cfg.Storage.Minio.Endpoint)
	}

	//init repo
	storeRepository := repository.NewStoreRepo(db)
	usersRepository := repository.NewUsers(db)
	tokensRepository := repository.NewTokens(db)

	//init service
	usersService := service.NewUsers(usersRepository, tokensRepository, time.Hour*24, "testgovna")
	storeService := service.NewStoreService(storeRepository, fileStore)

	//init handlers
	userHandler := rest.NewAuthHandler(usersService)
	fileHandler := rest.NewFileHandler(storeService)

	//service := service.NewService(repo, fileStore)
	//handler := rest.NewHandler(service)

	//init router
	router := gin.Default()

	// Add logging middleware
	config := cors2.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors2.New(config))
	router.Use(rest.LoggingMiddleware())

	// Setup routes
	userHandler.InjectRoutes(router)
	fileHandler.InjectRoutes(router, userHandler.AuthMiddleware())

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: router,
	}
	log.Print("starting server on port 8080")

	log.Printf("Starting server on %s", 8080)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Could not listen on %s: %v", cfg.ServerAddress, err)
	}
}

//} else {
//	fileStore, err = local.NewStore(cfg.Storage.Local.Path)
//	if err != nil {
//		log.Fatal(err)
//	}
//	log.Printf("Using local storage at %s", cfg.Storage.Local.Path)
//}

//path := "C:\\localhost\\"
//log.Printf("repo path: %s", path)
//fileStore, err := local.NewStore(path)
//if err != nil {
//	log.Fatal(err)
//}
//

//fileStore, err := minio.NewStore(
//"localhost:9000",
//"minioadmin",
//"minioadmin",
//"mybucket",
//false,
//)
//if err != nil {
//log.Fatal(err)
//}

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

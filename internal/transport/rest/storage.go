package rest

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type FileHandler struct {
	service StorageService
}

func NewFileHandler(service StorageService) *FileHandler {
	return &FileHandler{
		service: service,
	}
}

func (h *FileHandler) InjectRoutes(r *gin.Engine, middlewares ...gin.HandlerFunc) {
	files := r.Group("/files").Use(middlewares...)
	{
		files.POST("", h.UploadFile)
		files.GET("", h.ListFiles)
		files.GET("/:id", h.DownloadFile)
	}
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	username, err := GetUsernameFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get username",
		})
		return
	}

	maxFileSize := int64(1 << 30) // 1GB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to get file: %v", err),
		})
		return
	}
	defer file.Close()

	if header.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File size is zero",
		})
		return
	}

	//username := c.GetString("username") // Предполагается, что username установлен в middleware

	fileInfo, err := h.service.UploadFile(username, header.Filename, file, header.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to upload file: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "File uploaded successfully",
		"file":    fileInfo,
	})
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File ID is required",
		})
		return
	}

	readSeeker, fileInfo, err := h.service.DownloadFile(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get file: %v", err),
		})
		return
	}
	defer readSeeker.Close()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileInfo.Name))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size))
	c.Header("Accept-Ranges", "bytes")

	http.ServeContent(c.Writer, c.Request, fileInfo.Name, time.Time{}, readSeeker)
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	//username := c.GetString("username") // Предполагается, что username установлен в middleware
	username, err := GetUsernameFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get username",
		})
		return
	}

	files, err := h.service.ListFiles(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to list files: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, files)
}

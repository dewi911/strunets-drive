package rest

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"path/filepath"
	"strunetsdrive/internal/models"
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
	{

		files.DELETE("/:id", h.DeleteFile)
		files.PUT("/:id", h.UpdateFile)
		files.POST("/:id/copy", h.CopyFile)
		files.PUT("/:id/move", h.MoveFile)
		files.GET("/:id/info", h.GetFileInfo)
	}

	folders := r.Group("/folders").Use(middlewares...)
	{
		folders.POST("", h.CreateFolder)
		folders.GET("", h.GetFolderContent)
		folders.GET("/:id", h.GetFolderContent)
		folders.GET("/:id/download", h.DownloadFolder)
		folders.GET("/hierarchy", h.GetFolderHierarchy)
		folders.GET("/complete", h.GetCompleteHierarchy)
	}
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	//username := c.GetString("username")
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

func (h *FileHandler) GetFolderContent(c *gin.Context) {
	username, err := GetUsernameFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get username",
		})
		return
	}

	folderID := c.Param("id")

	folderContent, err := h.service.GetFolderContent(folderID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folderContent)
}

func (h *FileHandler) UploadFolderStructure(c *gin.Context) {
	username, err := GetUsernameFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get username",
		})
		return
	}

	parentFolderID := c.PostForm("parent_folder_id")

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to parse multipart form: %v", err),
		})
		return
	}
	defer form.RemoveAll()

	if parentFolderID == "" {
		rootFolder, err := h.service.GetRootFolder(username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get root folder",
			})
			return
		}
		parentFolderID = rootFolder.ID
	}

	uploadedFiles := form.File["files"]
	var uploadResults []*models.FileResponse
	var uploadErrors []string

	folderMap := make(map[string]*models.Folder)

	for _, fileHeader := range uploadedFiles {
		relativePath := fileHeader.Filename

		dir := filepath.Dir(relativePath)
		filename := filepath.Base(relativePath)

		currentParentID := parentFolderID
		if dir != "." {

			currentParentID, err = h.createNestedFolders(username, dir, parentFolderID, folderMap)
			if err != nil {
				uploadErrors = append(uploadErrors, fmt.Sprintf("Failed to create folder for %s: %v", relativePath, err))
				continue
			}
		}

		file, err := fileHeader.Open()
		if err != nil {
			uploadErrors = append(uploadErrors, fmt.Sprintf("Failed to open file %s: %v", relativePath, err))
			continue
		}
		defer file.Close()

		fileInfo, err := h.service.UploadFile(username, filename, file, fileHeader.Size, currentParentID)
		if err != nil {
			uploadErrors = append(uploadErrors, fmt.Sprintf("Failed to upload file %s: %v", relativePath, err))
			continue
		}

		uploadResults = append(uploadResults, &models.FileResponse{
			ID:         fileInfo.ID,
			Name:       fileInfo.Name,
			Size:       fileInfo.Size,
			UploadedAt: fileInfo.UploadedAt,
		})
	}

	if len(uploadErrors) > 0 {
		c.JSON(http.StatusPartialContent, gin.H{
			"uploaded_files": uploadResults,
			"errors":         uploadErrors,
		})
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"uploaded_files": uploadResults,
		})
	}
}

func (h *FileHandler) createNestedFolders(username, relativePath, parentFolderID string, folderMap map[string]*models.Folder) (string, error) {

	if folder, exists := folderMap[relativePath]; exists {
		return folder.ID, nil
	}

	pathParts := filepath.SplitList(relativePath)

	currentParentID := parentFolderID
	currentPath := ""

	for _, part := range pathParts {
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = filepath.Join(currentPath, part)
		}

		if folder, exists := folderMap[currentPath]; exists {
			currentParentID = folder.ID
			continue
		}

		folder, err := h.service.CreateFolder(username, part, currentParentID)
		if err != nil {
			return "", err
		}

		folderMap[currentPath] = folder

		currentParentID = folder.ID
	}

	return currentParentID, nil
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	username, err := GetUsernameFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get username",
		})
		return
	}

	folderID := c.PostForm("folderID")

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

	//username := c.GetString("username")

	fileInfo, err := h.service.UploadFile(username, header.Filename, file, header.Size, folderID)
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

func (h *FileHandler) DownloadFolder(c *gin.Context) {
	username, err := GetUsernameFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get username",
		})
		return
	}

	folderID := c.Param("id")

	zipReader, err := h.service.DownloadFolderAsZip(folderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer zipReader.Close()

	folderContent, err := h.service.GetFolderContent(folderID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", folderContent.Name))

	c.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, zipReader)
		return err == nil
	})
}

func (h *FileHandler) CreateFolder(c *gin.Context) {
	username, err := GetUsernameFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get username"})
		return
	}

	var folderInput struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&folderInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.service.CreateFolder(username, folderInput.Name, folderInput.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, folder)
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

func (h *FileHandler) DeleteFile(c *gin.Context) {
	username, err := GetUsernameFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get username"})
		return
	}

	fileID := c.Param("id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	if err := h.service.DeleteFile(username, fileID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

func (h *FileHandler) GetFolderHierarchy(c *gin.Context) {
	username, err := GetUsernameFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get username",
		})
		return
	}

	hierarchy, err := h.service.GetFolderHierarchy(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get folder hierarchy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, hierarchy)
}

func (h *FileHandler) GetCompleteHierarchy(c *gin.Context) {
	username, err := GetUsernameFromContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get username",
		})
		return
	}

	hierarchy, err := h.service.GetCompleteHierarchy(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to get complete hierarchy: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, hierarchy)
}

func (h *FileHandler) CopyFile(c *gin.Context) {
	//fileID := c.Param("id")
	//if fileID == "" {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": "File ID is required",
	//	})
	//	return
	//}
	//
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"error": "Failed to get username",
	//	})
	//	return
	//}
	//
	//var copyInput struct {
	//	NewFileName string `json:"new_filename"`
	//}
	//if err := c.ShouldBindJSON(&copyInput); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": "Invalid input",
	//	})
	//	return
	//}
	//
	//// If no new filename provided, generate a default one
	//if copyInput.NewFileName == "" {
	//	copyInput.NewFileName = fmt.Sprintf("Copy_of_%s", fileID)
	//}
	//
	//copiedFileInfo, err := h.service.CopyFile(username, fileID, copyInput.NewFileName)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"error": fmt.Sprintf("Failed to copy file: %v", err),
	//	})
	//	return
	//}
	//
	//c.JSON(http.StatusCreated, gin.H{
	//	"message": "File copied successfully",
	//	"file":    copiedFileInfo,
	//})
}

func (h *FileHandler) GetFileInfo(c *gin.Context) {
	//fileID := c.Param("id")
	//if fileID == "" {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": "File ID is required",
	//	})
	//	return
	//}
	//
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"error": "Failed to get username",
	//	})
	//	return
	//}
	//
	//fileInfo, err := h.service.GetFileInfo(username, fileID)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"error": fmt.Sprintf("Failed to retrieve file info: %v", err),
	//	})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, fileInfo)
}

func (h *FileHandler) MoveFile(c *gin.Context) {
	//fileID := c.Param("id")
	//if fileID == "" {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": "File ID is required",
	//	})
	//	return
	//}
	//
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"error": "Failed to get username",
	//	})
	//	return
	//}
	//
	//var moveInput struct {
	//	DestinationPath string `json:"destination_path"`
	//}
	//if err := c.ShouldBindJSON(&moveInput); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": "Invalid input",
	//	})
	//	return
	//}
	//
	//if moveInput.DestinationPath == "" {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": "Destination path is required",
	//	})
	//	return
	//}
	//
	//movedFileInfo, err := h.service.MoveFile(username, fileID, moveInput.DestinationPath)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"error": fmt.Sprintf("Failed to move file: %v", err),
	//	})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, gin.H{
	//	"message": "File moved successfully",
	//	"file":    movedFileInfo,
	//})
}

func (h *FileHandler) UpdateFile(c *gin.Context) {
	//fileID := c.Param("id")
	//if fileID == "" {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": "File ID is required",
	//	})
	//	return
	//}
	//
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"error": "Failed to get username",
	//	})
	//	return
	//}
	//
	//var updateInput struct {
	//	Name string   `json:"name"`
	//	Tags []string `json:"tags"`
	//}
	//if err := c.ShouldBindJSON(&updateInput); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"error": "Invalid input",
	//	})
	//	return
	//}
	//
	//updatedFileInfo, err := h.service.UpdateFileMetadata(username, fileID, updateInput.Name, updateInput.Tags)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"error": fmt.Sprintf("Failed to update file: %v", err),
	//	})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, gin.H{
	//	"message": "File metadata updated successfully",
	//	"file":    updatedFileInfo,
	//})
}

//fileId := c.Param("id")
//if fileId == "" {
//	c.JSON(http.StatusBadRequest, gin.H{
//		"error": "File ID is required",
//	})
//	return
//}
//
//var input struct {
//	Name string   `json:"name"`
//	Tags []string `json:"tags"`
//}
//if err := c.ShouldBindJSON(&input); err != nil {
//	c.JSON(http.StatusBadRequest, gin.H{
//		"error": "Invalid input",
//	})
//}
//
//username, err := GetUsernameFromContext(c)
//if err != nil {
//	c.JSON(http.StatusInternalServerError, gin.H{
//		"error": "Failed to get username",
//	})
//	return
//}
//
//c.JSON(http.StatusOK, gin.H{})

func (h *FileHandler) SearchFiles(c *gin.Context) {
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get username"})
	//	return
	//}
	//
	//query := c.Query("q")
	//fileType := c.Query("type")
	//startDate := c.Query("start_date")
	//endDate := c.Query("end_date")
	//tags := c.QueryArray("tags")
	//folder := c.Query("folder")
	//
	//searchParams := SearchParams{
	//	Query:     query,
	//	FileType:  fileType,
	//	StartDate: startDate,
	//	EndDate:   endDate,
	//	Tags:      tags,
	//	Folder:    folder,
	//}
	//
	//results, err := h.service.SearchFiles(username, searchParams)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, results)
}

func (h *FileHandler) GetRecentFiles(c *gin.Context) {
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get username"})
	//	return
	//}
	//
	//limit := c.DefaultQuery("limit", "20")
	//files, err := h.service.GetRecentFiles(username, limit)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, files)
}

func (h *FileHandler) BatchMoveFiles(c *gin.Context) {
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get username"})
	//	return
	//}
	//
	//var input struct {
	//	FileIDs  []string `json:"file_ids"`
	//	TargetID string   `json:"target_folder_id"`
	//}
	//
	//if err := c.ShouldBindJSON(&input); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
	//	return
	//}
	//
	//if err := h.service.BatchMoveFiles(username, input.FileIDs, input.TargetID); err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, gin.H{"message": "Files moved successfully"})
}

func (h *FileHandler) UpdateFileTags(c *gin.Context) {
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get username"})
	//	return
	//}
	//
	//fileID := c.Param("id")
	//var input struct {
	//	Tags []string `json:"tags"`
	//}
	//
	//if err := c.ShouldBindJSON(&input); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
	//	return
	//}
	//
	//if err := h.service.UpdateFileTags(username, fileID, input.Tags); err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, gin.H{"message": "Tags updated successfully"})
}

func (h *FileHandler) GetStorageQuota(c *gin.Context) {
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get username"})
	//	return
	//}
	//
	//quota, err := h.service.GetStorageQuota(username)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, quota)
}

func (h *FileHandler) CompressFiles(c *gin.Context) {
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get username"})
	//	return
	//}
	//
	//var input struct {
	//	FileIDs []string `json:"file_ids"`
	//	Format  string   `json:"format"` // zip, tar.gz, etc.
	//	Name    string   `json:"name"`
	//}
	//
	//if err := c.ShouldBindJSON(&input); err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
	//	return
	//}
	//
	//archiveInfo, err := h.service.CompressFiles(username, input.FileIDs, input.Format, input.Name)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, archiveInfo)
}

func (h *FileHandler) GetUserActivity(c *gin.Context) {
	//username, err := GetUsernameFromContext(c)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get username"})
	//	return
	//}
	//
	//startDate := c.Query("start_date")
	//endDate := c.Query("end_date")
	//
	//activity, err := h.service.GetUserActivity(username, startDate, endDate)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, activity)
}

//// Поиск и метаданные
//files.GET("/search", h.SearchFiles)
//files.GET("/recent", h.GetRecentFiles)
//files.GET("/starred", h.GetStarredFiles)
//files.PUT("/:id/star", h.ToggleStarFile)
//files.PUT("/:id/tags", h.UpdateFileTags)
//files.GET("/by-type/:type", h.GetFilesByType)
//
//// Расширенные операции с файлами
//files.POST("/batch/delete", h.BatchDeleteFiles)
//files.POST("/batch/move", h.BatchMoveFiles)
//files.POST("/batch/copy", h.BatchCopyFiles)
//files.POST("/:id/compress", h.CompressFiles)
//files.POST("/decompress/:id", h.DecompressArchive)
//
//// Шаринг и совместный доступ
//files.POST("/:id/share", h.ShareFile)
//files.DELETE("/:id/share", h.RevokeShare)
//files.GET("/shared", h.ListSharedFiles)
//files.GET("/shared/:shareId", h.GetSharedFile)
//files.PUT("/:id/share/permissions", h.UpdateSharePermissions)
//files.GET("/:id/share/links", h.GetShareLinks)
//files.POST("/:id/share/link", h.CreateShareLink)
//files.DELETE("/:id/share/link/:linkId", h.DeleteShareLink)
//
//// Работа с папками
//files.POST("/folder", h.CreateFolder)
//files.GET("/folder/:id", h.GetFolderContents)
//files.DELETE("/folder/:id", h.DeleteFolder)
//files.PUT("/folder/:id", h.RenameFolder)
//files.GET("/folder/:id/tree", h.GetFolderTree)
//files.GET("/folder/:id/size", h.GetFolderSize)
//
//// Версионирование
//files.GET("/:id/versions", h.ListFileVersions)
//files.POST("/:id/restore/:versionId", h.RestoreFileVersion)
//files.DELETE("/:id/versions/:versionId", h.DeleteFileVersion)
//
//// Квоты и статистика
//files.GET("/quota", h.GetStorageQuota)
//files.GET("/stats/usage", h.GetStorageUsageStats)
//files.GET("/stats/activity", h.GetUserActivity)

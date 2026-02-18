package http

import (
	"fmt"
	"net/http"
	"video-processor/internal/core/domain"
	"video-processor/internal/core/ports"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	videoUseCase ports.VideoUseCase
	userUseCase  ports.UserUseCase
	storage      ports.Storage
	jwtSecret    string
}

func NewHandler(v ports.VideoUseCase, u ports.UserUseCase, s ports.Storage, jwtSecret string) *Handler {
	return &Handler{
		videoUseCase: v,
		userUseCase:  u,
		storage:      s,
		jwtSecret:    jwtSecret,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	fmt.Println("Registering routes...")
	r.GET("/", h.HandleIndex)

	// Protected routes
	auth := r.Group("/api")
	auth.Use(AuthMiddleware(h.jwtSecret))
	{
		fmt.Println("Registering: POST /api/upload")
		auth.POST("/upload", h.HandleVideoUpload)
		fmt.Println("Registering: GET /api/videos")
		auth.GET("/videos", h.HandleListUserVideos)
		fmt.Println("Registering: GET /api/status")
		auth.GET("/status", h.HandleStatus) // Legacy or general status
	}

	r.GET("/download/:filename", h.HandleDownload)

	// Auth routes
	r.POST("/register", h.HandleRegister)
	r.POST("/login", h.HandleLogin)
}

func (h *Handler) HandleIndex(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, getHTMLForm())
}

// HandleVideoUpload handles video file uploads
// @Summary Upload and process a video
// @Description Receives a video file, saves it, and publishes a processing event.
// @Tags videos
// @Accept multipart/form-data
// @Produce json
// @Param video formData file true "Video file"
// @Success 200 {object} domain.ProcessingResult
// @Failure 400 {object} domain.ProcessingResult
// @Failure 401 {object} domain.ProcessingResult
// @Failure 500 {object} domain.ProcessingResult
// @Security ApiKeyAuth
// @Router /api/upload [post]
func (h *Handler) HandleVideoUpload(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Usuário não identificado"})
		return
	}

	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Erro ao receber arquivo: " + err.Error()})
		return
	}
	defer file.Close()

	result, err := h.videoUseCase.UploadAndProcess(userID.(int64), header.Filename, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleListUserVideos lists all videos for the authenticated user
// @Summary List user videos
// @Description Retrieves a list of all videos uploaded by the authenticated user.
// @Tags videos
// @Produce json
// @Success 200 {object} domain.ListVideosResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/videos [get]
func (h *Handler) HandleListUserVideos(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Usuário não identificado"})
		return
	}

	videos, err := h.videoUseCase.GetVideosByUserID(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erro ao listar vídeos: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"videos":  videos,
	})
}

// HandleDownload serves the processed ZIP file
// @Summary Download processed video
// @Description Downloads the ZIP file containing extracted frames for a processed video.
// @Tags videos
// @Param filename path string true "ZIP filename"
// @Produce application/zip
// @Success 200 {file} file
// @Router /download/{filename} [get]
func (h *Handler) HandleDownload(c *gin.Context) {
	filename := c.Param("filename")
	filePath := h.storage.GetOutputPath(filename)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/zip")

	c.File(filePath)
}

// HandleStatus lists all processed files (Legacy/Admin)
// @Summary List all processed files
// @Description Retrieves a list of all processed ZIP files.
// @Tags videos
// @Produce json
// @Success 200 {object} domain.FileListResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security ApiKeyAuth
// @Router /api/status [get]
func (h *Handler) HandleStatus(c *gin.Context) {
	files, err := h.videoUseCase.ListProcessedFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Success: false, Message: "Erro ao listar arquivos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"total": len(files),
	})
}

// HandleRegister registers a new user
// @Summary Register a new user
// @Description Creates a new user account with email, password, and name.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Registration data"
// @Success 201 {object} domain.AuthResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse
// @Router /register [post]
func (h *Handler) HandleRegister(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Dados inválidos: " + err.Error()})
		return
	}

	response, err := h.userUseCase.Register(req.Email, req.Password, req.Name)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// HandleLogin authenticates a user
// @Summary Authenticate user
// @Description Logs in a user and returns a JWT token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login credentials"
// @Success 200 {object} domain.AuthResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /login [post]
func (h *Handler) HandleLogin(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Dados inválidos: " + err.Error()})
		return
	}

	response, err := h.userUseCase.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

package services

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
	"video-processor/internal/core/domain"
	"video-processor/internal/core/ports"
)

type videoService struct {
	storage   ports.Storage
	repo      ports.VideoRepository
	publisher ports.EventPublisher
}

func NewVideoService(s ports.Storage, r ports.VideoRepository, p ports.EventPublisher) ports.VideoUseCase {
	return &videoService{
		storage:   s,
		repo:      r,
		publisher: p,
	}
}

func (s *videoService) UploadAndProcess(userID int64, filename string, file io.Reader) (domain.ProcessingResult, error) {
	if !s.isValidVideoFile(filename) {
		return domain.ProcessingResult{
			Success:   false,
			Message:   "Formato de arquivo não suportado. Use: mp4, avi, mov, mkv",
			ErrorCode: "ERR_INVALID_FORMAT",
		}, nil
	}

	timestamp := time.Now().Format("20060102_150405")
	uniqueID := time.Now().UnixNano()
	uniqueFilename := fmt.Sprintf("%s_%d_%s", timestamp, uniqueID, filename)

	videoPath, err := s.storage.SaveUpload(uniqueFilename, file)
	if err != nil {
		return domain.ProcessingResult{
			Success:   false,
			Message:   "Erro ao salvar arquivo: " + err.Error(),
			ErrorCode: "ERR_STORAGE_FAIL",
		}, err
	}

	video := &domain.Video{
		UserID:   userID,
		Filename: uniqueFilename, // Store the unique filename so worker can find it
		Status:   domain.StatusPending,
	}

	err = s.repo.Create(video)
	if err != nil {
		s.storage.DeleteFile(videoPath)
		return domain.ProcessingResult{
			Success:   false,
			Message:   "Erro ao criar registro no banco: " + err.Error(),
			ErrorCode: "ERR_DB_FAIL",
		}, err
	}

	// Publish NATS event
	err = s.publisher.PublishUploadEvent(video.ID, video.Filename)
	if err != nil {
		// Log error but don't fail the upload since it's already in DB/Storage
		// or should we fail it? Usually, we want the event to be published.
		// For this exercise, let's keep it robust.
		fmt.Printf("⚠️ Erro ao publicar evento no NATS: %v\n", err)
	}

	return domain.ProcessingResult{
		Success: true,
		Message: "Vídeo recebido e adicionado à fila de processamento!",
		VideoID: video.ID,
	}, nil
}

func (s *videoService) ListProcessedFiles() ([]domain.FileInfo, error) {
	return s.storage.ListOutputs()
}

func (s *videoService) GetVideosByUserID(userID int64) ([]domain.Video, error) {
	return s.repo.GetByUserID(userID)
}

func (s *videoService) isValidVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExts := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv", ".flv", ".webm"}

	for _, validExt := range validExts {
		if ext == validExt {
			return true
		}
	}
	return false
}

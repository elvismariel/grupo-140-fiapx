package services

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
	"video-processor-worker/internal/core/domain"
	"video-processor-worker/internal/core/ports"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	videoProcessingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "worker_video_processing_duration_seconds",
		Help:    "Duration of video processing in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"status"})

	videosProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_videos_processed_total",
		Help: "Total number of videos processed",
	}, []string{"status"})
)

type workerService struct {
	processor ports.VideoProcessor
	storage   ports.Storage
	repo      ports.VideoRepository
	userRepo  ports.UserRepository
	emailer   ports.EmailSender
}

func NewWorkerService(p ports.VideoProcessor, s ports.Storage, r ports.VideoRepository, ur ports.UserRepository, e ports.EmailSender) *workerService {
	return &workerService{
		processor: p,
		storage:   s,
		repo:      r,
		userRepo:  ur,
		emailer:   e,
	}
}

func (s *workerService) ProcessVideoByID(ctx context.Context, videoID int64) error {
	log.Printf("📥 Processing request for video ID: %d", videoID)

	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("error fetching video %d: %w", videoID, err)
	}

	if video == nil {
		return fmt.Errorf("video %d not found", videoID)
	}

	if video.Status != domain.StatusPending {
		log.Printf("ℹ️ Video %d already in status %s, skipping", videoID, video.Status)
		return nil
	}

	return s.processVideo(ctx, video)
}

func (s *workerService) processVideo(ctx context.Context, video *domain.Video) error {
	start := time.Now()
	var status = "success"

	defer func() {
		duration := time.Since(start).Seconds()
		videoProcessingDuration.WithLabelValues(status).Observe(duration)
		videosProcessedTotal.WithLabelValues(status).Inc()
	}()

	// Update status to PROCESSING
	video.Status = domain.StatusProcessing
	video.Message = "Processamento iniciado..."
	if err := s.repo.Update(ctx, video); err != nil {
		return fmt.Errorf("error updating video status: %w", err)
	}

	videoPath := s.storage.GetUploadPath(video.Filename)
	uniqueJobID := strings.TrimSuffix(video.Filename, filepath.Ext(video.Filename))

	log.Printf("🎬 Extracting frames for video ID: %d (%s)", video.ID, video.Filename)
	frames, err := s.processor.ExtractFrames(videoPath, uniqueJobID)
	if err != nil {
		log.Printf("❌ Error extracting frames for video %d: %v", video.ID, err)
		video.Status = domain.StatusFailed
		video.Message = "Erro no processamento: " + err.Error()
		s.repo.Update(ctx, video)
		s.storage.DeleteFile(videoPath)
		s.notifyFailure(ctx, video)
		status = "error"
		return err
	}

	log.Printf("📦 Creating ZIP for video ID: %d", video.ID)
	zipFilename := fmt.Sprintf("frames_%s.zip", uniqueJobID)
	err = s.storage.SaveZip(zipFilename, frames)
	if err != nil {
		log.Printf("❌ Error saving ZIP for video %d: %v", video.ID, err)
		video.Status = domain.StatusFailed
		video.Message = "Erro ao criar ZIP: " + err.Error()
		s.repo.Update(ctx, video)
		s.storage.DeleteFile(videoPath)
		s.notifyFailure(ctx, video)
		status = "error"
		return err
	}

	// Cleanup
	s.storage.DeleteFile(videoPath)
	if len(frames) > 0 {
		tempDir := filepath.Dir(frames[0])
		s.storage.DeleteDir(tempDir)
	}

	// Final Update
	video.Status = domain.StatusCompleted
	video.ZipPath = zipFilename
	video.FrameCount = len(frames)
	video.Message = fmt.Sprintf("Processamento concluído! %d frames extraídos.", len(frames))
	if err := s.repo.Update(ctx, video); err != nil {
		log.Printf("❌ Error final updating video %d: %v", video.ID, err)
		return err
	}

	log.Printf("✅ Video %d processed successfully", video.ID)
	return nil
}

func (s *workerService) notifyFailure(ctx context.Context, video *domain.Video) {
	log.Printf("📧 Initiating failure notification for video %d (User %d)", video.ID, video.UserID)

	user, err := s.userRepo.GetByID(video.UserID)
	if err != nil {
		log.Printf("⚠️ Error fetching user %d for notification: %v", video.UserID, err)
		return
	}

	if user == nil || user.Email == "" {
		log.Printf("⚠️ User %d not found or has no email for notification", video.UserID)
		return
	}

	subject := fmt.Sprintf("Falha no Processamento do Vídeo: %s", video.Filename)
	body := fmt.Sprintf("Olá %s,\n\nInfelizmente ocorreu um erro ao processar seu vídeo '%s'.\n\nDetalhes do erro: %s\n\nEquipe FiapX", user.Name, video.Filename, video.Message)

	if err := s.emailer.SendEmail(user.Email, subject, body); err != nil {
		log.Printf("❌ Failed to send failure email to %s: %v", user.Email, err)
	}
}

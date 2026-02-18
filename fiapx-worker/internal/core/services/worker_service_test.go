package services

import (
	"context"
	"errors"
	"testing"
	"video-processor-worker/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWorkerService_ProcessVideoByID(t *testing.T) {
	ctx := context.Background()

	t.Run("video not found", func(t *testing.T) {
		processor := new(MockVideoProcessor)
		storage := new(MockStorage)
		repo := new(MockVideoRepository)
		userRepo := new(MockUserRepository)
		emailer := new(MockEmailSender)
		service := NewWorkerService(processor, storage, repo, userRepo, emailer)

		repo.On("GetByID", ctx, int64(1)).Return(nil, nil)

		err := service.ProcessVideoByID(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "video 1 not found")
	})

	t.Run("video already processed", func(t *testing.T) {
		processor := new(MockVideoProcessor)
		storage := new(MockStorage)
		repo := new(MockVideoRepository)
		userRepo := new(MockUserRepository)
		emailer := new(MockEmailSender)
		service := NewWorkerService(processor, storage, repo, userRepo, emailer)

		video := &domain.Video{ID: 1, Status: domain.StatusCompleted}
		repo.On("GetByID", ctx, int64(1)).Return(video, nil)

		err := service.ProcessVideoByID(ctx, 1)

		assert.NoError(t, err)
		repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
	})

	t.Run("success processing", func(t *testing.T) {
		processor := new(MockVideoProcessor)
		storage := new(MockStorage)
		repo := new(MockVideoRepository)
		userRepo := new(MockUserRepository)
		emailer := new(MockEmailSender)
		service := NewWorkerService(processor, storage, repo, userRepo, emailer)

		video := &domain.Video{ID: 1, Status: domain.StatusPending, Filename: "video.mp4"}
		repo.On("GetByID", ctx, int64(1)).Return(video, nil)
		repo.On("Update", ctx, mock.MatchedBy(func(v *domain.Video) bool {
			return v.Status == domain.StatusProcessing
		})).Return(nil)

		storage.On("GetUploadPath", "video.mp4").Return("/uploads/video.mp4")
		processor.On("ExtractFrames", "/uploads/video.mp4", "video").Return([]string{"/tmp/frame1.jpg", "/tmp/frame2.jpg"}, nil)
		storage.On("SaveZip", "frames_video.zip", []string{"/tmp/frame1.jpg", "/tmp/frame2.jpg"}).Return(nil)

		storage.On("DeleteFile", "/uploads/video.mp4").Return(nil)
		storage.On("DeleteDir", "/tmp").Return(nil)

		repo.On("Update", ctx, mock.MatchedBy(func(v *domain.Video) bool {
			return v.Status == domain.StatusCompleted && v.FrameCount == 2 && v.ZipPath == "frames_video.zip"
		})).Return(nil)

		err := service.ProcessVideoByID(ctx, 1)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
		processor.AssertExpectations(t)
		storage.AssertExpectations(t)
	})

	t.Run("extraction failure", func(t *testing.T) {
		processor := new(MockVideoProcessor)
		storage := new(MockStorage)
		repo := new(MockVideoRepository)
		userRepo := new(MockUserRepository)
		emailer := new(MockEmailSender)
		service := NewWorkerService(processor, storage, repo, userRepo, emailer)

		video := &domain.Video{ID: 1, UserID: 10, Status: domain.StatusPending, Filename: "video.mp4"}
		user := &domain.User{ID: 10, Name: "Test User", Email: "test@example.com"}

		repo.On("GetByID", ctx, int64(1)).Return(video, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*domain.Video")).Return(nil)

		storage.On("GetUploadPath", "video.mp4").Return("/uploads/video.mp4")
		processor.On("ExtractFrames", "/uploads/video.mp4", "video").Return([]string{}, errors.New("ffmpeg error"))

		repo.On("Update", ctx, mock.MatchedBy(func(v *domain.Video) bool {
			return v.Status == domain.StatusFailed && assert.Contains(t, v.Message, "ffmpeg error")
		})).Return(nil)
		storage.On("DeleteFile", "/uploads/video.mp4").Return(nil)

		// Notification expectations
		userRepo.On("GetByID", int64(10)).Return(user, nil)
		emailer.On("SendEmail", "test@example.com", mock.Anything, mock.Anything).Return(nil)

		err := service.ProcessVideoByID(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, "ffmpeg error", err.Error())
		userRepo.AssertExpectations(t)
		emailer.AssertExpectations(t)
	})

	t.Run("zipping failure", func(t *testing.T) {
		processor := new(MockVideoProcessor)
		storage := new(MockStorage)
		repo := new(MockVideoRepository)
		userRepo := new(MockUserRepository)
		emailer := new(MockEmailSender)
		service := NewWorkerService(processor, storage, repo, userRepo, emailer)

		video := &domain.Video{ID: 1, UserID: 10, Status: domain.StatusPending, Filename: "video.mp4"}
		user := &domain.User{ID: 10, Name: "Test User", Email: "test@example.com"}

		repo.On("GetByID", ctx, int64(1)).Return(video, nil)
		repo.On("Update", ctx, mock.AnythingOfType("*domain.Video")).Return(nil)

		storage.On("GetUploadPath", "video.mp4").Return("/uploads/video.mp4")
		processor.On("ExtractFrames", "/uploads/video.mp4", "video").Return([]string{"/tmp/f1.jpg"}, nil)
		storage.On("SaveZip", "frames_video.zip", []string{"/tmp/f1.jpg"}).Return(errors.New("zip error"))

		repo.On("Update", ctx, mock.MatchedBy(func(v *domain.Video) bool {
			return v.Status == domain.StatusFailed && assert.Contains(t, v.Message, "zip error")
		})).Return(nil)
		storage.On("DeleteFile", "/uploads/video.mp4").Return(nil)

		// Notification expectations
		userRepo.On("GetByID", int64(10)).Return(user, nil)
		emailer.On("SendEmail", "test@example.com", mock.Anything, mock.Anything).Return(nil)

		err := service.ProcessVideoByID(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, "zip error", err.Error())
		userRepo.AssertExpectations(t)
		emailer.AssertExpectations(t)
	})
}

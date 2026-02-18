package services

import (
	"bytes"
	"errors"
	"testing"
	"video-processor/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVideoService_UploadAndProcess(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		storage := new(MockStorage)
		repo := new(MockVideoRepository)
		publisher := new(MockEventPublisher)
		service := NewVideoService(storage, repo, publisher)

		userID := int64(1)
		filename := "video.mp4"
		fileContent := []byte("fake video content")
		reader := bytes.NewReader(fileContent)

		storage.On("SaveUpload", mock.AnythingOfType("string"), mock.Anything).Return("/path/to/video.mp4", nil)
		repo.On("Create", mock.AnythingOfType("*domain.Video")).Return(nil).Run(func(args mock.Arguments) {
			video := args.Get(0).(*domain.Video)
			video.ID = 100
		})
		publisher.On("PublishUploadEvent", int64(100), mock.AnythingOfType("string")).Return(nil)

		resp, err := service.UploadAndProcess(userID, filename, reader)

		assert.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, int64(100), resp.VideoID)
		storage.AssertExpectations(t)
		repo.AssertExpectations(t)
		publisher.AssertExpectations(t)
	})

	t.Run("invalid file format", func(t *testing.T) {
		service := NewVideoService(nil, nil, nil)

		resp, err := service.UploadAndProcess(1, "test.txt", bytes.NewReader([]byte("txt")))

		assert.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "Formato de arquivo não suportado")
	})

	t.Run("storage error", func(t *testing.T) {
		storage := new(MockStorage)
		service := NewVideoService(storage, nil, nil)

		storage.On("SaveUpload", mock.AnythingOfType("string"), mock.Anything).Return("", errors.New("storage fail"))

		resp, err := service.UploadAndProcess(1, "video.mp4", bytes.NewReader([]byte("video")))

		assert.Error(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "storage fail", err.Error())
	})

	t.Run("repo error", func(t *testing.T) {
		storage := new(MockStorage)
		repo := new(MockVideoRepository)
		service := NewVideoService(storage, repo, nil)

		storage.On("SaveUpload", mock.AnythingOfType("string"), mock.Anything).Return("/path/to/video.mp4", nil)
		storage.On("DeleteFile", "/path/to/video.mp4").Return(nil)
		repo.On("Create", mock.AnythingOfType("*domain.Video")).Return(errors.New("db error"))

		resp, err := service.UploadAndProcess(1, "video.mp4", bytes.NewReader([]byte("video")))

		assert.Error(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "db error", err.Error())
	})
}

func TestVideoService_ListProcessedFiles(t *testing.T) {
	storage := new(MockStorage)
	service := NewVideoService(storage, nil, nil)

	expectedFiles := []domain.FileInfo{{Name: "file1.zip"}, {Name: "file2.zip"}}
	storage.On("ListOutputs").Return(expectedFiles, nil)

	files, err := service.ListProcessedFiles()

	assert.NoError(t, err)
	assert.Equal(t, expectedFiles, files)
}

func TestVideoService_GetVideosByUserID(t *testing.T) {
	repo := new(MockVideoRepository)
	service := NewVideoService(nil, repo, nil)

	userID := int64(1)
	expectedVideos := []domain.Video{{ID: 1, UserID: userID}, {ID: 2, UserID: userID}}
	repo.On("GetByUserID", userID).Return(expectedVideos, nil)

	videos, err := service.GetVideosByUserID(userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedVideos, videos)
}

func TestVideoService_IsValidVideoFile(t *testing.T) {
	service := &videoService{}

	tests := []struct {
		filename string
		want     bool
	}{
		{"video.mp4", true},
		{"VIDEO.MP4", true},
		{"movie.avi", true},
		{"clip.mov", true},
		{"test.mkv", true},
		{"document.pdf", false},
		{"image.jpg", false},
		{"archive.zip", false},
		{"noext", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.want, service.isValidVideoFile(tt.filename))
		})
	}
}

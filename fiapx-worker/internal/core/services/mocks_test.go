package services

import (
	"context"
	"io"
	"video-processor-worker/internal/core/domain"

	"github.com/stretchr/testify/mock"
)

type MockVideoProcessor struct {
	mock.Mock
}

func (m *MockVideoProcessor) ExtractFrames(videoPath string, timestamp string) ([]string, error) {
	args := m.Called(videoPath, timestamp)
	return args.Get(0).([]string), args.Error(1)
}

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SaveUpload(filename string, data io.Reader) (string, error) {
	args := m.Called(filename, data)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) SaveZip(zipFilename string, files []string) error {
	args := m.Called(zipFilename, files)
	return args.Error(0)
}

func (m *MockStorage) DeleteFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockStorage) DeleteDir(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockStorage) ListOutputs() ([]domain.FileInfo, error) {
	args := m.Called()
	return args.Get(0).([]domain.FileInfo), args.Error(1)
}

func (m *MockStorage) GetOutputPath(filename string) string {
	args := m.Called(filename)
	return args.String(0)
}

func (m *MockStorage) GetUploadPath(filename string) string {
	args := m.Called(filename)
	return args.String(0)
}

type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) Update(ctx context.Context, video *domain.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *MockVideoRepository) GetByID(ctx context.Context, id int64) (*domain.Video, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Video), args.Error(1)
}

func (m *MockVideoRepository) GetPending(ctx context.Context) ([]domain.Video, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Video), args.Error(1)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id int64) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

package ports

import (
	"io"
	"video-processor/internal/core/domain"
)

// VideoUseCase is the Inbound Port
type VideoUseCase interface {
	UploadAndProcess(userID int64, filename string, file io.Reader) (domain.ProcessingResult, error)
	ListProcessedFiles() ([]domain.FileInfo, error)
	GetVideosByUserID(userID int64) ([]domain.Video, error)
}

// Storage is the Outbound Port for file operations
type Storage interface {
	SaveUpload(filename string, data io.Reader) (string, error)
	SaveZip(zipFilename string, files []string) error
	DeleteFile(path string) error
	DeleteDir(path string) error
	ListOutputs() ([]domain.FileInfo, error)
	GetOutputPath(filename string) string
	GetUploadPath(filename string) string
}

// VideoRepository is the Outbound Port for video data persistence
type VideoRepository interface {
	Create(video *domain.Video) error
	Update(video *domain.Video) error
	GetByID(id int64) (*domain.Video, error)
	GetByUserID(userID int64) ([]domain.Video, error)
}

// UserUseCase is the Inbound Port for user logic
type UserUseCase interface {
	Register(email, password, name string) (domain.AuthResponse, error)
	Login(email, password string) (domain.AuthResponse, error)
}

// UserRepository is the Outbound Port for user data persistence
type UserRepository interface {
	Create(user *domain.User) error
	GetByEmail(email string) (*domain.User, error)
	GetByID(id int64) (*domain.User, error)
}

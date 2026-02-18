package domain

import "time"

const (
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
)

type Video struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Filename   string    `json:"filename"`
	Status     string    `json:"status"`
	ZipPath    string    `json:"zip_path,omitempty"`
	FrameCount int       `json:"frame_count"`
	Message    string    `json:"message,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ProcessingResult struct {
	Success    bool     `json:"success"`
	Message    string   `json:"message"`
	ErrorCode  string   `json:"error_code,omitempty"`
	VideoID    int64    `json:"video_id,omitempty"`
	ZipPath    string   `json:"zip_path,omitempty"`
	FrameCount int      `json:"frame_count,omitempty"`
	Images     []string `json:"images,omitempty"`
}

type FileInfo struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"filename"`
	Size        int64  `json:"size"`
	CreatedAt   string `json:"created_at"`
	DownloadURL string `json:"download_url"`
	Status      string `json:"status,omitempty"`
}

type ListVideosResponse struct {
	Success bool    `json:"success"`
	Videos  []Video `json:"videos"`
}

type FileListResponse struct {
	Files []FileInfo `json:"files"`
	Total int        `json:"total"`
}

type ErrorResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode string `json:"error_code,omitempty"`
}

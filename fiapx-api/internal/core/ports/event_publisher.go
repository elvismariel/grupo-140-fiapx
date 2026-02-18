package ports

type EventPublisher interface {
	PublishUploadEvent(videoID int64, filename string) error
}

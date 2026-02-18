package repository

import (
	"context"
	"video-processor/internal/core/domain"
	"video-processor/internal/core/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresVideoRepository struct {
	db *pgxpool.Pool
}

func NewPostgresVideoRepository(db *pgxpool.Pool) ports.VideoRepository {
	return &postgresVideoRepository{
		db: db,
	}
}

func (r *postgresVideoRepository) Create(video *domain.Video) error {
	query := `
		INSERT INTO videos (user_id, filename, status, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(context.Background(), query, video.UserID, video.Filename, video.Status).
		Scan(&video.ID, &video.CreatedAt, &video.UpdatedAt)
	return err
}

func (r *postgresVideoRepository) Update(video *domain.Video) error {
	query := `
		UPDATE videos
		SET status = $1, zip_path = $2, frame_count = $3, message = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	err := r.db.QueryRow(context.Background(), query, video.Status, video.ZipPath, video.FrameCount, video.Message, video.ID).
		Scan(&video.UpdatedAt)
	return err
}

func (r *postgresVideoRepository) GetByID(id int64) (*domain.Video, error) {
	query := `SELECT id, user_id, filename, status, zip_path, frame_count, message, created_at, updated_at FROM videos WHERE id = $1`
	video := &domain.Video{}
	err := r.db.QueryRow(context.Background(), query, id).
		Scan(&video.ID, &video.UserID, &video.Filename, &video.Status, &video.ZipPath, &video.FrameCount, &video.Message, &video.CreatedAt, &video.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return video, err
}

func (r *postgresVideoRepository) GetByUserID(userID int64) ([]domain.Video, error) {
	query := `SELECT id, user_id, filename, status, zip_path, frame_count, message, created_at, updated_at FROM videos WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []domain.Video
	for rows.Next() {
		var v domain.Video
		err := rows.Scan(&v.ID, &v.UserID, &v.Filename, &v.Status, &v.ZipPath, &v.FrameCount, &v.Message, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, nil
}

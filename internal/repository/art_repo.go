package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ArtPiece struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Medium     string    `json:"medium"`
	Dimensions string    `json:"dimensions"`
	ImageURL   string    `json:"image_url"`
	CreatedAt  time.Time `json:"created_at"`
}

type ArtRepository struct {
	DB *pgxpool.Pool
}

func NewArtRepository(db *pgxpool.Pool) *ArtRepository {
	return &ArtRepository{DB: db}
}

// GetGallery fetches all art records for your public front-page gallery
func (r *ArtRepository) GetGallery(ctx context.Context) ([]ArtPiece, error) {
	query := `SELECT id, title, medium, dimensions, image_url, created_at FROM art_pieces ORDER BY created_at DESC`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gallery []ArtPiece
	for rows.Next() {
		var art ArtPiece
		err := rows.Scan(&art.ID, &art.Title, &art.Medium, &art.Dimensions, &art.ImageURL, &art.CreatedAt)
		if err != nil {
			return nil, err
		}
		gallery = append(gallery, art)
	}
	return gallery, nil
}

// AddArt inserts a fresh painting into the database via your private route
func (r *ArtRepository) AddArt(ctx context.Context, title, medium, dims, imgURL string) error {
	query := `INSERT INTO art_pieces (title, medium, dimensions, image_url) VALUES ($1, $2, $3, $4)`
	_, err := r.DB.Exec(ctx, query, title, medium, dims, imgURL)
	return err
}

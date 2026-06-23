package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Commission represents a completed sold artwork with collector testimonial
type Commission struct {
	ID            int       `json:"id"`
	CustomerName  string    `json:"customer_name"`
	Review        string    `json:"review"`
	Price         float64   `json:"price"`
	ImageFilename string    `json:"image_filename"`
	ImageMimeType string    `json:"image_mime_type,omitempty"`
	ImageData     []byte    `json:"-"` // never serialize to JSON
	CreatedAt     time.Time `json:"created_at"`
}

type CommissionRepository struct {
	DB *pgxpool.Pool
}

func NewCommissionRepository(db *pgxpool.Pool) *CommissionRepository {
	return &CommissionRepository{DB: db}
}

// GetAll returns all commissions (metadata only, no image bytes)
func (r *CommissionRepository) GetAll(ctx context.Context) ([]Commission, error) {
	query := `SELECT id, customer_name, review, price, image_filename, created_at
	          FROM commissions ORDER BY created_at DESC`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commissions []Commission
	for rows.Next() {
		var c Commission
		if err := rows.Scan(&c.ID, &c.CustomerName, &c.Review, &c.Price, &c.ImageFilename, &c.CreatedAt); err != nil {
			return nil, err
		}
		commissions = append(commissions, c)
	}
	return commissions, nil
}

// GetImage returns the raw image bytes and MIME type for a commission
func (r *CommissionRepository) GetImage(ctx context.Context, id int) ([]byte, string, error) {
	var data []byte
	var mimeType string
	query := `SELECT image_data, image_mime_type FROM commissions WHERE id = $1`
	err := r.DB.QueryRow(ctx, query, id).Scan(&data, &mimeType)
	return data, mimeType, err
}

// Create inserts a new commission record
func (r *CommissionRepository) Create(ctx context.Context, customerName, review string, price float64, filename, mimeType string, data []byte) (int, error) {
	var id int
	query := `INSERT INTO commissions (customer_name, review, price, image_filename, image_mime_type, image_data)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := r.DB.QueryRow(ctx, query, customerName, review, price, filename, mimeType, data).Scan(&id)
	return id, err
}

// Delete removes a commission record
func (r *CommissionRepository) Delete(ctx context.Context, id int) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM commissions WHERE id = $1`, id)
	return err
}

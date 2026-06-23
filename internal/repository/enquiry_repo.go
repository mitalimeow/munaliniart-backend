package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Enquiry struct {
	ID          int       `json:"id"`
	SenderName  string    `json:"sender_name"`
	SenderEmail string    `json:"sender_email"`
	Message     string    `json:"message"`
	IsRead      bool      `json:"is_read"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type EnquiryRepository struct {
	DB *pgxpool.Pool
}

func NewEnquiryRepository(db *pgxpool.Pool) *EnquiryRepository {
	return &EnquiryRepository{DB: db}
}

func (r *EnquiryRepository) CreateEnquiry(ctx context.Context, name, email, message string) error {
	query := `INSERT INTO enquiries (sender_name, sender_email, message) VALUES ($1, $2, $3)`
	_, err := r.DB.Exec(ctx, query, name, email, message)
	return err
}

func (r *EnquiryRepository) GetAll(ctx context.Context) ([]Enquiry, error) {
	query := `SELECT id, sender_name, sender_email, message, is_read, submitted_at FROM enquiries ORDER BY submitted_at DESC`
	rows, err := r.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Enquiry
	for rows.Next() {
		var eq Enquiry
		err := rows.Scan(&eq.ID, &eq.SenderName, &eq.SenderEmail, &eq.Message, &eq.IsRead, &eq.SubmittedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, eq)
	}
	return logs, nil
}

func (r *EnquiryRepository) DeleteEnquiry(ctx context.Context, id int) error {
	query := `DELETE FROM enquiries WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	return err
}

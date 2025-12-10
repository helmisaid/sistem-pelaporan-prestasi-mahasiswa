package repository

import (
	"context"
	"database/sql"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
)

type ILecturerRepository interface {
	Create(ctx context.Context, tx *sql.Tx, userID, lecturerID, department string) error
	Update(ctx context.Context, tx *sql.Tx, userID string, lecturerID, department *string) error
	GetByUserID(ctx context.Context, userID string) (*model.LecturerInfo, error)
	Delete(ctx context.Context, tx *sql.Tx, userID string) error
	CheckLecturerIDExists(ctx context.Context, lecturerID string, excludeUserID *string) (bool, error)
}

type lecturerRepository struct {
	db *sql.DB
}

func NewLecturerRepository(db *sql.DB) ILecturerRepository {
	return &lecturerRepository{db: db}
}

func (r *lecturerRepository) Create(ctx context.Context, tx *sql.Tx, userID, lecturerID, department string) error {
	query := `
		INSERT INTO lecturers (user_id, lecturer_id, department)
		VALUES ($1, $2, $3)
	`
	
	var executor interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	
	if tx != nil {
		executor = tx
	} else {
		executor = r.db
	}
	
	_, err := executor.ExecContext(ctx, query, userID, lecturerID, department)
	return err
}

func (r *lecturerRepository) Update(ctx context.Context, tx *sql.Tx, userID string, lecturerID, department *string) error {
	query := `UPDATE lecturers SET department = COALESCE($1, department), lecturer_id = COALESCE($2, lecturer_id) WHERE user_id = $3`
	
	var executor interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	
	if tx != nil {
		executor = tx
	} else {
		executor = r.db
	}
	
	_, err := executor.ExecContext(ctx, query, department, lecturerID, userID)
	return err
}

func (r *lecturerRepository) GetByUserID(ctx context.Context, userID string) (*model.LecturerInfo, error) {
	query := `SELECT id, lecturer_id, department FROM lecturers WHERE user_id = $1`
	
	var info model.LecturerInfo
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&info.ID, &info.LecturerID, &info.Department)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &info, nil
}

func (r *lecturerRepository) Delete(ctx context.Context, tx *sql.Tx, userID string) error {
	query := `DELETE FROM lecturers WHERE user_id = $1`
	
	var executor interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	
	if tx != nil {
		executor = tx
	} else {
		executor = r.db
	}
	
	_, err := executor.ExecContext(ctx, query, userID)
	return err
}

func (r *lecturerRepository) CheckLecturerIDExists(ctx context.Context, lecturerID string, excludeUserID *string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM lecturers WHERE lecturer_id = $1 AND ($2::uuid IS NULL OR user_id != $2))`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, lecturerID, excludeUserID).Scan(&exists)
	return exists, err
}

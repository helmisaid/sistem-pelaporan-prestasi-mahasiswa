package repository

import (
	"context"
	"database/sql"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
)

type IStudentRepository interface {
	Create(ctx context.Context, tx *sql.Tx, userID string, studentID, programStudy, academicYear string, advisorID *string) error
	Update(ctx context.Context, tx *sql.Tx, userID string, studentID, programStudy, academicYear *string, advisorID *string) error
	GetByUserID(ctx context.Context, userID string) (*model.StudentInfo, error)
	Delete(ctx context.Context, tx *sql.Tx, userID string) error
	CheckStudentIDExists(ctx context.Context, studentID string, excludeUserID *string) (bool, error)
}

type studentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) IStudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) Create(ctx context.Context, tx *sql.Tx, userID string, studentID, programStudy, academicYear string, advisorID *string) error {
	query := `
		INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	var executor interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	
	if tx != nil {
		executor = tx
	} else {
		executor = r.db
	}
	
	_, err := executor.ExecContext(ctx, query, userID, studentID, programStudy, academicYear, advisorID)
	return err
}

func (r *studentRepository) Update(ctx context.Context, tx *sql.Tx, userID string, studentID, programStudy, academicYear *string, advisorID *string) error {
	query := `
		UPDATE students 
		SET program_study = COALESCE($1, program_study),
		    academic_year = COALESCE($2, academic_year),
		    advisor_id = $3,
		    student_id = COALESCE($4, student_id)
		WHERE user_id = $5
	`
	
	var executor interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	}
	
	if tx != nil {
		executor = tx
	} else {
		executor = r.db
	}
	
	_, err := executor.ExecContext(ctx, query, programStudy, academicYear, advisorID, studentID, userID)
	return err
}

func (r *studentRepository) GetByUserID(ctx context.Context, userID string) (*model.StudentInfo, error) {
	query := `
		SELECT s.student_id, s.program_study, s.academic_year, s.advisor_id,
		       l.lecturer_id as advisor_name
		FROM students s
		LEFT JOIN lecturers l ON s.advisor_id = l.id
		WHERE s.user_id = $1
	`
	
	var info model.StudentInfo
	var advisorID, advisorName sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&info.StudentID, &info.ProgramStudy, &info.AcademicYear,
		&advisorID, &advisorName,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	if advisorID.Valid {
		info.AdvisorID = &advisorID.String
	}
	if advisorName.Valid {
		info.AdvisorName = &advisorName.String
	}
	
	return &info, nil
}

func (r *studentRepository) Delete(ctx context.Context, tx *sql.Tx, userID string) error {
	query := `DELETE FROM students WHERE user_id = $1`
	
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

func (r *studentRepository) CheckStudentIDExists(ctx context.Context, studentID string, excludeUserID *string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM students WHERE student_id = $1 AND ($2::uuid IS NULL OR user_id != $2))`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, studentID, excludeUserID).Scan(&exists)
	return exists, err
}

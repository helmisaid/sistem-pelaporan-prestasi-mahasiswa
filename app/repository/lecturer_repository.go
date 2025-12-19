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
	CheckExistsByID(ctx context.Context, id string) (bool, error)
	GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) ([]model.LecturerListDTO, int64, error)
	GetAdvisees(ctx context.Context, lecturerID string, page, pageSize int) ([]model.StudentListDTO, int64, error)
}

type lecturerRepository struct {
	db *sql.DB
}

func NewLecturerRepository(db *sql.DB) ILecturerRepository {
	return &lecturerRepository{db: db}
}

// CreateLecturer
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

// UpdateLecturer
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

// GetLecturerByUserID
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

// DeleteLecturer
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

// CheckLecturerIDExists
func (r *lecturerRepository) CheckLecturerIDExists(ctx context.Context, lecturerID string, excludeUserID *string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM lecturers WHERE lecturer_id = $1 AND ($2::uuid IS NULL OR user_id != $2))`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, lecturerID, excludeUserID).Scan(&exists)
	return exists, err
}

// CheckExistsByID
func (r *lecturerRepository) CheckExistsByID(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM lecturers WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

// GetAllLecturer
func (r *lecturerRepository) GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) ([]model.LecturerListDTO, int64, error) {
	offset := (page - 1) * pageSize

	countQuery := `
		SELECT COUNT(*)
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE u.is_active = true
		  AND (u.full_name ILIKE $1 OR l.lecturer_id ILIKE $1)
	`

	var total int64
	searchPattern := "%" + search + "%"
	err := r.db.QueryRowContext(ctx, countQuery, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT 
			u.id, l.lecturer_id, u.full_name, u.email,
			l.department, u.is_active
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE u.is_active = true
		  AND (u.full_name ILIKE $1 OR l.lecturer_id ILIKE $1)
		ORDER BY ` + sortBy + ` ` + sortOrder + `
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, dataQuery, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lecturers []model.LecturerListDTO
	for rows.Next() {
		var lecturer model.LecturerListDTO
		err := rows.Scan(
			&lecturer.ID, &lecturer.LecturerID, &lecturer.FullName, &lecturer.Email,
			&lecturer.Department, &lecturer.IsActive,
		)
		if err != nil {
			return nil, 0, err
		}
		lecturers = append(lecturers, lecturer)
	}

	return lecturers, total, rows.Err()
}

// GetAdvisees
func (r *lecturerRepository) GetAdvisees(ctx context.Context, lecturerID string, page, pageSize int) ([]model.StudentListDTO, int64, error) {
	offset := (page - 1) * pageSize

	countQuery := `
		SELECT COUNT(*)
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE s.advisor_id = $1 AND u.is_active = true
	`

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, lecturerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT 
			u.id, s.student_id, u.full_name, u.email,
			s.program_study, s.academic_year, s.advisor_id,
			u_advisor.full_name as advisor_name, u.is_active
		FROM students s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN lecturers l ON s.advisor_id = l.id
		LEFT JOIN users u_advisor ON l.user_id = u_advisor.id
		WHERE s.advisor_id = $1 AND u.is_active = true
		ORDER BY u.full_name ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, dataQuery, lecturerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.StudentListDTO
	for rows.Next() {
		var student model.StudentListDTO
		var advisorID, advisorName sql.NullString

		err := rows.Scan(
			&student.ID, &student.StudentID, &student.FullName, &student.Email,
			&student.ProgramStudy, &student.AcademicYear, &advisorID,
			&advisorName, &student.IsActive,
		)
		if err != nil {
			return nil, 0, err
		}

		if advisorID.Valid {
			student.AdvisorID = &advisorID.String
		}
		if advisorName.Valid {
			student.AdvisorName = &advisorName.String
		}

		students = append(students, student)
	}

	return students, total, rows.Err()
}

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
	GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) ([]model.StudentListDTO, int64, error)
	GetDetailByID(ctx context.Context, studentID string) (*model.StudentDetailDTO, error)
	UpdateAdvisor(ctx context.Context, studentID string, advisorID *string) error
}

type studentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) IStudentRepository {
	return &studentRepository{db: db}
}

// CreateStudent
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

// UpdateStudent
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

// GetByUserID
func (r *studentRepository) GetByUserID(ctx context.Context, userID string) (*model.StudentInfo, error) {
	query := `
		SELECT s.id, s.student_id, s.program_study, s.academic_year, s.advisor_id,
		       l.lecturer_id as advisor_name
		FROM students s
		LEFT JOIN lecturers l ON s.advisor_id = l.id
		WHERE s.user_id = $1
	`
	
	var info model.StudentInfo
	var advisorID, advisorName sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&info.ID, &info.StudentID, &info.ProgramStudy, &info.AcademicYear,
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

// DeleteStudent
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

// CheckStudentIDExists
func (r *studentRepository) CheckStudentIDExists(ctx context.Context, studentID string, excludeUserID *string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM students WHERE student_id = $1 AND ($2::uuid IS NULL OR user_id != $2))`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, studentID, excludeUserID).Scan(&exists)
	return exists, err
}

// GetAllStudent
func (r *studentRepository) GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) ([]model.StudentListDTO, int64, error) {
	offset := (page - 1) * pageSize

	countQuery := `
		SELECT COUNT(*)
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE u.is_active = true
		  AND (u.full_name ILIKE $1 OR s.student_id ILIKE $1)
	`

	var total int64
	searchPattern := "%" + search + "%"
	err := r.db.QueryRowContext(ctx, countQuery, searchPattern).Scan(&total)
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
		WHERE u.is_active = true
		  AND (u.full_name ILIKE $1 OR s.student_id ILIKE $1)
		ORDER BY ` + sortBy + ` ` + sortOrder + `
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, dataQuery, searchPattern, pageSize, offset)
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

func (r *studentRepository) GetDetailByID(ctx context.Context, id string) (*model.StudentDetailDTO, error) {
	query := `
		SELECT 
			u.id, u.username, u.email, u.full_name,
			s.student_id, s.program_study, s.academic_year,
			s.advisor_id, u_advisor.full_name as advisor_name,
			u.is_active, u.created_at, u.updated_at
		FROM users u
		JOIN students s ON u.id = s.user_id
		LEFT JOIN lecturers l ON s.advisor_id = l.id
		LEFT JOIN users u_advisor ON l.user_id = u_advisor.id
		WHERE s.id = $1 AND u.is_active = true
	`

	var detail model.StudentDetailDTO
	var advisorID, advisorName sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&detail.ID, &detail.Username, &detail.Email, &detail.FullName,
		&detail.StudentID, &detail.ProgramStudy, &detail.AcademicYear,
		&advisorID, &advisorName,
		&detail.IsActive, &detail.CreatedAt, &detail.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if advisorID.Valid {
		detail.AdvisorID = &advisorID.String
	}
	if advisorName.Valid {
		detail.AdvisorName = &advisorName.String
	}

	return &detail, nil
}

func (r *studentRepository) UpdateAdvisor(ctx context.Context, studentID string, advisorID *string) error {
	query := `UPDATE students SET advisor_id = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, advisorID, studentID)
	return err
}

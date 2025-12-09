package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sistem-pelaporan-prestasi-mahasiswa/app/model"

	"github.com/lib/pq"
)

type IUserRepository interface {
	GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) ([]model.User, int64, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, id string, user *model.User) error
	Delete(ctx context.Context, id string) error
	UpdateRole(ctx context.Context, userID, roleID string) error
	CheckUsernameExists(ctx context.Context, username string, excludeUserID *string) (bool, error)
	CheckEmailExists(ctx context.Context, email string, excludeUserID *string) (bool, error)
	CheckRoleExists(ctx context.Context, roleID string) (bool, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) IUserRepository {
	return &userRepository{db: db}
}

// GetAllUsers
func (r *userRepository) GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) ([]model.User, int64, error) {
	offset := (page - 1) * pageSize

	baseQuery := `
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.is_active = true
	`

	var args []interface{}
	argCounter := 1

	if search != "" {
		baseQuery += fmt.Sprintf(" AND (u.username ILIKE $%d OR u.email ILIKE $%d OR u.full_name ILIKE $%d)", argCounter, argCounter, argCounter)
		args = append(args, "%"+search+"%")
		argCounter++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	selectQuery := `
		SELECT 
			u.id, u.username, u.email, u.password_hash, u.full_name, 
			u.role_id, u.is_active, u.created_at, u.updated_at,
			r.id, r.name,
			COALESCE(
				(SELECT array_agg(p.name) 
				 FROM role_permissions rp 
				 JOIN permissions p ON rp.permission_id = p.id 
				 WHERE rp.role_id = r.id), 
				'{}'
			) as permissions
	` + baseQuery

	validCols := map[string]string{
		"created_at": "u.created_at",
		"username":   "u.username",
		"full_name":  "u.full_name",
		"email":      "u.email",
	}

	dbCol, ok := validCols[sortBy]
	if !ok {
		dbCol = "u.created_at"
	}

	selectQuery += fmt.Sprintf(" ORDER BY %s %s", dbCol, sortOrder)

	selectQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
			&user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
			&user.Role.ID, &user.Role.Name,
			pq.Array(&user.Permissions),
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

// GetUserByID 
func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT 
			u.id, u.username, u.email, u.password_hash, u.full_name, 
			u.role_id, u.is_active, u.created_at, u.updated_at,
			r.id, r.name,
			COALESCE(
				(SELECT array_agg(p.name) 
				 FROM role_permissions rp 
				 JOIN permissions p ON rp.permission_id = p.id 
				 WHERE rp.role_id = r.id), 
				'{}'
			) as permissions
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`
	
	var user model.User
	row := r.db.QueryRowContext(ctx, query, id)
	
	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&user.Role.ID, &user.Role.Name,
		pq.Array(&user.Permissions),
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &user, nil
}

// Create 
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRowContext(
		ctx, query,
		user.Username, user.Email, user.PasswordHash, user.FullName, user.RoleID, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	return err
}

// Update 
func (r *userRepository) Update(ctx context.Context, id string, user *model.User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, full_name = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`
	
	_, err := r.db.ExecContext(ctx, query, user.Username, user.Email, user.FullName, user.IsActive, id)
	return err
}

// Soft Delete 
func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// UpdateRole 
func (r *userRepository) UpdateRole(ctx context.Context, userID, roleID string) error {
	query := `UPDATE users SET role_id = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, roleID, userID)
	return err
}

// CheckUsernameExists 
func (r *userRepository) CheckUsernameExists(ctx context.Context, username string, excludeUserID *string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND ($2::uuid IS NULL OR id != $2))`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, username, excludeUserID).Scan(&exists)
	return exists, err
}

// CheckEmailExists 
func (r *userRepository) CheckEmailExists(ctx context.Context, email string, excludeUserID *string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND ($2::uuid IS NULL OR id != $2))`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email, excludeUserID).Scan(&exists)
	return exists, err
}

// CheckRoleExists
func (r *userRepository) CheckRoleExists(ctx context.Context, roleID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, roleID).Scan(&exists)
	return exists, err
}

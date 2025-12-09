package repository

import (
	"context"
	"database/sql"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"

	"github.com/lib/pq"
)

type IAuthRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error) 
}

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) IAuthRepository {
	return &authRepository{db: db}
}

// GetUserByUsername
func (r *authRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {

	query := `
		SELECT 
			u.id, 
			u.username, 
			u.email, 
			u.password_hash, 
			u.full_name, 
			u.role_id, 
			u.is_active,
			u.created_at,
			u.updated_at,
			r.id,
			r.name,
			COALESCE(
				(SELECT array_agg(p.name) 
				 FROM role_permissions rp 
				 JOIN permissions p ON rp.permission_id = p.id 
				 WHERE rp.role_id = r.id), 
				'{}'
			) as permissions
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1 OR u.email = $1
	`
	
	var user model.User

	row := r.db.QueryRowContext(ctx, query, username)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,   
		&user.Role.Name,
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

// GetUserByID
func (r *authRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
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
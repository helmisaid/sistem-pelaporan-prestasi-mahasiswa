package service

import (
	"context"
	"database/sql"
	"math"
	"strings"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
	"sistem-pelaporan-prestasi-mahasiswa/utils"

	"github.com/google/uuid"
)

type IUserService interface {
	GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) (*model.PaginatedUsers, error)
	GetByID(ctx context.Context, id string) (*model.UserDetailDTO, error)
	Create(ctx context.Context, req model.CreateUserRequest) (*model.UserDetailDTO, error)
	Update(ctx context.Context, id string, req model.UpdateUserRequest) (*model.UserDetailDTO, error)
	Delete(ctx context.Context, id string) error
	UpdateRole(ctx context.Context, id string, req model.UpdateRoleRequest) error
}

type UserService struct {
	userRepo    repository.IUserRepository
	studentSvc  IStudentService
	lecturerSvc ILecturerService
	db          *sql.DB
}

func NewUserService(
	userRepo repository.IUserRepository,
	studentSvc IStudentService,
	lecturerSvc ILecturerService,
	db *sql.DB,
) IUserService {
	return &UserService{
		userRepo:    userRepo,
		studentSvc:  studentSvc,
		lecturerSvc: lecturerSvc,
		db:          db,
	}
}

// GetAllUsers
func (s *UserService) GetAll(ctx context.Context, page, pageSize int, search, sortBy, sortOrder string) (*model.PaginatedUsers, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	validSortColumns := map[string]bool{
		"username":   true,
		"email":      true,
		"full_name":  true,
		"created_at": true,
	}

	if !validSortColumns[sortBy] {
		sortBy = "created_at"
	}

	sortOrder = strings.ToUpper(sortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	users, total, err := s.userRepo.GetAll(ctx, page, pageSize, search, sortBy, sortOrder)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	userDTOs := make([]model.UserListDTO, len(users))
	for i, user := range users {
		userDTOs[i] = user.ToListDTO()
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &model.PaginatedUsers{
		Data:       userDTOs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserDetailByID 
func (s *UserService) GetByID(ctx context.Context, id string) (*model.UserDetailDTO, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if user == nil {
		return nil, model.ErrUserNotFound
	}

	dto := user.ToDetailDTO()

	// Get student
	if user.Role.Name == "Mahasiswa" {
		studentInfo, err := s.studentSvc.GetProfile(ctx, id)
		if err == nil && studentInfo != nil {
			dto.Student = studentInfo
		}
	}

	// Get lecturer 
	if user.Role.Name == "Dosen Wali" {
		lecturerInfo, err := s.lecturerSvc.GetProfile(ctx, id)
		if err == nil && lecturerInfo != nil {
			dto.Lecturer = lecturerInfo
		}
	}

	return &dto, nil
}

// Create 
func (s *UserService) Create(ctx context.Context, req model.CreateUserRequest) (*model.UserDetailDTO, error) {
	exists, err := s.userRepo.CheckUsernameExists(ctx, req.Username, nil)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if exists {
		return nil, model.NewValidationError("username sudah digunakan")
	}

	exists, err = s.userRepo.CheckEmailExists(ctx, req.Email, nil)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if exists {
		return nil, model.NewValidationError("email sudah digunakan")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, model.ErrTokenGenerationFailed
	}

	// Validate and parse role ID
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, model.NewValidationError("format Role ID tidak valid")
	}

	// Check if role exists
	roleExists, err := s.userRepo.CheckRoleExists(ctx, req.RoleID)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if !roleExists {
		return nil, model.NewValidationError("role tidak ditemukan")
	}

	// Get role details to check role name
	role, err := s.userRepo.GetRoleByID(ctx, req.RoleID)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if role == nil {
		return nil, model.NewValidationError("role tidak ditemukan")
	}

	// Validate required fields based on role
	switch role.Name {
	case "Mahasiswa":
		if req.StudentID == nil || *req.StudentID == "" {
			return nil, model.NewValidationError("student_id wajib diisi untuk role Mahasiswa")
		}
		if req.ProgramStudy == nil || *req.ProgramStudy == "" {
			return nil, model.NewValidationError("program_study wajib diisi untuk role Mahasiswa")
		}
		if req.AcademicYear == nil || *req.AcademicYear == "" {
			return nil, model.NewValidationError("academic_year wajib diisi untuk role Mahasiswa")
		}

	case "Dosen Wali":
		if req.LecturerID == nil || *req.LecturerID == "" {
			return nil, model.NewValidationError("lecturer_id wajib diisi untuk role Dosen Wali")
		}
		if req.Department == nil || *req.Department == "" {
			return nil, model.NewValidationError("department wajib diisi untuk role Dosen Wali")
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	defer tx.Rollback()

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		RoleID:       roleID,
		IsActive:     true,
	}

	err = s.userRepo.Create(ctx, tx, user)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	// Create student profile if Mahasiswa role
	if req.StudentID != nil && *req.StudentID != "" {
		if req.AdvisorID != nil && *req.AdvisorID != "" {
			exists, err := s.lecturerSvc.CheckExistsByID(ctx, *req.AdvisorID)
			if err != nil {
				return nil, model.ErrDatabaseError
			}
			if !exists {
				return nil, model.NewValidationError("Dosen wali dengan ID tersebut tidak ditemukan")
			}
		}

		studentReq := model.CreateStudentProfileRequest{
			StudentID:    *req.StudentID,
			ProgramStudy: utils.GetStringOrDefault(req.ProgramStudy, ""),
			AcademicYear: utils.GetStringOrDefault(req.AcademicYear, ""),
			AdvisorID:    req.AdvisorID,
		}
		err = s.studentSvc.CreateProfile(ctx, tx, user.ID.String(), studentReq)
		if err != nil {
			return nil, err
		}
	}

	// Create lecturer profile if Dosen Wali role
	if req.LecturerID != nil && *req.LecturerID != "" {
		lecturerReq := model.CreateLecturerProfileRequest{
			LecturerID: *req.LecturerID,
			Department: utils.GetStringOrDefault(req.Department, ""),
		}
		err = s.lecturerSvc.CreateProfile(ctx, tx, user.ID.String(), lecturerReq)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, model.ErrDatabaseError
	}

	return s.GetByID(ctx, user.ID.String())
}

// Update 
func (s *UserService) Update(ctx context.Context, id string, req model.UpdateUserRequest) (*model.UserDetailDTO, error) {
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if existingUser == nil {
		return nil, model.ErrUserNotFound
	}

	// Validate username if changed
	if req.Username != nil && *req.Username != existingUser.Username {
		exists, err := s.userRepo.CheckUsernameExists(ctx, *req.Username, &id)
		if err != nil {
			return nil, model.ErrDatabaseError
		}
		if exists {
			return nil, model.NewValidationError("username sudah digunakan")
		}
	}

	// Validate email if changed
	if req.Email != nil && *req.Email != existingUser.Email {
		exists, err := s.userRepo.CheckEmailExists(ctx, *req.Email, &id)
		if err != nil {
			return nil, model.ErrDatabaseError
		}
		if exists {
			return nil, model.NewValidationError("email sudah digunakan")
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	defer tx.Rollback()

	updatedUser := &model.User{
		Username: utils.GetStringOrDefault(req.Username, existingUser.Username),
		Email:    utils.GetStringOrDefault(req.Email, existingUser.Email),
		FullName: utils.GetStringOrDefault(req.FullName, existingUser.FullName),
		IsActive: utils.GetBoolOrDefault(req.IsActive, existingUser.IsActive),
	}

	err = s.userRepo.Update(ctx, id, updatedUser)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	// Update student profile if Mahasiswa role
	if existingUser.Role.Name == "Mahasiswa" {
		if req.StudentID != nil || req.ProgramStudy != nil || req.AcademicYear != nil || req.AdvisorID != nil {
			updateReq := model.UpdateStudentProfileRequest{
				StudentID:    req.StudentID,
				ProgramStudy: req.ProgramStudy,
				AcademicYear: req.AcademicYear,
				AdvisorID:    req.AdvisorID,
			}
			err = s.studentSvc.UpdateProfile(ctx, tx, id, updateReq)
			if err != nil {
				return nil, err
			}
		}
	}

	// Update lecturer profile if Dosen Wali role
	if existingUser.Role.Name == "Dosen Wali" {
		if req.LecturerID != nil || req.Department != nil {
			updateReq := model.UpdateLecturerProfileRequest{
				LecturerID: req.LecturerID,
				Department: req.Department,
			}
			err = s.lecturerSvc.UpdateProfile(ctx, tx, id, updateReq)
			if err != nil {
				return nil, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, model.ErrDatabaseError
	}

	return s.GetByID(ctx, id)
}

// Soft deletes 
func (s *UserService) Delete(ctx context.Context, id string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return model.ErrDatabaseError
	}
	if user == nil {
		return model.ErrUserNotFound
	}

	err = s.userRepo.Delete(ctx, id)
	if err != nil {
		return model.ErrDatabaseError
	}

	return nil
}

// UpdateRole 
func (s *UserService) UpdateRole(ctx context.Context, id string, req model.UpdateRoleRequest) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return model.ErrDatabaseError
	}
	if user == nil {
		return model.ErrUserNotFound
	}

	err = s.userRepo.UpdateRole(ctx, id, req.RoleID)
	if err != nil {
		return model.ErrDatabaseError
	}

	return nil
}


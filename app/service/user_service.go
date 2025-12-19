package service

import (
	"database/sql"
	"math"
	"strings"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
	"sistem-pelaporan-prestasi-mahasiswa/helper"
	"sistem-pelaporan-prestasi-mahasiswa/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type IUserService interface {
	GetAll(c *fiber.Ctx) error
	GetByID(c *fiber.Ctx) error
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	UpdateRole(c *fiber.Ctx) error
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

// GetAll godoc
// @Summary List all users
// @Description Get paginated list of users with optional filtering
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by name, email, or username"
// @Param sort_by query string false "Sort by field" Enums(created_at, full_name, email)
// @Param sort_order query string false "Sort order" Enums(asc, desc)
// @Success 200 {object} helper.Response{data=model.PaginatedUsers} "Users retrieved"
// @Failure 401 {object} helper.ErrorResponse "Unauthorized"
// @Router /users [get]
func (s *UserService) GetAll(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	sortBy := c.Query("sortBy", "created_at")
	sortOrder := c.Query("order", "desc")

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

	users, total, err := s.userRepo.GetAll(c.Context(), page, pageSize, search, sortBy, sortOrder)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	userDTOs := make([]model.UserListDTO, len(users))
	for i, user := range users {
		userDTOs[i] = user.ToListDTO()
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	result := &model.PaginatedUsers{
		Data:       userDTOs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	return helper.Success(c, "Daftar user berhasil diambil", result)
}

// GetByID godoc
// @Summary Get user by ID
// @Description Get detailed information about a specific user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} helper.Response{data=model.UserDetailDTO} "User details retrieved"
// @Failure 400 {object} helper.ErrorResponse "Invalid ID format"
// @Failure 404 {object} helper.ErrorResponse "User not found"
// @Router /users/{id} [get]
func (s *UserService) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := s.userRepo.GetByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if user == nil {
		return helper.HandleError(c, model.ErrUserNotFound)
	}

	dto := user.ToDetailDTO()

	if user.Role.Name == "Mahasiswa" {
		studentInfo, err := s.studentSvc.GetProfile(c.Context(), id)
		if err == nil && studentInfo != nil {
			dto.Student = studentInfo
		}
	}

	if user.Role.Name == "Dosen Wali" {
		lecturerInfo, err := s.lecturerSvc.GetProfile(c.Context(), id)
		if err == nil && lecturerInfo != nil {
			dto.Lecturer = lecturerInfo
		}
	}

	return helper.Success(c, "Detail user berhasil diambil", dto)
}

// Create godoc
// @Summary Create new user
// @Description Create a new user with role-specific profile (Student or Lecturer)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.CreateUserRequest true "User creation data"
// @Success 201 {object} helper.Response{data=model.UserDetailDTO} "User created successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request or validation error"
// @Failure 409 {object} helper.ErrorResponse "Username or email already exists"
// @Router /users [post]
func (s *UserService) Create(c *fiber.Ctx) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Format request tidak valid", nil)
	}

	exists, err := s.userRepo.CheckUsernameExists(c.Context(), req.Username, nil)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if exists {
		return helper.HandleError(c, model.NewValidationError("username sudah digunakan"))
	}

	exists, err = s.userRepo.CheckEmailExists(c.Context(), req.Email, nil)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if exists {
		return helper.HandleError(c, model.NewValidationError("email sudah digunakan"))
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return helper.HandleError(c, model.ErrTokenGenerationFailed)
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return helper.HandleError(c, model.NewValidationError("format Role ID tidak valid"))
	}

	roleExists, err := s.userRepo.CheckRoleExists(c.Context(), req.RoleID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if !roleExists {
		return helper.HandleError(c, model.NewValidationError("role tidak ditemukan"))
	}

	role, err := s.userRepo.GetRoleByID(c.Context(), req.RoleID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if role == nil {
		return helper.HandleError(c, model.NewValidationError("role tidak ditemukan"))
	}

	switch role.Name {
	case "Mahasiswa":
		if req.StudentID == nil || *req.StudentID == "" {
			return helper.HandleError(c, model.NewValidationError("student_id wajib diisi untuk role Mahasiswa"))
		}
		if req.ProgramStudy == nil || *req.ProgramStudy == "" {
			return helper.HandleError(c, model.NewValidationError("program_study wajib diisi untuk role Mahasiswa"))
		}
		if req.AcademicYear == nil || *req.AcademicYear == "" {
			return helper.HandleError(c, model.NewValidationError("academic_year wajib diisi untuk role Mahasiswa"))
		}

	case "Dosen Wali":
		if req.LecturerID == nil || *req.LecturerID == "" {
			return helper.HandleError(c, model.NewValidationError("lecturer_id wajib diisi untuk role Dosen Wali"))
		}
		if req.Department == nil || *req.Department == "" {
			return helper.HandleError(c, model.NewValidationError("department wajib diisi untuk role Dosen Wali"))
		}
	}

	tx, err := s.db.BeginTx(c.Context(), nil)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
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

	err = s.userRepo.Create(c.Context(), tx, user)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	if req.StudentID != nil && *req.StudentID != "" {
		if req.AdvisorID != nil && *req.AdvisorID != "" {
			exists, err := s.lecturerSvc.CheckExistsByID(c.Context(), *req.AdvisorID)
			if err != nil {
				return helper.HandleError(c, model.ErrDatabaseError)
			}
			if !exists {
				return helper.HandleError(c, model.NewValidationError("Dosen wali dengan ID tersebut tidak ditemukan"))
			}
		}

		studentReq := model.CreateStudentProfileRequest{
			StudentID:    *req.StudentID,
			ProgramStudy: utils.GetStringOrDefault(req.ProgramStudy, ""),
			AcademicYear: utils.GetStringOrDefault(req.AcademicYear, ""),
			AdvisorID:    req.AdvisorID,
		}
		err = s.studentSvc.CreateProfile(c.Context(), tx, user.ID.String(), studentReq)
		if err != nil {
			return helper.HandleError(c, err)
		}
	}

	if req.LecturerID != nil && *req.LecturerID != "" {
		lecturerReq := model.CreateLecturerProfileRequest{
			LecturerID: *req.LecturerID,
			Department: utils.GetStringOrDefault(req.Department, ""),
		}
		err = s.lecturerSvc.CreateProfile(c.Context(), tx, user.ID.String(), lecturerReq)
		if err != nil {
			return helper.HandleError(c, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	createdUser, err := s.userRepo.GetByID(c.Context(), user.ID.String())
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	dto := createdUser.ToDetailDTO()

	if createdUser.Role.Name == "Mahasiswa" {
		studentInfo, err := s.studentSvc.GetProfile(c.Context(), user.ID.String())
		if err == nil && studentInfo != nil {
			dto.Student = studentInfo
		}
	}

	if createdUser.Role.Name == "Dosen Wali" {
		lecturerInfo, err := s.lecturerSvc.GetProfile(c.Context(), user.ID.String())
		if err == nil && lecturerInfo != nil {
			dto.Lecturer = lecturerInfo
		}
	}

	return helper.Created(c, "User berhasil dibuat", dto)
}

// Update godoc
// @Summary Update user
// @Description Update user information and role-specific profile
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body model.UpdateUserRequest true "User update data"
// @Success 200 {object} helper.Response{data=model.UserDetailDTO} "User updated successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request"
// @Failure 404 {object} helper.ErrorResponse "User not found"
// @Router /users/{id} [put]
func (s *UserService) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Format request tidak valid", nil)
	}

	existingUser, err := s.userRepo.GetByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if existingUser == nil {
		return helper.HandleError(c, model.ErrUserNotFound)
	}

	if req.Username != nil && *req.Username != existingUser.Username {
		exists, err := s.userRepo.CheckUsernameExists(c.Context(), *req.Username, &id)
		if err != nil {
			return helper.HandleError(c, model.ErrDatabaseError)
		}
		if exists {
			return helper.HandleError(c, model.NewValidationError("username sudah digunakan"))
		}
	}

	if req.Email != nil && *req.Email != existingUser.Email {
		exists, err := s.userRepo.CheckEmailExists(c.Context(), *req.Email, &id)
		if err != nil {
			return helper.HandleError(c, model.ErrDatabaseError)
		}
		if exists {
			return helper.HandleError(c, model.NewValidationError("email sudah digunakan"))
		}
	}

	tx, err := s.db.BeginTx(c.Context(), nil)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	defer tx.Rollback()

	updatedUser := &model.User{
		Username: utils.GetStringOrDefault(req.Username, existingUser.Username),
		Email:    utils.GetStringOrDefault(req.Email, existingUser.Email),
		FullName: utils.GetStringOrDefault(req.FullName, existingUser.FullName),
		IsActive: utils.GetBoolOrDefault(req.IsActive, existingUser.IsActive),
	}

	err = s.userRepo.Update(c.Context(), id, updatedUser)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	if existingUser.Role.Name == "Mahasiswa" {
		if req.StudentID != nil || req.ProgramStudy != nil || req.AcademicYear != nil || req.AdvisorID != nil {
			updateReq := model.UpdateStudentProfileRequest{
				StudentID:    req.StudentID,
				ProgramStudy: req.ProgramStudy,
				AcademicYear: req.AcademicYear,
				AdvisorID:    req.AdvisorID,
			}
			err = s.studentSvc.UpdateProfile(c.Context(), tx, id, updateReq)
			if err != nil {
				return helper.HandleError(c, err)
			}
		}
	}

	if existingUser.Role.Name == "Dosen Wali" {
		if req.LecturerID != nil || req.Department != nil {
			updateReq := model.UpdateLecturerProfileRequest{
				LecturerID: req.LecturerID,
				Department: req.Department,
			}
			err = s.lecturerSvc.UpdateProfile(c.Context(), tx, id, updateReq)
			if err != nil {
				return helper.HandleError(c, err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	user, err := s.userRepo.GetByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	dto := user.ToDetailDTO()

	if user.Role.Name == "Mahasiswa" {
		studentInfo, err := s.studentSvc.GetProfile(c.Context(), id)
		if err == nil && studentInfo != nil {
			dto.Student = studentInfo
		}
	}

	if user.Role.Name == "Dosen Wali" {
		lecturerInfo, err := s.lecturerSvc.GetProfile(c.Context(), id)
		if err == nil && lecturerInfo != nil {
			dto.Lecturer = lecturerInfo
		}
	}

	return helper.Success(c, "User berhasil diupdate", dto)
}

// Delete godoc
// @Summary Delete user
// @Description Soft delete a user and their associated profile
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} helper.Response "User deleted successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid ID"
// @Failure 404 {object} helper.ErrorResponse "User not found"
// @Router /users/{id} [delete]
func (s *UserService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	user, err := s.userRepo.GetByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if user == nil {
		return helper.HandleError(c, model.ErrUserNotFound)
	}

	err = s.userRepo.Delete(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Success(c, "User berhasil dihapus", nil)
}

// UpdateRole godoc
// @Summary Update user role
// @Description Change user's role (requires profile recreation if role type changes)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body model.UpdateRoleRequest true "New role ID"
// @Success 200 {object} helper.Response "Role updated successfully"
// @Failure 400 {object} helper.ErrorResponse "Invalid request"
// @Failure 404 {object} helper.ErrorResponse "User or role not found"
// @Router /users/{id}/role [put]
func (s *UserService) UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Format request tidak valid", nil)
	}

	user, err := s.userRepo.GetByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if user == nil {
		return helper.HandleError(c, model.ErrUserNotFound)
	}

	err = s.userRepo.UpdateRole(c.Context(), id, req.RoleID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Success(c, "Role user berhasil diupdate", nil)
}

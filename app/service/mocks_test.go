package service

import (
	"context"
	"database/sql"
	"mime/multipart"
	"sistem-pelaporan-prestasi-mahasiswa/app/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// --- MOCK USER REPOSITORY ---
type MockUserRepository struct {
	users          map[string]*model.User
	roles          map[string]*model.Role
	usernameExists bool
	emailExists    bool
	roleExists     bool
}

func (m *MockUserRepository) GetAll(ctx context.Context, p, ps int, s, sb, so string) ([]model.User, int64, error) {
	return nil, 0, nil
}
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, nil
}
func (m *MockUserRepository) CheckUsernameExists(ctx context.Context, u string, ex *string) (bool, error) {
	return m.usernameExists, nil
}
func (m *MockUserRepository) CheckEmailExists(ctx context.Context, e string, ex *string) (bool, error) {
	return m.emailExists, nil
}
func (m *MockUserRepository) CheckRoleExists(ctx context.Context, rID string) (bool, error) {
	return m.roleExists, nil
}
func (m *MockUserRepository) GetRoleByID(ctx context.Context, rID string) (*model.Role, error) {
	if r, ok := m.roles[rID]; ok {
		return r, nil
	}
	return nil, nil
}
func (m *MockUserRepository) Create(ctx context.Context, tx *sql.Tx, u *model.User) error {
	u.ID = uuid.New()
	return nil
}
func (m *MockUserRepository) Update(ctx context.Context, id string, u *model.User) error { return nil }
func (m *MockUserRepository) Delete(ctx context.Context, id string) error                { return nil }
func (m *MockUserRepository) UpdateRole(ctx context.Context, uID, rID string) error      { return nil }

// --- MOCK STUDENT SERVICE ---
type MockStudentService struct {
	studentInfo *model.StudentInfo
	exists      bool
}

func (m *MockStudentService) GetProfile(ctx context.Context, id string) (*model.StudentInfo, error) {
	return m.studentInfo, nil
}
func (m *MockStudentService) CreateProfile(ctx context.Context, tx *sql.Tx, id string, req model.CreateStudentProfileRequest) error {
	return nil
}
func (m *MockStudentService) UpdateProfile(ctx context.Context, tx *sql.Tx, id string, req model.UpdateStudentProfileRequest) error {
	return nil
}
func (m *MockStudentService) DeleteProfile(ctx context.Context, tx *sql.Tx, id string) error {
	return nil
}
func (m *MockStudentService) ValidateStudentID(ctx context.Context, sID string, ex *string) error {
	return nil
}

// HTTP handler methods to match IStudentService interface
func (m *MockStudentService) GetAll(c *fiber.Ctx) error {
	c.Status(200).JSON(fiber.Map{"success": true})
	return nil
}
func (m *MockStudentService) GetByID(c *fiber.Ctx) error {
	c.Status(200).JSON(fiber.Map{"success": true})
	return nil
}
func (m *MockStudentService) UpdateAdvisor(c *fiber.Ctx) error {
	c.Status(200).JSON(fiber.Map{"success": true})
	return nil
}

// --- MOCK LECTURER SERVICE ---
type MockLecturerService struct {
	lecturerInfo *model.LecturerInfo
	exists       bool
}

func (m *MockLecturerService) GetProfile(ctx context.Context, id string) (*model.LecturerInfo, error) {
	return m.lecturerInfo, nil
}
func (m *MockLecturerService) CreateProfile(ctx context.Context, tx *sql.Tx, id string, req model.CreateLecturerProfileRequest) error {
	return nil
}
func (m *MockLecturerService) UpdateProfile(ctx context.Context, tx *sql.Tx, id string, req model.UpdateLecturerProfileRequest) error {
	return nil
}
func (m *MockLecturerService) DeleteProfile(ctx context.Context, tx *sql.Tx, id string) error {
	return nil
}
func (m *MockLecturerService) CheckExistsByID(ctx context.Context, id string) (bool, error) {
	return m.exists, nil
}
func (m *MockLecturerService) ValidateLecturerID(ctx context.Context, lID string, ex *string) error {
	return nil
}
// HTTP handler methods to match ILecturerService interface
func (m *MockLecturerService) GetAll(c *fiber.Ctx) error {
	c.Status(200).JSON(fiber.Map{"success": true})
	return nil
}
func (m *MockLecturerService) GetAdvisees(c *fiber.Ctx) error {
	c.Status(200).JSON(fiber.Map{"success": true})
	return nil
}

// --- MOCK LECTURER REPOSITORY ---
type MockLecturerRepository struct {
	lecturers map[string]*model.LecturerInfo
	idExists  bool
	exists    bool
}

func (m *MockLecturerRepository) Create(ctx context.Context, tx *sql.Tx, uID, lID, d string) error {
	return nil
}
func (m *MockLecturerRepository) Update(ctx context.Context, tx *sql.Tx, uID string, lID, d *string) error {
	return nil
}
func (m *MockLecturerRepository) GetByUserID(ctx context.Context, uID string) (*model.LecturerInfo, error) {
	if l, ok := m.lecturers[uID]; ok {
		return l, nil
	}
	return nil, nil
}
func (m *MockLecturerRepository) Delete(ctx context.Context, tx *sql.Tx, uID string) error { return nil }
func (m *MockLecturerRepository) CheckLecturerIDExists(ctx context.Context, id string, ex *string) (bool, error) {
	return m.idExists, nil
}
func (m *MockLecturerRepository) CheckExistsByID(ctx context.Context, id string) (bool, error) {
	return true, nil
}
func (m *MockLecturerRepository) GetAll(ctx context.Context, p, ps int, s, sb, so string) ([]model.LecturerListDTO, int64, error) {
	return nil, 0, nil
}
func (m *MockLecturerRepository) GetAdvisees(ctx context.Context, id string, p, ps int) ([]model.StudentListDTO, int64, error) {
	return nil, 0, nil
}

// --- MOCK STUDENT REPOSITORY ---
type MockStudentRepository struct {
	students map[string]*model.StudentInfo
	idExists bool
	detailID string
}

func (m *MockStudentRepository) Create(ctx context.Context, tx *sql.Tx, uID, sID, ps, ay string, advID *string) error {
	return nil
}
func (m *MockStudentRepository) Update(ctx context.Context, tx *sql.Tx, uID string, sID, ps, ay *string, advID *string) error {
	return nil
}
func (m *MockStudentRepository) GetByUserID(ctx context.Context, uID string) (*model.StudentInfo, error) {
	if s, ok := m.students[uID]; ok {
		return s, nil
	}
	return nil, nil
}
func (m *MockStudentRepository) Delete(ctx context.Context, tx *sql.Tx, uID string) error { return nil }
func (m *MockStudentRepository) CheckStudentIDExists(ctx context.Context, id string, ex *string) (bool, error) {
	return m.idExists, nil
}
func (m *MockStudentRepository) GetAll(ctx context.Context, p, ps int, s, sb, so string) ([]model.StudentListDTO, int64, error) {
	return nil, 0, nil
}
func (m *MockStudentRepository) GetDetailByID(ctx context.Context, id string) (*model.StudentDetailDTO, error) {
	if id == "valid-student" || id == m.detailID {
		return &model.StudentDetailDTO{
			ID:           uuid.New(),
			StudentID:    "S12345",
			FullName:     "Test Student",
			Email:        "test@example.com",
			ProgramStudy: "Informatika",
		}, nil
	}
	return nil, nil
}
func (m *MockStudentRepository) UpdateAdvisor(ctx context.Context, sID string, advID *string) error {
	return nil
}

// --- MOCK ACHIEVEMENT REPOSITORY ---
type MockAchievementRepository struct {
	achRefs   map[string]*model.AchievementReference
	achDetail *model.AchievementDetailDTO
}

func (m *MockAchievementRepository) Create(ctx context.Context, r *model.AchievementReference, mo *model.AchievementMongo) error {
	r.ID = uuid.New().String()
	m.achRefs[r.ID] = r
	return nil
}

func (m *MockAchievementRepository) GetRefByID(ctx context.Context, id string) (*model.AchievementReference, error) {
	if ref, ok := m.achRefs[id]; ok {
		return ref, nil
	}
	return nil, nil
}

func (m *MockAchievementRepository) AddAttachment(ctx context.Context, mID string, a model.AchievementAttachment) error {
	return nil
}
func (m *MockAchievementRepository) GetAll(ctx context.Context, p, ps int, s, sf, af, stf string) ([]model.AchievementListDTO, int64, error) {
	return nil, 0, nil
}

func (m *MockAchievementRepository) GetDetailByID(ctx context.Context, id string) (*model.AchievementDetailDTO, error) {
	return m.achDetail, nil
}
func (m *MockAchievementRepository) Update(ctx context.Context, id, mID string, req *model.UpdateAchievementRequest) error {
	return nil
}
func (m *MockAchievementRepository) Submit(ctx context.Context, id string) error     { return nil }
func (m *MockAchievementRepository) SoftDelete(ctx context.Context, id string) error { return nil }
func (m *MockAchievementRepository) Verify(ctx context.Context, id, lID string, pts int) error {
	return nil
}
func (m *MockAchievementRepository) Reject(ctx context.Context, id, lID string, note string) error {
	return nil
}
func (m *MockAchievementRepository) UploadAttachment(ctx context.Context, id, uID string, fh *multipart.FileHeader) (*model.AchievementAttachment, error) {
	return nil, nil
}

// --- MOCK AUTH REPOSITORY ---
type MockAuthRepository struct {
	users map[string]*model.User
}

func (m *MockAuthRepository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, nil
}

func (m *MockAuthRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, nil
}
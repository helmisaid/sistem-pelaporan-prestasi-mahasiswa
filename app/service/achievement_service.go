package service

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
	"sistem-pelaporan-prestasi-mahasiswa/helper"

	"github.com/gofiber/fiber/v2"
)

type IAchievementService interface {
	Create(c *fiber.Ctx) error
	UploadAttachment(c *fiber.Ctx) error
	GetAll(c *fiber.Ctx) error
	GetDetail(c *fiber.Ctx) error
	Edit(c *fiber.Ctx) error
	Submit(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	Verify(c *fiber.Ctx) error
	Reject(c *fiber.Ctx) error
	GetByStudent(c *fiber.Ctx) error
}

type AchievementService struct {
	achRepo     repository.IAchievementRepository
	studentRepo repository.IStudentRepository
	lecturerSvc ILecturerService
}

func NewAchievementService(
	achRepo repository.IAchievementRepository,
	studentRepo repository.IStudentRepository,
	lecturerSvc ILecturerService,
) IAchievementService {
	return &AchievementService{
		achRepo:     achRepo,
		studentRepo: studentRepo,
		lecturerSvc: lecturerSvc,
	}
}

// Create godoc
// @Summary Create new achievement
// @Description Create a new achievement in draft status
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.CreateAchievementRequest true "Achievement data"
// @Success 201 {object} helper.Response{data=model.AchievementReference} "Achievement created"
// @Failure 400 {object} helper.ErrorResponse "Invalid request"
// @Failure 401 {object} helper.ErrorResponse "unauthorized - student only"
// @Router /achievements [post]
func (s *AchievementService) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req model.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Format request tidak valid", nil)
	}

	studentInfo, err := s.studentRepo.GetByUserID(c.Context(), userID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if studentInfo == nil {
		return helper.HandleError(c, model.NewValidationError("Hanya mahasiswa yang boleh melapor prestasi"))
	}

	achMongo := &model.AchievementMongo{
		StudentID:       studentInfo.ID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Tags:            req.Tags,
		Attachments:     []model.AchievementAttachment{},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	achRef := &model.AchievementReference{
		StudentID: studentInfo.ID,
	}

	err = s.achRepo.Create(c.Context(), achRef, achMongo)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Created(c, "Prestasi berhasil dibuat (Draft). Silakan upload bukti.", achRef)
}

// Edit godoc
// @Summary Edit achievement
// @Description Edit draft achievement details
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Param request body model.UpdateAchievementRequest true "Updated data"
// @Success 200 {object} helper.Response "Achievement updated"
// @Failure 400 {object} helper.ErrorResponse "Can only edit draft achievements"
// @Router /achievements/{id} [put]
func (s *AchievementService) Edit(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req model.UpdateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Format data tidak valid", nil)
	}

	achRef, err := s.achRepo.GetRefByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if achRef == nil {
		return helper.HandleError(c, model.NewNotFoundError("Prestasi tidak ditemukan"))
	}

	studentInfo, err := s.studentRepo.GetByUserID(c.Context(), userID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if studentInfo == nil || achRef.StudentID != studentInfo.ID {
		return helper.HandleError(c, model.NewValidationError("Anda tidak berhak mengedit prestasi ini"))
	}

	if achRef.Status != "draft" {
		return helper.HandleError(c, model.NewValidationError("Hanya prestasi status Draft yang boleh diedit."))
	}

	err = s.achRepo.Update(c.Context(), id, achRef.MongoAchievementID, &req)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Success(c, "Prestasi berhasil diperbarui", nil)
}

// UploadAttachment godoc
// @Summary Upload achievement attachment
// @Description Upload proof file for achievement (PDF/Image, max 5MB)
// @Tags Achievements
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Param file formData file true "Attachment file"
// @Success 200 {object} helper.Response{data=model.AchievementAttachment} "File uploaded"
// @Failure 400 {object} helper.ErrorResponse "Invalid file"
// @Router /achievements/{id}/attachments [post]
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return helper.BadRequest(c, "File tidak ditemukan.", nil)
	}

	maxSize := int64(5 * 1024 * 1024)
	allowedTypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/jpg",
		"image/png",
	}

	if err := helper.ValidateFile(fileHeader, maxSize, allowedTypes); err != nil {
		return helper.HandleError(c, err)
	}

	achRef, err := s.achRepo.GetRefByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if achRef == nil {
		return helper.HandleError(c, model.NewNotFoundError("Prestasi tidak ditemukan"))
	}

	studentInfo, err := s.studentRepo.GetByUserID(c.Context(), userID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if studentInfo == nil || achRef.StudentID != studentInfo.ID {
		return helper.HandleError(c, model.NewValidationError("Anda tidak berhak mengedit prestasi ini"))
	}

	if achRef.Status != "draft" {
		return helper.HandleError(c, model.NewValidationError("Perubahan data tidak diizinkan. Prestasi ini sedang dalam proses verifikasi atau telah disetujui oleh Dosen Wali."))
	}

	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("ACH-%s-%d%s", id, time.Now().Unix(), ext)
	uploadDir := "./uploads/achievements"

	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	savePath := filepath.Join(uploadDir, filename)
	src, err := fileHeader.Open()
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	attachmentData := model.AchievementAttachment{
		FileName:   filename,
		FileURL:    "/uploads/achievements/" + filename,
		FileType:   fileHeader.Header.Get("Content-Type"),
		UploadedAt: time.Now(),
	}

	err = s.achRepo.AddAttachment(c.Context(), achRef.MongoAchievementID, attachmentData)
	if err != nil {
		os.Remove(savePath)
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Success(c, "File berhasil diupload", attachmentData)
}

// Submit godoc
// @Summary Submit achievement
// @Description Submit achievement for advisor verification
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Success 200 {object} helper.Response "Achievement submitted"
// @Failure 400 {object} helper.ErrorResponse "Can only submit draft achievements"
// @Router /achievements/{id}/submit [post]
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	achRef, err := s.achRepo.GetRefByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if achRef == nil {
		return helper.HandleError(c, model.NewNotFoundError("Prestasi tidak ditemukan"))
	}

	studentInfo, err := s.studentRepo.GetByUserID(c.Context(), userID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if studentInfo == nil || achRef.StudentID != studentInfo.ID {
		return helper.HandleError(c, model.NewValidationError("Akses ditolak"))
	}

	if achRef.Status != "draft" {
		return helper.HandleError(c, model.NewValidationError("Hanya prestasi berstatus Draft yang dapat disubmit."))
	}

	err = s.achRepo.Submit(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Success(c, "Prestasi berhasil disubmit ke Dosen Wali", nil)
}

// Delete godoc
// @Summary Delete achievement
// @Description Soft delete draft achievement
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Success 200 {object} helper.Response "Achievement deleted"
// @Failure 400 {object} helper.ErrorResponse "Can only delete draft achievements"
// @Router /achievements/{id} [delete]
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	achRef, err := s.achRepo.GetRefByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if achRef == nil {
		return helper.HandleError(c, model.NewNotFoundError("Prestasi tidak ditemukan"))
	}

	studentInfo, err := s.studentRepo.GetByUserID(c.Context(), userID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if studentInfo == nil || achRef.StudentID != studentInfo.ID {
		return helper.HandleError(c, model.NewValidationError("Akses ditolak"))
	}

	if achRef.Status != "draft" {
		return helper.HandleError(c, model.NewValidationError("Hanya prestasi berstatus Draft yang dapat dihapus."))
	}

	err = s.achRepo.SoftDelete(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Success(c, "Prestasi berhasil dihapus", nil)
}

// Verify godoc
// @Summary Verify achievement
// @Description Verify and approve achievement with points (Advisor only)
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Param request body model.VerifyAchievementRequest true "Points to award"
// @Success 200 {object} helper.Response "Achievement verified"
// @Failure 403 {object} helper.ErrorResponse "Not your advisee"
// @Router /achievements/{id}/verify [post]
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req model.VerifyAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Format data tidak valid (poin diperlukan)", nil)
	}

	achRef, err := s.achRepo.GetRefByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if achRef == nil {
		return helper.HandleError(c, model.NewNotFoundError("Prestasi tidak ditemukan"))
	}

	achDetail, err := s.achRepo.GetDetailByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	lecturerInfo, err := s.lecturerSvc.GetProfile(c.Context(), userID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if lecturerInfo == nil {
		return helper.HandleError(c, model.NewValidationError("Akses ditolak. User bukan dosen."))
	}

	if achDetail.Student.AdvisorID == nil || *achDetail.Student.AdvisorID != lecturerInfo.ID {
		return helper.HandleError(c, model.NewValidationError("Anda tidak berhak memverifikasi prestasi ini (Bukan mahasiswa bimbingan anda)"))
	}

	if achRef.Status != "submitted" {
		return helper.HandleError(c, model.NewValidationError("Hanya prestasi yang sudah disubmit yang dapat diverifikasi."))
	}

	err = s.achRepo.Verify(c.Context(), id, userID, req.Points)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Success(c, "Prestasi berhasil diverifikasi dan poin disimpan", nil)
}

// Reject godoc
// @Summary Reject achievement
// @Description Reject achievement with reason (Advisor only)
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Param request body model.RejectAchievementRequest true "Rejection reason"
// @Success 200 {object} helper.Response "Achievement rejected"
// @Failure 403 {object} helper.ErrorResponse "Not your advisee"
// @Router /achievements/{id}/reject [post]
func (s *AchievementService) Reject(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req model.RejectAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest(c, "Format data tidak valid (catatan penolakan diperlukan)", nil)
	}

	achRef, err := s.achRepo.GetRefByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if achRef == nil {
		return helper.HandleError(c, model.NewNotFoundError("Prestasi tidak ditemukan"))
	}

	achDetail, err := s.achRepo.GetDetailByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	lecturerInfo, err := s.lecturerSvc.GetProfile(c.Context(), userID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if lecturerInfo == nil {
		return helper.HandleError(c, model.NewValidationError("Akses ditolak. User bukan dosen."))
	}

	if achDetail.Student.AdvisorID == nil || *achDetail.Student.AdvisorID != lecturerInfo.ID {
		return helper.HandleError(c, model.NewValidationError("Anda tidak berhak menolak prestasi ini (Bukan mahasiswa bimbingan anda)"))
	}

	if achRef.Status != "submitted" {
		return helper.HandleError(c, model.NewValidationError("Hanya prestasi yang sudah disubmit yang dapat ditolak."))
	}

	err = s.achRepo.Reject(c.Context(), id, userID, req.RejectionNote)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	return helper.Success(c, "Prestasi ditolak dan dikembalikan ke mahasiswa", nil)
}

// GetAll godoc
// @Summary List achievements
// @Description Get paginated list of achievements with role-based filtering
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by title"
// @Param status query string false "Filter by status" Enums(draft, submitted, verified, rejected)
// @Success 200 {object} helper.Response{data=model.PaginatedAchievements} "Achievements retrieved"
// @Router /achievements [get]
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	roleName := c.Locals("role").(string)

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 10)
	search := c.Query("search", "")
	status := c.Query("status", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var studentIDFilter, advisorIDFilter string

	switch roleName {
	case "Mahasiswa":
		studentInfo, err := s.studentRepo.GetByUserID(c.Context(), userID)
		if err != nil {
			return helper.HandleError(c, model.ErrDatabaseError)
		}
		if studentInfo == nil {
			return helper.HandleError(c, model.NewValidationError("Data mahasiswa tidak ditemukan"))
		}
		studentIDFilter = studentInfo.ID

	case "Dosen Wali":
		lecturerInfo, err := s.lecturerSvc.GetProfile(c.Context(), userID)
		if err != nil {
			return helper.HandleError(c, model.ErrDatabaseError)
		}
		if lecturerInfo == nil {
			return helper.HandleError(c, model.NewValidationError("Data dosen tidak ditemukan"))
		}
		advisorIDFilter = lecturerInfo.ID
	default:
	}

	data, total, err := s.achRepo.GetAll(c.Context(), page, pageSize, search, studentIDFilter, advisorIDFilter, status)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	result := &model.PaginatedAchievements{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	return helper.Success(c, "Daftar prestasi berhasil diambil", result)
}

// GetDetail godoc
// @Summary Get achievement detail
// @Description Get detailed information about specific achievement
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Success 200 {object} helper.Response{data=model.AchievementDetailDTO} "Achievement details"
// @Failure 403 {object} helper.ErrorResponse "Forbidden"
// @Failure 404 {object} helper.ErrorResponse "Not found"
// @Router /achievements/{id} [get]
func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)
	roleName := c.Locals("role").(string)

	detail, err := s.achRepo.GetDetailByID(c.Context(), id)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if detail == nil {
		return helper.HandleError(c, model.NewNotFoundError("Prestasi tidak ditemukan"))
	}

	if roleName == "Mahasiswa" {
		studentInfo, err := s.studentRepo.GetByUserID(c.Context(), userID)
		if err != nil {
			return helper.HandleError(c, model.ErrDatabaseError)
		}

		if studentInfo == nil || studentInfo.ID != detail.Student.ID.String() {
			return helper.HandleError(c, model.NewValidationError("Anda tidak berhak melihat prestasi ini"))
		}
	} else if roleName == "Dosen Wali" {
		lecturerInfo, err := s.lecturerSvc.GetProfile(c.Context(), userID)
		if err != nil {
			return helper.HandleError(c, model.ErrDatabaseError)
		}

		if detail.Student.AdvisorID != nil && *detail.Student.AdvisorID != lecturerInfo.ID {
			return helper.HandleError(c, model.NewValidationError("Mahasiswa ini bukan bimbingan anda"))
		}
	}

	return helper.Success(c, "Detail prestasi berhasil diambil", detail)
}

// GetByStudent godoc
// @Summary Get student achievement history
// @Description Get all achievements for a specific student
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Filter by status"
// @Success 200 {object} helper.Response{data=model.PaginatedAchievements} "Student achievements"
// @Failure 403 {object} helper.ErrorResponse "Forbidden"
// @Router /achievements/{id}/history [get]
// @Router /students/{id}/achievements [get]
func (s *AchievementService) GetByStudent(c *fiber.Ctx) error {
	targetID := c.Params("id")
	viewerID := c.Locals("user_id").(string)
	viewerRole := c.Locals("role").(string)

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("limit", 10)
	status := c.Query("status", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	targetStudent, err := s.studentRepo.GetByUserID(c.Context(), targetID)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}
	if targetStudent == nil {
		return helper.HandleError(c, model.NewNotFoundError("Mahasiswa tidak ditemukan"))
	}

	switch viewerRole {
	case "Mahasiswa":
		if viewerID != targetID {
			return helper.HandleError(c, model.NewValidationError("Anda tidak memiliki akses untuk melihat prestasi mahasiswa lain"))
		}

	case "Dosen Wali":
		lecturer, err := s.lecturerSvc.GetProfile(c.Context(), viewerID)
		if err != nil {
			return helper.HandleError(c, model.ErrDatabaseError)
		}
		if targetStudent.AdvisorID == nil || *targetStudent.AdvisorID != lecturer.ID {
			return helper.HandleError(c, model.NewValidationError("Mahasiswa ini bukan bimbingan Anda"))
		}

	case "Admin":

	default:
		return helper.HandleError(c, model.NewValidationError("Role tidak dikenali"))
	}

	data, total, err := s.achRepo.GetAll(c.Context(), page, pageSize, "", targetStudent.ID, "", status)
	if err != nil {
		return helper.HandleError(c, model.ErrDatabaseError)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	result := &model.PaginatedAchievements{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	return helper.Success(c, "Daftar prestasi mahasiswa berhasil diambil", result)
}
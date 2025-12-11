package service

import (
	"context"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"sistem-pelaporan-prestasi-mahasiswa/app/model"
	"sistem-pelaporan-prestasi-mahasiswa/app/repository"
	"sistem-pelaporan-prestasi-mahasiswa/helper"
)

type IAchievementService interface {
	Create(ctx context.Context, userID string, req model.CreateAchievementRequest) (*model.AchievementReference, error)
	UploadAttachment(ctx context.Context, achievementID string, userID string, fileHeader *multipart.FileHeader) (*model.AchievementAttachment, error)
	GetAll(ctx context.Context, userID, roleName string, page, pageSize int, search, status string) (*model.PaginatedAchievements, error)
	GetDetail(ctx context.Context, id, userID, roleName string) (*model.AchievementDetailDTO, error)
	Edit(ctx context.Context, id, userID string, req model.UpdateAchievementRequest) error
	Submit(ctx context.Context, id, userID string) error
	Delete(ctx context.Context, id, userID string) error
	Verify(ctx context.Context, id, userID string, req model.VerifyAchievementRequest) error
    Reject(ctx context.Context, id, userID string, req model.RejectAchievementRequest) error
	GetByStudent(ctx context.Context, targetUserID, viewerUserID, viewerRole string, page, pageSize int, status string) (*model.PaginatedAchievements, error)
}

type AchievementService struct {
	achRepo      repository.IAchievementRepository
	studentRepo  repository.IStudentRepository
	lecturerSvc  ILecturerService
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

// Create Achievement
func (s *AchievementService) Create(ctx context.Context, userID string, req model.CreateAchievementRequest) (*model.AchievementReference, error) {
	studentInfo, err := s.studentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if studentInfo == nil {
		return nil, model.NewValidationError("Hanya mahasiswa yang boleh melapor prestasi")
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

	err = s.achRepo.Create(ctx, achRef, achMongo)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	return achRef, nil
}

// Edit Achievement
func (s *AchievementService) Edit(ctx context.Context, id, userID string, req model.UpdateAchievementRequest) error {
    achRef, err := s.achRepo.GetRefByID(ctx, id)
    if err != nil { return model.ErrDatabaseError }
    if achRef == nil { return model.NewNotFoundError("Prestasi tidak ditemukan") }

    studentInfo, err := s.studentRepo.GetByUserID(ctx, userID)
    if err != nil { return model.ErrDatabaseError }
    if studentInfo == nil || achRef.StudentID != studentInfo.ID {
        return model.NewValidationError("Anda tidak berhak mengedit prestasi ini")
    }

    if achRef.Status != "draft" {
        return model.NewValidationError("Hanya prestasi status Draft yang boleh diedit.")
    }

    return s.achRepo.Update(ctx, id, achRef.MongoAchievementID, &req)
}

// UploadAttachment

func (s *AchievementService) UploadAttachment(ctx context.Context, achievementID string, userID string, fileHeader *multipart.FileHeader) (*model.AchievementAttachment, error) {

    maxSize := int64(5 * 1024 * 1024)

    allowedTypes := []string{
        "application/pdf",
        "image/jpeg",
        "image/jpg",
        "image/png",
    }

    if err := helper.ValidateFile(fileHeader, maxSize, allowedTypes); err != nil {
        return nil, err
    }

    achRef, err := s.achRepo.GetRefByID(ctx, achievementID)

    if err != nil {
        return nil, model.ErrDatabaseError
    }

    if achRef == nil {
        return nil, model.NewNotFoundError("Prestasi tidak ditemukan")
    }

    studentInfo, err := s.studentRepo.GetByUserID(ctx, userID)

    if err != nil {
        return nil, model.ErrDatabaseError
    }
    if studentInfo == nil || achRef.StudentID != studentInfo.ID {
        return nil, model.NewValidationError("Anda tidak berhak mengedit prestasi ini")
    }

    if achRef.Status != "draft" {
        return nil, model.NewValidationError("Perubahan data tidak diizinkan. Prestasi ini sedang dalam proses verifikasi atau telah disetujui oleh Dosen Wali.")
    }

    ext := filepath.Ext(fileHeader.Filename)

    filename := fmt.Sprintf("ACH-%s-%d%s", achievementID, time.Now().Unix(), ext)

    uploadDir := "./uploads/achievements"

    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        os.MkdirAll(uploadDir, 0755)
    }

    savePath := filepath.Join(uploadDir, filename)
    src, err := fileHeader.Open()

    if err != nil {
        return nil, model.ErrDatabaseError
    }
    defer src.Close()

    dst, err := os.Create(savePath)

    if err != nil {
        return nil, model.ErrDatabaseError
    }
    defer dst.Close()

    if _, err = io.Copy(dst, src); err != nil {
        return nil, model.ErrDatabaseError
    }

    attachmentData := model.AchievementAttachment{
        FileName:   filename,
        FileURL:    "/uploads/achievements/" + filename,
        FileType:   fileHeader.Header.Get("Content-Type"),
        UploadedAt: time.Now(),
    }

    err = s.achRepo.AddAttachment(ctx, achRef.MongoAchievementID, attachmentData)

    if err != nil {
        os.Remove(savePath)
        return nil, model.ErrDatabaseError
    }
    return &attachmentData, nil

}

// Submit
func (s *AchievementService) Submit(ctx context.Context, id, userID string) error {
	achRef, err := s.achRepo.GetRefByID(ctx, id)
	if err != nil {
		return model.ErrDatabaseError
	}
	if achRef == nil {
		return model.NewNotFoundError("Prestasi tidak ditemukan")
	}

	studentInfo, err := s.studentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return model.ErrDatabaseError
	}
	if studentInfo == nil || achRef.StudentID != studentInfo.ID {
		return model.NewValidationError("Akses ditolak")
	}

	if achRef.Status != "draft" {
		return model.NewValidationError("Hanya prestasi berstatus Draft yang dapat disubmit.")
	}

	return s.achRepo.Submit(ctx, id)
}

// Delete 
func (s *AchievementService) Delete(ctx context.Context, id, userID string) error {
	achRef, err := s.achRepo.GetRefByID(ctx, id)
	if err != nil {
		return model.ErrDatabaseError
	}
	if achRef == nil {
		return model.NewNotFoundError("Prestasi tidak ditemukan")
	}

	studentInfo, err := s.studentRepo.GetByUserID(ctx, userID)
	if err != nil {
		return model.ErrDatabaseError
	}
	if studentInfo == nil || achRef.StudentID != studentInfo.ID {
		return model.NewValidationError("Akses ditolak")
	}

	if achRef.Status != "draft" {
		return model.NewValidationError("Hanya prestasi berstatus Draft yang dapat dihapus.")
	}

	return s.achRepo.SoftDelete(ctx, id)
}

// Verify 
func (s *AchievementService) Verify(ctx context.Context, id, userID string, req model.VerifyAchievementRequest) error {
    achRef, err := s.achRepo.GetRefByID(ctx, id)
    if err != nil { return model.ErrDatabaseError }
    if achRef == nil { return model.NewNotFoundError("Prestasi tidak ditemukan") }

    achDetail, err := s.achRepo.GetDetailByID(ctx, id)
    if err != nil { return model.ErrDatabaseError }

	// Get lecturer info
    lecturerInfo, err := s.lecturerSvc.GetProfile(ctx, userID)

    if err != nil { return model.ErrDatabaseError }
    if lecturerInfo == nil { return model.NewValidationError("Akses ditolak. User bukan dosen.") }

    if achDetail.Student.AdvisorID == nil || *achDetail.Student.AdvisorID != lecturerInfo.ID {
        return model.NewValidationError("Anda tidak berhak memverifikasi prestasi ini (Bukan mahasiswa bimbingan anda)")
    }

    if achRef.Status != "submitted" {
        return model.NewValidationError("Hanya prestasi yang sudah disubmit yang dapat diverifikasi.")
    }

    return s.achRepo.Verify(ctx, id, userID, req.Points)
}

// Reject
func (s *AchievementService) Reject(ctx context.Context, id, userID string, req model.RejectAchievementRequest) error {
    achRef, err := s.achRepo.GetRefByID(ctx, id)
    if err != nil { return model.ErrDatabaseError }
    if achRef == nil { return model.NewNotFoundError("Prestasi tidak ditemukan") }

    achDetail, err := s.achRepo.GetDetailByID(ctx, id)
    if err != nil { return model.ErrDatabaseError }

    lecturerInfo, err := s.lecturerSvc.GetProfile(ctx, userID)
    if err != nil { return model.ErrDatabaseError }
    if lecturerInfo == nil { return model.NewValidationError("Akses ditolak. User bukan dosen.") }

    if achDetail.Student.AdvisorID == nil || *achDetail.Student.AdvisorID != lecturerInfo.ID {
        return model.NewValidationError("Anda tidak berhak menolak prestasi ini (Bukan mahasiswa bimbingan anda)")
    }

    if achRef.Status != "submitted" {
        return model.NewValidationError("Hanya prestasi yang sudah disubmit yang dapat ditolak.")
    }

    return s.achRepo.Reject(ctx, id, userID, req.RejectionNote)
}

// GetAll 
func (s *AchievementService) GetAll(ctx context.Context, userID, roleName string, page, pageSize int, search, status string) (*model.PaginatedAchievements, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var studentIDFilter, advisorIDFilter string

	switch roleName {
	case "Mahasiswa":
		studentInfo, err := s.studentRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, model.ErrDatabaseError
		}
		if studentInfo == nil {
			return nil, model.NewValidationError("Data mahasiswa tidak ditemukan")
		}
		studentIDFilter = studentInfo.ID

	case "Dosen Wali":
		lecturerInfo, err := s.lecturerSvc.GetProfile(ctx, userID)
		if err != nil {
			return nil, model.ErrDatabaseError
		}
		if lecturerInfo == nil {
			return nil, model.NewValidationError("Data dosen tidak ditemukan")
		}
		advisorIDFilter = lecturerInfo.ID
	default:
	}

	data, total, err := s.achRepo.GetAll(ctx, page, pageSize, search, studentIDFilter, advisorIDFilter, status)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &model.PaginatedAchievements{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetDetail
func (s *AchievementService) GetDetail(ctx context.Context, id, userID, roleName string) (*model.AchievementDetailDTO, error) {
	detail, err := s.achRepo.GetDetailByID(ctx, id)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if detail == nil {
		return nil, model.NewNotFoundError("Prestasi tidak ditemukan")
	}

	// Security Check based on Role
	if roleName == "Mahasiswa" {
		studentInfo, err := s.studentRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, model.ErrDatabaseError
		}

		if studentInfo == nil || studentInfo.ID != detail.Student.ID.String() {
			return nil, model.NewValidationError("Anda tidak berhak melihat prestasi ini")
		}
	} else if roleName == "Dosen Wali" {
		lecturerInfo, err := s.lecturerSvc.GetProfile(ctx, userID)
		if err != nil {
			return nil, model.ErrDatabaseError
		}
		
		if detail.Student.AdvisorID != nil && *detail.Student.AdvisorID != lecturerInfo.ID {
			return nil, model.NewValidationError("Mahasiswa ini bukan bimbingan anda")
		}
	}

	return detail, nil
}

// Get Achievement by Student ID
func (s *AchievementService) GetByStudent(ctx context.Context, targetUserID, viewerUserID, viewerRole string, page, pageSize int, status string) (*model.PaginatedAchievements, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	targetStudent, err := s.studentRepo.GetByUserID(ctx, targetUserID)
	if err != nil {
		return nil, model.ErrDatabaseError
	}
	if targetStudent == nil {
		return nil, model.NewNotFoundError("Mahasiswa tidak ditemukan")
	}

	// Validate viewer access based on role
	switch viewerRole {
	case "Mahasiswa":
		if viewerUserID != targetUserID {
			return nil, model.NewValidationError("Anda tidak memiliki akses untuk melihat prestasi mahasiswa lain")
		}

	case "Dosen Wali":
		lecturer, err := s.lecturerSvc.GetProfile(ctx, viewerUserID)
		if err != nil {
			return nil, model.ErrDatabaseError
		}
		if targetStudent.AdvisorID == nil || *targetStudent.AdvisorID != lecturer.ID {
			return nil, model.NewValidationError("Mahasiswa ini bukan bimbingan Anda")
		}

	case "Admin":
		// Admin has full access - no additional validation needed

	default:
		return nil, model.NewValidationError("Role tidak dikenali")
	}

	data, total, err := s.achRepo.GetAll(ctx, page, pageSize, "", targetStudent.ID, "", status)
	if err != nil {
		return nil, model.ErrDatabaseError
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &model.PaginatedAchievements{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
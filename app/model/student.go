package model

type CreateStudentProfileRequest struct {
	StudentID    string  `json:"student_id" validate:"required"`
	ProgramStudy string  `json:"program_study" validate:"required"`
	AcademicYear string  `json:"academic_year" validate:"required"`
	AdvisorID    *string `json:"advisor_id,omitempty" validate:"omitempty,uuid"`
}

type UpdateStudentProfileRequest struct {
	StudentID    *string `json:"student_id,omitempty"`
	ProgramStudy *string `json:"program_study,omitempty"`
	AcademicYear *string `json:"academic_year,omitempty"`
	AdvisorID    *string `json:"advisor_id,omitempty" validate:"omitempty,uuid"`
}
